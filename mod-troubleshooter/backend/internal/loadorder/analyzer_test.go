package loadorder

import (
	"context"
	"testing"

	"github.com/mod-troubleshooter/backend/internal/plugin"
)

func TestAnalyzer_Analyze_NoIssues(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()

	// Valid load order: masters load before their dependents
	plugins := []PluginFile{
		{
			Filename: "Skyrim.esm",
			Header: &plugin.PluginHeader{
				Filename: "Skyrim.esm",
				Type:     plugin.PluginTypeESM,
				Flags:    plugin.PluginFlags{IsMaster: true},
				Masters:  []plugin.Master{},
			},
		},
		{
			Filename: "Update.esm",
			Header: &plugin.PluginHeader{
				Filename: "Update.esm",
				Type:     plugin.PluginTypeESM,
				Flags:    plugin.PluginFlags{IsMaster: true},
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}},
			},
		},
		{
			Filename: "MyMod.esp",
			Header: &plugin.PluginHeader{
				Filename: "MyMod.esp",
				Type:     plugin.PluginTypeESP,
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}, {Filename: "Update.esm"}},
			},
		},
	}

	result, err := analyzer.Analyze(ctx, plugins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Issues) != 0 {
		t.Errorf("expected no issues, got %d: %+v", len(result.Issues), result.Issues)
	}

	if result.Stats.TotalPlugins != 3 {
		t.Errorf("expected 3 plugins, got %d", result.Stats.TotalPlugins)
	}

	if result.Stats.ESMCount != 2 {
		t.Errorf("expected 2 ESM, got %d", result.Stats.ESMCount)
	}

	if result.Stats.ESPCount != 1 {
		t.Errorf("expected 1 ESP, got %d", result.Stats.ESPCount)
	}
}

func TestAnalyzer_Analyze_MissingMaster(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()

	plugins := []PluginFile{
		{
			Filename: "Skyrim.esm",
			Header: &plugin.PluginHeader{
				Filename: "Skyrim.esm",
				Type:     plugin.PluginTypeESM,
				Flags:    plugin.PluginFlags{IsMaster: true},
				Masters:  []plugin.Master{},
			},
		},
		{
			Filename: "MyMod.esp",
			Header: &plugin.PluginHeader{
				Filename: "MyMod.esp",
				Type:     plugin.PluginTypeESP,
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}, {Filename: "MissingMod.esm"}},
			},
		},
	}

	result, err := analyzer.Analyze(ctx, plugins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(result.Issues))
	}

	issue := result.Issues[0]
	if issue.Type != IssueMissingMaster {
		t.Errorf("expected issue type %s, got %s", IssueMissingMaster, issue.Type)
	}
	if issue.Plugin != "MyMod.esp" {
		t.Errorf("expected plugin MyMod.esp, got %s", issue.Plugin)
	}
	if issue.RelatedPlugin != "MissingMod.esm" {
		t.Errorf("expected related plugin MissingMod.esm, got %s", issue.RelatedPlugin)
	}

	if result.Stats.MissingMasters != 1 {
		t.Errorf("expected 1 missing master, got %d", result.Stats.MissingMasters)
	}
}

func TestAnalyzer_Analyze_WrongOrder(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()

	// Wrong order: MyMod.esp loads before its master Update.esm
	plugins := []PluginFile{
		{
			Filename: "Skyrim.esm",
			Header: &plugin.PluginHeader{
				Filename: "Skyrim.esm",
				Type:     plugin.PluginTypeESM,
				Flags:    plugin.PluginFlags{IsMaster: true},
				Masters:  []plugin.Master{},
			},
		},
		{
			Filename: "MyMod.esp",
			Header: &plugin.PluginHeader{
				Filename: "MyMod.esp",
				Type:     plugin.PluginTypeESP,
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}, {Filename: "Update.esm"}},
			},
		},
		{
			Filename: "Update.esm",
			Header: &plugin.PluginHeader{
				Filename: "Update.esm",
				Type:     plugin.PluginTypeESM,
				Flags:    plugin.PluginFlags{IsMaster: true},
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}},
			},
		},
	}

	result, err := analyzer.Analyze(ctx, plugins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %+v", len(result.Issues), result.Issues)
	}

	issue := result.Issues[0]
	if issue.Type != IssueWrongOrder {
		t.Errorf("expected issue type %s, got %s", IssueWrongOrder, issue.Type)
	}
	if issue.Plugin != "MyMod.esp" {
		t.Errorf("expected plugin MyMod.esp, got %s", issue.Plugin)
	}
	if issue.RelatedPlugin != "Update.esm" {
		t.Errorf("expected related plugin Update.esm, got %s", issue.RelatedPlugin)
	}

	if result.Stats.WrongOrderCount != 1 {
		t.Errorf("expected 1 wrong order issue, got %d", result.Stats.WrongOrderCount)
	}
}

func TestAnalyzer_Analyze_CaseInsensitive(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()

	// Masters with different cases should still match
	plugins := []PluginFile{
		{
			Filename: "Skyrim.esm",
			Header: &plugin.PluginHeader{
				Filename: "Skyrim.esm",
				Type:     plugin.PluginTypeESM,
				Flags:    plugin.PluginFlags{IsMaster: true},
				Masters:  []plugin.Master{},
			},
		},
		{
			Filename: "MyMod.esp",
			Header: &plugin.PluginHeader{
				Filename: "MyMod.esp",
				Type:     plugin.PluginTypeESP,
				Masters:  []plugin.Master{{Filename: "SKYRIM.ESM"}}, // Different case
			},
		},
	}

	result, err := analyzer.Analyze(ctx, plugins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Issues) != 0 {
		t.Errorf("expected no issues (case insensitive match), got %d: %+v", len(result.Issues), result.Issues)
	}
}

func TestAnalyzer_Analyze_MultipleIssues(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()

	plugins := []PluginFile{
		{
			Filename: "Skyrim.esm",
			Header: &plugin.PluginHeader{
				Filename: "Skyrim.esm",
				Type:     plugin.PluginTypeESM,
				Flags:    plugin.PluginFlags{IsMaster: true},
				Masters:  []plugin.Master{},
			},
		},
		{
			Filename: "Mod1.esp",
			Header: &plugin.PluginHeader{
				Filename: "Mod1.esp",
				Type:     plugin.PluginTypeESP,
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}, {Filename: "Missing.esm"}},
			},
		},
		{
			Filename: "Mod2.esp",
			Header: &plugin.PluginHeader{
				Filename: "Mod2.esp",
				Type:     plugin.PluginTypeESP,
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}, {Filename: "Mod3.esp"}}, // Wrong order
			},
		},
		{
			Filename: "Mod3.esp",
			Header: &plugin.PluginHeader{
				Filename: "Mod3.esp",
				Type:     plugin.PluginTypeESP,
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}},
			},
		},
	}

	result, err := analyzer.Analyze(ctx, plugins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Issues) != 2 {
		t.Errorf("expected 2 issues, got %d: %+v", len(result.Issues), result.Issues)
	}

	if result.Stats.PluginsWithIssues != 2 {
		t.Errorf("expected 2 plugins with issues, got %d", result.Stats.PluginsWithIssues)
	}

	if result.Stats.ErrorCount != 2 {
		t.Errorf("expected 2 errors, got %d", result.Stats.ErrorCount)
	}
}

func TestAnalyzer_Analyze_DependencyGraph(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()

	plugins := []PluginFile{
		{
			Filename: "Skyrim.esm",
			Header: &plugin.PluginHeader{
				Filename: "Skyrim.esm",
				Type:     plugin.PluginTypeESM,
				Masters:  []plugin.Master{},
			},
		},
		{
			Filename: "Update.esm",
			Header: &plugin.PluginHeader{
				Filename: "Update.esm",
				Type:     plugin.PluginTypeESM,
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}},
			},
		},
		{
			Filename: "MyMod.esp",
			Header: &plugin.PluginHeader{
				Filename: "MyMod.esp",
				Type:     plugin.PluginTypeESP,
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}, {Filename: "Update.esm"}},
			},
		},
	}

	result, err := analyzer.Analyze(ctx, plugins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check dependency graph
	if len(result.DependencyGraph) != 2 {
		t.Errorf("expected 2 entries in dependency graph, got %d", len(result.DependencyGraph))
	}

	if masters, ok := result.DependencyGraph["Update.esm"]; ok {
		if len(masters) != 1 || masters[0] != "Skyrim.esm" {
			t.Errorf("unexpected masters for Update.esm: %v", masters)
		}
	} else {
		t.Error("expected Update.esm in dependency graph")
	}

	if masters, ok := result.DependencyGraph["MyMod.esp"]; ok {
		if len(masters) != 2 {
			t.Errorf("expected 2 masters for MyMod.esp, got %d", len(masters))
		}
	} else {
		t.Error("expected MyMod.esp in dependency graph")
	}
}

func TestAnalyzer_Analyze_ESLPlugins(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()

	plugins := []PluginFile{
		{
			Filename: "Skyrim.esm",
			Header: &plugin.PluginHeader{
				Filename: "Skyrim.esm",
				Type:     plugin.PluginTypeESM,
				Masters:  []plugin.Master{},
			},
		},
		{
			Filename: "LightMod.esl",
			Header: &plugin.PluginHeader{
				Filename: "LightMod.esl",
				Type:     plugin.PluginTypeESL,
				Flags:    plugin.PluginFlags{IsLight: true},
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}},
			},
		},
		{
			Filename: "FlaggedLight.esp",
			Header: &plugin.PluginHeader{
				Filename: "FlaggedLight.esp",
				Type:     plugin.PluginTypeESL, // Flagged as light
				Flags:    plugin.PluginFlags{IsLight: true},
				Masters:  []plugin.Master{{Filename: "Skyrim.esm"}},
			},
		},
	}

	result, err := analyzer.Analyze(ctx, plugins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Stats.ESLCount != 2 {
		t.Errorf("expected 2 ESL plugins, got %d", result.Stats.ESLCount)
	}

	if len(result.Issues) != 0 {
		t.Errorf("expected no issues, got %d", len(result.Issues))
	}
}

func TestAnalyzer_AnalyzeFromHeaders(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()

	headers := []*plugin.PluginHeader{
		{
			Filename: "Skyrim.esm",
			Type:     plugin.PluginTypeESM,
			Masters:  []plugin.Master{},
		},
		{
			Filename: "MyMod.esp",
			Type:     plugin.PluginTypeESP,
			Masters:  []plugin.Master{{Filename: "Skyrim.esm"}},
		},
	}

	result, err := analyzer.AnalyzeFromHeaders(ctx, headers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(result.Plugins))
	}

	if len(result.Issues) != 0 {
		t.Errorf("expected no issues, got %d", len(result.Issues))
	}
}

func TestDetermineTypeFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected plugin.PluginType
	}{
		{"Skyrim.esm", plugin.PluginTypeESM},
		{"SKYRIM.ESM", plugin.PluginTypeESM},
		{"MyMod.esp", plugin.PluginTypeESP},
		{"MyMod.ESP", plugin.PluginTypeESP},
		{"Light.esl", plugin.PluginTypeESL},
		{"LIGHT.ESL", plugin.PluginTypeESL},
		{"unknown.txt", plugin.PluginTypeESP}, // Default to ESP
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := determineTypeFromFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
