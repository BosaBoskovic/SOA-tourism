using System.Text;
using System.Text.Json;
using RabbitMQ.Client;
using Microsoft.Extensions.Configuration;

namespace Payments.Infrastructure.Messaging;

public class RabbitMqPublisher
{
    private readonly string _host;

    public RabbitMqPublisher(IConfiguration configuration)
    {
        _host = configuration["RABBITMQ_HOST"] ?? "localhost";
    }

    public void Publish<T>(string queueName, T message)
    {
        var factory = new ConnectionFactory { HostName = _host };

        using var connection = factory.CreateConnection();
        using var channel = connection.CreateModel();

        channel.QueueDeclare(
            queue: queueName,
            durable: true,
            exclusive: false,
            autoDelete: false);

        var json = JsonSerializer.Serialize(message);
        var body = Encoding.UTF8.GetBytes(json);

        channel.BasicPublish(
            exchange: "",
            routingKey: queueName,
            basicProperties: null,
            body: body);
    }
}