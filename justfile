generate-api: clean
    mkdir -p ./internal/api
    oapi-codegen -package api -generate types -o ./internal/api/models.go ./api/openapi/openapi.yaml
    oapi-codegen -package api -generate chi-server -o ./internal/api/server.go ./api/openapi/openapi.yaml    

clean:
    rm -f ./internal/api/server.go ./internal/api/models.go
