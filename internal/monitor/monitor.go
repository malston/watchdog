// Package monitor provides functionality for monitoring internet connection status
package monitor

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Config holds configuration settings for the connection monitor
type Config struct {
	PingTarget    string
	CheckInterval int
	LogFile       string
	PingCount     int
	PingTimeout   int
}

// DefaultConfig returns a Config with reasonable default values
func DefaultConfig() Config {
	return Config{
		PingTarget:    "8.8.8.8",
		CheckInterval: 30,
		LogFile:       "connection_log.csv",
		PingCount:     3,
		PingTimeout:   5,
	}
}

// Validate checks if the configuration values are valid
func (c Config) Validate() error {
	if c.PingTarget == "" {
		return fmt.Errorf("ping-target cannot be empty")
	}
	if c.CheckInterval <= 0 {
		return fmt.Errorf("check-interval must be greater than 0")
	}
	if c.LogFile == "" {
		return fmt.Errorf("log-file cannot be empty")
	}
	if c.PingCount <= 0 {
		return fmt.Errorf("ping-count must be greater than 0")
	}
	if c.PingTimeout <= 0 {
		return fmt.Errorf("ping-timeout must be greater than 0")
	}
	return nil
}

// ConnectionState tracks the current network connection state
type ConnectionState struct {
	LastStatus        string
	LastStatusTime    time.Time
	LastDownTime      time.Time
	LastUpTime        time.Time
	CurrentUptime     time.Duration
	PreviousDowntime  time.Duration
	PreviousUptime    time.Duration
	ConnectionChanges int
}

// NewConnectionState creates a new ConnectionState with initial values
func NewConnectionState() *ConnectionState {
	return &ConnectionState{
		LastStatus:     "UNKNOWN",
		LastStatusTime: time.Now(),
	}
}

// CheckResult represents the result of a connection check
type CheckResult struct {
	Timestamp   time.Time
	Status      string
	Latency     int64
	UptimeStr   string
	DowntimeStr string
	Message     string
	Changes     int
}

// Check performs a single connection check and returns the result
func Check(config Config, state *ConnectionState) CheckResult {
	now := time.Now()
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

		if state.LastStatus == "UP" || state.LastStatus == "UNKNOWN" {
			state.LastDownTime = now
			state.PreviousUptime = now.Sub(state.LastUpTime)
			state.ConnectionChanges++
			message = fmt.Sprintf("Connection lost after %s uptime. %s", FormatDuration(state.PreviousUptime), message)
		} else {
			state.CurrentUptime = 0
			message = "Connection still down. " + message
		}

		uptimeStr = FormatDuration(state.PreviousUptime)
		downtimeStr = FormatDuration(now.Sub(state.LastDownTime))

		fmt.Printf("\033[31m✖ Connection DOWN at %s (Down for: %s)\033[0m\n",
			now.Format(time.RFC3339), downtimeStr)
	} else {
		status = "UP"
		latency = parseLatency(outputStr)

		if state.LastStatus == "DOWN" || state.LastStatus == "UNKNOWN" {
			state.LastUpTime = now
			state.PreviousDowntime = now.Sub(state.LastDownTime)
			state.ConnectionChanges++
			message = fmt.Sprintf("Connection restored after %s downtime", FormatDuration(state.PreviousDowntime))
		} else {
			state.CurrentUptime = now.Sub(state.LastUpTime)
			message = "Connection stable"
		}

		uptimeStr = FormatDuration(now.Sub(state.LastUpTime))
		downtimeStr = FormatDuration(state.PreviousDowntime)

		fmt.Printf("\033[32m✓ Connection UP at %s (Up for: %s, Latency: %dms)\033[0m\n",
			now.Format(time.RFC3339), uptimeStr, latency)
	}

	state.LastStatus = status
	state.LastStatusTime = now

	message = strings.ReplaceAll(message, "\"", "\"\"")

	return CheckResult{
		Timestamp:   now,
		Status:      status,
		Latency:     latency,
		UptimeStr:   uptimeStr,
		DowntimeStr: downtimeStr,
		Message:     message,
		Changes:     state.ConnectionChanges,
	}
}

// parseLatency extracts the average latency from ping output
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

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
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
