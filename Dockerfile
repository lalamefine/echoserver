FROM golang:1.23-alpine
ENV PORT=80
ENV PRINT_DELAY=1
ENV MODE=count
ENV SCT_TOKEN=""
EXPOSE 80

WORKDIR /app
COPY . .

RUN go build -o echoserver main.go

ENTRYPOINT ["./echoserver"]