# Product Requirements Document: Mod Troubleshooter

## Executive Summary

Mod Troubleshooter is a web-based diagnostic and visualization tool designed for Skyrim Special Edition mod users. It extends the existing SolidKitsune collection viewer with deep analysis capabilities including FOMOD installer visualization, load order analysis, and file conflict detection. The application helps users understand, troubleshoot, and optimize their mod setups through an intuitive dark-themed gaming interface.

---

## Product Overview

### Vision Statement

Empower Skyrim SE mod users to understand and resolve mod conflicts, visualize complex FOMOD installers, and optimize their load orders through an intuitive, comprehensive troubleshooting platform that transforms opaque modding problems into clear, actionable insights.

### Product Name

**Mod Troubleshooter** (part of the SolidKitsune Project)

### Problem Statement

Skyrim Special Edition modding involves managing hundreds of mods with complex interdependencies. Users face several critical challenges:

1. **Opaque FOMOD Installers**: FOMOD installers present options without clear visibility into what files each choice installs, making it difficult to understand the impact of selections or compare alternatives.

2. **Load Order Complexity**: Plugin dependencies create intricate chains where incorrect ordering causes crashes, missing content, or broken functionality. Users lack tools to visualize these dependencies clearly.

3. **Hidden File Conflicts**: Multiple mods often modify the same game files, but users cannot easily identify which mods conflict or which version "wins" based on install order.

4. **Troubleshooting Difficulty**: When mod setups break, users have no systematic way to diagnose problems, leading to hours of trial-and-error debugging or complete setup rebuilds.

5. **Collection Comprehension**: Nexus Mods collections can contain 400+ mods, making it nearly impossible to understand the full scope without dedicated visualization tools.

---

## Goals and Objectives

### Primary Goals

1. **Simplify Mod Diagnostics**: Reduce the time users spend troubleshooting mod issues by 80% through clear visualization and automated conflict detection.

2. **Demystify FOMOD Installers**: Allow users to fully understand FOMOD installer logic before and after installation, including file destinations and conditional dependencies.

3. **Visualize Dependencies**: Present plugin load order and master dependencies as interactive graphs that make complex relationships immediately comprehensible.

4. **Identify Conflicts Early**: Detect file-level conflicts before they cause in-game issues, with clear severity ratings and resolution suggestions.

### Success Metrics

- Users can identify the root cause of a mod conflict within 5 minutes
- FOMOD structure for any mod is fully visualized within 30 seconds of request
- Load order warnings achieve 95% accuracy for common issues (missing masters, incorrect ordering)
- Conflict detection covers 100% of file-level overwrites with correct winner identification

---

## Scope and Capabilities

### Core Features

#### 1. Collection Browser (Enhanced)

**Purpose**: View and navigate Nexus Mods collections with enhanced metadata and categorization.

**Capabilities**:
- Display all mods in a collection with Essential/Optional categorization
- Show mod metadata: author, version, description, category
- Dependency chain visualization between mods
- Conflict warning indicators on mod cards
- Multi-collection selection and comparison
- Real-time search and filtering
- Grid, list, and carousel view modes

**Data Source**: Nexus Mods GraphQL API via Go backend proxy

#### 2. FOMOD Visualizer (New)

**Purpose**: Parse, display, and simulate FOMOD installer structures.

**Capabilities**:
- Parse `fomod/info.xml` for mod metadata
- Parse `fomod/ModuleConfig.xml` for installation logic
- Display installation steps as navigable wizard or tree view
- Show option groups with selection type constraints (SelectExactlyOne, SelectAny, etc.)
- Visualize conditional logic and flag dependencies
- Display file destinations for each option
- Simulate different selection combinations
- Preview files that will be installed based on current selections
- Export/import option selections as JSON
- Dependency graph view showing option relationships

**Technical Requirements**:
- Backend archive download and extraction (zip, 7z, rar)
- XML parsing for FOMOD schema
- Image extraction for option previews
- Result caching in SQLite (7-day TTL)

#### 3. Load Order Analyzer (New)

**Purpose**: Visualize and validate plugin load order with dependency tracking.

**Capabilities**:
- Display all plugins (.esm, .esp, .esl) with type indicators
- Show master file dependencies for each plugin
- Visualize dependency chains as interactive graph
- Detect and warn about:
  - Missing master files
  - Incorrect load order (master after dependent)
  - Duplicate plugins
  - Approaching 254 plugin slot limit
- Display slot usage statistics (X/254 full plugins, Y light plugins)
- Filter and search plugins
- Group by type, mod, or category
- Export in plugins.txt and loadorder.txt formats

**Data Sources**:
- Collection metadata (recommended load order)
- Plugin header parsing (master dependencies)
- Future: LOOT masterlist integration

#### 4. Conflict Detector (New)

**Purpose**: Identify and visualize file-level conflicts between mods.

**Capabilities**:
- Extract file manifests from mod archives
- Identify files modified by multiple mods
- Classify conflicts by file type (texture, mesh, script, plugin, etc.)
- Assign severity ratings (critical, high, medium, low, info)
- Show "winner" based on mod install order/priority
- Visualize conflict relationships between mods
- Filter by severity, file type, or specific mod
- Provide resolution suggestions
- Group conflicts by file, mod, or severity

**Severity Classification**:
| File Type | Default Severity | Rationale |
|-----------|-----------------|-----------|
| Script (.pex) | Critical | Often causes CTDs |
| Plugin (.esp/.esm) | High | Record conflicts need patches |
| Mesh (.nif) | Medium | Visual issues |
| Texture (.dds) | Low | Aesthetic preference only |
| Sound/Config | Low | Minor impact |

---

## Technical Requirements

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    React Frontend                            │
│  (Vite + React 19 + TypeScript + TanStack Query)            │
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
│  SQLite (cached data) │ File System (archives) │ JSON       │
└─────────────────────────────────────────────────────────────┘
```

### Frontend Stack

- **Build Tool**: Vite
- **Framework**: React 19 with TypeScript (strict mode)
- **Data Fetching**: TanStack Query
- **Styling**: Tailwind CSS with gaming dark theme
- **Visualization**: React Flow or D3.js for dependency graphs
- **State Management**: React Context + TanStack Query cache

### Backend Stack

- **Language**: Go 1.22+
- **HTTP Server**: Standard library `net/http`
- **Database**: SQLite for caching
- **Archive Support**: zip, 7z, rar via `github.com/mholt/archiver/v4`
- **CORS**: `github.com/rs/cors`

### External APIs

- **Nexus Mods GraphQL API**: Collection metadata, mod information
- **Nexus Mods REST API**: File downloads (Premium required)

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/collections/:gameId/:slug` | GET | Fetch collection metadata |
| `/api/collections/:gameId/:slug/loadorder` | GET | Get recommended load order |
| `/api/fomod/analyze` | POST | Analyze FOMOD structure |
| `/api/conflicts/analyze` | POST | Analyze file conflicts |
| `/api/settings` | GET/PUT | User settings (API key) |
| `/api/games` | GET | Supported games list |

### Non-Functional Requirements

#### Performance
- Collection metadata loads in <3 seconds
- FOMOD analysis completes in <30 seconds (including download)
- Conflict analysis for 50 mods completes in <2 minutes
- Frontend maintains 60fps during interactions

#### Caching
- Collection metadata: 24-hour TTL
- FOMOD analysis results: 7-day TTL
- File manifests: 7-day TTL
- Invalidate on mod version change

#### Accessibility
- WCAG 2.2 AA compliance
- Full keyboard navigation
- Screen reader support with ARIA labels
- Visible focus indicators
- Minimum 44×44px touch targets
- Respect `prefers-reduced-motion`

#### Security
- API keys stored securely (not in localStorage)
- No sensitive data in URLs
- Input sanitization for all user inputs
- CORS configured for frontend origin only

---

## User Interface Requirements

### Design System

The application uses the SolidKitsune Gaming design system:

- **Theme**: Dark mode with obsidian backgrounds (#121215, #1a1a1f)
- **Accent**: Glowing cyan (#00e5e5) for interactive elements
- **Secondary**: Deep purple (#6b4d8a) for magical/special elements
- **Typography**: Rajdhani/Exo 2 for headers, Inter for body text
- **Components**: Sharp corners (4-8px radius), layered shadows, glow effects for emphasis

### Key UI Patterns

1. **Card-based layouts** for mod and collection display
2. **Tabbed navigation** between features (Collection, FOMOD, Load Order, Conflicts)
3. **Wizard interface** for FOMOD step-by-step navigation
4. **Tree/graph views** for dependency visualization
5. **Severity-coded badges** for warnings and conflicts
6. **Collapsible panels** for detailed information
7. **Real-time search** with keyboard shortcuts (Ctrl/Cmd+K)

---

## Success Criteria

### Functional Acceptance

- [ ] User can enter a Nexus collection URL and view all mods with metadata
- [ ] User can view FOMOD installer structure for any mod with FOMOD
- [ ] User can simulate different FOMOD option selections and see resulting files
- [ ] User can view the recommended load order from a collection
- [ ] User can see plugin dependencies visualized as an interactive graph
- [ ] User can identify missing masters and load order issues automatically
- [ ] User can analyze file conflicts between selected mods
- [ ] User can see which mod "wins" for each conflicting file
- [ ] User can filter and search all data views
- [ ] User can export load order in standard formats

### Quality Acceptance

- [ ] All interactive elements are keyboard accessible
- [ ] Loading states shown for all async operations
- [ ] Error states provide actionable recovery options
- [ ] UI maintains gaming aesthetic consistently
- [ ] Application works on desktop and tablet viewports
- [ ] All features work without page refresh

---

## Constraints and Assumptions

### Constraints

1. **Nexus Premium Required**: File downloads require Nexus Premium membership
2. **Rate Limits**: Nexus API has daily request limits (~2500/day for Premium)
3. **Game Scope**: Initial release targets Skyrim Special Edition only
4. **Archive Formats**: Must support zip, 7z, and rar formats

### Assumptions

1. Users have their own Nexus Mods API key
2. Users have basic familiarity with mod managers (Vortex, MO2)
3. Users understand concepts like load order and file overwrites
4. Collections being analyzed are publicly accessible on Nexus Mods

---

## Future Considerations

These items are explicitly out of scope for initial release but noted for future planning:

- LOOT masterlist integration for load order suggestions
- Support for additional Bethesda games (Fallout 4, Starfield)
- Mod manager integration (direct Vortex/MO2 communication)
- User accounts and saved configurations
- Community-contributed compatibility patches database
- Real-time collaboration features