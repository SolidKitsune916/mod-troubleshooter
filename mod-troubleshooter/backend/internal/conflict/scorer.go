package conflict

import (
	"regexp"
	"strings"

	"github.com/mod-troubleshooter/backend/internal/manifest"
)

// ScoreRange defines the possible score values.
const (
	// MaxScore is the maximum possible conflict score.
	MaxScore = 100
	// MinScore is the minimum possible conflict score.
	MinScore = 0
)

// Base scores for file types (0-100 scale).
var fileTypeBaseScores = map[manifest.FileType]int{
	manifest.FileTypePlugin:    90, // Plugin conflicts are very serious
	manifest.FileTypeBSA:       75, // BSA conflicts can cause missing assets
	manifest.FileTypeScript:    70, // Script conflicts affect functionality
	manifest.FileTypeMesh:      50, // Mesh conflicts cause visual issues
	manifest.FileTypeTexture:   45, // Texture conflicts are visual only
	manifest.FileTypeInterface: 55, // Interface conflicts can break UI
	manifest.FileTypeSound:     25, // Sound conflicts are usually minor
	manifest.FileTypeSEQ:       30, // SEQ conflicts affect animations
	manifest.FileTypeOther:     20, // Other files are low priority
}

// Score modifiers.
const (
	// identicalFileDiscount is subtracted for conflicts with identical content.
	identicalFileDiscount = 80
	// multiModBonus is added per additional mod beyond 2 in a conflict.
	multiModBonus = 5
	// ruleMatchBonus is the base bonus for matching an incompatibility rule.
	ruleMatchBonus = 10
)

// RuleMatchType defines how a rule matches file paths or mods.
type RuleMatchType string

const (
	// RuleMatchExact matches exact file paths or mod IDs.
	RuleMatchExact RuleMatchType = "exact"
	// RuleMatchPrefix matches paths/IDs starting with the pattern.
	RuleMatchPrefix RuleMatchType = "prefix"
	// RuleMatchSuffix matches paths/IDs ending with the pattern.
	RuleMatchSuffix RuleMatchType = "suffix"
	// RuleMatchContains matches paths/IDs containing the pattern.
	RuleMatchContains RuleMatchType = "contains"
	// RuleMatchRegex matches paths/IDs using a regex pattern.
	RuleMatchRegex RuleMatchType = "regex"
)

// IncompatibilityRule defines a known incompatibility pattern.
type IncompatibilityRule struct {
	// ID is a unique identifier for the rule.
	ID string `json:"id"`
	// Name is a human-readable name for the rule.
	Name string `json:"name"`
	// Description explains why this incompatibility exists.
	Description string `json:"description"`
	// ScoreBonus is added to the conflict score when this rule matches.
	// Positive values increase severity, negative values decrease it.
	ScoreBonus int `json:"scoreBonus"`
	// PathPattern matches against the conflicting file path.
	PathPattern string `json:"pathPattern,omitempty"`
	// PathMatchType defines how PathPattern is matched.
	PathMatchType RuleMatchType `json:"pathMatchType,omitempty"`
	// ModPatterns matches against mod IDs involved in the conflict.
	// If multiple patterns are specified, ALL must match different mods.
	ModPatterns []string `json:"modPatterns,omitempty"`
	// ModMatchType defines how ModPatterns are matched.
	ModMatchType RuleMatchType `json:"modMatchType,omitempty"`
	// FileTypes restricts the rule to specific file types.
	// Empty means all file types.
	FileTypes []manifest.FileType `json:"fileTypes,omitempty"`
	// compiled regex pattern (internal use)
	compiledPathRegex *regexp.Regexp
	// compiled mod patterns (internal use)
	compiledModRegexes []*regexp.Regexp
}

// Scorer calculates conflict severity scores.
type Scorer struct {
	rules []*IncompatibilityRule
}

// NewScorer creates a new scorer with the default incompatibility rules.
func NewScorer() *Scorer {
	return &Scorer{
		rules: defaultRules(),
	}
}

// NewScorerWithRules creates a scorer with custom rules.
func NewScorerWithRules(rules []*IncompatibilityRule) *Scorer {
	// Compile regex patterns
	for _, rule := range rules {
		if rule.PathMatchType == RuleMatchRegex && rule.PathPattern != "" {
			rule.compiledPathRegex, _ = regexp.Compile(rule.PathPattern)
		}
		if rule.ModMatchType == RuleMatchRegex {
			for _, pattern := range rule.ModPatterns {
				if re, err := regexp.Compile(pattern); err == nil {
					rule.compiledModRegexes = append(rule.compiledModRegexes, re)
				}
			}
		}
	}
	return &Scorer{rules: rules}
}

// Score calculates the severity score for a conflict.
// Returns a score from 0-100 and a list of matched rule IDs.
func (s *Scorer) Score(conflict *Conflict) (int, []string) {
	// Start with base score for file type
	score := s.getBaseScore(conflict.FileType)

	// Apply identical file discount
	if conflict.IsIdentical {
		score -= identicalFileDiscount
	}

	// Apply multi-mod bonus (more mods = more complex conflict)
	if len(conflict.Sources) > 2 {
		score += (len(conflict.Sources) - 2) * multiModBonus
	}

	// Check incompatibility rules
	matchedRules := s.matchRules(conflict)
	for _, rule := range matchedRules {
		score += rule.ScoreBonus
	}

	// Clamp to valid range
	if score > MaxScore {
		score = MaxScore
	}
	if score < MinScore {
		score = MinScore
	}

	// Extract rule IDs
	ruleIDs := make([]string, len(matchedRules))
	for i, rule := range matchedRules {
		ruleIDs[i] = rule.ID
	}

	return score, ruleIDs
}

// getBaseScore returns the base score for a file type.
func (s *Scorer) getBaseScore(fileType manifest.FileType) int {
	if score, ok := fileTypeBaseScores[fileType]; ok {
		return score
	}
	return fileTypeBaseScores[manifest.FileTypeOther]
}

// matchRules finds all rules that match the conflict.
func (s *Scorer) matchRules(conflict *Conflict) []*IncompatibilityRule {
	var matched []*IncompatibilityRule

	for _, rule := range s.rules {
		if s.ruleMatches(rule, conflict) {
			matched = append(matched, rule)
		}
	}

	return matched
}

// ruleMatches checks if a single rule matches the conflict.
func (s *Scorer) ruleMatches(rule *IncompatibilityRule, conflict *Conflict) bool {
	// Check file type restriction
	if len(rule.FileTypes) > 0 {
		found := false
		for _, ft := range rule.FileTypes {
			if ft == conflict.FileType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check path pattern
	if rule.PathPattern != "" {
		if !s.matchPattern(rule.PathPattern, conflict.Path, rule.PathMatchType, rule.compiledPathRegex) {
			return false
		}
	}

	// Check mod patterns (all patterns must match different mods)
	if len(rule.ModPatterns) > 0 {
		modIDs := make([]string, len(conflict.Sources))
		for i, src := range conflict.Sources {
			modIDs[i] = src.ModID
		}

		if !s.matchModPatterns(rule, modIDs) {
			return false
		}
	}

	return true
}

// matchPattern checks if a value matches a pattern using the specified match type.
func (s *Scorer) matchPattern(pattern, value string, matchType RuleMatchType, compiledRegex *regexp.Regexp) bool {
	// Normalize for case-insensitive matching
	patternLower := strings.ToLower(pattern)
	valueLower := strings.ToLower(value)

	switch matchType {
	case RuleMatchExact:
		return valueLower == patternLower
	case RuleMatchPrefix:
		return strings.HasPrefix(valueLower, patternLower)
	case RuleMatchSuffix:
		return strings.HasSuffix(valueLower, patternLower)
	case RuleMatchContains:
		return strings.Contains(valueLower, patternLower)
	case RuleMatchRegex:
		if compiledRegex != nil {
			return compiledRegex.MatchString(value)
		}
		return false
	default:
		return strings.Contains(valueLower, patternLower)
	}
}

// matchModPatterns checks if mod patterns match the mod IDs.
// All patterns must match different mods.
func (s *Scorer) matchModPatterns(rule *IncompatibilityRule, modIDs []string) bool {
	if len(rule.ModPatterns) == 0 {
		return true
	}

	// Each pattern must match at least one unique mod
	matchedMods := make(map[int]bool)

	for patternIdx, pattern := range rule.ModPatterns {
		var compiledRegex *regexp.Regexp
		if rule.ModMatchType == RuleMatchRegex && patternIdx < len(rule.compiledModRegexes) {
			compiledRegex = rule.compiledModRegexes[patternIdx]
		}

		found := false
		for modIdx, modID := range modIDs {
			if matchedMods[modIdx] {
				continue // Already matched to another pattern
			}
			if s.matchPattern(pattern, modID, rule.ModMatchType, compiledRegex) {
				matchedMods[modIdx] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// GetRules returns the configured incompatibility rules.
func (s *Scorer) GetRules() []*IncompatibilityRule {
	return s.rules
}

// defaultRules returns the built-in incompatibility rules.
// These are common patterns for Bethesda game modding.
func defaultRules() []*IncompatibilityRule {
	rules := []*IncompatibilityRule{
		// Critical script paths
		{
			ID:            "skyui-scripts",
			Name:          "SkyUI Script Conflict",
			Description:   "Scripts in the SkyUI path are critical for UI functionality",
			ScoreBonus:    15,
			PathPattern:   "scripts/skyui",
			PathMatchType: RuleMatchPrefix,
			FileTypes:     []manifest.FileType{manifest.FileTypeScript},
		},
		{
			ID:            "skse-scripts",
			Name:          "SKSE Script Conflict",
			Description:   "SKSE plugin scripts are critical for extended functionality",
			ScoreBonus:    15,
			PathPattern:   "scripts/source",
			PathMatchType: RuleMatchPrefix,
			FileTypes:     []manifest.FileType{manifest.FileTypeScript},
		},

		// Critical interface paths
		{
			ID:            "skyui-interface",
			Name:          "SkyUI Interface Conflict",
			Description:   "SkyUI interface files are essential for the modded UI",
			ScoreBonus:    15,
			PathPattern:   "interface/skyui",
			PathMatchType: RuleMatchPrefix,
			FileTypes:     []manifest.FileType{manifest.FileTypeInterface},
		},
		{
			ID:            "mcm-interface",
			Name:          "MCM Interface Conflict",
			Description:   "Mod Configuration Menu interface conflicts",
			ScoreBonus:    10,
			PathPattern:   "interface/quest_journal",
			PathMatchType: RuleMatchContains,
			FileTypes:     []manifest.FileType{manifest.FileTypeInterface},
		},

		// Body/skeleton conflicts
		{
			ID:            "skeleton-conflict",
			Name:          "Skeleton Conflict",
			Description:   "Character skeleton conflicts can break animations and cause crashes",
			ScoreBonus:    20,
			PathPattern:   "skeleton",
			PathMatchType: RuleMatchContains,
			FileTypes:     []manifest.FileType{manifest.FileTypeMesh},
		},
		{
			ID:            "body-mesh",
			Name:          "Body Mesh Conflict",
			Description:   "Character body mesh conflicts affect all NPCs",
			ScoreBonus:    15,
			PathPattern:   "actors/character/character assets",
			PathMatchType: RuleMatchContains,
			FileTypes:     []manifest.FileType{manifest.FileTypeMesh},
		},

		// Animation conflicts
		{
			ID:            "animation-behavior",
			Name:          "Animation Behavior Conflict",
			Description:   "Behavior file conflicts can break character animations",
			ScoreBonus:    20,
			PathPattern:   ".hkx",
			PathMatchType: RuleMatchSuffix,
		},
		{
			ID:            "fnis-nemesis",
			Name:          "Animation Framework Conflict",
			Description:   "Animation framework output conflicts (FNIS/Nemesis)",
			ScoreBonus:    25,
			PathPattern:   "meshes/actors/character/behaviors",
			PathMatchType: RuleMatchPrefix,
		},

		// Combat and gameplay
		{
			ID:            "combat-style",
			Name:          "Combat Style Conflict",
			Description:   "Combat-related conflicts may affect game balance",
			ScoreBonus:    10,
			PathPattern:   "combat",
			PathMatchType: RuleMatchContains,
			FileTypes:     []manifest.FileType{manifest.FileTypeScript},
		},

		// Texture replacers
		{
			ID:            "face-texture",
			Name:          "Face Texture Conflict",
			Description:   "Face texture conflicts can cause visual inconsistencies",
			ScoreBonus:    5,
			PathPattern:   "actors/character/facegendata",
			PathMatchType: RuleMatchPrefix,
			FileTypes:     []manifest.FileType{manifest.FileTypeTexture},
		},

		// Plugin level conflicts
		{
			ID:            "plugin-overwrite",
			Name:          "Plugin Overwrite",
			Description:   "Plugin files being overwritten may lose custom patches",
			ScoreBonus:    10,
			FileTypes:     []manifest.FileType{manifest.FileTypePlugin},
		},
	}

	// Compile regex patterns
	for _, rule := range rules {
		if rule.PathMatchType == RuleMatchRegex && rule.PathPattern != "" {
			rule.compiledPathRegex, _ = regexp.Compile(rule.PathPattern)
		}
	}

	return rules
}
