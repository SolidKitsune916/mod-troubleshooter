# Implementation Plan - Mod Troubleshooter

## Phase 1: Foundation

### Backend Setup

- [ ] **US-001**: Initialize Go module and project structure
  - Create `backend/` directory structure
  - Initialize `go.mod` with module name
  - Create `cmd/server/main.go` entry point
  - Add Makefile with build/run/test targets
  - Priority: 1

- [ ] **US-002**: Implement basic HTTP server with health endpoint
  - Configure server with timeouts
  - Add graceful shutdown
  - Implement `/api/health` endpoint
  - Add CORS middleware for frontend
  - Priority: 2

- [ ] **US-003**: Create configuration management
  - Load from environment variables
  - Support `.env` file
  - Config struct for NEXUS_API_KEY, PORT, DATA_DIR
  - Priority: 3

- [ ] **US-004**: Implement Nexus GraphQL client
  - Create nexus package in internal/
  - Implement GraphQL query execution
  - Add API key header handling
  - Add rate limiting/backoff
  - Priority: 4

- [ ] **US-005**: Implement collection fetching endpoint
  - `GET /api/collections/:gameId/:slug`
  - Parse GraphQL response
  - Transform to frontend-friendly format
  - Return mod list with metadata
  - Priority: 5

### Frontend Migration

- [ ] **US-006**: Set up new frontend project structure
  - Copy relevant code from viewer-app
  - Set up Vite + React + TypeScript
  - Configure TanStack Query
  - Set up path aliases
  - Priority: 6

- [ ] **US-007**: Create API service layer
  - Axios or fetch wrapper
  - Type-safe API client
  - Error handling
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
- [x] **US-002**: Implement basic HTTP server with health endpoint
- [x] **US-003**: Create configuration management
- [x] **US-004**: Implement Nexus GraphQL client
- [x] **US-005**: Implement collection fetching endpoint
- [x] **US-006**: Set up new frontend project structure
- [x] **US-007**: Create API service layer
- [x] **US-008**: Migrate collection browser to use Go backend
- [x] **US-009**: Add settings page for API key
- [x] **US-010**: Implement Nexus download link fetching
- [x] **US-011**: Implement archive downloader
- [x] **US-012**: Implement archive extractor
- [x] **US-015**: Create FomodViewer container component
- [x] **US-016**: Implement FomodStepView component
- [x] **US-017**: Implement selection logic
- [x] **US-018**: Implement conditional step visibility
- [x] **US-019**: Implement file preview panel
  - Tree view of files to be installed
  - Shows required, selected, and conditional files
  - Collapsible folder hierarchy
  - File counts by category
  - Support for .zip, .7z, .rar via mholt/archiver/v4
  - Path-specific extraction (fomod/ directory)
  - File size limits and zip slip protection
  - ListFiles and HasFomod helper methods
- [x] **US-013**: Implement FOMOD XML parser
  - Parse info.xml for mod metadata (name, author, version, description, website)
  - Parse ModuleConfig.xml with full FOMOD schema support
  - Handle all FOMOD elements: installSteps, optionalFileGroups, groups, plugins
  - Support all group types: SelectExactlyOne, SelectAtMostOne, SelectAtLeastOne, SelectAny, SelectAll
  - Support all plugin types: Required, Optional, Recommended, NotUsable, CouldBeUsable
  - Parse condition flags and dependency-based type descriptors
  - Parse composite dependencies (And/Or with file, flag, game, fomm dependencies)
  - Parse conditional file installs
  - Case-insensitive directory and filename handling
  - Support for different XML encodings via x/net/html/charset
- [x] **US-014**: Create FOMOD analysis endpoint
  - `POST /api/fomod/analyze` endpoint for full FOMOD analysis
  - Orchestrates: download → extract fomod/ → parse XML → return JSON
  - SQLite caching via modernc.org/sqlite (pure Go)
  - Cache stores analysis results keyed by game/modId/fileId
  - Millisecond-precision TTL for cache expiration
  - Proper cleanup of temp files after analysis
  - Error handling for Premium-only downloads, invalid archives, missing FOMOD
- [x] **US-020**: Add tree view mode
  - FomodTreeView component with collapsible tree hierarchy
  - Shows all FOMOD elements: info, dependencies, steps, groups, plugins, files, flags
  - Expand/collapse all buttons for navigation
  - Type-specific icons and badge indicators
  - View mode toggle in FomodViewer to switch between wizard and tree modes
- [x] **US-021**: Implement plugin header parser
  - Parse TES4 record header from .esp, .esm, .esl plugin files
  - Extract plugin flags: Master (ESM), Light (ESL), Localized
  - Extract MAST subrecords for master file dependencies with sizes
  - Extract CNAM/SNAM for author and description metadata
  - Determine plugin type from flags and file extension
  - Support for Skyrim SE/AE form version (24-byte record headers)
  - Comprehensive test suite with synthetic plugin generation
- [x] **US-022**: Create load order analysis endpoint
  - `POST /api/loadorder/analyze` endpoint for manual plugin list analysis
  - `GET /api/collections/{slug}/revisions/{revision}/loadorder` for collection analysis
  - New loadorder package with Analyzer type for dependency detection
  - Detects missing masters (plugins require masters not in load order)
  - Detects wrong order (plugins load before their masters)
  - Returns dependency graph for frontend visualization
  - Statistics: total plugins, ESM/ESP/ESL counts, issues by severity
  - Downloads and parses plugin headers from Nexus archives
  - Case-insensitive master matching
  - Caching of collection analysis results
  - Comprehensive test suite for analyzer
- [x] **US-023**: Create LoadOrderView container
  - Layout with list and details panels in responsive grid (2/3 + 1/3)
  - Stats header with plugin counts (total, ESM, ESP, ESL) and issue metrics
  - Loading skeleton for async state
  - Error display with API error handling (404, 401, 403, 402, 500+)
  - Empty state for collections without plugins
  - Screen reader announcements for data load
- [x] **US-024**: Implement LoadOrderList component
  - PluginRow component with index, type badges, filename, master count
  - Selection state with visual feedback (border highlight)
  - Issue indicators per plugin showing count and severity
  - Scrollable list with max height constraint
  - Keyboard accessible selection (aria-pressed)
- [x] **US-025**: Implement warning detection UI
  - WarningPanel component showing all load order issues
  - Issue cards with severity badges (error/warning) and type labels
  - Click-to-select plugin from issue panel
  - Success state when no issues found
  - LoadOrderDetails panel showing plugin-specific issues
  - Issue type labels: Missing Master, Wrong Order, Duplicate Plugin

## Discovered Issues

<!-- Document bugs and issues found during implementation -->

## Future Considerations

- LOOT integration for load order suggestions
- Tauri wrapper for local Vortex/MO2 access
- Plugin record-level conflict detection
- Automated conflict resolution suggestions
- Mod comparison tool
- Collection export/import
