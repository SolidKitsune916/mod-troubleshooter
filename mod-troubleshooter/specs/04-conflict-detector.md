# Conflict Detector

## Overview

Identify and visualize file-level conflicts between mods in a collection. Help users understand which mods overwrite each other's files and why their setup might have issues.

## Conflict Types

### 1. File Overwrites
Same file path from multiple mods. Only one version can "win".
- **Textures**: Same texture replaced by multiple mods
- **Meshes**: Same model modified by multiple mods
- **Scripts**: Same Papyrus script from multiple mods
- **Records**: Same plugin record modified (requires plugin analysis)

### 2. Resource Conflicts
Mods that modify the same game systems.
- Same leveled list modifications
- Same NPC appearance changes
- Same location modifications

### 3. Compatibility Issues
Known incompatibilities between specific mods.
- Hard incompatibilities (crash/break)
- Soft incompatibilities (visual glitches, minor issues)
- Patch required (works with additional patch mod)

## Data Types

```typescript
interface ConflictAnalysis {
  collectionSlug: string;
  analyzedAt: string;
  modsAnalyzed: number;
  conflicts: FileConflict[];
  summary: ConflictSummary;
}

interface FileConflict {
  id: string;
  filePath: string;
  fileType: FileType;
  severity: ConflictSeverity;
  mods: ConflictingMod[];
  winner: ConflictingMod;      // Based on load order
  resolution?: string;         // Suggested fix
}

interface ConflictingMod {
  modId: number;
  fileId: number;
  modName: string;
  priority: number;           // Mod's position in install order
  fileSize: number;
  fileHash?: string;
}

type FileType =
  | 'texture'      // .dds, .png
  | 'mesh'         // .nif
  | 'script'       // .pex, .psc
  | 'plugin'       // .esp, .esm, .esl
  | 'interface'    // .swf
  | 'sound'        // .wav, .xwm, .fuz
  | 'animation'    // .hkx
  | 'config'       // .ini, .json
  | 'other';

type ConflictSeverity =
  | 'critical'     // Will likely crash or break game
  | 'high'         // Significant visual/gameplay issue
  | 'medium'       // Noticeable but not breaking
  | 'low'          // Minor, might not even notice
  | 'info';        // Just informational, expected overlap

interface ConflictSummary {
  totalConflicts: number;
  bySeverity: Record<ConflictSeverity, number>;
  byFileType: Record<FileType, number>;
  topConflictingMods: Array<{
    modId: number;
    modName: string;
    conflictCount: number;
  }>;
}
```

## Conflict Detection Algorithm

### 1. Extract File Manifests
For each mod in collection:
```go
func extractFileManifest(archivePath string) ([]FileEntry, error) {
    // Open archive (zip, 7z, rar)
    // List all files with paths
    // Calculate file hashes for dedup
    // Return normalized paths
}

type FileEntry struct {
    Path       string  // Normalized: lowercase, forward slashes
    Size       int64
    Hash       string  // MD5 or CRC32
    ModID      int
    FileID     int
}
```

### 2. Build Conflict Map
```go
func findConflicts(manifests []ModManifest) []Conflict {
    // Map: filepath -> []ModID
    fileMap := make(map[string][]ModSource)
    
    for _, manifest := range manifests {
        for _, file := range manifest.Files {
            path := normalizePath(file.Path)
            fileMap[path] = append(fileMap[path], ModSource{
                ModID:  manifest.ModID,
                FileID: manifest.FileID,
                Size:   file.Size,
                Hash:   file.Hash,
            })
        }
    }
    
    // Filter to paths with multiple sources
    var conflicts []Conflict
    for path, sources := range fileMap {
        if len(sources) > 1 {
            conflicts = append(conflicts, Conflict{
                FilePath: path,
                Sources:  sources,
                FileType: classifyFile(path),
            })
        }
    }
    return conflicts
}
```

### 3. Determine Winners
Based on mod install order (priority):
```go
func determineWinner(conflict Conflict, loadOrder []int) ModSource {
    // Mod installed later = higher priority = wins
    var winner ModSource
    highestPriority := -1
    
    for _, source := range conflict.Sources {
        priority := indexOf(loadOrder, source.ModID)
        if priority > highestPriority {
            highestPriority = priority
            winner = source
        }
    }
    return winner
}
```

### 4. Classify Severity
```go
func classifySeverity(conflict Conflict) ConflictSeverity {
    switch conflict.FileType {
    case "script":
        return "critical"  // Script conflicts often cause CTDs
    case "plugin":
        return "high"      // Record conflicts need patches
    case "mesh":
        return "medium"    // Visual issues
    case "texture":
        return "low"       // Just different textures
    case "sound":
        return "low"
    default:
        return "info"
    }
}
```

## UI Components

### ConflictView (Main Container)
```tsx
<ConflictView collectionSlug="qdurkx">
  <ConflictHeader />        {/* Summary stats */}
  <ConflictFilters />       {/* Filter by type, severity */}
  <ConflictList />          {/* Main conflict list */}
  <ConflictDetails />       {/* Selected conflict details */}
</ConflictView>
```

### ConflictHeader
```
┌─────────────────────────────────────────────────────────────┐
│  Conflict Analysis                                          │
│                                                             │
│  ⚠️ 127 conflicts found across 45 mods                      │
│                                                             │
│  🔴 Critical: 3   🟠 High: 12   🟡 Medium: 42   🟢 Low: 70  │
│                                                             │
│  [Analyze Again]                                            │
└─────────────────────────────────────────────────────────────┘
```

### ConflictFilters
- Filter by severity
- Filter by file type
- Filter by mod (show only conflicts involving X)
- Search by file path
- Group by: File / Mod / Severity

### ConflictList
```
┌─────────────────────────────────────────────────────────────┐
│ 🟡 textures/armor/steel/cuirass.dds                         │
│    3 mods: Armor Retex → Steel Overhaul → Complete Pack     │
│    Winner: Complete Pack (priority 45)                      │
├─────────────────────────────────────────────────────────────┤
│ 🔴 scripts/actor/player/playerheadtracking.pex             │
│    2 mods: Immersive First Person → Camera Overhaul         │
│    Winner: Camera Overhaul (priority 38)                    │
│    ⚠️ Script conflict - may cause issues                    │
├─────────────────────────────────────────────────────────────┤
│ 🟢 meshes/weapons/iron/sword.nif                           │
│    2 mods: Weapon Mesh Improvement → True Weapons           │
│    Winner: True Weapons (priority 52)                       │
└─────────────────────────────────────────────────────────────┘
```

### ConflictDetails
When a conflict is selected:
```
┌─────────────────────────────────────────────────────────────┐
│  textures/armor/steel/cuirass.dds                           │
│                                                             │
│  File Type: Texture (.dds)                                  │
│  Severity: Medium                                           │
│                                                             │
│  CONFLICTING MODS                                           │
│  ┌─────────────────────────────────────────────────────────┐│
│  │ #15 Armor Retexture Pack       │ 2.4 MB │ Overwrites    ││
│  │ #32 Steel Armor Overhaul       │ 1.8 MB │ Overwrites    ││
│  │ #45 Complete Texture Pack      │ 2.1 MB │ ✓ WINNER      ││
│  └─────────────────────────────────────────────────────────┘│
│                                                             │
│  RESOLUTION OPTIONS                                         │
│  • Keep winner (Complete Texture Pack)                      │
│  • Change mod order to prefer different version             │
│  • Install a compatibility patch                            │
│                                                             │
│  [Preview Texture A] [Preview Texture B] [Preview Winner]   │
└─────────────────────────────────────────────────────────────┘
```

### ConflictGraph
Visual representation of mod relationships:

```
     ┌──────────────┐
     │  Mod A       │
     │  (12 files)  │
     └──────┬───────┘
            │ overwrites
            ▼
     ┌──────────────┐
     │  Mod B       │◄───── overwrites ─────┐
     │  (8 files)   │                       │
     └──────┬───────┘                       │
            │ overwrites              ┌─────┴──────┐
            ▼                         │   Mod D    │
     ┌──────────────┐                 │  (5 files) │
     │  Mod C       │                 └────────────┘
     │  (15 files)  │
     └──────────────┘
```

## Special Conflict Handling

### Texture Conflicts
- Usually low severity
- Offer image preview comparison
- Winner is just aesthetic preference

### Script Conflicts
- High/Critical severity
- Often no resolution without patch
- Warn user prominently

### Plugin Record Conflicts
- Requires plugin record analysis
- Show which records conflict
- Recommend merged/bashed patches

### Known Incompatibilities
- Check against known incompatibility database
- Display specific warnings for known bad combinations
- Suggest patches or alternatives

## Performance Considerations

### Lazy Loading
- Don't analyze all mods at once
- Analyze on-demand or in background
- Show progress for large collections

### Caching
- Cache file manifests
- Cache conflict results
- Invalidate on mod version change

### Optimization
```go
// Use goroutines for parallel archive extraction
func analyzeCollection(mods []Mod) (*ConflictAnalysis, error) {
    manifests := make(chan ModManifest, len(mods))
    errors := make(chan error, len(mods))
    
    // Parallel extraction
    var wg sync.WaitGroup
    for _, mod := range mods {
        wg.Add(1)
        go func(m Mod) {
            defer wg.Done()
            manifest, err := extractManifest(m)
            if err != nil {
                errors <- err
                return
            }
            manifests <- manifest
        }(mod)
    }
    // ... collect and analyze
}
```

## API Endpoints

```
POST /api/conflicts/analyze
{
  "collectionSlug": "qdurkx",
  "gameId": "skyrimspecialedition",
  "modIds": [12345, 12346, ...]  // Optional: subset of mods
}
```

```
GET /api/conflicts/:analysisId
```
Get cached conflict analysis results.

```
GET /api/conflicts/:analysisId/export
```
Export conflict report as JSON/CSV.

## Acceptance Criteria

- [ ] Extract file manifests from mod archives
- [ ] Identify file path conflicts between mods
- [ ] Classify conflicts by severity
- [ ] Determine winner based on load order
- [ ] Display conflicts grouped by file/mod/severity
- [ ] Show conflict details with resolution suggestions
- [ ] Filter and search conflicts
- [ ] Export conflict report
- [ ] Handle large collections efficiently
- [ ] Cache results for performance
- [ ] Responsive dark theme UI
