# Mod Troubleshooter - Project Overview

## Vision

A web-based tool for Skyrim SE mod users to visualize, analyze, and troubleshoot mod collections from Nexus Mods. Extends the existing SolidKitsune viewer with deep FOMOD analysis, load order visualization, and conflict detection.

## Target User

- Skyrim SE mod users with Nexus Premium
- Uses Vortex or MO2 as mod manager
- Wants to understand why their mod setup isn't working
- Needs to visualize complex FOMOD installer options
- Needs to understand mod conflicts and load order issues

## Core Features

### 1. Collection Browser (Existing - Enhanced)
- View mods in a collection
- Essential vs Optional categorization
- Mod metadata (author, version, description)
- **NEW**: Dependency chain visualization
- **NEW**: Conflict warnings between mods

### 2. FOMOD Visualizer (New)
- Parse and display FOMOD installer structure
- Visualize installation steps as a tree/flow
- Show conditional logic (dependencies between options)
- Display which files each option installs
- Simulate different option selections
- Compare your selections vs recommended

### 3. Load Order Analyzer (New)
- Parse collection's recommended load order
- Visualize plugin dependencies (.esp/.esm/.esl)
- Highlight master file requirements
- Show potential conflicts (plugins modifying same records)
- Integration with LOOT rules (future)

### 4. Conflict Detector (New)
- Identify file-level conflicts (same file from multiple mods)
- Show overwrite order based on mod priority
- Visualize which mod "wins" for each file
- Flag potential compatibility issues

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    React Frontend                            │
│  (Vite + TypeScript + TanStack Query)                       │
├─────────────────────────────────────────────────────────────┤
│  Collection  │  FOMOD        │  Load Order  │  Conflicts    │
│  Browser     │  Visualizer   │  Analyzer    │  Detector     │
└──────────────┴───────────────┴──────────────┴───────────────┘
                              │
                              │ REST API
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Go Backend                               │
├─────────────────────────────────────────────────────────────┤
│  Nexus API   │  Archive      │  FOMOD       │  Conflict     │
│  Client      │  Extractor    │  Parser      │  Analyzer     │
└──────────────┴───────────────┴──────────────┴───────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Data Storage                              │
│  - SQLite for cached mod data                               │
│  - File system for downloaded/extracted archives            │
│  - JSON for collection metadata                             │
└─────────────────────────────────────────────────────────────┘
```

## Tech Stack

### Frontend
- Vite + React 19 + TypeScript (existing)
- TanStack Query for data fetching
- Tailwind CSS for styling
- React Flow or D3 for dependency visualization
- Gaming dark theme (existing design system)

### Backend
- Go 1.22+ with standard library HTTP server
- SQLite for caching
- Archive extraction (zip, 7z, rar support)
- XML parsing for FOMOD

### External APIs
- Nexus Mods GraphQL API (collections, mods)
- Nexus Mods REST API (downloads)

## Data Flow

1. **User enters collection URL** → Frontend
2. **Fetch collection metadata** → Go backend → Nexus GraphQL API
3. **Display mod list** → Frontend renders collection
4. **User requests FOMOD analysis** → Go backend downloads archive
5. **Extract FOMOD XML** → Go backend parses ModuleConfig.xml
6. **Visualize FOMOD** → Frontend renders interactive tree
7. **Analyze conflicts** → Go backend compares file lists
8. **Display results** → Frontend shows conflict report

## Project Structure

```
mod-troubleshooter/
├── frontend/                 # React app (based on viewer-app)
│   ├── src/
│   │   ├── components/
│   │   ├── features/
│   │   │   ├── collections/  # Enhanced collection browser
│   │   │   ├── fomod/        # FOMOD visualizer
│   │   │   ├── loadorder/    # Load order analyzer
│   │   │   └── conflicts/    # Conflict detector
│   │   ├── services/         # API clients
│   │   ├── hooks/
│   │   └── types/
│   └── public/
├── backend/                  # Go API server
│   ├── cmd/server/
│   ├── internal/
│   │   ├── nexus/           # Nexus API client
│   │   ├── fomod/           # FOMOD parser
│   │   ├── archive/         # Archive extraction
│   │   ├── conflict/        # Conflict analysis
│   │   └── handlers/        # HTTP handlers
│   └── pkg/
├── specs/                    # Feature specifications
└── data/                     # Local data storage
```

## Phases

### Phase 1: Foundation
- Set up Go backend with Nexus API client
- Migrate existing viewer to new project structure
- Add API key configuration
- Basic collection fetching through backend

### Phase 2: FOMOD Visualizer
- Archive download and extraction
- FOMOD XML parser
- Frontend tree visualization
- Option selection simulation

### Phase 3: Load Order
- Parse plugin metadata from archives
- Extract master dependencies
- Visualize plugin load order
- Collection recommended order display

### Phase 4: Conflict Detection
- File manifest extraction from archives
- Conflict identification algorithm
- Conflict visualization UI
- Resolution suggestions

## Success Criteria

- User can enter a Nexus collection URL and see all mods
- User can view FOMOD installer structure for any mod
- User can see the recommended load order
- User can identify which mods conflict with each other
- UI maintains gaming aesthetic from existing viewer
