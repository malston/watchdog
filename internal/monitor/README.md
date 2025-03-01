# Monitor Package

This package provides functionality for monitoring internet connection status.

## Components

- `Config`: Configuration settings for the monitor
- `ConnectionState`: Tracks the current network connection state
- `CheckResult`: Represents the result of a connection check
- `Logger`: Handles logging connection status to a CSV file

## Usage

```go
import "github.com/malston/watchdog/internal/monitor"

// Create default configuration
config := monitor.DefaultConfig()

// Create a connection state tracker
state := monitor.NewConnectionState()

// Initialize logger
logger, err := monitor.NewLogger(config.LogFile)
if err != nil {
    // Handle error
}

// Perform a connection check
result := monitor.Check(config, state)

// Log the result
err = logger.LogResult(result)
if err != nil {
    // Handle error
}
```