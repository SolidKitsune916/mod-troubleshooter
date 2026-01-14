package loadorder

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/mod-troubleshooter/backend/internal/plugin"
)

// Analyzer performs load order analysis on a set of plugins.
type Analyzer struct {
	parser *plugin.Parser
}

// NewAnalyzer creates a new load order analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		parser: plugin.NewParser(),
	}
}

// Analyze performs load order analysis on the given plugins.
// The plugins should be in their intended load order (index 0 loads first).
func (a *Analyzer) Analyze(ctx context.Context, plugins []PluginFile) (*AnalysisResult, error) {
	result := &AnalysisResult{
		Plugins:         make([]PluginInfo, 0, len(plugins)),
		Issues:          make([]Issue, 0),
		DependencyGraph: make(map[string][]string),
	}

	// Map of lowercase plugin filename to its index in the load order
	pluginIndex := make(map[string]int)
	// Map of lowercase plugin filename to its PluginInfo
	pluginInfoMap := make(map[string]*PluginInfo)

	// First pass: parse all plugins and build the index
	for i, pf := range plugins {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		info := PluginInfo{
			Filename: pf.Filename,
			Index:    i,
			Masters:  []string{},
		}

		// Use pre-parsed header if available
		if pf.Header != nil {
			info.Type = pf.Header.Type
			info.Flags = pf.Header.Flags
			info.Author = pf.Header.Author
			info.Description = pf.Header.Description
			for _, m := range pf.Header.Masters {
				info.Masters = append(info.Masters, m.Filename)
			}
		} else {
			// Determine type from filename extension if no header
			info.Type = determineTypeFromFilename(pf.Filename)
		}

		lowername := strings.ToLower(pf.Filename)
		pluginIndex[lowername] = i
		result.Plugins = append(result.Plugins, info)
		pluginInfoMap[lowername] = &result.Plugins[i]

		// Build dependency graph
		if len(info.Masters) > 0 {
			result.DependencyGraph[pf.Filename] = info.Masters
		}
	}

	// Second pass: detect issues
	for i := range result.Plugins {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		info := &result.Plugins[i]
		issues := a.detectIssuesForPlugin(info, pluginIndex)

		for _, issue := range issues {
			result.Issues = append(result.Issues, issue)
			info.HasIssues = true
			info.IssueCount++
		}
	}

	// Calculate stats
	result.Stats = a.calculateStats(result)

	return result, nil
}

// AnalyzeFromReaders parses plugins from readers and performs analysis.
// Each reader should provide the plugin file content.
func (a *Analyzer) AnalyzeFromReaders(ctx context.Context, files []struct {
	Filename string
	Reader   io.Reader
}) (*AnalysisResult, error) {
	plugins := make([]PluginFile, 0, len(files))

	for _, f := range files {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Only parse if it's a plugin file
		if !plugin.IsPluginFile(f.Filename) {
			continue
		}

		pf := PluginFile{
			Filename: f.Filename,
		}

		// Try to parse the plugin header
		if f.Reader != nil {
			header, err := a.parser.Parse(ctx, f.Reader, f.Filename)
			if err == nil {
				pf.Header = header
			}
			// If parsing fails, we'll still include the plugin with just the filename
		}

		plugins = append(plugins, pf)
	}

	return a.Analyze(ctx, plugins)
}

// AnalyzeFromHeaders performs analysis on pre-parsed plugin headers.
func (a *Analyzer) AnalyzeFromHeaders(ctx context.Context, headers []*plugin.PluginHeader) (*AnalysisResult, error) {
	plugins := make([]PluginFile, len(headers))
	for i, h := range headers {
		plugins[i] = PluginFile{
			Filename: h.Filename,
			Header:   h,
		}
	}
	return a.Analyze(ctx, plugins)
}

// detectIssuesForPlugin checks for issues with a single plugin.
func (a *Analyzer) detectIssuesForPlugin(info *PluginInfo, pluginIndex map[string]int) []Issue {
	var issues []Issue

	for _, master := range info.Masters {
		masterLower := strings.ToLower(master)
		masterIdx, exists := pluginIndex[masterLower]

		if !exists {
			// Missing master
			issues = append(issues, Issue{
				Type:          IssueMissingMaster,
				Severity:      SeverityError,
				Plugin:        info.Filename,
				RelatedPlugin: master,
				Message:       fmt.Sprintf("Missing required master: %s", master),
				Index:         info.Index,
			})
		} else if masterIdx > info.Index {
			// Master loads after this plugin (wrong order)
			issues = append(issues, Issue{
				Type:          IssueWrongOrder,
				Severity:      SeverityError,
				Plugin:        info.Filename,
				RelatedPlugin: master,
				Message:       fmt.Sprintf("Master %s loads after this plugin", master),
				Index:         info.Index,
			})
		}
	}

	return issues
}

// calculateStats computes summary statistics from the analysis result.
func (a *Analyzer) calculateStats(result *AnalysisResult) Stats {
	stats := Stats{
		TotalPlugins: len(result.Plugins),
		TotalIssues:  len(result.Issues),
	}

	pluginsWithIssues := make(map[string]bool)

	for _, p := range result.Plugins {
		switch p.Type {
		case plugin.PluginTypeESM:
			stats.ESMCount++
		case plugin.PluginTypeESP:
			stats.ESPCount++
		case plugin.PluginTypeESL:
			stats.ESLCount++
		}
	}

	for _, issue := range result.Issues {
		switch issue.Severity {
		case SeverityError:
			stats.ErrorCount++
		case SeverityWarning:
			stats.WarningCount++
		}

		switch issue.Type {
		case IssueMissingMaster:
			stats.MissingMasters++
		case IssueWrongOrder:
			stats.WrongOrderCount++
		}

		pluginsWithIssues[strings.ToLower(issue.Plugin)] = true
	}

	stats.PluginsWithIssues = len(pluginsWithIssues)

	return stats
}

// determineTypeFromFilename determines plugin type from file extension.
func determineTypeFromFilename(filename string) plugin.PluginType {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".esm"):
		return plugin.PluginTypeESM
	case strings.HasSuffix(lower, ".esl"):
		return plugin.PluginTypeESL
	default:
		return plugin.PluginTypeESP
	}
}
