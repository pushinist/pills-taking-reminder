FROM golang:1.24-alpine3.21

WORKDIR /app

COPY . .

RUN go get -d -v ./...

RUN go build -o app cmd/pills-taking-reminder/main.go

EXPOSE 8080

CMD ["./app"]