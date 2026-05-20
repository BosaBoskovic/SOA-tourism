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
