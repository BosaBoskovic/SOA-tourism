using Microsoft.EntityFrameworkCore;
using Payments.API.Grpc;
using Payments.Application.Services;
using Payments.Infrastructure.Data;
using Payments.Infrastructure.Repositories;

var builder = WebApplication.CreateBuilder(args);

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

app.MapControllers();
app.MapGrpcService<PaymentsGrpcService>();

app.Run();