# Server Package

This package provides the HTTP API for the watchdog application.

## Components

- `Server`: Represents the HTTP server for the watchdog API

## Endpoints

- `GET /api/connection-data`: Returns connection monitoring data from the log file

## Usage

```go
import "github.com/malston/watchdog/internal/server"

// Create a new server instance
srv := server.NewServer("connection_log.csv", 8080)

// Start the server
err := srv.Start()
if err != nil {
    // Handle error
}
```