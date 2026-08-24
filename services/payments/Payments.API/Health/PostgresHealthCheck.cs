using Microsoft.Extensions.Diagnostics.HealthChecks;
using Payments.Infrastructure.Data;

namespace Payments.API.Health;

/// <summary>
/// Reports whether the payments service can actually reach Postgres, so a
/// monitoring app can tell "process is up" apart from "process is up but
/// its database is unreachable".
/// </summary>
public class PostgresHealthCheck : IHealthCheck
{
    private readonly PaymentsDbContext _db;

    public PostgresHealthCheck(PaymentsDbContext db)
    {
        _db = db;
    }

    public async Task<HealthCheckResult> CheckHealthAsync(HealthCheckContext context, CancellationToken cancellationToken = default)
    {
        try
        {
            var canConnect = await _db.Database.CanConnectAsync(cancellationToken);
            return canConnect
                ? HealthCheckResult.Healthy("postgres reachable")
                : HealthCheckResult.Unhealthy("postgres not reachable");
        }
        catch (Exception ex)
        {
            return HealthCheckResult.Unhealthy("postgres check failed", ex);
        }
    }
}
