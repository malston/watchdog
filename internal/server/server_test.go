//go:build leaktest

package server

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// TestLeakServer verifies that server startup and shutdown don't leak goroutines
func TestLeakServer(t *testing.T) {
	// Set up the goroutine leak detection
	defer goleak.VerifyNone(t)

	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "test_server_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	fileName := tmpFile.Name()
	defer os.Remove(fileName)
	
	// Write header to the file
	_, err = tmpFile.WriteString("timestamp,status,latency,uptime,downtime,total_changes,message\n")
	if err != nil {
		t.Fatalf("Failed to write header: %v", err)
	}
	tmpFile.Close()

	// Set up a test server
	port := 8888 // Use a different port for testing
	server := NewServer(fileName, port)

	// Start the server
	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Make a test request
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8888/api/connection-data", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 2 * time.Second}
	_, _ = client.Do(req) // Ignore errors, we're just testing for leaks
	
	// Properly stop the server
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	err = server.Stop(stopCtx)
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
	
	// Wait a moment to ensure any goroutines have time to complete
	time.Sleep(500 * time.Millisecond)
	
	// goroutine leak detection happens in the defer goleak.VerifyNone(t)
}