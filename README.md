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
- **Structured JSON logs** — every service logs JSON to stdout (`service`, `level`, `msg`, plus `trace_id`/`span_id` when a span is active). Promtail ships container logs to Loki using Docker service discovery (relabeled from the `com.docker.compose.service` container label, not the container name, so it stays correct once a service runs as multiple replicas), labels them by `service` and `level`, and stores `trace_id`/`span_id` as Loki structured metadata (not labels, to avoid blowing up the index) — from a log line in Grafana you can jump straight to its Jaeger trace. payments' (.NET) trace/span IDs land nested under the JSON `Scopes` array rather than top-level; Promtail's pipeline extracts both shapes (flat top-level for Go/Node/Java, `Scopes[].TraceId`/`SpanId` for payments) into the same `trace_id`/`span_id` metadata either way.
- **gRPC tracing gap closed** — blog's `net.devh` gRPC server now extracts inbound W3C trace context via a custom `GrpcTracingInterceptor` (`services/blog/.../config/GrpcTracingInterceptor.java`), so a trace entering blog over gRPC continues the caller's trace instead of starting a new one.

**Grafana** (`:3001`, admin/admin) dashboards: **Services Overview** (up/down, request rate, 5xx rate, p95 latency, RabbitMQ backlog), **Container Metrics** (per-container CPU/RAM/network, plus per-*service* memory/CPU % of its `deploy.resources.limits` with 75%/90% thresholds, instance count, OOM-kill count, and restart/flap detection - all still correct once a service is scaled to N replicas), **Host Machine Metrics**, **Agregacija Logova** (log volume, all-service error feed, and a `$service` dropdown driving one log panel for whichever service you pick), and **Messaging & Databases** (RabbitMQ per-queue depth/consumers/publish-deliver rates, Postgres/Mongo up-status and connection counts via `postgres_exporter`/`mongodb_exporter`).

Alert rules (on top of the existing host CPU/RAM/disk alerts): service down, >5% 5xx rate, >1s p95 latency, RabbitMQ queue backlog, notification dead-letter queue non-empty, and a **Resource & Scaling Alerts** group - per-service memory/CPU >75% (warning) and >90% (critical), any OOM-kill event, and service flapping (up/down more than twice in 5 minutes).

### Resource limits (Faza 1)

Every app service now has a `deploy.resources.limits` in `docker-compose.yml` (memory + cpus), sized from an idle-usage baseline captured via `docker stats`: gateway 256M/0.5, stakeholders/tours 192M/0.5, encounters 128M/0.3, followers 256M/0.5, blog 512M/1.0 (JVM heap capped via `JDK_JAVA_OPTIONS=-XX:MaxRAMPercentage=75.0` so it fits under the container limit instead of a fixed `-Xmx`), payments 384M/0.75. Previously **no service had any resource limit** - cAdvisor's `container_spec_memory_limit_bytes`/`container_spec_cpu_quota` were unset, so a "% of limit" query (and therefore any "is this service critical" signal) was impossible. The DB/broker containers also got generous limits; the observability stack itself (Prometheus/Grafana/Loki/Promtail/Jaeger/cAdvisor/node-exporter) is deliberately left unconstrained so it keeps working under host memory pressure.

A separate monitoring app should read this data straight from the existing backends rather than have services push to it separately: **Prometheus** HTTP API (`:9090`) for metrics, **Loki** HTTP API (`:3100`) for logs, **Jaeger** HTTP API (`:16686`) for traces.

## Operations

### Scaling a service

`stakeholders`, `blog`, `followers`, `tours`, `encounters`, and `payments` can run as multiple replicas - none of them pin a fixed `container_name` anymore, so Docker's embedded DNS resolves their bare service name (e.g. `stakeholders`) to every healthy replica instead of just one. `gateway` stays single-instance (it's the only published host port, `8080:8080`).

```
docker compose up -d --scale stakeholders=3
```

This is a **one-off override** - the next plain `docker compose up -d <anything>` (with no `--scale` flag) reconciles every service back to its default count of 1 and removes the extra replicas. Repeat `--scale` on every `up` while you want it, or add a permanent `deploy.replicas:` to the service in `docker-compose.yml` instead.

How each path picks up new replicas differs, and this matters for how quickly scaling actually helps:
- **REST** (gateway's reverse proxies, and the outbound HTTP clients between services) re-resolve DNS on every single request (keep-alives are deliberately disabled for exactly this - see `scalableServiceTransport()` in `services/gateway/main.go`), so a newly-started replica starts receiving traffic on its very next request. No restart needed.
- **gRPC** (gateway's 4 downstream gRPC clients - stakeholders/tours/payments/blog) only re-resolves DNS periodically in the background (grpc-go's default interval is long enough that it won't help mid-load-test). **Restart the gateway after scaling** (`docker compose restart gateway`) to have it dial fresh and pick up the new replica count immediately - verified empirically: scaling stakeholders to 3 without restarting gateway kept 100% of gRPC traffic on the original replica; restarting gateway afterward spread it evenly across all 3.

Watch it work on the **Container Metrics** Grafana dashboard - the "Broj instanci po servisu" panel reflects the new replica count, and per-service memory/CPU % (aggregated across replicas) should drop as load spreads out. RabbitMQ-consuming services (stakeholders' notification delivery, tours' `purchase-completed` consumer) also drain their queue faster with more replicas, since RabbitMQ treats every replica's consumers as one competing-consumer pool automatically - visible on the **Messaging & Databases** dashboard.

payments' EF Core migration (`db.Database.Migrate()` on startup) is safe to scale too: it's now guarded by a Postgres advisory lock, so N replicas starting at once serialize on it instead of racing each other, and every replica after the first is a no-op.

### Login / change-password throttling

`/stakeholders/login` and `PUT /stakeholders/password` are protected by a progressive-backoff limiter (`services/stakeholders/ratelimit/backoff_limiter.go`), keyed per identity (username/email for login, the authenticated username for change-password): the first 2 failures are free, then every failure after that doubles the lockout - 1s, 2s, 4s, 8s, ... capped at 15 minutes - instead of a flat "N strikes, fixed timeout" rule. A locked-out identity gets `429 Too Many Requests` with a `Retry-After` header (and the same value as `retryAfterSeconds` in the JSON body); a successful attempt resets the streak to zero. Repeatedly hitting the endpoint *during* a lockout does not extend it further - only a request that actually reaches the credential check (i.e. after the lockout has expired) can do that - so an attacker flooding a locked-out account can't use that flood to indefinitely deny the real user access to their own account; it just wastes the attacker's requests. Single-instance/in-memory only, like the observability caveats above - if stakeholders is scaled to N replicas, each replica tracks its own counters.

### Simulating overload and failure

The pieces above exist specifically so this is safe and observable once you're ready to try it - none of the following is built into the app itself (no chaos-injection endpoints, no built-in load generator), it's a manual runbook:

- **Overload a service**: point a load-testing tool (e.g. [k6](https://k6.io) or [Locust](https://locust.io)) at the gateway (`http://localhost:8080`) and ramp concurrency until a service's Phase-1 `deploy.resources.limits` gets tight. Watch the **Resource & Scaling Alerts** fire in Grafana's Alerting view at 75%, then 90%; if it's memory-bound and pushed past its limit, the container gets OOM-killed by the kernel (`restart: on-failure` brings it back) - visible on the "OOM-kill događaji po servisu" panel and as a gap in that service's logs on the Loki dashboard. Scale the service (see above) and re-run the same load to see the difference.
- **Simulate a crash**: `docker compose pause stakeholders` (hangs it without killing - good for testing timeouts/retries downstream) or `docker compose stop stakeholders` (clean shutdown, `restart: on-failure` brings it back) or `docker kill soa-tourism-stakeholders-1` (hard kill, closest to an actual crash). Watch the "Restart / flap po servisu" panel and the `alert-service-flapping`/`alert-service-down` alerts, and check Jaeger for how callers (gateway, tours, blog, followers) handled the failure - do their traces show retries, timeouts, or an unhandled error propagating to the client?
- **Watch the RabbitMQ side specifically**: stop every `stakeholders` replica while load is still flowing through tours/payments that publish to `notification-delivery` - messages queue up (visible on the Messaging dashboard and the RabbitMQ backlog alert) instead of being lost, then drain once stakeholders comes back.

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
