using System.Text.Json;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Diagnostics.HealthChecks;

namespace Payments.API.Health;

/// <summary>
/// JSON response writer for the health check middleware, matching the
/// {"service","status","checks"} shape every other service returns from /health.
/// </summary>
public static class HealthCheckJson
{
    public static Task WriteResponse(HttpContext context, HealthReport report)
    {
        context.Response.ContentType = "application/json";

        var checks = new Dictionary<string, string>();
        foreach (var entry in report.Entries)
        {
            checks[entry.Key] = entry.Value.Status == HealthStatus.Healthy
                ? "ok"
                : $"error: {entry.Value.Description ?? entry.Value.Exception?.Message ?? entry.Value.Status.ToString()}";
        }

        var payload = new
        {
            service = "payments",
            status = report.Status == HealthStatus.Healthy ? "ok" : "degraded",
            checks,
        };

        return context.Response.WriteAsync(JsonSerializer.Serialize(payload));
    }
}
