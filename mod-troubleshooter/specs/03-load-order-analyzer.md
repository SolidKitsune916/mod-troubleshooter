# Load Order Analyzer

## Overview

Visualize and analyze plugin load order for Bethesda games (Skyrim SE). Shows master dependencies, identifies issues, and helps users understand why their plugins need to be in a specific order.

## Skyrim Plugin System

### Plugin Types
| Extension | Type | Limit | Description |
|-----------|------|-------|-------------|
| `.esm` | Master | 254 | Elder Scrolls Master - always loads first |
| `.esp` | Plugin | 254 | Elder Scrolls Plugin - regular mod plugins |
| `.esl` | Light | 4096 | Light plugin - doesn't count toward 254 limit |
| `.esp` (flagged) | Light | 4096 | ESP flagged as ESL in header |

### Load Order Rules
1. **Masters first**: All .esm files load before .esp files
2. **Master requirements**: If plugin A requires plugin B as master, B must load before A
3. **Total limit**: Maximum 254 full plugins (esm + esp), unlimited light plugins
4. **Official files**: Skyrim.esm, Update.esm, Dawnguard.esm, HearthFires.esm, Dragonborn.esm always first

### Plugin Header Info
Each plugin contains:
- Plugin name
- Author
- Description  
- Master file list (dependencies)
- Record count
- Flags (ESL, localized, etc.)

## Data Types

```typescript
interface Plugin {
  filename: string;
  type: 'esm' | 'esp' | 'esl';
  isLightFlagged: boolean;
  modId?: number;           // Nexus mod ID if known
  modName?: string;         // Human-readable mod name
  author?: string;
  description?: string;
  masters: string[];        // Required master files
  loadIndex: number;        // Position in load order
}

interface LoadOrder {
  plugins: Plugin[];
  warnings: LoadOrderWarning[];
  stats: {
    totalPlugins: number;
    masterCount: number;
    espCount: number;
    eslCount: number;
    slotsUsed: number;      // Out of 254
    slotsRemaining: number;
  };
}

interface LoadOrderWarning {
  type: WarningType;
  severity: 'error' | 'warning' | 'info';
  plugin: string;
  message: string;
  details?: string;
}

type WarningType =
  | 'missing_master'        // Required master not in load order
  | 'master_after_dependent'// Master loads after plugin that needs it
  | 'duplicate_plugin'      // Same plugin listed twice
  | 'slot_limit'            // Approaching 254 limit
  | 'outdated_version'      // Known outdated version
  | 'incompatible'          // Known incompatibility with another mod
  | 'load_order_suggestion';// LOOT-style suggestion
```

## Data Sources

### 1. Collection Metadata (Primary)
Some collections include recommended load order in their revision data.

```graphql
query CollectionLoadOrder($slug: String!, $revision: Int) {
  collectionRevision(slug: $slug, revision: $revision) {
    modFiles {
      fileId
      file {
        name
        mod { modId name }
      }
    }
    # Collections may include load order info
    # in externalResources or as bundled file
  }
}
```

### 2. Plugin Analysis (Secondary)
Extract master dependencies from downloaded plugin files.

```go
// Read plugin header to get masters
func parsePluginHeader(reader io.Reader) (*PluginInfo, error) {
    // TES4 record header parsing
    // Extract MAST subrecords for dependencies
}
```

### 3. LOOT Masterlist (Future)
https://github.com/loot/loot - Load Order Optimization Tool rules.

## UI Components

### LoadOrderView (Main Container)
```tsx
<LoadOrderView collectionSlug="qdurkx">
  <LoadOrderHeader />       {/* Stats, warnings summary */}
  <LoadOrderToolbar />      {/* View options, search */}
  <LoadOrderList />         {/* Plugin list */}
  <LoadOrderDetails />      {/* Selected plugin details */}
</LoadOrderView>
```

### LoadOrderHeader
- Total plugin count with breakdown by type
- Slot usage visualization (X/254)
- Warning/error count badges

### LoadOrderToolbar
- View toggle: List / Graph
- Search/filter plugins
- Group by: Type / Mod / Category
- Show/hide official plugins

### LoadOrderList
```
┌─────────────────────────────────────────────────────────────┐
│ # │ Type │ Plugin Name              │ Mod           │ Status│
├───┼──────┼──────────────────────────┼───────────────┼───────┤
│ 00│ ESM  │ Skyrim.esm               │ Official      │   ✓   │
│ 01│ ESM  │ Update.esm               │ Official      │   ✓   │
│ 02│ ESM  │ Dawnguard.esm            │ Official      │   ✓   │
│...│      │                          │               │       │
│ 10│ ESM  │ USSEP.esm                │ USSEP         │   ✓   │
│ 11│ ESP  │ SkyUI_SE.esp             │ SkyUI         │   ✓   │
│ 12│ ESP  │ EnhancedLights.esp       │ ELFX          │   ⚠️   │
│   │      │ └─ Warning: Load after   │               │       │
│   │      │    Realistic Lighting    │               │       │
└─────────────────────────────────────────────────────────────┘
```

Features:
- Drag-and-drop reordering (simulation mode)
- Color-coded by type
- Inline warnings with expandable details
- Click to select and show details panel

### LoadOrderDetails
- Full plugin information
- Master dependencies (with links)
- Dependent plugins (what depends on this)
- Conflicts/warnings
- Nexus link

### LoadOrderGraph
Alternative visualization showing dependency relationships.

```
[Skyrim.esm]
     │
     ├──────────────────────┐
     │                      │
     ▼                      ▼
[USSEP.esm]           [SkyUI_SE.esp]
     │                      
     ├───────────┐          
     │           │          
     ▼           ▼          
[ModA.esp]  [ModB.esp]      
```

Using React Flow or D3 for interactive graph.

## Warning Detection Logic

### Missing Master
```typescript
function checkMissingMasters(plugins: Plugin[]): Warning[] {
  const available = new Set(plugins.map(p => p.filename.toLowerCase()));
  const warnings: Warning[] = [];
  
  for (const plugin of plugins) {
    for (const master of plugin.masters) {
      if (!available.has(master.toLowerCase())) {
        warnings.push({
          type: 'missing_master',
          severity: 'error',
          plugin: plugin.filename,
          message: `Missing required master: ${master}`,
        });
      }
    }
  }
  return warnings;
}
```

### Load Order Issues
```typescript
function checkLoadOrder(plugins: Plugin[]): Warning[] {
  const warnings: Warning[] = [];
  const positions = new Map(plugins.map((p, i) => [p.filename.toLowerCase(), i]));
  
  for (const plugin of plugins) {
    const pluginPos = positions.get(plugin.filename.toLowerCase())!;
    
    for (const master of plugin.masters) {
      const masterPos = positions.get(master.toLowerCase());
      if (masterPos !== undefined && masterPos > pluginPos) {
        warnings.push({
          type: 'master_after_dependent',
          severity: 'error',
          plugin: plugin.filename,
          message: `Master "${master}" loads after this plugin`,
        });
      }
    }
  }
  return warnings;
}
```

## Features

### 1. Load Order Validation
- Check all master dependencies are met
- Verify correct ordering
- Identify slot limit issues

### 2. Dependency Visualization
- Show what each plugin needs
- Show what depends on each plugin
- Trace dependency chains

### 3. Comparison
- Compare collection's order vs your current order
- Highlight differences
- Suggest corrections

### 4. Export
- Export as plugins.txt format
- Export as loadorder.txt format
- Copy to clipboard

## State Management

```typescript
interface LoadOrderState {
  // Data
  loadOrder: LoadOrder | null;
  loading: boolean;
  error: string | null;
  
  // View
  viewMode: 'list' | 'graph';
  groupBy: 'none' | 'type' | 'mod';
  showOfficial: boolean;
  searchQuery: string;
  
  // Selection
  selectedPlugin: string | null;
  
  // Simulation
  simulationMode: boolean;
  simulatedOrder: Plugin[];  // For drag-drop reordering
}
```

## API Endpoints

```
GET /api/collections/:gameId/:slug/loadorder
```
Returns parsed load order from collection.

```
POST /api/plugins/analyze
```
Analyze uploaded plugin file(s) for master dependencies.

## Acceptance Criteria

- [ ] Display all plugins from a collection
- [ ] Show plugin type badges (ESM/ESP/ESL)
- [ ] Display master dependencies for each plugin
- [ ] Detect and show missing master warnings
- [ ] Detect and show load order issues
- [ ] Visualize dependencies as interactive graph
- [ ] Show slot usage statistics
- [ ] Filter/search plugins
- [ ] Export load order in standard formats
- [ ] Responsive dark theme UI
