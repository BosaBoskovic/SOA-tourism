# Builder stage
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY services/tours/go.mod services/tours/go.sum ./
COPY proto/gen/go /proto/gen/go
RUN go mod download

COPY services/tours .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o tours .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/tours .

EXPOSE 8085

ENTRYPOINT ["./tours"]