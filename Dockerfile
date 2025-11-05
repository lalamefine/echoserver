FROM golang:1.23-alpine
ENV PORT=8080

WORKDIR /app
COPY . .

RUN go build -o echoserver main.go

ENTRYPOINT ["./echoserver"]