package conflict

import (
	"context"
	"testing"

	"github.com/mod-troubleshooter/backend/internal/manifest"
)

func TestAnalyzer_Analyze_NoConflicts(t *testing.T) {
	analyzer := NewAnalyzer()

	mods := []ModManifest{
		{
			ModID:     "mod1",
			ModName:   "Mod One",
			LoadOrder: 0,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "textures/texture1.dds", Size: 1000, Type: manifest.FileTypeTexture},
				},
			},
		},
		{
			ModID:     "mod2",
			ModName:   "Mod Two",
			LoadOrder: 1,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "textures/texture2.dds", Size: 2000, Type: manifest.FileTypeTexture},
				},
			},
		},
	}

	result, err := analyzer.Analyze(context.Background(), mods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d", len(result.Conflicts))
	}

	if result.Stats.TotalConflicts != 0 {
		t.Errorf("expected TotalConflicts=0, got %d", result.Stats.TotalConflicts)
	}

	if result.Stats.UniqueFiles != 2 {
		t.Errorf("expected UniqueFiles=2, got %d", result.Stats.UniqueFiles)
	}

	if result.Stats.TotalFiles != 2 {
		t.Errorf("expected TotalFiles=2, got %d", result.Stats.TotalFiles)
	}
}

func TestAnalyzer_Analyze_SimpleConflict(t *testing.T) {
	analyzer := NewAnalyzer()

	mods := []ModManifest{
		{
			ModID:     "mod1",
			ModName:   "Mod One",
			LoadOrder: 0,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "textures/shared.dds", Size: 1000, Type: manifest.FileTypeTexture},
				},
			},
		},
		{
			ModID:     "mod2",
			ModName:   "Mod Two",
			LoadOrder: 1,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "textures/shared.dds", Size: 2000, Type: manifest.FileTypeTexture},
				},
			},
		},
	}

	result, err := analyzer.Analyze(context.Background(), mods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
	}

	conflict := result.Conflicts[0]
	if conflict.Path != "textures/shared.dds" {
		t.Errorf("expected path 'textures/shared.dds', got %q", conflict.Path)
	}

	if conflict.Winner == nil {
		t.Fatal("expected winner to be set")
	}

	if conflict.Winner.ModID != "mod2" {
		t.Errorf("expected winner to be mod2, got %q", conflict.Winner.ModID)
	}

	if len(conflict.Losers) != 1 || conflict.Losers[0].ModID != "mod1" {
		t.Errorf("expected loser to be mod1, got %v", conflict.Losers)
	}

	if conflict.Type != ConflictTypeOverwrite {
		t.Errorf("expected type Overwrite, got %q", conflict.Type)
	}

	if conflict.FileType != manifest.FileTypeTexture {
		t.Errorf("expected fileType Texture, got %q", conflict.FileType)
	}
}

func TestAnalyzer_Analyze_ThreeWayConflict(t *testing.T) {
	analyzer := NewAnalyzer()

	mods := []ModManifest{
		{
			ModID:     "mod1",
			ModName:   "Mod One",
			LoadOrder: 0,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "meshes/shared.nif", Size: 1000, Type: manifest.FileTypeMesh},
				},
			},
		},
		{
			ModID:     "mod2",
			ModName:   "Mod Two",
			LoadOrder: 1,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "meshes/shared.nif", Size: 2000, Type: manifest.FileTypeMesh},
				},
			},
		},
		{
			ModID:     "mod3",
			ModName:   "Mod Three",
			LoadOrder: 2,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "meshes/shared.nif", Size: 3000, Type: manifest.FileTypeMesh},
				},
			},
		},
	}

	result, err := analyzer.Analyze(context.Background(), mods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
	}

	conflict := result.Conflicts[0]
	if len(conflict.Sources) != 3 {
		t.Errorf("expected 3 sources, got %d", len(conflict.Sources))
	}

	if conflict.Winner.ModID != "mod3" {
		t.Errorf("expected winner to be mod3, got %q", conflict.Winner.ModID)
	}

	if len(conflict.Losers) != 2 {
		t.Errorf("expected 2 losers, got %d", len(conflict.Losers))
	}
}

func TestAnalyzer_Analyze_IdenticalFiles(t *testing.T) {
	analyzer := NewAnalyzer()

	mods := []ModManifest{
		{
			ModID:     "mod1",
			ModName:   "Mod One",
			LoadOrder: 0,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "textures/shared.dds", Size: 1000, Hash: "abc123", Type: manifest.FileTypeTexture},
				},
			},
		},
		{
			ModID:     "mod2",
			ModName:   "Mod Two",
			LoadOrder: 1,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "textures/shared.dds", Size: 1000, Hash: "abc123", Type: manifest.FileTypeTexture},
				},
			},
		},
	}

	result, err := analyzer.Analyze(context.Background(), mods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
	}

	conflict := result.Conflicts[0]
	if !conflict.IsIdentical {
		t.Error("expected IsIdentical to be true")
	}

	if conflict.Type != ConflictTypeDuplicate {
		t.Errorf("expected type Duplicate, got %q", conflict.Type)
	}

	if conflict.Severity != SeverityInfo {
		t.Errorf("expected severity Info for identical files, got %q", conflict.Severity)
	}

	if result.Stats.IdenticalConflicts != 1 {
		t.Errorf("expected IdenticalConflicts=1, got %d", result.Stats.IdenticalConflicts)
	}
}

func TestAnalyzer_Analyze_PluginConflictIsCritical(t *testing.T) {
	analyzer := NewAnalyzer()

	mods := []ModManifest{
		{
			ModID:     "mod1",
			ModName:   "Mod One",
			LoadOrder: 0,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "plugin.esp", Size: 1000, Type: manifest.FileTypePlugin},
				},
			},
		},
		{
			ModID:     "mod2",
			ModName:   "Mod Two",
			LoadOrder: 1,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "plugin.esp", Size: 2000, Type: manifest.FileTypePlugin},
				},
			},
		},
	}

	result, err := analyzer.Analyze(context.Background(), mods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
	}

	conflict := result.Conflicts[0]
	if conflict.Severity != SeverityCritical {
		t.Errorf("expected severity Critical for plugin conflict, got %q", conflict.Severity)
	}

	if result.Stats.CriticalCount != 1 {
		t.Errorf("expected CriticalCount=1, got %d", result.Stats.CriticalCount)
	}
}

func TestAnalyzer_Analyze_SeverityByFileType(t *testing.T) {
	tests := []struct {
		fileType         manifest.FileType
		extension        string
		expectedSeverity Severity
	}{
		{manifest.FileTypePlugin, ".esp", SeverityCritical},
		{manifest.FileTypeBSA, ".bsa", SeverityHigh},
		{manifest.FileTypeScript, ".pex", SeverityHigh},
		{manifest.FileTypeMesh, ".nif", SeverityMedium},
		{manifest.FileTypeTexture, ".dds", SeverityMedium},
		{manifest.FileTypeInterface, ".swf", SeverityMedium},
		{manifest.FileTypeSound, ".wav", SeverityLow},
		{manifest.FileTypeSEQ, ".seq", SeverityLow},
		{manifest.FileTypeOther, ".txt", SeverityLow},
	}

	analyzer := NewAnalyzer()

	for _, tt := range tests {
		t.Run(string(tt.fileType), func(t *testing.T) {
			mods := []ModManifest{
				{
					ModID:     "mod1",
					ModName:   "Mod One",
					LoadOrder: 0,
					Manifest: &manifest.Manifest{
						Files: []manifest.FileEntry{
							{Path: "file" + tt.extension, Size: 1000, Type: tt.fileType},
						},
					},
				},
				{
					ModID:     "mod2",
					ModName:   "Mod Two",
					LoadOrder: 1,
					Manifest: &manifest.Manifest{
						Files: []manifest.FileEntry{
							{Path: "file" + tt.extension, Size: 2000, Type: tt.fileType},
						},
					},
				},
			}

			result, err := analyzer.Analyze(context.Background(), mods)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Conflicts) != 1 {
				t.Fatalf("expected 1 conflict, got %d", len(result.Conflicts))
			}

			if result.Conflicts[0].Severity != tt.expectedSeverity {
				t.Errorf("expected severity %q for %s, got %q", tt.expectedSeverity, tt.fileType, result.Conflicts[0].Severity)
			}
		})
	}
}

func TestAnalyzer_Analyze_ModSummaries(t *testing.T) {
	analyzer := NewAnalyzer()

	mods := []ModManifest{
		{
			ModID:     "mod1",
			ModName:   "Mod One",
			LoadOrder: 0,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "textures/a.dds", Size: 1000, Type: manifest.FileTypeTexture},
					{Path: "textures/b.dds", Size: 1000, Type: manifest.FileTypeTexture},
				},
			},
		},
		{
			ModID:     "mod2",
			ModName:   "Mod Two",
			LoadOrder: 1,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "textures/a.dds", Size: 2000, Type: manifest.FileTypeTexture},
					{Path: "textures/c.dds", Size: 2000, Type: manifest.FileTypeTexture},
				},
			},
		},
		{
			ModID:     "mod3",
			ModName:   "Mod Three",
			LoadOrder: 2,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "textures/a.dds", Size: 3000, Type: manifest.FileTypeTexture},
					{Path: "textures/b.dds", Size: 3000, Type: manifest.FileTypeTexture},
				},
			},
		},
	}

	result, err := analyzer.Analyze(context.Background(), mods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 2 conflicts: a.dds (3 mods), b.dds (2 mods)
	if len(result.Conflicts) != 2 {
		t.Fatalf("expected 2 conflicts, got %d", len(result.Conflicts))
	}

	if len(result.ModSummaries) != 3 {
		t.Fatalf("expected 3 mod summaries, got %d", len(result.ModSummaries))
	}

	// Check mod3 wins both conflicts
	var mod3Summary *ModConflictSummary
	for i := range result.ModSummaries {
		if result.ModSummaries[i].ModID == "mod3" {
			mod3Summary = &result.ModSummaries[i]
			break
		}
	}

	if mod3Summary == nil {
		t.Fatal("mod3 summary not found")
	}

	if mod3Summary.WinCount != 2 {
		t.Errorf("expected mod3 WinCount=2, got %d", mod3Summary.WinCount)
	}

	if mod3Summary.LoseCount != 0 {
		t.Errorf("expected mod3 LoseCount=0, got %d", mod3Summary.LoseCount)
	}

	// Check mod1 loses both conflicts
	var mod1Summary *ModConflictSummary
	for i := range result.ModSummaries {
		if result.ModSummaries[i].ModID == "mod1" {
			mod1Summary = &result.ModSummaries[i]
			break
		}
	}

	if mod1Summary == nil {
		t.Fatal("mod1 summary not found")
	}

	if mod1Summary.LoseCount != 2 {
		t.Errorf("expected mod1 LoseCount=2, got %d", mod1Summary.LoseCount)
	}
}

func TestAnalyzer_Analyze_FileToMods(t *testing.T) {
	analyzer := NewAnalyzer()

	mods := []ModManifest{
		{
			ModID:     "mod1",
			ModName:   "Mod One",
			LoadOrder: 0,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "shared.dds", Size: 1000, Type: manifest.FileTypeTexture},
				},
			},
		},
		{
			ModID:     "mod2",
			ModName:   "Mod Two",
			LoadOrder: 1,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "shared.dds", Size: 2000, Type: manifest.FileTypeTexture},
				},
			},
		},
	}

	result, err := analyzer.Analyze(context.Background(), mods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	modIDs, ok := result.FileToMods["shared.dds"]
	if !ok {
		t.Fatal("expected shared.dds in FileToMods")
	}

	if len(modIDs) != 2 {
		t.Errorf("expected 2 mod IDs, got %d", len(modIDs))
	}

	// Should be sorted by load order
	if modIDs[0] != "mod1" || modIDs[1] != "mod2" {
		t.Errorf("unexpected mod IDs order: %v", modIDs)
	}
}

func TestAnalyzer_Analyze_EmptyManifest(t *testing.T) {
	analyzer := NewAnalyzer()

	mods := []ModManifest{
		{
			ModID:     "mod1",
			ModName:   "Mod One",
			LoadOrder: 0,
			Manifest:  nil, // No manifest
		},
		{
			ModID:     "mod2",
			ModName:   "Mod Two",
			LoadOrder: 1,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{},
			},
		},
	}

	result, err := analyzer.Analyze(context.Background(), mods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d", len(result.Conflicts))
	}

	if result.Stats.TotalFiles != 0 {
		t.Errorf("expected TotalFiles=0, got %d", result.Stats.TotalFiles)
	}
}

func TestAnalyzer_Analyze_ContextCancellation(t *testing.T) {
	analyzer := NewAnalyzer()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mods := []ModManifest{
		{
			ModID:     "mod1",
			ModName:   "Mod One",
			LoadOrder: 0,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "textures/shared.dds", Size: 1000, Type: manifest.FileTypeTexture},
				},
			},
		},
		{
			ModID:     "mod2",
			ModName:   "Mod Two",
			LoadOrder: 1,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "textures/shared.dds", Size: 2000, Type: manifest.FileTypeTexture},
				},
			},
		},
	}

	_, err := analyzer.Analyze(ctx, mods)
	if err == nil {
		t.Error("expected context cancellation error")
	}
}

func TestAnalyzer_Analyze_ConflictsSortedBySeverity(t *testing.T) {
	analyzer := NewAnalyzer()

	mods := []ModManifest{
		{
			ModID:     "mod1",
			ModName:   "Mod One",
			LoadOrder: 0,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "low.txt", Size: 1000, Type: manifest.FileTypeOther},
					{Path: "plugin.esp", Size: 1000, Type: manifest.FileTypePlugin},
					{Path: "texture.dds", Size: 1000, Type: manifest.FileTypeTexture},
				},
			},
		},
		{
			ModID:     "mod2",
			ModName:   "Mod Two",
			LoadOrder: 1,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "low.txt", Size: 2000, Type: manifest.FileTypeOther},
					{Path: "plugin.esp", Size: 2000, Type: manifest.FileTypePlugin},
					{Path: "texture.dds", Size: 2000, Type: manifest.FileTypeTexture},
				},
			},
		},
	}

	result, err := analyzer.Analyze(context.Background(), mods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Conflicts) != 3 {
		t.Fatalf("expected 3 conflicts, got %d", len(result.Conflicts))
	}

	// First should be critical (plugin)
	if result.Conflicts[0].Severity != SeverityCritical {
		t.Errorf("expected first conflict to be Critical, got %q", result.Conflicts[0].Severity)
	}

	// Second should be medium (texture)
	if result.Conflicts[1].Severity != SeverityMedium {
		t.Errorf("expected second conflict to be Medium, got %q", result.Conflicts[1].Severity)
	}

	// Third should be low (other)
	if result.Conflicts[2].Severity != SeverityLow {
		t.Errorf("expected third conflict to be Low, got %q", result.Conflicts[2].Severity)
	}
}

func TestAnalyzer_Analyze_StatsCount(t *testing.T) {
	analyzer := NewAnalyzer()

	mods := []ModManifest{
		{
			ModID:     "mod1",
			ModName:   "Mod One",
			LoadOrder: 0,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "plugin.esp", Size: 1000, Type: manifest.FileTypePlugin},
					{Path: "archive.bsa", Size: 1000, Type: manifest.FileTypeBSA},
					{Path: "texture.dds", Size: 1000, Type: manifest.FileTypeTexture},
					{Path: "sound.wav", Size: 1000, Type: manifest.FileTypeSound},
				},
			},
		},
		{
			ModID:     "mod2",
			ModName:   "Mod Two",
			LoadOrder: 1,
			Manifest: &manifest.Manifest{
				Files: []manifest.FileEntry{
					{Path: "plugin.esp", Size: 2000, Type: manifest.FileTypePlugin},
					{Path: "archive.bsa", Size: 2000, Type: manifest.FileTypeBSA},
					{Path: "texture.dds", Size: 2000, Type: manifest.FileTypeTexture},
					{Path: "sound.wav", Size: 2000, Type: manifest.FileTypeSound},
				},
			},
		},
	}

	result, err := analyzer.Analyze(context.Background(), mods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats := result.Stats

	if stats.TotalConflicts != 4 {
		t.Errorf("expected TotalConflicts=4, got %d", stats.TotalConflicts)
	}

	if stats.CriticalCount != 1 {
		t.Errorf("expected CriticalCount=1, got %d", stats.CriticalCount)
	}

	if stats.HighCount != 1 {
		t.Errorf("expected HighCount=1, got %d", stats.HighCount)
	}

	if stats.MediumCount != 1 {
		t.Errorf("expected MediumCount=1, got %d", stats.MediumCount)
	}

	if stats.LowCount != 1 {
		t.Errorf("expected LowCount=1, got %d", stats.LowCount)
	}

	if stats.ModsAnalyzed != 2 {
		t.Errorf("expected ModsAnalyzed=2, got %d", stats.ModsAnalyzed)
	}

	if stats.ModsWithConflicts != 2 {
		t.Errorf("expected ModsWithConflicts=2, got %d", stats.ModsWithConflicts)
	}

	// Check ByFileType counts
	if stats.ByFileType[manifest.FileTypePlugin] != 1 {
		t.Errorf("expected ByFileType[plugin]=1, got %d", stats.ByFileType[manifest.FileTypePlugin])
	}

	if stats.ByFileType[manifest.FileTypeBSA] != 1 {
		t.Errorf("expected ByFileType[bsa]=1, got %d", stats.ByFileType[manifest.FileTypeBSA])
	}
}
