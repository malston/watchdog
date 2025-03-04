package monitor

import (
	"fmt"
	"os"
	"time"
)

// Logger handles logging connection status to a CSV file
type Logger struct {
	filePath string
}

// NewLogger creates a new logger that writes to the specified file path
func NewLogger(filePath string) (*Logger, error) {
	// Initialize log file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("error creating log file: %v", err)
		}
		defer file.Close()

		_, err = file.WriteString("timestamp,status,latency,uptime,downtime,total_changes,message\n")
		if err != nil {
			return nil, fmt.Errorf("error writing header to log file: %v", err)
		}
		fmt.Printf("Created log file: %s\n", filePath)
	}

	return &Logger{
		filePath: filePath,
	}, nil
}

// LogResult writes a check result to the log file
func (l *Logger) LogResult(result CheckResult) error {
	entry := fmt.Sprintf("%s,%s,%d,%s,%s,%d,\"%s\"\n",
		result.Timestamp.Format(time.RFC3339),
		result.Status,
		result.Latency,
		result.UptimeStr,
		result.DowntimeStr,
		result.Changes,
		result.Message,
	)

	file, err := os.OpenFile(l.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(entry); err != nil {
		return fmt.Errorf("error writing to log file: %v", err)
	}

	return nil
}
