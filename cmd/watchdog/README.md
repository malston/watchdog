# WatchDog Command

This is the main entry point for the WatchDog application.

## Usage

```
watchdog [flags]
```

### Flags

```
  -api-port int
        Port for the HTTP API server (default 8080)
  -check-interval int
        Interval between checks in seconds (default 30)
  -log-file string
        Log file path (default "connection_log.csv")
  -ping-count int
        Number of ping packets to send (default 3)
  -ping-target string
        Target to ping (default "8.8.8.8")
  -ping-timeout int
        Ping timeout in seconds (default 5)
```

## Building

From the project root:

```
make build-backend
```

## Running

From the project root:

```
make run-backend
```

Or directly:

```
./watchdog
```