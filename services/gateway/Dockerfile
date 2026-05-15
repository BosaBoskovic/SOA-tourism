# Builder stage: kompajlira Go aplikaciju
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Kopiraj go.mod (gateway nema eksternih zavisnosti, go.sum nije potreban)
COPY go.mod ./

# Preuzmi zavisnosti (nema ih, ali ostavljamo za konzistentnost)
RUN go mod download

# Kopiraj izvorni kod
COPY . .

# Kompajliraj aplikaciju
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway .

# Runtime stage: minimalni kontejner sa samo binarnom datotekom
FROM alpine:latest

# Instaliraj ca-certificates za HTTPS konektivnost
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Kopiraj kompajlirani binar iz builder stage-a
COPY --from=builder /build/gateway .

# Izloži port koji koristi gateway
EXPOSE 8080

# Pokreni aplikaciju
ENTRYPOINT ["./gateway"]