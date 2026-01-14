package fomod

// GroupType represents the selection behavior of a plugin group.
type GroupType string

const (
	GroupSelectAtLeastOne  GroupType = "SelectAtLeastOne"
	GroupSelectAtMostOne   GroupType = "SelectAtMostOne"
	GroupSelectExactlyOne  GroupType = "SelectExactlyOne"
	GroupSelectAll         GroupType = "SelectAll"
	GroupSelectAny         GroupType = "SelectAny"
)

// PluginType represents the installation status/recommendation of a plugin.
type PluginType string

const (
	PluginRequired      PluginType = "Required"
	PluginOptional      PluginType = "Optional"
	PluginRecommended   PluginType = "Recommended"
	PluginNotUsable     PluginType = "NotUsable"
	PluginCouldBeUsable PluginType = "CouldBeUsable"
)

// FileState represents the state of a file dependency.
type FileState string

const (
	FileStateMissing  FileState = "Missing"
	FileStateInactive FileState = "Inactive"
	FileStateActive   FileState = "Active"
)

// DependencyOperator represents the logical operator for combining dependencies.
type DependencyOperator string

const (
	DependencyOperatorAnd DependencyOperator = "And"
	DependencyOperatorOr  DependencyOperator = "Or"
)

// Order represents the ordering of steps, groups, or plugins.
type Order string

const (
	OrderAscending  Order = "Ascending"
	OrderDescending Order = "Descending"
	OrderExplicit   Order = "Explicit"
)

// Info represents the metadata from info.xml.
type Info struct {
	Name        string `json:"name,omitempty"`
	Author      string `json:"author,omitempty"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	Website     string `json:"website,omitempty"`
	ID          string `json:"id,omitempty"`
}

// ModuleConfig represents the complete parsed ModuleConfig.xml.
type ModuleConfig struct {
	ModuleName              string                   `json:"moduleName"`
	ModuleImage             *HeaderImage             `json:"moduleImage,omitempty"`
	ModuleDependencies      *CompositeDependency     `json:"moduleDependencies,omitempty"`
	RequiredInstallFiles    *FileList                `json:"requiredInstallFiles,omitempty"`
	InstallSteps            []InstallStep            `json:"installSteps,omitempty"`
	ConditionalFileInstalls []ConditionalInstallItem `json:"conditionalFileInstalls,omitempty"`
}

// HeaderImage represents the module image configuration.
type HeaderImage struct {
	Path     string `json:"path"`
	ShowFade bool   `json:"showFade"`
	Height   int    `json:"height,omitempty"`
}

// InstallStep represents a single installation step/page.
type InstallStep struct {
	Name         string           `json:"name"`
	Visible      *Dependency      `json:"visible,omitempty"`
	OptionGroups []OptionGroup    `json:"optionGroups,omitempty"`
}

// OptionGroup represents a group of plugins with selection constraints.
type OptionGroup struct {
	Name    string    `json:"name"`
	Type    GroupType `json:"type"`
	Plugins []Plugin  `json:"plugins,omitempty"`
}

// Plugin represents a single selectable option in the installer.
type Plugin struct {
	Name           string          `json:"name"`
	Description    string          `json:"description,omitempty"`
	Image          string          `json:"image,omitempty"`
	Files          *FileList       `json:"files,omitempty"`
	ConditionFlags []ConditionFlag `json:"conditionFlags,omitempty"`
	TypeDescriptor *TypeDescriptor `json:"typeDescriptor,omitempty"`
}

// TypeDescriptor describes how the plugin type is determined.
type TypeDescriptor struct {
	// Type is set for simple type descriptors
	Type PluginType `json:"type,omitempty"`

	// DependencyType is set for conditional type descriptors
	DependencyType *DependencyPluginType `json:"dependencyType,omitempty"`
}

// DependencyPluginType describes a plugin type that depends on conditions.
type DependencyPluginType struct {
	DefaultType PluginType            `json:"defaultType"`
	Patterns    []DependencyPattern   `json:"patterns,omitempty"`
}

// DependencyPattern maps a dependency condition to a plugin type.
type DependencyPattern struct {
	Dependencies *Dependency `json:"dependencies"`
	Type         PluginType  `json:"type"`
}

// ConditionFlag represents a flag that gets set when a plugin is selected.
type ConditionFlag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// FileList contains files and folders to install.
type FileList struct {
	Files   []FileInstall   `json:"files,omitempty"`
	Folders []FolderInstall `json:"folders,omitempty"`
}

// FileInstall represents a single file to install.
type FileInstall struct {
	Source          string `json:"source"`
	Destination     string `json:"destination,omitempty"`
	Priority        int    `json:"priority,omitempty"`
	AlwaysInstall   bool   `json:"alwaysInstall,omitempty"`
	InstallIfUsable bool   `json:"installIfUsable,omitempty"`
}

// FolderInstall represents a folder to install.
type FolderInstall struct {
	Source          string `json:"source"`
	Destination     string `json:"destination,omitempty"`
	Priority        int    `json:"priority,omitempty"`
	AlwaysInstall   bool   `json:"alwaysInstall,omitempty"`
	InstallIfUsable bool   `json:"installIfUsable,omitempty"`
}

// Dependency represents a condition that must be met.
// It can be a single condition or a composite of multiple conditions.
type Dependency struct {
	// Operator is set for composite dependencies
	Operator DependencyOperator `json:"operator,omitempty"`

	// Children contains nested dependencies when Operator is set
	Children []Dependency `json:"children,omitempty"`

	// FileDependency is set for file-based conditions
	FileDependency *FileDependency `json:"fileDependency,omitempty"`

	// FlagDependency is set for flag-based conditions
	FlagDependency *FlagDependency `json:"flagDependency,omitempty"`

	// GameDependency is set for game version conditions
	GameDependency *VersionDependency `json:"gameDependency,omitempty"`

	// FommDependency is set for FOMM version conditions
	FommDependency *VersionDependency `json:"fommDependency,omitempty"`
}

// FileDependency represents a condition on a file's state.
type FileDependency struct {
	File  string    `json:"file"`
	State FileState `json:"state"`
}

// FlagDependency represents a condition on a flag's value.
type FlagDependency struct {
	Flag  string `json:"flag"`
	Value string `json:"value"`
}

// VersionDependency represents a condition on a version.
type VersionDependency struct {
	Version string `json:"version"`
}

// ConditionalInstallItem represents conditional file installation based on dependencies.
type ConditionalInstallItem struct {
	Dependencies *Dependency `json:"dependencies"`
	Files        *FileList   `json:"files"`
}

// CompositeDependency is an alias for Dependency used at the module level.
type CompositeDependency = Dependency

// FomodData represents the complete parsed FOMOD data from both XML files.
type FomodData struct {
	Info   *Info         `json:"info,omitempty"`
	Config *ModuleConfig `json:"config"`
}
