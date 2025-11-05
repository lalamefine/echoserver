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
var mode string = "count"

func handler(w http.ResponseWriter, r *http.Request) {
	body := make([]byte, r.ContentLength)
	r.Body.Read(body)
	w.Write(body)
	switch mode {
	case "log":
		fmt.Printf("------> %s %s %s\n", time.Now().Format(time.DateTime), r.Method, r.URL.String())
		for name, values := range r.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", name, value)
			}
		}
		fmt.Printf("--------------------\n")
		fmt.Println(string(body))
		fmt.Printf("--------------------\n")
	case "count":
		reqCounter++
	}
}

func main() {
	port := 80
	if p, ok := os.LookupEnv("PORT"); ok {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}
	delayBetweenPrints := 1
	if d, ok := os.LookupEnv("PRINT_DELAY"); ok {
		if parsed, err := strconv.Atoi(d); err == nil {
			delayBetweenPrints = parsed
		}
	}
	if m, ok := os.LookupEnv("MODE"); ok {
		mode = m
	}
	portFlag := flag.Int("p", port, "port number to listen on")
	delayFlag := flag.Int("d", delayBetweenPrints, "delay between prints in seconds")
	modeFlag := flag.String("m", mode, "mode of operation (count, logbody or logheaders)")
	flag.Parse()
	port = *portFlag
	delayBetweenPrints = *delayFlag
	mode = *modeFlag
	if mode != "count" && mode != "log" {
		fmt.Println("Invalid mode specified. Valid modes are 'count' and 'log'.")
		return
	}

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
	if mode == "count" {
		go func() {
			for {
				time.Sleep(time.Duration(delayBetweenPrints) * time.Second)
				logUsage()
			}
		}()
	}
	wg.Wait()
}

func logUsage() {
	if reqCounter > 0 {
		fmt.Printf("%s: %d Requests\n", time.Now().Format(time.DateTime), reqCounter)
	}
	reqCounter = 0
}
