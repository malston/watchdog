# Watch Dog

A lightweight, cross-platform tool for monitoring your internet connection stability and visualizing connection drops.

![Connection Monitor Dashboard](https://api.placeholder.com/400/320)

## Features

- **Real-time monitoring** of internet connectivity using ping
- **Cross-platform support** (Windows, macOS, Linux)
- **Detailed logging** with timestamps of all connection changes
- **Uptime/downtime tracking** with duration of connection issues
- **Interactive dashboard** for visualizing connection quality
- **CSV export** for further analysis in spreadsheet software

## Components

This project consists of two main components:

1. **Backend Monitor** (Go): Performs regular connection checks and logs results
2. **Frontend Dashboard** (React): Visualizes the connection data in a user-friendly interface

## Installation

### Prerequisites

- [Go](https://golang.org/dl/) 1.16 or higher
- [Node.js](https://nodejs.org/) 14 or higher (for React frontend)

### Backend Setup

1. Clone this repository

   ```sh
   git clone https://github.com/malston/watchdog.git
   cd watchdog
   ```

2. Build the Go monitor

   ```sh
   go build -o watchdog monitor.go
   ```

### Frontend Setup

1. Navigate to the frontend directory

   ```sh
   cd frontend
   ```

2. Install dependencies

   ```sh
   npm install
   ```

3. Start the development server

   ```sh
   npm start
   ```

## Usage

### Running the Backend Monitor

Start the connection monitor with:

```sh
./watchdog
```

By default, it will:

- Ping Google DNS (8.8.8.8) every 30 seconds
- Save results to `connection_log.csv` in the current directory
- Show colored status indicators in the terminal

### Configuration

You can modify these settings in the `config` struct in `monitor.go`:

```go
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
```

### Understanding the Log Output

The CSV log file contains the following columns:

- `timestamp`: Date and time of the check
- `status`: Connection status (UP/DOWN)
- `latency`: Ping response time in milliseconds (-1 for DOWN)
- `uptime`: Duration the connection has been up
- `downtime`: Duration the connection has been down
- `total_changes`: Number of status changes (UP to DOWN or DOWN to UP)
- `message`: Description of the status

### Using the Dashboard

Open a web browser and navigate to:

```sh
http://localhost:3000
```

The dashboard will automatically display:

- Current connection status
- Historical latency graph
- Connection uptime/downtime statistics
- Detailed event timeline

## Deployment

For continuous monitoring, you can:

### Backend

- Run the Go monitor as a service using systemd (Linux), launchd (macOS), or Windows Services
- Set up the monitor to start automatically on boot

### Frontend

- Build the React app for production with `npm run build`
- Serve the static files using a lightweight server like Nginx or use GitHub Pages

## Troubleshooting

### Common Issues

1. **Permission denied for ping**
   - Run the application with sudo/administrator privileges
   - On Linux, use `setcap cap_net_raw=+ep ./watchdog`

2. **CSV file not updating**
   - Check write permissions for the directory
   - Use an absolute path for the log file

3. **Frontend not showing data**
   - Verify the backend is running and generating the CSV
   - Check the browser console for any errors

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- Thanks to all contributors
- Inspiration from various network monitoring tools
