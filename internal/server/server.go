//Package server provides the HTTP API for the watchdog application
package server

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

// Server represents the HTTP server for the watchdog API
type Server struct {
	logFilePath string
	port        int
	httpServer  *http.Server
	wg          sync.WaitGroup
}

// NewServer creates a new API server instance
func NewServer(logFilePath string, port int) *Server {
	return &Server{
		logFilePath: logFilePath,
		port:        port,
	}
}

// Start initializes and starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/connection-data", s.handleConnectionData)

	addr := fmt.Sprintf(":%d", s.port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		fmt.Printf("Starting HTTP server on %s\n", addr)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		err := s.httpServer.Shutdown(ctx)
		s.wg.Wait()
		return err
	}
	return nil
}

// handleConnectionData processes requests for connection data
func (s *Server) handleConnectionData(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Read the CSV file
	file, err := os.Open(s.logFilePath)
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
		_, writeErr := w.Write([]byte("[]"))
		if writeErr != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
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
	jsonErr := json.NewEncoder(w).Encode(result)
	if jsonErr != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}
}