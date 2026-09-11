using System.Diagnostics;
using Payments.Infrastructure.Messaging;
using Xunit;

namespace Payments.Tests;

// RabbitMqPublisher.InjectTraceContext is the piece that lets the
// purchase-completed trace continue into tours' consumer instead of
// starting a disconnected trace there - see tours' ExtractAMQPContext.
// Tested directly against a plain header dictionary, no broker needed.
public class RabbitMqTraceContextTests : IDisposable
{
    private readonly ActivityListener _listener;
    private readonly ActivitySource _source = new("Payments.Messaging.Tests");

    public RabbitMqTraceContextTests()
    {
        // Activity.Current/StartActivity only actually produces an Activity
        // when something is listening - without this, the test would be
        // asserting on a null activity regardless of the code under test.
        _listener = new ActivityListener
        {
            ShouldListenTo = _ => true,
            // AllDataAndRecorded (not just AllData) is what actually sets
            // the W3C "recorded" flag - otherwise traceparent's flags byte
            // comes out "00" regardless of what Inject writes.
            Sample = (ref ActivityCreationOptions<ActivityContext> _) => ActivitySamplingResult.AllDataAndRecorded,
        };
        ActivitySource.AddActivityListener(_listener);
    }

    public void Dispose()
    {
        _listener.Dispose();
        _source.Dispose();
    }

    [Fact]
    public void InjectTraceContext_WritesTraceparentMatchingTheActivity()
    {
        using var activity = _source.StartActivity("purchase-completed publish", ActivityKind.Producer);
        Assert.NotNull(activity);

        var headers = new Dictionary<string, object>();
        RabbitMqPublisher.InjectTraceContext(activity, headers);

        var traceparent = Assert.IsType<string>(headers["traceparent"]);
        Assert.Equal($"00-{activity!.TraceId}-{activity.SpanId}-01", traceparent);
    }

    [Fact]
    public void InjectTraceContext_NoActivity_LeavesHeadersEmpty()
    {
        var headers = new Dictionary<string, object>();

        RabbitMqPublisher.InjectTraceContext(null, headers);

        Assert.Empty(headers);
    }
}
