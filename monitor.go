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

// Track connection state
var connectionState = struct {
	lastStatus        string
	lastStatusTime    time.Time
	lastDownTime      time.Time
	lastUpTime        time.Time
	currentUptime     time.Duration
	previousDowntime  time.Duration
	previousUptime    time.Duration
	connectionChanges int
}{
	lastStatus: "UNKNOWN",
}

func main() {
	// Initialize log file if it doesn't exist
	if _, err := os.Stat(config.LogFile); os.IsNotExist(err) {
		file, err := os.Create(config.LogFile)
		if err != nil {
			fmt.Printf("Error creating log file: %v\n", err)
			return
		}
		file.WriteString("timestamp,status,latency,uptime,downtime,total_changes,message\n")
		file.Close()
		fmt.Printf("Created log file: %s\n", config.LogFile)
	}

	// Initialize state tracking
	connectionState.lastStatusTime = time.Now()

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
	now := time.Now()
	timestamp := now.Format(time.RFC3339)
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
	var uptimeStr string
	var downtimeStr string

	// We don't need to calculate this since we're tracking specific up/down times instead

	if err != nil {
		status = "DOWN"
		latency = -1
		message = err.Error()

		// Update timing for status change
		if connectionState.lastStatus == "UP" || connectionState.lastStatus == "UNKNOWN" {
			// Connection just went down
			connectionState.lastDownTime = now
			connectionState.previousUptime = now.Sub(connectionState.lastUpTime)
			connectionState.connectionChanges++
			message = fmt.Sprintf("Connection lost after %s uptime", formatDuration(connectionState.previousUptime))
		} else {
			// Still down
			connectionState.currentUptime = 0
			message = "Connection still down"
		}

		// Format durations for logging
		uptimeStr = formatDuration(connectionState.previousUptime)
		downtimeStr = formatDuration(now.Sub(connectionState.lastDownTime))

		fmt.Printf("\033[31m✖ Connection DOWN at %s (Down for: %s)\033[0m\n",
			timestamp, downtimeStr)
	} else {
		status = "UP"
		// Parse the ping output to get latency - different formats for different OS
		latency = parseLatency(outputStr)

		// Update timing for status change
		if connectionState.lastStatus == "DOWN" || connectionState.lastStatus == "UNKNOWN" {
			// Connection just recovered
			connectionState.lastUpTime = now
			connectionState.previousDowntime = now.Sub(connectionState.lastDownTime)
			connectionState.connectionChanges++
			message = fmt.Sprintf("Connection restored after %s downtime", formatDuration(connectionState.previousDowntime))
		} else {
			// Still up
			connectionState.currentUptime = now.Sub(connectionState.lastUpTime)
			message = "Connection stable"
		}

		// Format durations for logging
		uptimeStr = formatDuration(now.Sub(connectionState.lastUpTime))
		downtimeStr = formatDuration(connectionState.previousDowntime)

		fmt.Printf("\033[32m✓ Connection UP at %s (Up for: %s, Latency: %dms)\033[0m\n",
			timestamp, uptimeStr, latency)
	}

	// Update state for next check
	connectionState.lastStatus = status
	connectionState.lastStatusTime = now

	// Escape double quotes in the message for CSV
	message = strings.ReplaceAll(message, "\"", "\"\"")

	// Log the result
	logEntry := fmt.Sprintf("%s,%s,%d,%s,%s,%d,\"%s\"\n",
		timestamp, status, latency, uptimeStr, downtimeStr, connectionState.connectionChanges, message)
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

// Helper to format duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "0s"
	}

	d = d.Round(time.Second)

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		m := d / time.Minute
		s := (d % time.Minute) / time.Second
		return fmt.Sprintf("%dm%ds", m, s)
	} else {
		h := d / time.Hour
		m := (d % time.Hour) / time.Minute
		s := (d % time.Minute) / time.Second
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
}
