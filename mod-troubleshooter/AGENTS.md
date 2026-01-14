# Mod Troubleshooter - Operations Guide

## Project Overview

Full-stack application for analyzing Nexus Mods collections:
- **Frontend**: React + Vite + TypeScript (port 5173)
- **Backend**: Go HTTP server (port 8080)

## Build & Run

### Backend (Go)

```bash
cd backend

# Run development server
go run ./cmd/server

# Build binary
go build -o bin/server ./cmd/server

# Run built binary
./bin/server
```

### Frontend (React)

```bash
cd frontend

# Install dependencies
npm install

# Development server
npm run dev

# Production build
npm run build
```

### Full Stack Development

```bash
# Terminal 1: Backend
cd backend && go run ./cmd/server

# Terminal 2: Frontend
cd frontend && npm run dev
```

## Validation Commands

### Backend

```bash
cd backend

# Run tests
go test -v ./...

# Type check / vet
go vet ./...

# Format code
go fmt ./...

# Lint (if golangci-lint installed)
golangci-lint run
```

### Frontend

```bash
cd frontend

# Type check
npm run typecheck

# Run tests
npm test

# Lint
npm run lint

# All checks
npm run typecheck && npm test && npm run lint
```

## Environment Variables

Create `.env` in project root:

```bash
# Required
NEXUS_API_KEY=your-nexus-api-key

# Optional (defaults shown)
PORT=8080
DATA_DIR=./data
CACHE_TTL_HOURS=168
```

## Project Structure

```
mod-troubleshooter/
├── backend/
│   ├── cmd/server/main.go      # Entry point
│   ├── internal/
│   │   ├── config/             # Configuration
│   │   ├── handlers/           # HTTP handlers
│   │   ├── nexus/              # Nexus API client
│   │   ├── fomod/              # FOMOD parser
│   │   ├── archive/            # Archive extraction
│   │   └── conflict/           # Conflict detection
│   ├── go.mod
│   └── Makefile
├── frontend/
│   ├── src/
│   │   ├── components/         # Reusable UI
│   │   ├── features/           # Feature modules
│   │   ├── services/           # API clients
│   │   ├── hooks/              # Custom hooks
│   │   └── types/              # TypeScript types
│   ├── package.json
│   └── vite.config.ts
├── specs/                      # Feature specifications
├── IMPLEMENTATION_PLAN.md
└── AGENTS.md
```

## Rule Files

Use cursor rules when implementing:

| File Type | Rule |
|-----------|------|
| `*.go` | `.cursor/rules/2000-golang-backend.mdc` |
| `src/**/*.tsx` | `.cursor/rules/1001-react-components.mdc` |
| `src/hooks/**/*.ts` | `.cursor/rules/1002-react-hooks.mdc` |
| `src/services/**/*.ts` | `.cursor/rules/1009-services.mdc` |
| `*.tsx` (forms) | `.cursor/rules/1003-react-forms.mdc` |
| All UI | `.cursor/rules/1004-accessibility-wcag.mdc` |
| All UI | `.cursor/rules/1005-qol-ux.mdc` |

## API Patterns

### Backend Response Envelope

```go
type Response struct {
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Message string      `json:"message,omitempty"`
}
```

### Frontend API Client

```typescript
// Use TanStack Query for data fetching
const { data, isLoading, error } = useQuery({
    queryKey: ['collection', gameId, slug],
    queryFn: () => api.getCollection(gameId, slug),
});
```

## Common Patterns

### Go HTTP Handler

```go
func (h *Handler) GetCollection(w http.ResponseWriter, r *http.Request) {
    gameId := r.PathValue("gameId")
    slug := r.PathValue("slug")
    
    collection, err := h.service.GetCollection(r.Context(), gameId, slug)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    
    writeJSON(w, http.StatusOK, collection)
}
```

### React Feature Component

```tsx
export function FeatureView({ id }: Props) {
    const { data, isLoading, error } = useFeatureData(id);
    
    if (isLoading) return <Skeleton />;
    if (error) return <ErrorDisplay error={error} />;
    if (!data) return null;
    
    return (
        <div className="feature-view">
            {/* content */}
        </div>
    );
}
```

## Nexus API Notes

- GraphQL endpoint: `https://api.nexusmods.com/v2/graphql`
- REST endpoint: `https://api.nexusmods.com/v1/...`
- All requests need `apikey` header
- Premium required for download links
- Rate limit: ~2500 requests/day for Premium

## Codebase Patterns

<!-- Ralph adds learnings here -->

## Operational Learnings

<!-- Ralph adds learnings here -->
