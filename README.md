# wheels

### gRPC Stream Example

```bash
$ protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpc/stream/fibonacci.proto

$ go mod tidy

$ go build -o server ./grpc/stream/server
$ go build -o client ./grpc/stream/client
```
