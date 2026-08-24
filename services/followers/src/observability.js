const client = require("prom-client");
const pino = require("pino");
const pinoHttp = require("pino-http");
const { trace } = require("@opentelemetry/api");

const SERVICE_NAME = process.env.OTEL_SERVICE_NAME || "followers-service";

// --- Structured logging --------------------------------------------------

const logger = pino({
  level: process.env.LOG_LEVEL || "info",
  base: { service: SERVICE_NAME },
  timestamp: pino.stdTimeFunctions.isoTime,
  // Emit "info"/"error"/... instead of pino's numeric levels, so Promtail
  // can extract a "level" label consistently with the other services.
  formatters: {
    level: (label) => ({ level: label }),
  },
});

// requestLogger emits one structured JSON line per request and tags it with
// the active OpenTelemetry trace ID so logs and Jaeger traces correlate.
const requestLogger = pinoHttp({
  logger,
  customProps: (req) => {
    const span = trace.getActiveSpan();
    if (!span) return {};
    const ctx = span.spanContext();
    return { trace_id: ctx.traceId, span_id: ctx.spanId };
  },
  customLogLevel: (_req, res, err) => {
    if (err || res.statusCode >= 500) return "error";
    if (res.statusCode >= 400) return "warn";
    return "info";
  },
});

// --- Prometheus metrics ---------------------------------------------------

client.collectDefaultMetrics({ prefix: "", labels: { service: SERVICE_NAME } });

const httpRequestsTotal = new client.Counter({
  name: "http_requests_total",
  help: "Total number of HTTP requests handled, labeled by route and status code.",
  labelNames: ["service", "method", "route", "status"],
});

const httpRequestDuration = new client.Histogram({
  name: "http_request_duration_seconds",
  help: "HTTP request duration in seconds.",
  labelNames: ["service", "method", "route"],
  buckets: client.exponentialBuckets(0.005, 2, 12),
});

function metricsMiddleware(req, res, next) {
  const start = process.hrtime.bigint();
  res.on("finish", () => {
    const durationSeconds = Number(process.hrtime.bigint() - start) / 1e9;
    const route = (req.route && req.baseUrl + req.route.path) || req.path;
    httpRequestsTotal.labels(SERVICE_NAME, req.method, route, String(res.statusCode)).inc();
    httpRequestDuration.labels(SERVICE_NAME, req.method, route).observe(durationSeconds);
  });
  next();
}

async function metricsHandler(_req, res) {
  res.set("Content-Type", client.register.contentType);
  res.end(await client.register.metrics());
}

// --- Health ----------------------------------------------------------------

// healthHandler reports liveness plus Neo4j connectivity, so a monitoring
// app can distinguish "process is up" from "process is up but its database
// is unreachable".
function healthHandler(driver) {
  return async (_req, res) => {
    let neo4jStatus = "ok";
    let overall = "ok";
    try {
      await driver.verifyConnectivity();
    } catch (err) {
      overall = "degraded";
      neo4jStatus = `error: ${err.message}`;
    }
    res.status(overall === "ok" ? 200 : 503).json({
      service: SERVICE_NAME,
      status: overall,
      checks: { neo4j: neo4jStatus },
    });
  };
}

module.exports = {
  logger,
  requestLogger,
  metricsMiddleware,
  metricsHandler,
  healthHandler,
};
