# FOMOD Visualizer

## Overview

Interactive visualization of FOMOD installer structure. Allows users to explore installation steps, understand conditional options, and simulate different selection paths.

## FOMOD Structure Reference

### File Structure
```
mod-archive.7z
├── fomod/
│   ├── info.xml           # Mod metadata
│   ├── ModuleConfig.xml   # Installation logic
│   └── images/            # Option preview images
├── option-a/
│   └── ...files
├── option-b/
│   └── ...files
└── required/
    └── ...files
```

### info.xml
```xml
<fomod>
  <Name>Example Mod</Name>
  <Author>Author Name</Author>
  <Version MachineVersion="1.0.0">1.0.0</Version>
  <Description>Mod description here</Description>
  <Website>https://nexusmods.com/...</Website>
</fomod>
```

### ModuleConfig.xml Structure
```xml
<config>
  <moduleName>Example Mod</moduleName>
  
  <!-- Files installed regardless of options -->
  <requiredInstallFiles>
    <folder source="required" destination="" priority="0"/>
  </requiredInstallFiles>
  
  <!-- Installation wizard steps -->
  <installSteps order="Explicit">
    <installStep name="Choose Version">
      <optionalFileGroups order="Explicit">
        <group name="Select an option" type="SelectExactlyOne">
          <plugins order="Explicit">
            <plugin name="Option A">
              <description>Description for Option A</description>
              <image path="fomod/images/option-a.png"/>
              <files>
                <folder source="option-a" destination=""/>
              </files>
              <typeDescriptor>
                <type name="Recommended"/>
              </typeDescriptor>
            </plugin>
          </plugins>
        </group>
      </optionalFileGroups>
    </installStep>
  </installSteps>
  
  <!-- Conditional file installation -->
  <conditionalFileInstalls>
    <patterns>
      <pattern>
        <dependencies operator="And">
          <flagDependency flag="option-a-selected" value="true"/>
        </dependencies>
        <files>
          <folder source="patches/option-a" destination=""/>
        </files>
      </pattern>
    </patterns>
  </conditionalFileInstalls>
</config>
```

## Data Types

### TypeScript Types
```typescript
interface FomodInfo {
  name: string;
  author: string;
  version: string;
  description?: string;
  website?: string;
}

interface FomodAnalysis {
  modId: number;
  fileId: number;
  modName: string;
  hasFomod: boolean;
  info: FomodInfo;
  requiredFiles: FileInstall[];
  steps: InstallStep[];
  conditionalInstalls: ConditionalPattern[];
}

interface InstallStep {
  name: string;
  visible?: DependencyGroup;  // Conditions for step visibility
  groups: OptionGroup[];
}

interface OptionGroup {
  name: string;
  type: GroupType;
  plugins: Plugin[];
}

type GroupType = 
  | 'SelectAtLeastOne'
  | 'SelectAtMostOne'
  | 'SelectExactlyOne'
  | 'SelectAny'
  | 'SelectAll';

interface Plugin {
  name: string;
  description?: string;
  image?: string;
  files: FileInstall[];
  flags: FlagSet[];
  typeDescriptor: PluginType;
  conditionFlags?: DependencyGroup;
}

type PluginType = 
  | 'Required'
  | 'Recommended'
  | 'Optional'
  | 'NotUsable'
  | 'CouldBeUsable';

interface FileInstall {
  source: string;
  destination: string;
  priority?: number;
  isFolder: boolean;
}

interface FlagSet {
  name: string;
  value: string;
}

interface DependencyGroup {
  operator: 'And' | 'Or';
  dependencies: Dependency[];
}

type Dependency =
  | { type: 'flag'; flag: string; value: string }
  | { type: 'file'; file: string; state: 'Active' | 'Inactive' | 'Missing' }
  | { type: 'version'; version: string }
  | { type: 'nested'; group: DependencyGroup };

interface ConditionalPattern {
  dependencies: DependencyGroup;
  files: FileInstall[];
}
```

## UI Components

### FomodViewer (Main Container)
```tsx
<FomodViewer modId={12345} fileId={67890}>
  <FomodHeader />           {/* Mod info, fetch status */}
  <FomodStepNavigator />    {/* Step tabs/breadcrumbs */}
  <FomodStepView />         {/* Current step content */}
  <FomodSummary />          {/* Selected options summary */}
  <FomodFilePreview />      {/* Files to be installed */}
</FomodViewer>
```

### FomodHeader
- Mod name, author, version
- FOMOD fetch status (loading, cached, error)
- Link to Nexus page

### FomodStepNavigator
- Horizontal step indicator
- Shows step names
- Indicates current step
- Grayed out if step is conditionally hidden

### FomodStepView
- Renders current step's groups
- Each group shows its type (radio, checkbox, etc.)
- Options displayed as cards with:
  - Name
  - Description
  - Preview image (if available)
  - Type badge (Recommended, Optional, etc.)
- Selection state managed

### FomodSummary
- Collapsible panel showing all selections
- Organized by step
- Quick overview of choices made

### FomodFilePreview
- Tree view of files that will be installed
- Based on current selections
- Shows source → destination mapping
- Highlights conflicts (same destination from multiple options)

## Visualization Modes

### 1. Wizard Mode (Default)
Step-by-step walkthrough matching mod manager experience.
- One step at a time
- Previous/Next navigation
- Selection persistence

### 2. Tree Mode
Hierarchical view of entire FOMOD structure.
```
📦 Example Mod
├── 📋 Required Files
│   └── 📁 required/ → /
├── 🔢 Step 1: Choose Version
│   └── 📌 Select an option (SelectExactlyOne)
│       ├── ⚪ Option A [Recommended]
│       └── ⚪ Option B [Optional]
└── 🔢 Step 2: Patches
    └── 📌 Select patches (SelectAny)
        ├── ☑️ Patch 1
        └── ☑️ Patch 2
```

### 3. Dependency Graph Mode
Visual graph showing option dependencies.
- Nodes = Options
- Edges = Dependencies (flag requirements)
- Highlights chains and conflicts

## User Interactions

### Selection
- Click option to select (respects group type)
- Shows immediate impact on:
  - Files to install
  - Other options (enables/disables)
  - Conditional patterns triggered

### Comparison
- Toggle "Compare Mode"
- Select two different option combinations
- Side-by-side file diff
- Useful for "What's different between Option A and B?"

### Search
- Filter options by name/description
- Find specific files in any option
- Highlight matching options

### Export
- Export selected options as JSON
- Import previous selections
- Share configuration with others

## State Management

```typescript
interface FomodState {
  // Data
  analysis: FomodAnalysis | null;
  loading: boolean;
  error: string | null;
  
  // Navigation
  currentStepIndex: number;
  viewMode: 'wizard' | 'tree' | 'graph';
  
  // Selections
  selections: Map<string, Set<string>>; // stepName → selected plugin names
  flags: Map<string, string>;           // flag name → value
  
  // Computed
  visibleSteps: InstallStep[];          // Steps visible based on conditions
  filesToInstall: FileInstall[];        // Final file list
  conflicts: FileConflict[];            // Internal FOMOD conflicts
}
```

## API Integration

### Fetch FOMOD
```typescript
const { data, isLoading, error } = useQuery({
  queryKey: ['fomod', modId, fileId],
  queryFn: () => api.analyzeFomod(modId, fileId),
  staleTime: 1000 * 60 * 60 * 24, // 24 hours
});
```

### Caching Strategy
- Cache parsed FOMOD data in React Query
- Backend also caches in SQLite
- Re-fetch only if mod version changes

## Accessibility

- Keyboard navigation through steps and options
- Screen reader announcements for selections
- Focus management when step changes
- Clear visual indicators for selection state
- ARIA labels for all interactive elements

## Acceptance Criteria

- [ ] Can fetch and display FOMOD structure for any mod
- [ ] Wizard mode allows step-by-step navigation
- [ ] Tree mode shows full structure at once
- [ ] Options correctly reflect group type constraints
- [ ] Selecting options updates file preview
- [ ] Conditional steps show/hide based on flags
- [ ] Can export/import option selections
- [ ] UI follows gaming dark theme
- [ ] Fully keyboard accessible
- [ ] Loading and error states handled gracefully
