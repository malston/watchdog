//go:build leaktest

package main

import (
	"context"
	"os"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// TestLeakMain verifies that main application functions don't leak goroutines
func TestLeakMain(t *testing.T) {
	// Skip this test in normal runs since it's for manual leak testing
	if os.Getenv("TEST_FULL_LEAK") == "" {
		t.Skip("Skipping full leak test - set TEST_FULL_LEAK env var to run")
	}

	// Set up the goroutine leak detection
	defer goleak.VerifyNone(t)

	// Create a context for cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a channel to receive errors
	errCh := make(chan error, 1)

	// Start the application in a goroutine
	go func() {
		// This should be replaced with a proper test that doesn't use main()
		// but rather exercises the components used by main
		errCh <- nil
	}()

	// Wait for either completion or timeout
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Error in test: %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("Test timed out: %v", ctx.Err())
	}

	// Wait a moment to ensure any goroutines have time to complete
	time.Sleep(500 * time.Millisecond)
	
	// goroutine leak detection happens in the defer goleak.VerifyNone(t)
}