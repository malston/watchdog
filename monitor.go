package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
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

func init() {
	flag.StringVar(&config.PingTarget, "ping-target", config.PingTarget, "Target to ping")
	flag.IntVar(&config.CheckInterval, "check-interval", config.CheckInterval, "Interval between checks in seconds")
	flag.StringVar(&config.LogFile, "log-file", config.LogFile, "Log file path")
	flag.IntVar(&config.PingCount, "ping-count", config.PingCount, "Number of ping packets to send")
	flag.IntVar(&config.PingTimeout, "ping-timeout", config.PingTimeout, "Ping timeout in seconds")
	flag.Parse()

	if err := validateConfig(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}
}

func validateConfig() error {
	if config.PingTarget == "" {
		return fmt.Errorf("ping-target cannot be empty")
	}
	if config.CheckInterval <= 0 {
		return fmt.Errorf("check-interval must be greater than 0")
	}
	if config.LogFile == "" {
		return fmt.Errorf("log-file cannot be empty")
	}
	if config.PingCount <= 0 {
		return fmt.Errorf("ping-count must be greater than 0")
	}
	if config.PingTimeout <= 0 {
		return fmt.Errorf("ping-timeout must be greater than 0")
	}
	return nil
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

	// Start the HTTP server to serve the data to the frontend
	startHTTPServer()

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

	if err != nil {
		status = "DOWN"
		latency = -1
		message = fmt.Sprintf("Error: %v", err)

		if connectionState.lastStatus == "UP" || connectionState.lastStatus == "UNKNOWN" {
			connectionState.lastDownTime = now
			connectionState.previousUptime = now.Sub(connectionState.lastUpTime)
			connectionState.connectionChanges++
			message = fmt.Sprintf("Connection lost after %s uptime. %s", formatDuration(connectionState.previousUptime), message)
		} else {
			connectionState.currentUptime = 0
			message = "Connection still down. " + message
		}

		uptimeStr = formatDuration(connectionState.previousUptime)
		downtimeStr = formatDuration(now.Sub(connectionState.lastDownTime))

		fmt.Printf("\033[31m✖ Connection DOWN at %s (Down for: %s)\033[0m\n",
			timestamp, downtimeStr)
	} else {
		status = "UP"
		latency = parseLatency(outputStr)

		if connectionState.lastStatus == "DOWN" || connectionState.lastStatus == "UNKNOWN" {
			connectionState.lastUpTime = now
			connectionState.previousDowntime = now.Sub(connectionState.lastDownTime)
			connectionState.connectionChanges++
			message = fmt.Sprintf("Connection restored after %s downtime", formatDuration(connectionState.previousDowntime))
		} else {
			connectionState.currentUptime = now.Sub(connectionState.lastUpTime)
			message = "Connection stable"
		}

		uptimeStr = formatDuration(now.Sub(connectionState.lastUpTime))
		downtimeStr = formatDuration(connectionState.previousDowntime)

		fmt.Printf("\033[32m✓ Connection UP at %s (Up for: %s, Latency: %dms)\033[0m\n",
			timestamp, uptimeStr, latency)
	}

	connectionState.lastStatus = status
	connectionState.lastStatusTime = now

	message = strings.ReplaceAll(message, "\"", "\"\"")

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

func startHTTPServer() {
	http.HandleFunc("/api/connection-data", func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Read the CSV file
		file, err := os.Open(config.LogFile)
		if err != nil {
			http.Error(w, "Unable to read log file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Parse the CSV
		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1 // Allow variable number of fields

		// Read all records
		records, err := reader.ReadAll()
		if err != nil {
			http.Error(w, "Error parsing CSV data", http.StatusInternalServerError)
			return
		}

		// Skip header row and convert to JSON
		if len(records) <= 1 {
			// Return empty array if only header exists
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
			return
		}

		// Get headers from first row
		headers := records[0]

		// Convert records to map
		var result []map[string]string
		for _, record := range records[1:] {
			row := make(map[string]string)
			for i, value := range record {
				if i < len(headers) {
					row[headers[i]] = value
				}
			}
			result = append(result, row)
		}

		// Convert to JSON and send
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// Start HTTP server on port 8080
	fmt.Println("Starting HTTP server on :8080")
	go http.ListenAndServe(":8080", nil)
}
