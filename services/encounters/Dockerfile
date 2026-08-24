# Builder stage: kompajlira Go aplikaciju
FROM golang:1.26-alpine AS builder

WORKDIR /build/services/encounters

# Kopiraj go.mod i go.sum
COPY services/encounters/go.mod services/encounters/go.sum ./

# Preuzmi zavisnosti
RUN go mod download

# Kopiraj izvorni kod
COPY services/encounters .

# Sinhronizuj go.mod/go.sum sa stvarnim importima (dodaje observability zavisnosti)
RUN go mod tidy

# Kompajliraj aplikaciju
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o encounters .

# Runtime stage: minimalni kontejner sa samo binarnom datotekom
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/services/encounters/encounters .

EXPOSE 8083

# Docker healthcheck oslanja se na /health endpoint (vidi observability.go)
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=5 \
    CMD wget -q -O- http://localhost:8083/health || exit 1

ENTRYPOINT ["./encounters"]
