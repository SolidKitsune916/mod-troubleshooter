# Project Operations

## Frontend (React/Vite)

```bash
# Development
npm run dev

# Production build
npm run build
```

### Frontend Validation

```bash
npm run typecheck
npm test
npm run lint
```

## Backend (Go)

```bash
# Development (with hot reload)
make dev
# Or: air
# Or: go run ./cmd/server

# Build
make build
```

### Backend Validation

```bash
go test -v ./...
go vet ./...
golangci-lint run
```

## Full Stack

```bash
# Run both (in separate terminals)
# Terminal 1: npm run dev
# Terminal 2: make dev

# Test everything
npm test && go test -v ./...
```

## Project Structure

```
src/                     # React frontend
├── components/
├── features/
├── hooks/
├── services/           # API client for Go backend
├── store/
├── types/
└── utils/

cmd/                     # Go backend entry points
├── server/
│   └── main.go

internal/                # Go private packages
├── handlers/
├── models/
├── repository/
└── services/
```

## Rule Files Location

Standards are defined in `.cursor/rules/*.mdc` - always load relevant rules before implementing.

## Operational Learnings

<!-- Ralph adds learnings here as discoveries are made -->
