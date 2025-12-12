## What it does : 

This is a server that listens for HTTP requests and echoes back any data it receives. 
It also logs the number of requests it has handled every second (in counter mode).
Or it can log detailed request information including method, URL, headers, and body (in log mode).
Primary use case is for debugging and load testing.

## Run with one of the following commands: 
```sh
go run main.go
go build -o echoserver main.go && ./echoserver
docker run -p 80:80 lalamefine/echoserver:latest
```
Accepted cmd flags: 
- `-p <port>` : Port to listen on (default 80)
- `-d <delay>` : Delay between prints in seconds (default 1)
- `-m <mode>` : Mode of operation: "count" (default) to log request count, "log" to log request method, URL, Headers and body
- `-sct_token <token>` : If set, the server will respond to `/.well-known/scale-test-claim-token.txt` with this token (plain text)

Accepted environment variables:
- `PORT` : Port to listen on (default 80)
- `PRINT_DELAY` : Delay between prints in seconds (default 1)
- `MODE` : Mode of operation: "count" (default) to log request count, "log" to log request method, URL, Headers and body
 - `SCT_TOKEN` : If set, the server will respond to `/.well-known/scale-test-claim-token.txt` with this token (plain text)

Flags take precedence over environment variables.

## Docker compose example:
```yaml
version: '3'
services:
  echoserver:
    image: lalamefine/echoserver:latest
    ports:
      - "80:80" # Map host port to container port if needed
    environment:
      - PORT=80 #(optional) port number to listen on
      - PRINT_DELAY=1 #(optional) delay between prints in seconds
      - MODE=count #(optional) other mode is log
```
### Sample output in count mode:
```
>echoserver -p 8080
Echoserver on :8080
2025-11-05 23:59:27: 2 Requests
2025-11-05 23:59:29: 8 Requests
2025-11-05 23:59:30: 3 Requests
2025-11-05 23:59:32: 1 Requests
```
### Sample output in log mode on a GET request (Body empty):
```
>echoserver -p 80 -m log
Echoserver on :80
------> 2025-11-06 00:06:54 GET /
Accept-Encoding: gzip, deflate, br, zstd
Upgrade-Insecure-Requests: 1
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36
Connection: keep-alive
Sec-Ch-Ua-Platform: "Windows"
Sec-Fetch-Site: none
Cache-Control: max-age=0
Sec-Ch-Ua: "Google Chrome";v="141", "Not?A_Brand";v="8", "Chromium";v="141"
Sec-Ch-Ua-Mobile: ?0
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7
Sec-Fetch-Dest: document
--------------------

--------------------
```