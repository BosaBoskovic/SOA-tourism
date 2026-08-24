// Must be required before any other module (see package.json "start" script,
// which preloads this file with `node -r`) so the OpenTelemetry
// auto-instrumentations can patch http/express before they are `require()`d.
//
// Service name and collector endpoint come from the standard OTEL_SERVICE_NAME /
// OTEL_EXPORTER_OTLP_ENDPOINT env vars (same convention as the Go services),
// which the SDK and exporter both pick up on their own.
const { NodeSDK } = require("@opentelemetry/sdk-node");
const { getNodeAutoInstrumentations } = require("@opentelemetry/auto-instrumentations-node");
const { OTLPTraceExporter } = require("@opentelemetry/exporter-trace-otlp-http");

const serviceName = process.env.OTEL_SERVICE_NAME || "followers-service";

const sdk = new NodeSDK({
  traceExporter: new OTLPTraceExporter(),
  instrumentations: [getNodeAutoInstrumentations()],
});

try {
  sdk.start();
} catch (err) {
  // eslint-disable-next-line no-console
  console.error(JSON.stringify({ level: "error", service: serviceName, msg: "opentelemetry init failed", error: String(err) }));
}

process.on("SIGTERM", () => {
  sdk.shutdown().finally(() => process.exit(0));
});
