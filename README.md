### What it does : 

This is a server that listens for HTTP requests and echoes back any data it receives. It also logs the number of requests it has handled every second.

### Run with : 
```sh
docker run -p 80:80 lalamefine/echoserver:latest
```
Accepted flags: 
- `-p <port>` : Port to listen on (default 80)
- `-d <delay>` : Delay between prints in seconds (default 1)

You can also set the `PORT` environment variable to change the listening port.