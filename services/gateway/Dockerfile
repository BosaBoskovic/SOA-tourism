# Builder stage: kompajlira Go aplikaciju
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Kopiraj shared proto modul
COPY proto/gen/go ./proto/gen/go

# Kopiraj go.mod i go.sum
COPY services/gateway/go.mod services/gateway/go.sum ./services/gateway/

WORKDIR /build/services/gateway

# Preuzmi zavisnosti
RUN go mod download

# Kopiraj izvorni kod
COPY services/gateway .

# Kompajliraj aplikaciju
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway .

# Runtime stage: minimalni kontejner sa samo binarnom datotekom
FROM alpine:latest

# Instaliraj ca-certificates za HTTPS konektivnost
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Kopiraj kompajlirani binar iz builder stage-a
COPY --from=builder /build/services/gateway/gateway .

# Izloži port koji koristi gateway
EXPOSE 8080

# Pokreni aplikaciju
ENTRYPOINT ["./gateway"]