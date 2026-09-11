using System.Diagnostics;
using System.Text;
using System.Text.Json;
using RabbitMQ.Client;
using Microsoft.Extensions.Configuration;

namespace Payments.Infrastructure.Messaging;

public class RabbitMqPublisher
{
    // A dedicated ActivitySource so every publish gets its own Producer
    // span in Jaeger, distinct from the ASP.NET Core request span that
    // triggered it. Registered via .AddSource("Payments.Messaging") in
    // Program.cs - without that registration, StartActivity below returns
    // null (no listener) and the span is silently dropped.
    private static readonly ActivitySource ActivitySource = new("Payments.Messaging");

    private readonly string _host;

    public RabbitMqPublisher(IConfiguration configuration)
    {
        _host = configuration["RABBITMQ_HOST"] ?? "localhost";
    }

    public void Publish<T>(string queueName, T message)
    {
        using var activity = ActivitySource.StartActivity($"{queueName} publish", ActivityKind.Producer);
        activity?.SetTag("messaging.system", "rabbitmq");
        activity?.SetTag("messaging.destination", queueName);

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

        var props = channel.CreateBasicProperties();
        props.Headers = new Dictionary<string, object>();
        InjectTraceContext(activity, props.Headers);

        channel.BasicPublish(
            exchange: "",
            routingKey: queueName,
            basicProperties: props,
            body: body);
    }

    // Public (not internal) and taking a plain header dictionary rather
    // than IBasicProperties so it's unit-testable without a broker or a
    // concrete RabbitMQ.Client properties implementation: given an
    // Activity (or null) and a set of AMQP headers, write the W3C
    // traceparent so a consumer on the other side (see tours'
    // ExtractAMQPContext) continues this trace instead of starting a new
    // one.
    public static void InjectTraceContext(Activity? activity, IDictionary<string, object> headers)
    {
        var propagationActivity = activity ?? Activity.Current;
        if (propagationActivity == null)
        {
            return;
        }

        DistributedContextPropagator.Current.Inject(
            propagationActivity,
            headers,
            static (carrier, key, value) => ((IDictionary<string, object>)carrier!)[key] = value);
    }
}
