FROM golang:1.24-alpine3.21 AS builder

WORKDIR /app

RUN apk add --no-cache protobuf protobuf-dev git just
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN just generate-all

RUN go build -o app cmd/pills-taking-reminder/main.go


EXPOSE 8080 8081

CMD ["./app"]
