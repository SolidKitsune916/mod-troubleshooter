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

- [x] **US-008**: Migrate collection browser to use Go backend
  - Update data fetching to use new API
  - Remove direct Nexus API calls from frontend
  - Maintain existing UI/UX
  - Priority: 8

- [x] **US-009**: Add settings page for API key
  - Settings form component
  - Store API key in backend
  - Validate key on save
  - Priority: 9

## Phase 2: FOMOD Visualizer

### Backend - Archive Processing

- [x] **US-010**: Implement Nexus download link fetching
  - `GET /api/games/{game}/mods/{modId}/files/{fileId}/download`
  - Handle Premium-only restriction (returns 403 with clear message)
  - Return download URLs array
  - Priority: 10

- [x] **US-011**: Implement archive downloader
  - Download to temp directory
  - Support large files (streaming)
  - Track download progress
  - Priority: 11

- [x] **US-012**: Implement archive extractor
  - Support .zip, .7z, .rar formats
  - Extract specific paths (fomod/ directory)
  - Use github.com/mholt/archiver
  - Priority: 12

- [x] **US-013**: Implement FOMOD XML parser
  - Parse info.xml for metadata
  - Parse ModuleConfig.xml for install steps
  - Handle all FOMOD elements (steps, groups, plugins, conditions)
  - Return structured JSON
  - Priority: 13

- [x] **US-014**: Create FOMOD analysis endpoint
  - `POST /api/fomod/analyze`
  - Orchestrate download → extract → parse
  - Cache results in SQLite
  - Priority: 14

### Frontend - FOMOD Visualization

- [x] **US-015**: Create FomodViewer container component
  - Layout with header, steps, summary panels
  - State management for selections
  - API integration
  - Priority: 15

- [x] **US-016**: Implement FomodStepView component
  - Render option groups
  - Handle different group types (SelectOne, SelectAny, etc.)
  - Option cards with image, description, type badge
  - Priority: 16

- [x] **US-017**: Implement selection logic
  - Track selections per step
  - Enforce group type constraints
  - Set/evaluate condition flags
  - Priority: 17

- [x] **US-018**: Implement conditional step visibility
  - Evaluate dependency conditions
  - Show/hide steps based on flags
  - Update when selections change
  - Priority: 18

- [x] **US-019**: Implement file preview panel
  - Show files that will be installed
  - Update based on current selections
  - Tree view of destination paths
  - Priority: 19

- [x] **US-020**: Add tree view mode
  - Full FOMOD structure as collapsible tree
  - Alternative to wizard mode
  - Priority: 20

## Phase 3: Load Order

### Backend

- [x] **US-021**: Implement plugin header parser
  - Read TES4 record from .esp/.esm/.esl
  - Extract master dependencies
  - Extract plugin flags
  - Priority: 21

- [x] **US-022**: Create load order analysis endpoint
  - `GET /api/collections/:gameId/:slug/loadorder`
  - Parse plugins from collection mods
  - Build dependency graph
  - Detect issues (missing masters, wrong order)
  - Priority: 22

### Frontend

- [x] **US-023**: Create LoadOrderView container
  - Layout with list and details panels
  - Stats header (slot usage, warnings)
  - Priority: 23

- [x] **US-024**: Implement LoadOrderList component
  - Display plugins with type badges
  - Show master dependencies inline
  - Highlight warnings
  - Priority: 24

- [x] **US-025**: Implement warning detection UI
  - Missing master indicators
  - Load order issue indicators
  - Expandable warning details
  - Priority: 25

- [x] **US-026**: Add dependency graph view
  - React Flow or D3 visualization
  - Interactive node selection
  - Priority: 26

## Phase 4: Conflict Detection

### Backend

- [x] **US-027**: Implement file manifest extraction
  - Extract file list from archives
  - Normalize paths
  - Calculate hashes for dedup
  - Priority: 27

- [x] **US-028**: Implement conflict detection algorithm
  - Build file → mod map
  - Identify multi-source files
  - Classify by file type
  - Priority: 28

- [x] **US-029**: Implement conflict severity scoring
  - Score based on file type
  - Consider known incompatibilities
  - Priority: 29

- [x] **US-030**: Create conflict analysis endpoint
  - `POST /api/conflicts/analyze`
  - Analyze subset or full collection
  - Return detailed conflict report
  - Priority: 30

### Frontend

- [x] **US-031**: Create ConflictView container
  - Layout with filters, list, details
  - Summary header
  - Priority: 31

- [x] **US-032**: Implement ConflictList component
  - Display conflicts with severity badges
  - Show conflicting mods and winner
  - Priority: 32

- [x] **US-033**: Implement conflict filters
  - By severity
  - By file type
  - By mod
  - Search by path
  - Priority: 33

- [x] **US-034**: Implement ConflictDetails panel
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

- [x] **US-008**: Migrate collection browser to use Go backend
  - Created CollectionSearch component with URL/slug parsing
  - Created CollectionHeader component for collection metadata display
  - Created ModCard component for individual mod file display
  - Created ModList component with Required/Optional mod grouping
  - Created CollectionBrowser container with loading skeletons, error states, and screen reader announcements
  - Integrated with existing API service layer and TanStack Query hooks
  - WCAG 2.2 AA compliance: semantic HTML, focus management, aria-live regions

- [x] **US-009**: Add settings page for API key
  - Created `internal/handlers/settings.go` with SettingsStore for runtime API key management
  - Implemented `GET /api/settings`, `POST /api/settings`, `POST /api/settings/validate` endpoints
  - Added `ValidateAPIKey` method to Nexus client for key validation
  - Refactored collection handler to use dynamic client getter pattern for runtime key updates
  - Created frontend `SettingsPage` component with WCAG 2.2 AA compliant form
  - Added navigation between Collections and Settings pages
  - Features: key validation before save, masked key display, show/hide toggle, clear key option
  - Fixed API response envelope schema mismatch between frontend and backend

- [x] **US-010**: Implement Nexus download link fetching
  - Added `DownloadLink` and `DownloadLinksResponse` types to `internal/nexus/types.go`
  - Added `RESTAPIBase` constant for Nexus v1 REST API
  - Added `ErrPremiumOnly` and `ErrForbidden` errors to nexus client
  - Implemented `GetModFileDownloadLinks` method in nexus client using REST API endpoint
  - Created `internal/handlers/download.go` with `DownloadHandler` and `GetModFileDownloadLinks` handler
  - Registered `GET /api/games/{game}/mods/{modId}/files/{fileId}/download` endpoint
  - Returns 403 with clear message for non-Premium accounts

- [x] **US-011**: Implement archive downloader
  - Created `internal/archive/downloader.go` with `Downloader` type
  - Streaming download support using `io.Copy` for memory efficiency with large files
  - `ProgressCallback` function type for tracking download progress (downloaded bytes, total size)
  - Temp directory management with `Cleanup()` and `CleanupPath()` methods
  - `DownloaderConfig` for configurable temp dir, HTTP client, max file size, user agent
  - File size limit enforcement via Content-Length header and streaming limit
  - Comprehensive unit tests covering success, errors, context cancellation, progress tracking
  - All tests passing, go vet clean

- [x] **US-015, US-016, US-017**: Create FomodViewer container and step components
  - Added FOMOD Zod schemas and TypeScript types to frontend/src/types/api.ts
  - Created fomodService.ts with `analyzeFomod` API function
  - Created useFomod.ts hook with TanStack Query integration (24hr cache)
  - Implemented FomodViewer container with FomodHeader, FomodStepNavigator, FomodStepView, FomodSummary
  - Support all FOMOD group types (SelectExactlyOne, SelectAtMostOne, SelectAtLeastOne, SelectAny, SelectAll)
  - Support all plugin types with visual badges (Required, Recommended, Optional, NotUsable, CouldBeUsable)
  - Selection logic handles radio vs checkbox behavior based on group type
  - Loading skeleton and error handling with retry capability
  - Full accessibility: ARIA labels, keyboard navigation, screen reader support
  - Typecheck and lint passing

- [x] **US-018**: Implement conditional step visibility
  - Added `evaluateDependency` function to evaluate recursive dependency conditions against flag state
  - Added `collectFlags` function to collect condition flags from selected plugins
  - Step navigator now shows visibility status with visual indication for hidden steps
  - Auto-navigation to next/previous visible step when current step becomes hidden
  - Screen reader announcements include visible step count
  - Handles flag dependencies, file dependencies (defaults), and composite And/Or conditions
  - Typecheck and lint passing

- [x] **US-023**: Create LoadOrderView container
  - Added LoadOrder Zod schemas and TypeScript types (PluginInfo, Issue, Stats, AnalysisResult)
  - Created loadorderService.ts with analyzeCollectionLoadOrder API function
  - Created useLoadOrder.ts hook with TanStack Query (24hr cache)
  - Implemented LoadOrderView container with LoadOrderHeader, LoadOrderList, LoadOrderDetails, WarningPanel
  - Stats header displays total plugins, ESM/ESP/ESL counts, error and warning counts
  - Plugin list with index, type badges, master count, and issue indicators
  - Details panel shows plugin info, flags, masters, and issues when plugin selected
  - Warning panel shows all issues with click-to-select functionality
  - View mode tabs in CollectionBrowser to switch between Mod Files and Load Order views
  - Loading skeleton, error handling, and empty state
  - Full accessibility: ARIA labels, keyboard navigation, screen reader announcements
  - Typecheck and lint passing

- [x] **US-024**: Implement LoadOrderList component
  - LoadOrderList component displays plugins with index numbers, type badges, master count
  - PluginRow component with selectable rows and issue indicators
  - Highlights warnings with colored badge showing issue count
  - Maximum 600px scrollable list for long plugin lists
  - Keyboard accessible with focus-visible outlines
  - Implemented as part of US-023 within LoadOrderView.tsx

- [x] **US-025**: Implement warning detection UI
  - WarningPanel component shows all load order issues
  - Issues displayed with severity badges (error/warning)
  - Click-to-select functionality navigates to associated plugin
  - Issue type labels (Missing Master, Wrong Order, Duplicate Plugin)
  - Error/warning counts in panel header
  - "No Issues Found" success state when load order is healthy
  - Implemented as part of US-023 within LoadOrderView.tsx

## Phase 5: Enhanced Features (High Priority)

### Backend Improvements

- [x] **US-035**: Implement rate limiting with exponential backoff
  - Track Nexus API quota via response headers
  - Automatic backoff when quota low
  - User-friendly quota display in UI (QuotaIndicator component)
  - Added GET /api/quota endpoint
  - Priority: 35

- [x] **US-036**: Add games endpoint for dynamic game support
  - `GET /api/games` - list supported games with IDs, labels, and Nexus domain names
  - Frontend service (gamesService.ts) and hook (useGames.ts) for consuming the endpoint
  - Backend tests for games handler
  - Priority: 36

### FOMOD Enhancements

- [ ] **US-037**: Implement FOMOD comparison mode
  - Compare two FOMOD selections side-by-side
  - Highlight differences in selections
  - Priority: 37

- [x] **US-038**: Add FOMOD export/import
  - Export selections to JSON file with version, metadata, and selections
  - Import selections from JSON file with validation and mod mismatch warning
  - SelectionsToolbar component with Export/Import buttons
  - Priority: 38

- [ ] **US-039**: Add FOMOD dependency graph visualization
  - Show option dependencies as graph
  - Highlight conditional relationships
  - Priority: 39

- [ ] **US-040**: Add FOMOD search functionality
  - Search options across all steps
  - Filter by option type
  - Priority: 40

### Load Order Enhancements

- [x] **US-043**: Export load order
  - Export to plugins.txt format (with asterisk prefix for enabled plugins)
  - Export to loadorder.txt format (plain filenames)
  - ExportToolbar component with download buttons
  - Filenames include collection name and timestamp
  - Priority: 43

- [ ] **US-044**: Load order comparison mode
  - Compare two load orders side-by-side
  - Highlight differences
  - Priority: 44

- [x] **US-046**: Slot limit warning
  - SlotLimitWarning component with progress bar visualization
  - Warning at 90% threshold (229+ plugins), critical at 98% (249+ plugins)
  - Shows used vs. remaining slots with actionable advice
  - Note that ESL plugins don't count toward limit
  - Priority: 46

### Conflict Enhancements

- [ ] **US-048**: Export conflict report
  - Export to CSV/JSON format
  - Include all conflict details
  - Priority: 48

- [ ] **US-049**: Conflict graph visualization
  - Show mod relationships as graph
  - Highlight problematic dependencies
  - Priority: 49

### UI/UX Improvements

- [ ] **US-055**: Keyboard shortcuts
  - Navigation shortcuts
  - Action shortcuts with help overlay
  - Priority: 55

- [ ] **US-056**: Loading skeletons everywhere
  - Consistent skeleton UI across all features
  - Priority: 56

### Accessibility

- [ ] **US-059**: Reduced motion support
  - Respect prefers-reduced-motion
  - Disable animations when preferred
  - Priority: 59

- [ ] **US-060**: Full keyboard navigation audit
  - Ensure all interactive elements focusable
  - Proper focus order
  - Priority: 60

## Discovered Issues

<!-- Document bugs and issues found during implementation -->

## Future Considerations

- LOOT integration for load order suggestions
- Tauri wrapper for local Vortex/MO2 access
- Plugin record-level conflict detection
- Automated conflict resolution suggestions
- Mod comparison tool
- Collection export/import
