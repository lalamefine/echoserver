package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var reqCounter int64 = 0

func handler(w http.ResponseWriter, r *http.Request) {
	body := make([]byte, r.ContentLength)
	r.Body.Read(body)
	w.Write(body)
	reqCounter++
}

func main() {
	port := 80
	if p, ok := os.LookupEnv("PORT"); ok {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}
	delayBetweenPrints := 1
	portFlag := flag.Int("p", port, "port number to listen on")
	delayFlag := flag.Int("d", delayBetweenPrints, "delay between prints in seconds")
	flag.Parse()
	port = *portFlag
	delayBetweenPrints = *delayFlag

	address := fmt.Sprintf(":%d", port)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	srv := &http.Server{
		Addr: address,
	}
	srv.Handler = http.HandlerFunc(handler)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer stop() // if context did not get canceled, cancel it so we don't block below
		fmt.Println("Echoserver on", address)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Println("Error starting server:", err)
		}
	}()
	go func() {
		<-ctx.Done()
		err := srv.Shutdown(context.Background())
		if err != nil {
			fmt.Println("Error shutting down server:", err)
		}
	}()
	go func() {
		for {
			time.Sleep(time.Duration(delayBetweenPrints) * time.Second)
			logUsage()
		}
	}()
	wg.Wait()
}

func logUsage() {
	if reqCounter > 0 {
		fmt.Printf("%s: %d Requests\n", time.Now().Format(time.DateTime), reqCounter)
	}
	reqCounter = 0
}
