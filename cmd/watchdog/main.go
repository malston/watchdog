// Package main is the entry point for the watchdog application
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/malston/watchdog/internal/monitor"
	"github.com/malston/watchdog/internal/server"
)

func main() {
	// Load default configuration
	config := monitor.DefaultConfig()

	// Parse command line flags
	flag.StringVar(&config.PingTarget, "ping-target", config.PingTarget, "Target to ping")
	flag.IntVar(&config.CheckInterval, "check-interval", config.CheckInterval, "Interval between checks in seconds")
	flag.StringVar(&config.LogFile, "log-file", config.LogFile, "Log file path")
	flag.IntVar(&config.PingCount, "ping-count", config.PingCount, "Number of ping packets to send")
	flag.IntVar(&config.PingTimeout, "ping-timeout", config.PingTimeout, "Ping timeout in seconds")

	// Add API server port
	apiPort := flag.Int("api-port", 8080, "Port for the HTTP API server")

	flag.Parse()

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize connection state
	state := monitor.NewConnectionState()

	// Initialize logger
	logger, err := monitor.NewLogger(config.LogFile)
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	// Start the HTTP API server
	apiServer := server.NewServer(config.LogFile, *apiPort)
	if err := apiServer.Start(); err != nil {
		fmt.Printf("Error starting API server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting connection monitor (checking every %d seconds)\n", config.CheckInterval)
	fmt.Printf("Pinging %s with %d packets every check\n", config.PingTarget, config.PingCount)
	fmt.Printf("Logging results to %s\n", config.LogFile)

	// Set up channel to handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Set up ticker for periodic checks
	ticker := time.NewTicker(time.Duration(config.CheckInterval) * time.Second)
	defer ticker.Stop()

	// Run the first check immediately
	result := monitor.Check(config, state)
	if err := logger.LogResult(result); err != nil {
		fmt.Printf("Error logging result: %v\n", err)
	}

	// Main loop
	for {
		select {
		case <-ticker.C:
			result := monitor.Check(config, state)
			if err := logger.LogResult(result); err != nil {
				fmt.Printf("Error logging result: %v\n", err)
			}
		case <-sigChan:
			fmt.Println("\nConnection monitor stopped")
			return
		}
	}
}
