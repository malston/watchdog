# WatchDog Development Guide

## Build & Run Commands
- **Go Backend:** `go build -o watchdog monitor.go` and `./watchdog`
- **Frontend:** `cd app && npm run dev` (uses turbopack)
- **Lint:** `cd app && npm run lint`
- **Build:** `cd app && npm run build`
- **Test:** For Go: `go test -v ./...` or single test: `go test -v -run TestName`

## Code Style Guidelines
- **TypeScript:** Use React FC type for functional components, explicit interfaces for props/data
- **Go:** Comprehensive error handling with descriptive messages
- **Imports:** Group imports (std lib first, then 3rd party, then local)
- **Naming:** camelCase for JS/TS variables, PascalCase for components/types, snake_case for Go
- **Formatting:** Use consistent 2-space indentation in TS/JS, gofmt for Go
- **Error Handling:** Always check error returns in Go, use try/catch sparingly in TS
- **State Management:** Prefer React hooks over class components, minimize prop drilling
- **CSS:** Use module.css files for component-specific styling

## Architecture
Frontend communicates with Go backend via HTTP API at port 8080