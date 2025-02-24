package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Configuration
var config = struct {
	PingTarget    string
	CheckInterval int
	LogFile       string
	PingCount     int
	PingTimeout   int
}{
	PingTarget:    "8.8.8.8",
	CheckInterval: 30,
	LogFile:       "connection_log.csv",
	PingCount:     3,
	PingTimeout:   5,
}

func main() {
	// Initialize log file if it doesn't exist
	if _, err := os.Stat(config.LogFile); os.IsNotExist(err) {
		file, err := os.Create(config.LogFile)
		if err != nil {
			fmt.Printf("Error creating log file: %v\n", err)
			return
		}
		file.WriteString("timestamp,status,latency,message\n")
		file.Close()
		fmt.Printf("Created log file: %s\n", config.LogFile)
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
	go checkConnection()

	// Main loop
	for {
		select {
		case <-ticker.C:
			go checkConnection()
		case <-sigChan:
			fmt.Println("\nConnection monitor stopped")
			return
		}
	}
}

func checkConnection() {
	timestamp := time.Now().Format(time.RFC3339)
	var cmd *exec.Cmd

	// Determine ping command based on OS
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", fmt.Sprintf("%d", config.PingCount),
			"-w", fmt.Sprintf("%d", config.PingTimeout*1000), config.PingTarget)
	} else {
		cmd = exec.Command("ping", "-c", fmt.Sprintf("%d", config.PingCount),
			"-W", fmt.Sprintf("%d", config.PingTimeout), config.PingTarget)
	}

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	var status string
	var latency int64
	var message string

	if err != nil {
		status = "DOWN"
		latency = -1
		message = err.Error()
		fmt.Printf("\033[31m✖ Connection DOWN at %s\033[0m\n", timestamp)
	} else {
		status = "UP"
		// Parse the ping output to get latency - different formats for different OS
		latency = parseLatency(outputStr)
		message = "Connection successful"
		fmt.Printf("\033[32m✓ Connection UP at %s (%dms)\033[0m\n", timestamp, latency)
	}

	// Escape double quotes in the message for CSV
	message = strings.ReplaceAll(message, "\"", "\"\"")

	// Log the result
	logEntry := fmt.Sprintf("%s,%s,%d,\"%s\"\n", timestamp, status, latency, message)
	appendToLog(logEntry)
}

func parseLatency(output string) int64 {
	// Try to extract average latency (works for most OS formats)
	var latencyPattern *regexp.Regexp

	if runtime.GOOS == "windows" {
		latencyPattern = regexp.MustCompile(`Average\s*=\s*(\d+)ms`)
	} else {
		// Linux / MacOS pattern
		latencyPattern = regexp.MustCompile(`min/avg/max/[^=]+=\s*[0-9.]+/([0-9.]+)/`)
	}

	matches := latencyPattern.FindStringSubmatch(output)
	if len(matches) >= 2 {
		// Convert to int, handle any decimal points (e.g., from macOS)
		parts := strings.Split(matches[1], ".")
		if val, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
			return val
		}
	}

	return 0
}

func appendToLog(entry string) {
	file, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(entry); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
	}
}
