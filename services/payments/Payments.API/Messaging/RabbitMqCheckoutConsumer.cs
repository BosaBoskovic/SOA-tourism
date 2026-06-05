using System.Text;
using System.Text.Json;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Payments.Application.Events;
using Payments.Domain.Entities;
using Payments.Infrastructure.Repositories;
using Payments.Infrastructure.Messaging;
using RabbitMQ.Client;
using RabbitMQ.Client.Events;

namespace Payments.API.Messaging;

public class RabbitMqCheckoutConsumer : BackgroundService
{
    private readonly IServiceScopeFactory _scopeFactory;
    private readonly RabbitMqPublisher _publisher;
    private readonly string _host;

    public RabbitMqCheckoutConsumer(
        IServiceScopeFactory scopeFactory,
        RabbitMqPublisher publisher,
        IConfiguration configuration)
    {
        _scopeFactory = scopeFactory;
        _publisher = publisher;
        _host = configuration["RABBITMQ_HOST"] ?? "localhost";
    }

    protected override Task ExecuteAsync(CancellationToken stoppingToken)
    {
        Task.Run(() => StartConsumer("checkout-approved", HandleApproved), stoppingToken);
        Task.Run(() => StartConsumer("checkout-rejected", HandleRejected), stoppingToken);

        return Task.CompletedTask;
    }

    private void StartConsumer(string queueName, Action<string> handler)
    {
        var factory = new ConnectionFactory { HostName = _host };

        var connection = factory.CreateConnection();
        var channel = connection.CreateModel();

        channel.QueueDeclare(queueName, durable: true, exclusive: false, autoDelete: false);

        var consumer = new EventingBasicConsumer(channel);

        consumer.Received += (_, ea) =>
        {
            var json = Encoding.UTF8.GetString(ea.Body.ToArray());

            try
            {
                handler(json);
                channel.BasicAck(ea.DeliveryTag, false);
            }
            catch
            {
                channel.BasicNack(ea.DeliveryTag, false, true);
            }
        };

        channel.BasicConsume(queueName, autoAck: false, consumer);
    }

    private void HandleApproved(string json)
    {
        var ev = JsonSerializer.Deserialize<CheckoutApprovedEvent>(json);

        if (ev == null)
            return;

        using var scope = _scopeFactory.CreateScope();

        var cartRepo = scope.ServiceProvider.GetRequiredService<ShoppingCartRepository>();
        var tokenRepo = scope.ServiceProvider.GetRequiredService<TourPurchaseTokenRepository>();

        var cart = cartRepo.GetByTouristIdAsync(ev.TouristId).Result;

        if (cart == null || cart.Items.Count == 0)
            return;

        var tokens = cart.Items.Select(item => new TourPurchaseToken
        {
            TouristId = ev.TouristId,
            TourId = item.TourId,
            TourName = item.TourName,
            Price = item.Price,
            PurchasedAt = DateTime.UtcNow
        }).ToList();

        tokenRepo.AddRangeAsync(tokens).Wait();
        cartRepo.ClearAsync(cart).Wait();

        var completedEvent = new PurchaseCompletedEvent
        {
            SagaId = ev.SagaId,
            TouristId = ev.TouristId,
            Items = tokens.Select(t => new PurchasedTourItem
            {
                TourId = t.TourId,
                TourName = t.TourName,
                Price = (double)t.Price
            }).ToList()
        };

        _publisher.Publish("purchase-completed", completedEvent);
    }

    private void HandleRejected(string json)
    {
        var ev = JsonSerializer.Deserialize<CheckoutRejectedEvent>(json);

        if (ev == null)
            return;

        Console.WriteLine($"Checkout saga rejected. SagaId={ev.SagaId}, reason={ev.Reason}");
    }
}