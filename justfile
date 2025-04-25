generate-restapi: clean-restapi
    mkdir -p ./internal/api
    oapi-codegen -package api -generate types -o ./internal/api/models.go ./api/openapi/openapi.yaml
    oapi-codegen -package api -generate chi-server -o ./internal/api/server.go ./api/openapi/openapi.yaml    

generate-grpc: clean-grpc
    mkdir -p ./internal/grpc/pb
    protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        api/proto/pills.proto
    mv api/proto/*.pb.go internal/grpc/pb
    
generate-all: generate-restapi generate-grpc

clean-restapi:
    rm -f ./internal/api/server.go ./internal/api/models.go

clean-grpc:
    rm -rf ./internal/grpc/pb


build:
    go build -o ./bin ./cmd/pills-taking-reminder/main.go 

run: build
    ./bin/main

