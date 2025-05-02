generate-restapi: clean-restapi
    mkdir -p ./internal/api/http/generated
    oapi-codegen -package api -generate types -o ./internal/api/http/models.go ./api/openapi/openapi.yaml
    oapi-codegen -package api -generate chi-server -o ./internal/api/http/server.go ./api/openapi/openapi.yaml    

generate-grpc: clean-grpc
    mkdir -p ./internal/api/grpc/pb
    protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        api/proto/pills.proto
    mv api/proto/*.pb.go internal/api/grpc/pb
    
generate-all: generate-restapi generate-grpc

clean-restapi:
    rm -f ./internal/api/http/server.go ./internal/api/http/models.go

clean-grpc:
    rm -rf ./internal/api/grpc/pb


build:
    go build -o ./bin ./cmd/pills-taking-reminder/main.go 

run: build
    ./bin/main

