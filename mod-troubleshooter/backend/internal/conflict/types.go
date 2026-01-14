package conflict

import "github.com/mod-troubleshooter/backend/internal/manifest"

// ConflictType represents the type of file conflict.
type ConflictType string

const (
	// ConflictTypeOverwrite indicates a file that will be overwritten by another mod.
	ConflictTypeOverwrite ConflictType = "overwrite"
	// ConflictTypeDuplicate indicates the same file from multiple mods.
	ConflictTypeDuplicate ConflictType = "duplicate"
)

// Severity represents the severity level of a conflict.
type Severity string

const (
	// SeverityCritical indicates conflicts that will likely break functionality.
	SeverityCritical Severity = "critical"
	// SeverityHigh indicates conflicts that may cause noticeable issues.
	SeverityHigh Severity = "high"
	// SeverityMedium indicates conflicts that could cause minor issues.
	SeverityMedium Severity = "medium"
	// SeverityLow indicates conflicts that are unlikely to cause problems.
	SeverityLow Severity = "low"
	// SeverityInfo indicates informational conflicts (e.g., identical files).
	SeverityInfo Severity = "info"
)

// ModFile represents a file from a specific mod.
type ModFile struct {
	// ModID is the unique identifier for the mod.
	ModID string `json:"modId"`
	// ModName is the display name of the mod.
	ModName string `json:"modName"`
	// Path is the normalized file path within the mod.
	Path string `json:"path"`
	// Size is the file size in bytes.
	Size int64 `json:"size"`
	// Hash is the content hash (if available).
	Hash string `json:"hash,omitempty"`
	// FileType is the type classification of the file.
	FileType manifest.FileType `json:"fileType"`
}

// Conflict represents a detected file conflict between mods.
type Conflict struct {
	// Path is the normalized file path that has conflicts.
	Path string `json:"path"`
	// Type indicates the kind of conflict.
	Type ConflictType `json:"type"`
	// Severity indicates how serious the conflict is.
	Severity Severity `json:"severity"`
	// Score is a numeric severity score from 0-100 for ranking conflicts.
	// Higher scores indicate more serious conflicts.
	Score int `json:"score"`
	// FileType is the type classification of the conflicting file.
	FileType manifest.FileType `json:"fileType"`
	// Sources lists all mods that provide this file.
	Sources []ModFile `json:"sources"`
	// Winner is the mod that will provide the final version (last in load order).
	Winner *ModFile `json:"winner"`
	// Losers are the mods whose files will be overwritten.
	Losers []ModFile `json:"losers"`
	// IsIdentical indicates if all conflicting files have the same content.
	// Only populated when content hashes are available.
	IsIdentical bool `json:"isIdentical"`
	// MatchedRules contains IDs of any incompatibility rules that matched this conflict.
	MatchedRules []string `json:"matchedRules,omitempty"`
	// Message is a human-readable description of the conflict.
	Message string `json:"message"`
}

// ModManifest represents a mod's file manifest with metadata.
type ModManifest struct {
	// ModID is the unique identifier for the mod.
	ModID string `json:"modId"`
	// ModName is the display name of the mod.
	ModName string `json:"modName"`
	// Manifest is the file listing from the mod archive.
	Manifest *manifest.Manifest `json:"manifest"`
	// LoadOrder is the mod's position in the load order (0 = loads first).
	// Higher numbers overwrite lower numbers.
	LoadOrder int `json:"loadOrder"`
}

// Stats contains summary statistics about detected conflicts.
type Stats struct {
	// TotalFiles is the total number of files across all mods.
	TotalFiles int `json:"totalFiles"`
	// UniqueFiles is the number of unique file paths.
	UniqueFiles int `json:"uniqueFiles"`
	// TotalConflicts is the total number of conflicts detected.
	TotalConflicts int `json:"totalConflicts"`
	// CriticalCount is the number of critical severity conflicts.
	CriticalCount int `json:"criticalCount"`
	// HighCount is the number of high severity conflicts.
	HighCount int `json:"highCount"`
	// MediumCount is the number of medium severity conflicts.
	MediumCount int `json:"mediumCount"`
	// LowCount is the number of low severity conflicts.
	LowCount int `json:"lowCount"`
	// InfoCount is the number of info severity conflicts.
	InfoCount int `json:"infoCount"`
	// IdenticalConflicts is the number of conflicts with identical files.
	IdenticalConflicts int `json:"identicalConflicts"`
	// RuleMatchCount is the number of conflicts that matched incompatibility rules.
	RuleMatchCount int `json:"ruleMatchCount"`
	// TotalScore is the sum of all conflict scores.
	TotalScore int `json:"totalScore"`
	// MaxScore is the highest individual conflict score.
	MaxScore int `json:"maxScore"`
	// AverageScore is the average conflict score.
	AverageScore float64 `json:"averageScore"`
	// ByFileType contains conflict counts grouped by file type.
	ByFileType map[manifest.FileType]int `json:"byFileType"`
	// ModsAnalyzed is the number of mods included in the analysis.
	ModsAnalyzed int `json:"modsAnalyzed"`
	// ModsWithConflicts is the number of mods that have at least one conflict.
	ModsWithConflicts int `json:"modsWithConflicts"`
}

// ModConflictSummary contains conflict information for a specific mod.
type ModConflictSummary struct {
	// ModID is the unique identifier for the mod.
	ModID string `json:"modId"`
	// ModName is the display name of the mod.
	ModName string `json:"modName"`
	// TotalConflicts is the number of conflicts involving this mod.
	TotalConflicts int `json:"totalConflicts"`
	// WinCount is the number of conflicts where this mod wins.
	WinCount int `json:"winCount"`
	// LoseCount is the number of conflicts where this mod loses.
	LoseCount int `json:"loseCount"`
	// CriticalCount is the number of critical conflicts for this mod.
	CriticalCount int `json:"criticalCount"`
	// HighCount is the number of high severity conflicts for this mod.
	HighCount int `json:"highCount"`
}

// AnalysisResult contains the complete conflict analysis results.
type AnalysisResult struct {
	// Conflicts is the list of all detected conflicts.
	Conflicts []Conflict `json:"conflicts"`
	// Stats contains summary statistics.
	Stats Stats `json:"stats"`
	// ModSummaries contains per-mod conflict summaries.
	ModSummaries []ModConflictSummary `json:"modSummaries"`
	// FileToMods maps file paths to the mods that provide them.
	// Used for quick lookups in the frontend.
	FileToMods map[string][]string `json:"fileToMods"`
}
