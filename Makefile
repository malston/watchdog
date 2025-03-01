.PHONY: build build-backend build-frontend clean run run-backend run-frontend all dev help

# Binary names
BINARY_NAME=watchdog

# Directories
BACKEND_DIR=.
FRONTEND_DIR=app
BINARY_OUT=$(BACKEND_DIR)/$(BINARY_NAME)

# Default target
all: build

help:
	@echo "WatchDog - Internet Connection Monitoring Tool"
	@echo ""
	@echo "Usage:"
	@echo "  make build         Build both backend and frontend"
	@echo "  make build-backend Build Go backend only"
	@echo "  make build-frontend Build React frontend only"
	@echo "  make run           Run both backend and frontend (in separate terminals)"
	@echo "  make run-backend   Run Go backend only"
	@echo "  make run-frontend  Run React frontend only"
	@echo "  make dev           Run both in development mode"
	@echo "  make clean         Remove build artifacts"
	@echo "  make help          Show this help message"
	@echo ""

# Build both backend and frontend
build: build-backend build-frontend

# Build Go backend
build-backend:
	@echo "Building backend..."
	@cd $(BACKEND_DIR) && go build -o $(BINARY_NAME) ./cmd/watchdog
	@echo "Backend built!"

# Build React frontend for production
build-frontend:
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && npm run build
	@echo "Frontend built!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@cd $(BACKEND_DIR) && rm -f $(BINARY_NAME)
	@cd $(FRONTEND_DIR) && rm -rf .next
	@echo "Cleaned!"

# Run both backend and frontend (This needs multiple terminals)
run: run-backend run-frontend

# Run Go backend
run-backend:
	@echo "Starting backend server..."
	@cd $(BACKEND_DIR) && ./$(BINARY_NAME)

# Run React frontend
run-frontend:
	@echo "Starting frontend server..."
	@cd $(FRONTEND_DIR) && npm run start

# Development mode
dev:
	@echo "Starting backend and frontend in development mode..."
	@echo "Note: This will start the backend in the current terminal."
	@echo "      Open a new terminal and run 'make run-frontend' for the frontend."
	@cd $(BACKEND_DIR) && ./$(BINARY_NAME)