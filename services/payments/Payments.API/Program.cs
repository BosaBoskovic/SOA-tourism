using Microsoft.AspNetCore.Diagnostics.HealthChecks;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Diagnostics.HealthChecks;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;
using Payments.API.Grpc;
using Payments.API.Health;
using Payments.Application.Services;
using Payments.Infrastructure.Data;
using Payments.Infrastructure.Repositories;
using Payments.Application.Clients;
using Payments.Infrastructure.Messaging;
using Payments.API.Messaging;

var builder = WebApplication.CreateBuilder(args);

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

builder.Services.AddHostedService<RabbitMqCheckoutConsumer>();

builder.Services.AddHttpClient<TourClient>(client =>
{
    var toursUrl = builder.Configuration["TOURS_URL"]
                   ?? Environment.GetEnvironmentVariable("TOURS_URL")
                   ?? "http://localhost:8085";

    client.BaseAddress = new Uri(toursUrl);
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
    db.Database.Migrate();
}

// --- Monitoring: RED metrics for every HTTP/gRPC request, exposed at /metrics ---
app.UseHttpMetrics();
app.MapMetrics("/metrics");

app.MapHealthChecks("/health", new HealthCheckOptions
{
    ResponseWriter = HealthCheckJson.WriteResponse,
});

app.MapControllers();
app.MapGrpcService<PaymentsGrpcService>();

app.Run();