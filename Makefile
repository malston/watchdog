.PHONY: build build-backend build-frontend clean run run-backend run-frontend all dev help lint lint-backend lint-frontend test test-backend

# Binary names
BINARY_NAME=watchdog

# Directories
BACKEND_DIR=.
FRONTEND_DIR=app
BINARY_OUT=$(BACKEND_DIR)/$(BINARY_NAME)

# Go package paths
GO_PACKAGES=./cmd/... ./internal/...
GO_LINT_FLAGS=--timeout=3m

# Default target
all: build

help:
	@echo "WatchDog - Internet Connection Monitoring Tool"
	@echo ""
	@echo "Usage:"
	@echo "  make build           Build both backend and frontend"
	@echo "  make build-backend   Build Go backend only"
	@echo "  make build-frontend  Build React frontend only"
	@echo "  make run             Run both backend and frontend (in separate terminals)"
	@echo "  make run-backend     Run Go backend only"
	@echo "  make run-frontend    Run React frontend only"
	@echo "  make dev             Run both in development mode"
	@echo "  make lint            Lint both backend and frontend code"
	@echo "  make lint-backend    Lint Go backend code"
	@echo "  make lint-frontend   Lint Next.js frontend code"
	@echo "  make test            Run tests for both backend and frontend"
	@echo "  make test-backend    Run Go tests"
	@echo "  make clean           Remove build artifacts"
	@echo "  make help            Show this help message"
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

# Lint both backend and frontend code
lint: lint-backend lint-frontend

# Lint Go backend code
lint-backend:
	@echo "Linting Go code..."
	@cd $(BACKEND_DIR) && golangci-lint run $(GO_LINT_FLAGS) $(GO_PACKAGES)
	@echo "Go code linting completed!"

# Lint Next.js frontend code
lint-frontend:
	@echo "Linting frontend code..."
	@cd $(FRONTEND_DIR) && npm run lint
	@echo "Frontend code linting completed!"

# Run tests for both backend and frontend
test: test-backend

# Run Go tests
test-backend:
	@echo "Running Go tests..."
	@cd $(BACKEND_DIR) && go test -v $(GO_PACKAGES)
	@echo "Go tests completed!"