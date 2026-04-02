# Builder stage: kompajlira Go aplikaciju
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Kopiraj go.mod i go.sum
COPY go.mod go.sum ./

# Preuzmi zavisnosti
RUN go mod download

# Kopiraj izvorni kod
COPY . .

# Kompajliraj aplikaciju
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o stakeholders .

# Runtime stage: minimal kontejner sa samo binarnom datotekom
FROM alpine:latest

# Instaliraj ca-certificates za HTTPS konektivnost
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Kopiraj kompajlirani binar iz builder stage-a
COPY --from=builder /build/stakeholders .

# Izloži port koji koristi servis
EXPOSE 8081

# Pokreni aplikaciju
ENTRYPOINT ["./stakeholders"]
