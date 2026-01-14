# Implementation Plan - Mod Troubleshooter

## Phase 1: Foundation

### Backend Setup

- [x] **US-001**: Initialize Go module and project structure
  - Create `backend/` directory structure
  - Initialize `go.mod` with module name
  - Create `cmd/server/main.go` entry point
  - Add Makefile with build/run/test targets
  - Priority: 1

- [x] **US-002**: Implement basic HTTP server with health endpoint
  - Configure server with timeouts
  - Add graceful shutdown
  - Implement `/api/health` endpoint
  - Add CORS middleware for frontend
  - Priority: 2

- [x] **US-003**: Create configuration management
  - Load from environment variables
  - Support `.env` file
  - Config struct for NEXUS_API_KEY, PORT, DATA_DIR
  - Priority: 3

- [x] **US-004**: Implement Nexus GraphQL client
  - Create nexus package in internal/
  - Implement GraphQL query execution
  - Add API key header handling
  - Add rate limiting/backoff
  - Priority: 4

- [x] **US-005**: Implement collection fetching endpoint
  - `GET /api/collections/{slug}` - fetches collection with latest revision mods
  - `GET /api/collections/{slug}/revisions` - fetches revision history
  - `GET /api/collections/{slug}/revisions/{revision}` - fetches specific revision mods
  - Implemented handlers package with standard JSON response envelope
  - Proper error mapping from Nexus client errors to HTTP status codes
  - Priority: 5

### Frontend Migration

- [x] **US-006**: Set up new frontend project structure
  - Set up Vite + React 19 + TypeScript with strict mode
  - Configure TanStack Query with QueryClientProvider
  - Set up Tailwind CSS v4 with gaming dark theme
  - Configure path aliases (@/, @components/, @features/, @hooks/, @services/, @utils/)
  - Created directory structure: components, features (collections, fomod, loadorder, conflicts), hooks, services, types, utils
  - Priority: 6

- [x] **US-007**: Create API service layer
  - Implemented fetch wrapper with type safety using Zod schemas
  - Created ApiError class for typed error handling
  - Created collectionService with fetchCollection, fetchCollectionRevisions, fetchCollectionRevisionMods
  - Created useCollections hooks for TanStack Query integration
  - Priority: 7

- [ ] **US-008**: Migrate collection browser to use Go backend
  - Update data fetching to use new API
  - Remove direct Nexus API calls from frontend
  - Maintain existing UI/UX
  - Priority: 8

- [ ] **US-009**: Add settings page for API key
  - Settings form component
  - Store API key in backend
  - Validate key on save
  - Priority: 9

## Phase 2: FOMOD Visualizer

### Backend - Archive Processing

- [ ] **US-010**: Implement Nexus download link fetching
  - `GET /api/mods/:gameId/:modId/files/:fileId/download`
  - Handle Premium-only restriction
  - Return download URLs
  - Priority: 10

- [ ] **US-011**: Implement archive downloader
  - Download to temp directory
  - Support large files (streaming)
  - Track download progress
  - Priority: 11

- [ ] **US-012**: Implement archive extractor
  - Support .zip, .7z, .rar formats
  - Extract specific paths (fomod/ directory)
  - Use github.com/mholt/archiver
  - Priority: 12

- [ ] **US-013**: Implement FOMOD XML parser
  - Parse info.xml for metadata
  - Parse ModuleConfig.xml for install steps
  - Handle all FOMOD elements (steps, groups, plugins, conditions)
  - Return structured JSON
  - Priority: 13

- [ ] **US-014**: Create FOMOD analysis endpoint
  - `POST /api/fomod/analyze`
  - Orchestrate download → extract → parse
  - Cache results in SQLite
  - Priority: 14

### Frontend - FOMOD Visualization

- [ ] **US-015**: Create FomodViewer container component
  - Layout with header, steps, summary panels
  - State management for selections
  - API integration
  - Priority: 15

- [ ] **US-016**: Implement FomodStepView component
  - Render option groups
  - Handle different group types (SelectOne, SelectAny, etc.)
  - Option cards with image, description, type badge
  - Priority: 16

- [ ] **US-017**: Implement selection logic
  - Track selections per step
  - Enforce group type constraints
  - Set/evaluate condition flags
  - Priority: 17

- [ ] **US-018**: Implement conditional step visibility
  - Evaluate dependency conditions
  - Show/hide steps based on flags
  - Update when selections change
  - Priority: 18

- [ ] **US-019**: Implement file preview panel
  - Show files that will be installed
  - Update based on current selections
  - Tree view of destination paths
  - Priority: 19

- [ ] **US-020**: Add tree view mode
  - Full FOMOD structure as collapsible tree
  - Alternative to wizard mode
  - Priority: 20

## Phase 3: Load Order

### Backend

- [ ] **US-021**: Implement plugin header parser
  - Read TES4 record from .esp/.esm/.esl
  - Extract master dependencies
  - Extract plugin flags
  - Priority: 21

- [ ] **US-022**: Create load order analysis endpoint
  - `GET /api/collections/:gameId/:slug/loadorder`
  - Parse plugins from collection mods
  - Build dependency graph
  - Detect issues (missing masters, wrong order)
  - Priority: 22

### Frontend

- [ ] **US-023**: Create LoadOrderView container
  - Layout with list and details panels
  - Stats header (slot usage, warnings)
  - Priority: 23

- [ ] **US-024**: Implement LoadOrderList component
  - Display plugins with type badges
  - Show master dependencies inline
  - Highlight warnings
  - Priority: 24

- [ ] **US-025**: Implement warning detection UI
  - Missing master indicators
  - Load order issue indicators
  - Expandable warning details
  - Priority: 25

- [ ] **US-026**: Add dependency graph view
  - React Flow or D3 visualization
  - Interactive node selection
  - Priority: 26

## Phase 4: Conflict Detection

### Backend

- [ ] **US-027**: Implement file manifest extraction
  - Extract file list from archives
  - Normalize paths
  - Calculate hashes for dedup
  - Priority: 27

- [ ] **US-028**: Implement conflict detection algorithm
  - Build file → mod map
  - Identify multi-source files
  - Classify by file type
  - Priority: 28

- [ ] **US-029**: Implement conflict severity scoring
  - Score based on file type
  - Consider known incompatibilities
  - Priority: 29

- [ ] **US-030**: Create conflict analysis endpoint
  - `POST /api/conflicts/analyze`
  - Analyze subset or full collection
  - Return detailed conflict report
  - Priority: 30

### Frontend

- [ ] **US-031**: Create ConflictView container
  - Layout with filters, list, details
  - Summary header
  - Priority: 31

- [ ] **US-032**: Implement ConflictList component
  - Display conflicts with severity badges
  - Show conflicting mods and winner
  - Priority: 32

- [ ] **US-033**: Implement conflict filters
  - By severity
  - By file type
  - By mod
  - Search by path
  - Priority: 33

- [ ] **US-034**: Implement ConflictDetails panel
  - Full conflict information
  - Resolution suggestions
  - Priority: 34

## Completed

- [x] **US-001**: Initialize Go module and project structure
  - Created `backend/` directory structure with internal packages (config, handlers, nexus, fomod, archive, conflict, cache, models) and pkg/response
  - Initialized `go.mod` with module `github.com/mod-troubleshooter/backend`
  - Created `cmd/server/main.go` entry point
  - Added Makefile with build/run/dev/test/lint/clean targets
  - Added `.air.toml` for hot reload development

- [x] **US-002**: Implement basic HTTP server with health endpoint
  - Server configured with ReadTimeout, WriteTimeout, IdleTimeout
  - Graceful shutdown on SIGINT/SIGTERM
  - `/api/health` endpoint returning `{"status":"ok"}`
  - CORS middleware via rs/cors for localhost:5173 and localhost:3000

- [x] **US-003**: Create configuration management
  - Created `internal/config/config.go` with Config struct
  - Supports environment variables: NEXUS_API_KEY, PORT, DATA_DIR, CACHE_TTL_HOURS, ENVIRONMENT, CORS_ORIGINS
  - Automatic `.env` file loading from current or parent directories
  - Validation (NEXUS_API_KEY required only in production)
  - Updated `main.go` to use config package
  - Added unit tests in `config_test.go`
  - Updated `.env.example` with all supported variables

- [x] **US-004**: Implement Nexus GraphQL client
  - Created `internal/nexus/` package with types.go, queries.go, client.go
  - Implemented GraphQL query execution with proper JSON request/response handling
  - API key header handling via `apikey` header on all requests
  - Rate limiting with configurable min request delay, adjusts based on remaining quota
  - Exponential backoff with configurable initial/max backoff and max retries
  - Helper methods: GetCollection, GetCollectionRevisions, GetCollectionRevisionMods
  - Comprehensive unit tests in client_test.go (100% pass)

## Discovered Issues

<!-- Document bugs and issues found during implementation -->

## Future Considerations

- LOOT integration for load order suggestions
- Tauri wrapper for local Vortex/MO2 access
- Plugin record-level conflict detection
- Automated conflict resolution suggestions
- Mod comparison tool
- Collection export/import
