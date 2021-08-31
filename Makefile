gen:
	protoc --proto_path=proto proto/*.proto --go_out=:. --go-grpc_out=require_unimplemented_servers=false:. --grpc-gateway_out=:. --openapiv2_out=:swagger
clean:
	rm pb/*go
run:
	go run main.go
server:
	go run cmd/server/main.go -port 8080
client:
	go run cmd/client/main.go -address 0.0.0.0:8080
test:
	go test -cover -race ./...
certs:
	cd certs; sh gen.sh; cd ..


.PHONY: gen clean server client test certs