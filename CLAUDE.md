# WatchDog Development Guide

## Build & Run Commands
- **Go Backend:** `make build-backend` and `make run-backend`
- **Frontend:** `cd app && npm run dev` or `make run-frontend`
- **Build All:** `make build`
- **Lint:** `make lint` (both backend and frontend)
- **Lint Backend Only:** `make lint-backend`
- **Lint Frontend Only:** `make lint-frontend`
- **Test:** `make test` or single test: `go test -v -run TestName ./internal/...`
- **Scan for Leaks:** `make scan-leaks` (checks for goroutine leaks)

## Code Style Guidelines
- **TypeScript:** Use React FC type for functional components, explicit interfaces for props/data
- **Go:** Comprehensive error handling with descriptive messages
- **Imports:** Group imports (std lib first, then 3rd party, then local)
- **Naming:** camelCase for JS/TS variables, PascalCase for components/types, snake_case for Go
- **Formatting:** Use consistent 2-space indentation in TS/JS, gofmt for Go
- **Error Handling:** Always check error returns in Go, use try/catch sparingly in TS
- **State Management:** Prefer React hooks over class components, minimize prop drilling
- **CSS:** Use module.css files for component-specific styling
- **Linting:** Use golangci-lint for Go, ESLint for TypeScript/JavaScript
- **Goroutines:** Always ensure goroutines can exit properly, use context for cancellation

## Project Structure
- **cmd/watchdog:** Main application entry point
- **internal/monitor:** Connection monitoring logic
- **internal/server:** HTTP API server
- **app/:** Next.js frontend application
- **scripts/:** Helper scripts

## Architecture
Frontend communicates with Go backend via HTTP API at port 8080

## Testing
- Regular tests: `go test ./...`
- Specific test: `go test -v -run TestName ./internal/...`
- Goroutine leak tests: `go test -tags=leaktest -v ./... -run TestLeak`