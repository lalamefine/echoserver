package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var reqCounter int64 = 0 // counts requests per print period
var totalRequests int64 = 0
var totalBytes int64 = 0
var periodBytes int64 = 0
var mode string = "count"
var startTime = time.Now()
var sctToken string = ""

func handler(w http.ResponseWriter, r *http.Request) {
	// read body to echo it back and to count bytes
	body, _ := io.ReadAll(r.Body)
	w.Write(body)

	// update counters atomically
	n := int64(len(body))
	atomic.AddInt64(&totalBytes, n)
	atomic.AddInt64(&totalRequests, 1)
	atomic.AddInt64(&reqCounter, 1)
	atomic.AddInt64(&periodBytes, n)

	if mode == "log" {
		fmt.Printf("------> %s %s %s\n", time.Now().Format(time.DateTime), r.Method, r.URL.String())
		for name, values := range r.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", name, value)
			}
		}
		fmt.Printf("--------------------\n")
		fmt.Println(string(body))
		fmt.Printf("--------------------\n")
	}
}

// statHandler returns traffic statistics since server start.
func statHandler(w http.ResponseWriter, r *http.Request) {
	totalReq := atomic.LoadInt64(&totalRequests)
	totalB := atomic.LoadInt64(&totalBytes)
	uptime := time.Since(startTime).Seconds()
	if uptime < 1e-9 {
		uptime = 1e-9
	}

	avgPerSec := float64(totalReq) / uptime
	avgBytesPerSec := float64(totalB) / uptime
	avgBytesPerReq := float64(0)
	if totalReq > 0 {
		avgBytesPerReq = float64(totalB) / float64(totalReq)
	}

	// prepare response
	resp := map[string]interface{}{
		"total_requests":            totalReq,
		"uptime_seconds":            uptime,
		"average_requests_per_sec":  avgPerSec,
		"total_bytes":               totalB,
		"average_bytes_per_sec":     avgBytesPerSec,
		"average_bytes_per_request": avgBytesPerReq,
	}

	// support HTML simple view when requested
	if r.URL.Query().Get("format") == "html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<html><head><title>Stats</title></head><body>")
		fmt.Fprintf(w, "<h1>Traffic stats</h1>")
		fmt.Fprintf(w, "<ul>")
		fmt.Fprintf(w, "<li>Total requests: %d</li>", totalReq)
		fmt.Fprintf(w, "<li>Uptime (s): %.2f</li>", uptime)
		fmt.Fprintf(w, "<li>Average requests/sec: %.4f</li>", avgPerSec)
		fmt.Fprintf(w, "<li>Total bytes: %d</li>", totalB)
		fmt.Fprintf(w, "<li>Average bytes/sec: %.4f</li>", avgBytesPerSec)
		fmt.Fprintf(w, "<li>Average bytes/request: %.4f</li>", avgBytesPerReq)
		fmt.Fprintf(w, "</ul></body></html>")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}

// sctHandler serves the SCT token when configured.
func sctHandler(w http.ResponseWriter, r *http.Request) {
	if sctToken == "" {
		http.NotFound(w, r)
		return
	}
	// Only allow GET
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(sctToken))
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
	if t, ok := os.LookupEnv("SCT_TOKEN"); ok {
		sctToken = t
	}
	portFlag := flag.Int("p", port, "port number to listen on")
	delayFlag := flag.Int("d", delayBetweenPrints, "delay between prints in seconds")
	modeFlag := flag.String("m", mode, "mode of operation (count, logbody or logheaders)")
	sctFlag := flag.String("sct_token", sctToken, "token to serve at /.well-known/scale-test-claim-token.txt or via SCT_TOKEN env var")
	flag.Parse()
	port = *portFlag
	delayBetweenPrints = *delayFlag
	mode = *modeFlag
	sctToken = *sctFlag
	if mode != "count" && mode != "log" {
		fmt.Println("Invalid mode specified. Valid modes are 'count' and 'log'.")
		return
	}

	address := fmt.Sprintf(":%d", port)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	srv := &http.Server{
		Addr: address,
	}

	// use a ServeMux to add /stat endpoint
	mux := http.NewServeMux()
	mux.HandleFunc("/stat", statHandler)
	mux.HandleFunc("/.well-known/scale-test-claim-token.txt", sctHandler)
	mux.HandleFunc("/", handler)
	srv.Handler = mux
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
	r := atomic.SwapInt64(&reqCounter, 0)
	b := atomic.SwapInt64(&periodBytes, 0)
	if r > 0 || b > 0 {
		fmt.Printf("%s: %d Requests, %d Bytes\n", time.Now().Format(time.DateTime), r, b)
	}
}
