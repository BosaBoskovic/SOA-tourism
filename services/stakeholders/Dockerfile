# Builder stage: kompajlira Go aplikaciju
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Kopiraj shared proto modul
COPY proto/gen/go ./proto/gen/go

# Kopiraj go.mod i go.sum
COPY services/stakeholders/go.mod services/stakeholders/go.sum ./services/stakeholders/

WORKDIR /build/services/stakeholders

# Preuzmi zavisnosti
RUN go mod download

# Kopiraj izvorni kod
COPY services/stakeholders .

# Sinhronizuj go.mod/go.sum sa stvarnim importima (dodaje observability zavisnosti)
RUN go mod tidy

# Kompajliraj aplikaciju
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o stakeholders .

# Runtime stage: minimal kontejner sa samo binarnom datotekom
FROM alpine:latest

# Instaliraj ca-certificates za HTTPS konektivnost
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Kopiraj kompajlirani binar iz builder stage-a
COPY --from=builder /build/services/stakeholders/stakeholders .

# Izloži port koji koristi servis
EXPOSE 8081

# Docker healthcheck oslanja se na /health endpoint (vidi observability.go)
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=5 \
    CMD wget -q -O- http://localhost:8081/health || exit 1

# Pokreni aplikaciju
ENTRYPOINT ["./stakeholders"]
