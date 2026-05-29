module tours

go 1.26.1

require (
	github.com/gorilla/mux v1.8.1
	go.mongodb.org/mongo-driver/v2 v2.6.0
	google.golang.org/grpc v1.81.1
	soa-tourism-proto v0.0.0
)

require (
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace soa-tourism-proto => ../../proto/gen/go
