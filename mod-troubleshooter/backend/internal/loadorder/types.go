package loadorder

import "github.com/mod-troubleshooter/backend/internal/plugin"

// IssueType represents the type of load order issue.
type IssueType string

const (
	// IssueMissingMaster indicates a plugin requires a master that is not present.
	IssueMissingMaster IssueType = "missing_master"
	// IssueWrongOrder indicates a plugin loads before one of its masters.
	IssueWrongOrder IssueType = "wrong_order"
	// IssueDuplicatePlugin indicates the same plugin appears multiple times.
	IssueDuplicatePlugin IssueType = "duplicate_plugin"
)

// IssueSeverity represents the severity level of an issue.
type IssueSeverity string

const (
	// SeverityError indicates the issue will cause problems.
	SeverityError IssueSeverity = "error"
	// SeverityWarning indicates the issue may cause problems.
	SeverityWarning IssueSeverity = "warning"
)

// Issue represents a detected load order problem.
type Issue struct {
	// Type identifies what kind of issue this is.
	Type IssueType `json:"type"`
	// Severity indicates how serious the issue is.
	Severity IssueSeverity `json:"severity"`
	// Plugin is the filename of the plugin with the issue.
	Plugin string `json:"plugin"`
	// RelatedPlugin is the filename of the related plugin (e.g., missing master).
	RelatedPlugin string `json:"relatedPlugin,omitempty"`
	// Message is a human-readable description of the issue.
	Message string `json:"message"`
	// Index is the position in the load order where the issue occurs.
	Index int `json:"index"`
}

// PluginInfo contains parsed plugin information with load order context.
type PluginInfo struct {
	// Filename is the plugin filename.
	Filename string `json:"filename"`
	// Type is the plugin type (ESM, ESP, ESL).
	Type plugin.PluginType `json:"type"`
	// Flags contains the plugin's flags.
	Flags plugin.PluginFlags `json:"flags"`
	// Author is the plugin author if available.
	Author string `json:"author,omitempty"`
	// Description is the plugin description if available.
	Description string `json:"description,omitempty"`
	// Masters is the list of master dependencies.
	Masters []string `json:"masters"`
	// Index is the position in the load order.
	Index int `json:"index"`
	// HasIssues indicates whether this plugin has any issues.
	HasIssues bool `json:"hasIssues"`
	// IssueCount is the number of issues affecting this plugin.
	IssueCount int `json:"issueCount"`
}

// Stats contains summary statistics about the load order.
type Stats struct {
	// TotalPlugins is the total number of plugins.
	TotalPlugins int `json:"totalPlugins"`
	// ESMCount is the number of ESM (master) plugins.
	ESMCount int `json:"esmCount"`
	// ESPCount is the number of ESP plugins.
	ESPCount int `json:"espCount"`
	// ESLCount is the number of ESL (light) plugins.
	ESLCount int `json:"eslCount"`
	// TotalIssues is the total number of detected issues.
	TotalIssues int `json:"totalIssues"`
	// ErrorCount is the number of error-severity issues.
	ErrorCount int `json:"errorCount"`
	// WarningCount is the number of warning-severity issues.
	WarningCount int `json:"warningCount"`
	// PluginsWithIssues is the count of plugins that have at least one issue.
	PluginsWithIssues int `json:"pluginsWithIssues"`
	// MissingMasters is the count of missing master issues.
	MissingMasters int `json:"missingMasters"`
	// WrongOrderCount is the count of wrong order issues.
	WrongOrderCount int `json:"wrongOrderCount"`
}

// AnalysisResult contains the complete load order analysis.
type AnalysisResult struct {
	// Plugins is the list of plugins in load order.
	Plugins []PluginInfo `json:"plugins"`
	// Issues is the list of detected problems.
	Issues []Issue `json:"issues"`
	// Stats contains summary statistics.
	Stats Stats `json:"stats"`
	// DependencyGraph maps plugin filenames to their masters.
	// Used for visualization in the frontend.
	DependencyGraph map[string][]string `json:"dependencyGraph"`
}

// PluginFile represents a plugin file to be analyzed.
type PluginFile struct {
	// Filename is the plugin filename.
	Filename string
	// Reader provides access to the plugin file contents.
	// If nil, only filename-based analysis is performed.
	Reader interface{}
	// Header contains pre-parsed header information if available.
	Header *plugin.PluginHeader
}
