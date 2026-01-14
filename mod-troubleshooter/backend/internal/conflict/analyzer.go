package conflict

import (
	"context"
	"fmt"
	"sort"

	"github.com/mod-troubleshooter/backend/internal/manifest"
)

// Analyzer detects file conflicts between mods.
type Analyzer struct {
	scorer *Scorer
}

// NewAnalyzer creates a new conflict analyzer with default scoring rules.
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		scorer: NewScorer(),
	}
}

// NewAnalyzerWithRules creates a new conflict analyzer with custom incompatibility rules.
func NewAnalyzerWithRules(rules []*IncompatibilityRule) *Analyzer {
	return &Analyzer{
		scorer: NewScorerWithRules(rules),
	}
}

// Analyze detects conflicts between the given mod manifests.
// Mods are expected to be in load order (index 0 = loads first, higher index = overwrites lower).
func (a *Analyzer) Analyze(ctx context.Context, mods []ModManifest) (*AnalysisResult, error) {
	result := &AnalysisResult{
		Conflicts:    make([]Conflict, 0),
		ModSummaries: make([]ModConflictSummary, 0, len(mods)),
		FileToMods:   make(map[string][]string),
		Stats: Stats{
			ByFileType: make(map[manifest.FileType]int),
		},
	}

	// Build file -> mods map
	fileMap := a.buildFileMap(mods)

	// Track total files and unique files
	result.Stats.UniqueFiles = len(fileMap)
	for _, files := range fileMap {
		result.Stats.TotalFiles += len(files)
	}

	// Initialize mod summaries
	modSummaryMap := make(map[string]*ModConflictSummary)
	for _, mod := range mods {
		summary := ModConflictSummary{
			ModID:   mod.ModID,
			ModName: mod.ModName,
		}
		modSummaryMap[mod.ModID] = &summary
	}

	// Detect conflicts (files with multiple sources)
	for path, files := range fileMap {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if len(files) < 2 {
			continue
		}

		// Sort by load order to determine winner/losers
		sort.Slice(files, func(i, j int) bool {
			return files[i].loadOrder < files[j].loadOrder
		})

		conflict := a.createConflict(path, files)
		result.Conflicts = append(result.Conflicts, conflict)

		// Update file to mods mapping
		modIDs := make([]string, len(files))
		for i, f := range files {
			modIDs[i] = f.modFile.ModID
		}
		result.FileToMods[path] = modIDs

		// Update mod summaries
		a.updateModSummaries(modSummaryMap, &conflict)
	}

	// Sort conflicts by severity (critical first), then by score (descending), then by path
	sort.Slice(result.Conflicts, func(i, j int) bool {
		if result.Conflicts[i].Severity != result.Conflicts[j].Severity {
			return severityOrder(result.Conflicts[i].Severity) < severityOrder(result.Conflicts[j].Severity)
		}
		if result.Conflicts[i].Score != result.Conflicts[j].Score {
			return result.Conflicts[i].Score > result.Conflicts[j].Score // Higher score first
		}
		return result.Conflicts[i].Path < result.Conflicts[j].Path
	})

	// Calculate stats
	result.Stats = a.calculateStats(result, len(mods))

	// Build mod summaries list
	for _, mod := range mods {
		if summary, ok := modSummaryMap[mod.ModID]; ok {
			result.ModSummaries = append(result.ModSummaries, *summary)
		}
	}

	return result, nil
}

// fileWithContext holds a ModFile with its load order context.
type fileWithContext struct {
	modFile   ModFile
	loadOrder int
}

// buildFileMap creates a map of file paths to all mods that provide them.
func (a *Analyzer) buildFileMap(mods []ModManifest) map[string][]fileWithContext {
	fileMap := make(map[string][]fileWithContext)

	for _, mod := range mods {
		if mod.Manifest == nil {
			continue
		}

		for _, entry := range mod.Manifest.Files {
			modFile := ModFile{
				ModID:    mod.ModID,
				ModName:  mod.ModName,
				Path:     entry.Path,
				Size:     entry.Size,
				Hash:     entry.Hash,
				FileType: entry.Type,
			}

			fileMap[entry.Path] = append(fileMap[entry.Path], fileWithContext{
				modFile:   modFile,
				loadOrder: mod.LoadOrder,
			})
		}
	}

	return fileMap
}

// createConflict creates a Conflict from a list of files (sorted by load order).
func (a *Analyzer) createConflict(path string, files []fileWithContext) Conflict {
	sources := make([]ModFile, len(files))
	for i, f := range files {
		sources[i] = f.modFile
	}

	// Winner is the last in load order (highest index)
	winner := files[len(files)-1].modFile

	// Losers are all others
	losers := make([]ModFile, len(files)-1)
	for i := 0; i < len(files)-1; i++ {
		losers[i] = files[i].modFile
	}

	// Determine file type (all should be the same for same path)
	fileType := files[0].modFile.FileType

	// Check if all files are identical (only if hashes are available)
	isIdentical := a.checkIdentical(files)

	// Determine severity based on file type and whether files are identical
	severity := a.determineSeverity(fileType, isIdentical)

	// Determine conflict type
	conflictType := ConflictTypeOverwrite
	if isIdentical {
		conflictType = ConflictTypeDuplicate
	}

	// Generate message
	message := a.generateMessage(path, &winner, losers, isIdentical)

	// Create conflict without score first (need full conflict to calculate score)
	conflict := Conflict{
		Path:        path,
		Type:        conflictType,
		Severity:    severity,
		FileType:    fileType,
		Sources:     sources,
		Winner:      &winner,
		Losers:      losers,
		IsIdentical: isIdentical,
		Message:     message,
	}

	// Calculate score using the scorer
	score, matchedRules := a.scorer.Score(&conflict)
	conflict.Score = score
	conflict.MatchedRules = matchedRules

	return conflict
}

// checkIdentical checks if all files have the same content hash.
func (a *Analyzer) checkIdentical(files []fileWithContext) bool {
	if len(files) < 2 {
		return true
	}

	// If any file doesn't have a hash, we can't determine
	firstHash := files[0].modFile.Hash
	if firstHash == "" {
		return false
	}

	for i := 1; i < len(files); i++ {
		if files[i].modFile.Hash == "" || files[i].modFile.Hash != firstHash {
			return false
		}
	}

	return true
}

// determineSeverity determines conflict severity based on file type and identity.
func (a *Analyzer) determineSeverity(fileType manifest.FileType, isIdentical bool) Severity {
	// Identical files are just informational
	if isIdentical {
		return SeverityInfo
	}

	// Severity based on file type
	switch fileType {
	case manifest.FileTypePlugin:
		// Plugin conflicts are critical - they can break the game
		return SeverityCritical
	case manifest.FileTypeBSA:
		// BSA conflicts can cause missing assets
		return SeverityHigh
	case manifest.FileTypeScript:
		// Script conflicts can cause functionality issues
		return SeverityHigh
	case manifest.FileTypeMesh:
		// Mesh conflicts can cause visual issues
		return SeverityMedium
	case manifest.FileTypeTexture:
		// Texture conflicts are usually visual only
		return SeverityMedium
	case manifest.FileTypeSound:
		// Sound conflicts are usually minor
		return SeverityLow
	case manifest.FileTypeInterface:
		// Interface conflicts can cause UI issues
		return SeverityMedium
	case manifest.FileTypeSEQ:
		// SEQ conflicts affect animations
		return SeverityLow
	default:
		return SeverityLow
	}
}

// generateMessage generates a human-readable conflict message.
func (a *Analyzer) generateMessage(path string, winner *ModFile, losers []ModFile, isIdentical bool) string {
	if isIdentical {
		return fmt.Sprintf("File '%s' is provided by %d mods with identical content", path, len(losers)+1)
	}

	if len(losers) == 1 {
		return fmt.Sprintf("File '%s' from '%s' overwrites '%s'", path, winner.ModName, losers[0].ModName)
	}

	return fmt.Sprintf("File '%s' from '%s' overwrites %d other mod(s)", path, winner.ModName, len(losers))
}

// updateModSummaries updates the mod summaries with conflict information.
func (a *Analyzer) updateModSummaries(summaries map[string]*ModConflictSummary, conflict *Conflict) {
	// Update winner
	if conflict.Winner != nil {
		if summary, ok := summaries[conflict.Winner.ModID]; ok {
			summary.TotalConflicts++
			summary.WinCount++
			if conflict.Severity == SeverityCritical {
				summary.CriticalCount++
			} else if conflict.Severity == SeverityHigh {
				summary.HighCount++
			}
		}
	}

	// Update losers
	for _, loser := range conflict.Losers {
		if summary, ok := summaries[loser.ModID]; ok {
			summary.TotalConflicts++
			summary.LoseCount++
			if conflict.Severity == SeverityCritical {
				summary.CriticalCount++
			} else if conflict.Severity == SeverityHigh {
				summary.HighCount++
			}
		}
	}
}

// calculateStats computes summary statistics from the analysis.
func (a *Analyzer) calculateStats(result *AnalysisResult, modCount int) Stats {
	stats := Stats{
		TotalFiles:     result.Stats.TotalFiles,
		UniqueFiles:    result.Stats.UniqueFiles,
		TotalConflicts: len(result.Conflicts),
		ModsAnalyzed:   modCount,
		ByFileType:     make(map[manifest.FileType]int),
	}

	modsWithConflicts := make(map[string]bool)

	for _, conflict := range result.Conflicts {
		// Count by severity
		switch conflict.Severity {
		case SeverityCritical:
			stats.CriticalCount++
		case SeverityHigh:
			stats.HighCount++
		case SeverityMedium:
			stats.MediumCount++
		case SeverityLow:
			stats.LowCount++
		case SeverityInfo:
			stats.InfoCount++
		}

		// Count identical conflicts
		if conflict.IsIdentical {
			stats.IdenticalConflicts++
		}

		// Count rule matches
		if len(conflict.MatchedRules) > 0 {
			stats.RuleMatchCount++
		}

		// Track score statistics
		stats.TotalScore += conflict.Score
		if conflict.Score > stats.MaxScore {
			stats.MaxScore = conflict.Score
		}

		// Count by file type
		stats.ByFileType[conflict.FileType]++

		// Track mods with conflicts
		for _, source := range conflict.Sources {
			modsWithConflicts[source.ModID] = true
		}
	}

	stats.ModsWithConflicts = len(modsWithConflicts)

	// Calculate average score
	if len(result.Conflicts) > 0 {
		stats.AverageScore = float64(stats.TotalScore) / float64(len(result.Conflicts))
	}

	return stats
}

// severityOrder returns a numeric order for sorting severities.
func severityOrder(s Severity) int {
	switch s {
	case SeverityCritical:
		return 0
	case SeverityHigh:
		return 1
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 3
	case SeverityInfo:
		return 4
	default:
		return 5
	}
}
