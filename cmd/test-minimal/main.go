package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("[TEST] Starting minimal HTTP server test...")
	
	// Create simple HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Create server
	server := &http.Server{
		Addr:         ":3456",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	// Handle shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	
	// Start server
	go func() {
		fmt.Println("[TEST] Starting server on :3456...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[TEST] Server error: %v\n", err)
		}
	}()
	
	fmt.Println("[TEST] Server started. Press Ctrl+C to stop.")
	
	// Wait for signal
	<-stop
	fmt.Println("[TEST] Shutting down...")
	
	// Shutdown
	server.Close()
	fmt.Println("[TEST] Server stopped.")
}