# SOA Tourism

## gRPC generation (stakeholders)

This repo generates Go gRPC stubs from the proto files under proto/.

### Prerequisites

- Go 1.26+

### Generate code

From the proto directory:

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
go run github.com/bufbuild/buf/cmd/buf@v1.31.0 generate
```

Generated code is written to proto/gen/go/ and is referenced via a Go module replace directive in the gateway and stakeholders services.

## gRPC runtime configuration

- Stakeholders gRPC server: STAKEHOLDERS_GRPC_PORT (default: 9091)
- Gateway gRPC client target: STAKEHOLDERS_GRPC_URL (default: localhost:9091)

## Observability

Every service (gateway, stakeholders, blog, followers, tours, payments, encounters) exposes:

- **`GET /health`** — liveness plus a real check of the service's own dependency (DB/driver connectivity). Returns `{"service","status","checks"}`, HTTP 200 when healthy, 503 when degraded. Used both by Docker healthchecks and by anything polling for uptime.
- **`GET /metrics`** — Prometheus exposition format. `http_requests_total` / `http_request_duration_seconds` (labeled by `service`, `method`, `route`, `status`) give RED metrics for every service; language runtime metrics (GC, heap, goroutines/threads) come along for free from each client library's default collectors.
- **Distributed tracing** — every inbound request gets an OpenTelemetry span, exported via OTLP to Jaeger (`http://soa-jaeger:4318`, UI on `:16686`). Trace context propagates across service boundaries over both REST (gateway's reverse proxies) and gRPC, so a single request through the gateway shows up as one connected trace. Known gap: blog's two gRPC-only endpoints (`net.devh` gRPC server) don't yet extract inbound trace context, so a trace that enters blog via gRPC starts a new trace there instead of continuing the gateway's.
- **Structured JSON logs** — every service logs JSON to stdout (`service`, `level`, `msg`, plus `trace_id`/`span_id` when a span is active). Promtail ships container logs to Loki using Docker service discovery, labels them by `service` and `level`, and stores `trace_id`/`span_id` as Loki structured metadata (not labels, to avoid blowing up the index) — from a log line in Grafana you can jump straight to its Jaeger trace. Note: payments' (.NET) trace/span IDs currently land nested under the JSON `Scopes` array rather than top-level, so Promtail doesn't lift them into structured metadata for that service; they're still visible in the raw log line and in Jaeger itself.

**Grafana** (`:3001`, admin/admin) has a **Services Overview** dashboard (up/down, request rate, 5xx rate, p95 latency, RabbitMQ backlog) alongside the existing host/container dashboards, plus new alert rules: service down, >5% 5xx rate, >1s p95 latency, and a RabbitMQ backlog warning — on top of the existing host CPU/RAM/disk alerts.

A separate monitoring app should read this data straight from the existing backends rather than have services push to it separately: **Prometheus** HTTP API (`:9090`) for metrics, **Loki** HTTP API (`:3100`) for logs, **Jaeger** HTTP API (`:16686`) for traces.

### Environment variables (new)

Every service reads:
- `OTEL_SERVICE_NAME` — service name reported to Jaeger/Prometheus/logs (defaults are set per service in docker-compose.yml).
- `OTEL_EXPORTER_OTLP_ENDPOINT` — OTLP collector base URL (defaults to Jaeger: `http://soa-jaeger:4318`).

### Verifying this builds

This change was made without a Go/Node/Java/Maven toolchain available in the editing environment, so nothing here has been build-verified. Before deploying:

- **Go services** (gateway, stakeholders, tours, encounters): their Dockerfiles now run `go mod tidy` during the build, so `docker compose build` should self-heal `go.mod`/`go.sum`. If building outside Docker, run `go mod tidy` in each service directory first.
- **followers** (Node): `npm install` (or the Docker build, which already runs it) will resolve the new dependencies from `package.json`.
- **blog** (Java): new Maven dependencies rely on Spring Boot's own dependency management (no explicit versions pinned for Spring-managed ones), which is the lower-risk path, but hasn't been run through `mvn package`.
- **payments** (.NET): new NuGet packages are pinned to specific versions; `dotnet restore`/`docker compose build` will confirm they resolve.
