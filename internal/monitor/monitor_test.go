//go:build leaktest

package monitor

import (
	"os"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// TestLeakMonitor verifies that monitor functions don't leak goroutines
func TestLeakMonitor(t *testing.T) {
	// Set up the goroutine leak detection
	defer goleak.VerifyNone(t)

	// Create a temporary log file
	tmpFile, err := os.CreateTemp("", "test_monitor_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create a test config
	config := DefaultConfig()
	config.CheckInterval = 1  // 1 second interval for faster testing
	config.PingTimeout = 1    // 1 second timeout
	config.LogFile = tmpFile.Name()
	
	// Create a state
	state := NewConnectionState()
	
	// Initialize logger
	logger, err := NewLogger(config.LogFile)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Perform a check
	result := Check(config, state)
	err = logger.LogResult(result)
	if err != nil {
		t.Fatalf("Failed to log result: %v", err)
	}
	
	// Wait a moment to ensure any goroutines have time to complete
	time.Sleep(500 * time.Millisecond)
	
	// goroutine leak detection happens in the defer goleak.VerifyNone(t)
}