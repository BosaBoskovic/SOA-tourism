using System.Text;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.AspNetCore.Diagnostics.HealthChecks;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Diagnostics.HealthChecks;
using Microsoft.IdentityModel.Tokens;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;
using Prometheus;
using Payments.API.Grpc;
using Payments.API.Health;
using Payments.Application.Services;
using Payments.Infrastructure.Data;
using Payments.Infrastructure.Repositories;
using Payments.Application.Clients;
using Payments.Infrastructure.Messaging;

var builder = WebApplication.CreateBuilder(args);

// --- Auth: verifies the same HS256 access tokens stakeholders issues ---
var jwtSecret = Environment.GetEnvironmentVariable("JWT_SECRET");
if (string.IsNullOrWhiteSpace(jwtSecret))
{
    throw new InvalidOperationException("JWT_SECRET is not set; payments cannot verify tokens");
}

builder.Services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
    .AddJwtBearer(options =>
    {
        options.MapInboundClaims = false; // keep "sub"/"role" as-is instead of remapping to long claim URIs
        options.TokenValidationParameters = new TokenValidationParameters
        {
            ValidateIssuerSigningKey = true,
            IssuerSigningKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(jwtSecret)),
            ValidateIssuer = false,
            ValidateAudience = false,
            ValidateLifetime = true,
            NameClaimType = "sub",
            RoleClaimType = "role",
        };
    });
builder.Services.AddAuthorization();

// --- Monitoring: structured JSON logs, tagged with the active trace/span ID ---
builder.Logging.ClearProviders();
builder.Logging.AddJsonConsole(options =>
{
    options.IncludeScopes = true;
    options.TimestampFormat = "yyyy-MM-ddTHH:mm:ss.fffZ";
    options.UseUtcTimestamp = true;
});
builder.Logging.Configure(options =>
{
    options.ActivityTrackingOptions = ActivityTrackingOptions.TraceId
        | ActivityTrackingOptions.SpanId
        | ActivityTrackingOptions.ParentId;
});

var serviceName = Environment.GetEnvironmentVariable("OTEL_SERVICE_NAME") ?? "payments-service";
var otlpEndpoint = Environment.GetEnvironmentVariable("OTEL_EXPORTER_OTLP_ENDPOINT") ?? "http://localhost:4318";

// --- Monitoring: distributed tracing exported to Jaeger (OTLP), same collector every service uses ---
builder.Services.AddOpenTelemetry()
    .ConfigureResource(resource => resource.AddService(serviceName))
    .WithTracing(tracing => tracing
        .AddSource("Payments.Messaging") // RabbitMqPublisher's producer spans - see InjectTraceContext
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation()
        .AddOtlpExporter(otlp => otlp.Endpoint = new Uri($"{otlpEndpoint}/v1/traces")));

// --- Monitoring: /health, checking Postgres connectivity ---
builder.Services.AddHealthChecks()
    .AddCheck<PostgresHealthCheck>("postgres");

builder.Services.AddControllers();
builder.Services.AddGrpc();

var connectionString = builder.Configuration.GetConnectionString("DefaultConnection")
    ?? "Host=localhost;Port=5433;Database=paymentsdb;Username=postgres;Password=postgres";

builder.Services.AddDbContext<PaymentsDbContext>(options =>
    options.UseNpgsql(connectionString));

builder.Services.AddScoped<ShoppingCartRepository>();
builder.Services.AddScoped<TourPurchaseTokenRepository>();
builder.Services.AddScoped<ShoppingCartService>();
builder.Services.AddScoped<CheckoutService>();

builder.Services.AddSingleton<RabbitMqPublisher>();

builder.Services.AddHttpClient<TourClient>(client =>
{
    var toursUrl = builder.Configuration["TOURS_URL"]
                   ?? Environment.GetEnvironmentVariable("TOURS_URL")
                   ?? "http://localhost:8085";

    client.BaseAddress = new Uri(toursUrl);
})
    // tours can run as multiple replicas behind Docker's embedded DNS
    // (docker-compose.yml no longer pins its container_name).
    // SocketsHttpHandler pools/reuses connections indefinitely by default,
    // which would pin every call to whichever replica answered first -
    // PooledConnectionLifetime forces periodic reconnect (and DNS
    // re-resolution), spreading calls across replicas over time. This is
    // the standard .NET fix for load-balancing across a DNS name that can
    // resolve to more than one address.
    .ConfigurePrimaryHttpMessageHandler(() => new SocketsHttpHandler
    {
        PooledConnectionLifetime = TimeSpan.FromSeconds(30),
    });

builder.WebHost.ConfigureKestrel(options =>
{
    options.ListenAnyIP(8086); // HTTP/1.1 za REST
    options.ListenAnyIP(9092, listenOptions =>
    {
        listenOptions.Protocols = Microsoft.AspNetCore.Server.Kestrel.Core.HttpProtocols.Http2; // gRPC
    });
});

var app = builder.Build();

using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<PaymentsDbContext>();

    // Multiple payments replicas can now start at once (docker compose up
    // -d --scale payments=N), which would otherwise race to run EF Core
    // migrations against the same Postgres database concurrently. A
    // Postgres advisory lock serializes them: each replica blocks here
    // until it's its turn; Migrate() is then a no-op for every replica
    // after the first, since EF Core tracks applied migrations in
    // __EFMigrationsHistory. The lock must be held on one physical
    // connection for the whole lock->migrate->unlock sequence, so the
    // connection is opened explicitly instead of letting EF Core open/
    // close a pooled connection per command.
    const long migrationLockKey = 4815162342;
    db.Database.OpenConnection();
    try
    {
        db.Database.ExecuteSqlRaw("SELECT pg_advisory_lock({0})", migrationLockKey);
        db.Database.Migrate();
    }
    finally
    {
        db.Database.ExecuteSqlRaw("SELECT pg_advisory_unlock({0})", migrationLockKey);
        db.Database.CloseConnection();
    }
}

// --- Monitoring: RED metrics for every HTTP/gRPC request, exposed at /metrics ---
app.UseHttpMetrics();
app.MapMetrics("/metrics");

app.MapHealthChecks("/health", new HealthCheckOptions
{
    ResponseWriter = HealthCheckJson.WriteResponse,
});

app.UseAuthentication();
app.UseAuthorization();

app.MapControllers();
app.MapGrpcService<PaymentsGrpcService>();

app.Run();