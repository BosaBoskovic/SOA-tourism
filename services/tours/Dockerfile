# Builder stage
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY services/tours/go.mod services/tours/go.sum ./
COPY proto/gen/go /proto/gen/go
RUN go mod download

COPY services/tours .

# Sinhronizuj go.mod/go.sum sa stvarnim importima (dodaje observability zavisnosti)
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o tours .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/tours .

EXPOSE 8085

# Docker healthcheck oslanja se na /health endpoint (vidi observability.go)
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=5 \
    CMD wget -q -O- http://localhost:8085/health || exit 1

ENTRYPOINT ["./tours"]
