package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("CCProxy Safe Mode - Minimal startup")
	fmt.Println("This version bypasses all provider initialization")
	
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)
	
	// Create minimal router
	router := gin.New()
	router.Use(gin.Recovery())
	
	// Add only health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"mode":   "safe",
			"message": "Running in safe mode - no providers initialized",
		})
	})
	
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "CCProxy Safe Mode",
			"version": "1.0.0-safe",
		})
	})
	
	// Create server with timeouts
	server := &http.Server{
		Addr:         ":3456",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	
	// Handle shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	
	// Start server
	go func() {
		fmt.Println("Starting server on :3456...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()
	
	fmt.Println("Server started in safe mode. Press Ctrl+C to stop.")
	fmt.Println("Test with: curl http://localhost:3456/health")
	
	// Wait for signal
	<-stop
	fmt.Println("Shutting down...")
	
	// Shutdown with timeout
	if err := server.Close(); err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	}
	
	fmt.Println("Server stopped.")
}