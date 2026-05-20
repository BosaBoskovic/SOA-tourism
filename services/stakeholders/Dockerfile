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

# Pokreni aplikaciju
ENTRYPOINT ["./stakeholders"]
