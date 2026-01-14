package conflict

import (
	"testing"

	"github.com/mod-troubleshooter/backend/internal/manifest"
)

func TestScorer_Score_BaseScores(t *testing.T) {
	scorer := NewScorer()

	tests := []struct {
		name     string
		fileType manifest.FileType
		minScore int
		maxScore int
	}{
		{"plugin", manifest.FileTypePlugin, 85, 100},
		{"bsa", manifest.FileTypeBSA, 70, 85},
		{"script", manifest.FileTypeScript, 65, 80},
		{"mesh", manifest.FileTypeMesh, 45, 60},
		{"texture", manifest.FileTypeTexture, 40, 55},
		{"interface", manifest.FileTypeInterface, 50, 65},
		{"sound", manifest.FileTypeSound, 20, 35},
		{"seq", manifest.FileTypeSEQ, 25, 40},
		{"other", manifest.FileTypeOther, 15, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflict := &Conflict{
				Path:     "test/file.ext",
				FileType: tt.fileType,
				Sources: []ModFile{
					{ModID: "mod1", ModName: "Mod One"},
					{ModID: "mod2", ModName: "Mod Two"},
				},
				IsIdentical: false,
			}

			score, _ := scorer.Score(conflict)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score in range [%d, %d], got %d", tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestScorer_Score_IdenticalDiscount(t *testing.T) {
	scorer := NewScorer()

	// Create two conflicts - one with identical files, one without
	baseConflict := &Conflict{
		Path:     "textures/test.dds",
		FileType: manifest.FileTypeTexture,
		Sources: []ModFile{
			{ModID: "mod1", ModName: "Mod One"},
			{ModID: "mod2", ModName: "Mod Two"},
		},
		IsIdentical: false,
	}

	identicalConflict := &Conflict{
		Path:     "textures/test.dds",
		FileType: manifest.FileTypeTexture,
		Sources: []ModFile{
			{ModID: "mod1", ModName: "Mod One"},
			{ModID: "mod2", ModName: "Mod Two"},
		},
		IsIdentical: true,
	}

	baseScore, _ := scorer.Score(baseConflict)
	identicalScore, _ := scorer.Score(identicalConflict)

	if identicalScore >= baseScore {
		t.Errorf("identical files should have lower score: base=%d, identical=%d", baseScore, identicalScore)
	}

	// Identical should be clamped to minimum
	if identicalScore < MinScore {
		t.Errorf("score should not go below MinScore, got %d", identicalScore)
	}
}

func TestScorer_Score_MultiModBonus(t *testing.T) {
	scorer := NewScorer()

	twoModConflict := &Conflict{
		Path:     "test.dds",
		FileType: manifest.FileTypeTexture,
		Sources: []ModFile{
			{ModID: "mod1", ModName: "Mod One"},
			{ModID: "mod2", ModName: "Mod Two"},
		},
		IsIdentical: false,
	}

	threeModConflict := &Conflict{
		Path:     "test.dds",
		FileType: manifest.FileTypeTexture,
		Sources: []ModFile{
			{ModID: "mod1", ModName: "Mod One"},
			{ModID: "mod2", ModName: "Mod Two"},
			{ModID: "mod3", ModName: "Mod Three"},
		},
		IsIdentical: false,
	}

	twoScore, _ := scorer.Score(twoModConflict)
	threeScore, _ := scorer.Score(threeModConflict)

	if threeScore <= twoScore {
		t.Errorf("three-mod conflict should have higher score: two=%d, three=%d", twoScore, threeScore)
	}
}

func TestScorer_Score_RuleMatching(t *testing.T) {
	scorer := NewScorer()

	tests := []struct {
		name          string
		path          string
		fileType      manifest.FileType
		expectRules   bool
		minRuleBonus  int
	}{
		{
			name:         "skyui script",
			path:         "scripts/skyui/something.pex",
			fileType:     manifest.FileTypeScript,
			expectRules:  true,
			minRuleBonus: 10,
		},
		{
			name:         "skeleton mesh",
			path:         "meshes/actors/character/character assets/skeleton.nif",
			fileType:     manifest.FileTypeMesh,
			expectRules:  true,
			minRuleBonus: 15,
		},
		{
			name:         "animation behavior",
			path:         "meshes/actors/character/behaviors/0_master.hkx",
			fileType:     manifest.FileTypeOther,
			expectRules:  true,
			minRuleBonus: 20,
		},
		{
			name:         "generic texture no rule",
			path:         "textures/landscape/grass.dds",
			fileType:     manifest.FileTypeTexture,
			expectRules:  false,
			minRuleBonus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflict := &Conflict{
				Path:     tt.path,
				FileType: tt.fileType,
				Sources: []ModFile{
					{ModID: "mod1", ModName: "Mod One"},
					{ModID: "mod2", ModName: "Mod Two"},
				},
				IsIdentical: false,
			}

			_, matchedRules := scorer.Score(conflict)

			if tt.expectRules && len(matchedRules) == 0 {
				t.Error("expected rules to match, but none did")
			}

			if !tt.expectRules && len(matchedRules) > 0 {
				t.Errorf("expected no rules to match, but got %v", matchedRules)
			}
		})
	}
}

func TestScorer_Score_MaxScore(t *testing.T) {
	scorer := NewScorer()

	// Create a conflict with many aggravating factors
	conflict := &Conflict{
		Path:     "scripts/skyui/test.pex",
		FileType: manifest.FileTypePlugin, // Critical file type
		Sources: []ModFile{
			{ModID: "mod1", ModName: "Mod One"},
			{ModID: "mod2", ModName: "Mod Two"},
			{ModID: "mod3", ModName: "Mod Three"},
			{ModID: "mod4", ModName: "Mod Four"},
			{ModID: "mod5", ModName: "Mod Five"},
		},
		IsIdentical: false,
	}

	score, _ := scorer.Score(conflict)

	if score > MaxScore {
		t.Errorf("score should not exceed MaxScore=%d, got %d", MaxScore, score)
	}
}

func TestScorer_Score_MinScore(t *testing.T) {
	scorer := NewScorer()

	// Create an identical conflict of a low-severity type
	conflict := &Conflict{
		Path:     "readme.txt",
		FileType: manifest.FileTypeOther,
		Sources: []ModFile{
			{ModID: "mod1", ModName: "Mod One"},
			{ModID: "mod2", ModName: "Mod Two"},
		},
		IsIdentical: true,
	}

	score, _ := scorer.Score(conflict)

	if score < MinScore {
		t.Errorf("score should not go below MinScore=%d, got %d", MinScore, score)
	}
}

func TestNewScorerWithRules_CustomRules(t *testing.T) {
	customRules := []*IncompatibilityRule{
		{
			ID:            "custom-rule",
			Name:          "Custom Test Rule",
			Description:   "A custom rule for testing",
			ScoreBonus:    50,
			PathPattern:   "custom",
			PathMatchType: RuleMatchContains,
		},
	}

	scorer := NewScorerWithRules(customRules)

	conflict := &Conflict{
		Path:     "textures/custom/test.dds",
		FileType: manifest.FileTypeTexture,
		Sources: []ModFile{
			{ModID: "mod1", ModName: "Mod One"},
			{ModID: "mod2", ModName: "Mod Two"},
		},
		IsIdentical: false,
	}

	score, matchedRules := scorer.Score(conflict)

	if len(matchedRules) != 1 || matchedRules[0] != "custom-rule" {
		t.Errorf("expected custom-rule to match, got %v", matchedRules)
	}

	// Base texture score (45) + custom rule bonus (50) = 95
	if score != 95 {
		t.Errorf("expected score 95, got %d", score)
	}
}

func TestScorer_matchPattern(t *testing.T) {
	scorer := NewScorer()

	tests := []struct {
		pattern   string
		value     string
		matchType RuleMatchType
		expected  bool
	}{
		// Exact matching
		{"test", "test", RuleMatchExact, true},
		{"test", "TEST", RuleMatchExact, true}, // Case insensitive
		{"test", "testing", RuleMatchExact, false},

		// Prefix matching
		{"scripts/", "scripts/skyui/test.pex", RuleMatchPrefix, true},
		{"scripts/", "interface/test.swf", RuleMatchPrefix, false},

		// Suffix matching
		{".hkx", "test.hkx", RuleMatchSuffix, true},
		{".hkx", "test.pex", RuleMatchSuffix, false},

		// Contains matching
		{"skeleton", "actors/character/skeleton.nif", RuleMatchContains, true},
		{"skeleton", "actors/character/body.nif", RuleMatchContains, false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.value, func(t *testing.T) {
			result := scorer.matchPattern(tt.pattern, tt.value, tt.matchType, nil)
			if result != tt.expected {
				t.Errorf("matchPattern(%q, %q, %v) = %v, want %v",
					tt.pattern, tt.value, tt.matchType, result, tt.expected)
			}
		})
	}
}

func TestScorer_matchPattern_Regex(t *testing.T) {
	scorer := NewScorer()

	rules := []*IncompatibilityRule{
		{
			ID:            "regex-test",
			Name:          "Regex Test",
			ScoreBonus:    10,
			PathPattern:   `\.esp$|\.esm$|\.esl$`,
			PathMatchType: RuleMatchRegex,
		},
	}
	customScorer := NewScorerWithRules(rules)

	tests := []struct {
		path     string
		expected bool
	}{
		{"plugin.esp", true},
		{"plugin.esm", true},
		{"plugin.esl", true},
		{"plugin.pex", false},
		{"esp/file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			conflict := &Conflict{
				Path:     tt.path,
				FileType: manifest.FileTypePlugin,
				Sources: []ModFile{
					{ModID: "mod1"},
					{ModID: "mod2"},
				},
			}

			_, matchedRules := customScorer.Score(conflict)
			hasMatch := len(matchedRules) > 0

			if hasMatch != tt.expected {
				t.Errorf("path %q: expected match=%v, got match=%v", tt.path, tt.expected, hasMatch)
			}
		})
	}

	// Test with default scorer (no regex)
	_ = scorer
}

func TestScorer_GetRules(t *testing.T) {
	scorer := NewScorer()
	rules := scorer.GetRules()

	if len(rules) == 0 {
		t.Error("expected default rules to be present")
	}

	// Check that essential rules are present
	ruleIDs := make(map[string]bool)
	for _, rule := range rules {
		ruleIDs[rule.ID] = true
	}

	expectedRules := []string{
		"skyui-scripts",
		"skeleton-conflict",
		"animation-behavior",
		"plugin-overwrite",
	}

	for _, id := range expectedRules {
		if !ruleIDs[id] {
			t.Errorf("expected rule %q to be present", id)
		}
	}
}

func TestScorer_FileTypeRestrictedRules(t *testing.T) {
	customRules := []*IncompatibilityRule{
		{
			ID:            "script-only",
			Name:          "Script Only Rule",
			ScoreBonus:    30,
			PathPattern:   "test",
			PathMatchType: RuleMatchContains,
			FileTypes:     []manifest.FileType{manifest.FileTypeScript},
		},
	}

	scorer := NewScorerWithRules(customRules)

	// Should match script file
	scriptConflict := &Conflict{
		Path:     "scripts/test.pex",
		FileType: manifest.FileTypeScript,
		Sources: []ModFile{
			{ModID: "mod1"},
			{ModID: "mod2"},
		},
	}

	_, matchedRules := scorer.Score(scriptConflict)
	if len(matchedRules) != 1 {
		t.Errorf("expected rule to match script file, got %v", matchedRules)
	}

	// Should not match texture file with same path pattern
	textureConflict := &Conflict{
		Path:     "textures/test.dds",
		FileType: manifest.FileTypeTexture,
		Sources: []ModFile{
			{ModID: "mod1"},
			{ModID: "mod2"},
		},
	}

	_, matchedRules = scorer.Score(textureConflict)
	if len(matchedRules) != 0 {
		t.Errorf("expected rule not to match texture file, got %v", matchedRules)
	}
}
