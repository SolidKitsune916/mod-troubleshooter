package fomod

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
)

// Common errors returned by the parser.
var (
	ErrNoFomodDir        = errors.New("fomod directory not found")
	ErrNoModuleConfig    = errors.New("ModuleConfig.xml not found")
	ErrInvalidXML        = errors.New("invalid XML format")
	ErrMissingModuleName = errors.New("moduleName is required")
)

// Parser handles parsing FOMOD XML files.
type Parser struct {
	fomodDir string
}

// NewParser creates a new FOMOD parser for the given extracted directory.
// The directory should contain a fomod/ subdirectory with the XML files.
func NewParser(extractedDir string) (*Parser, error) {
	// Look for fomod directory (case-insensitive)
	fomodDir, err := findFomodDir(extractedDir)
	if err != nil {
		return nil, err
	}

	return &Parser{fomodDir: fomodDir}, nil
}

// findFomodDir locates the fomod directory within the extracted archive.
func findFomodDir(baseDir string) (string, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return "", fmt.Errorf("read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.EqualFold(entry.Name(), "fomod") {
			return filepath.Join(baseDir, entry.Name()), nil
		}
	}

	return "", ErrNoFomodDir
}

// Parse parses both info.xml and ModuleConfig.xml from the fomod directory.
func (p *Parser) Parse() (*FomodData, error) {
	config, err := p.ParseModuleConfig()
	if err != nil {
		return nil, err
	}

	info, err := p.ParseInfo()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return &FomodData{
		Info:   info,
		Config: config,
	}, nil
}

// ParseInfo parses the info.xml file if present.
func (p *Parser) ParseInfo() (*Info, error) {
	infoPath := p.findFile("info.xml")
	if infoPath == "" {
		return nil, os.ErrNotExist
	}

	data, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, fmt.Errorf("read info.xml: %w", err)
	}

	var xmlData xmlInfo
	if err := decodeXML(data, &xmlData); err != nil {
		return nil, fmt.Errorf("parse info.xml: %w", err)
	}

	return &Info{
		Name:        strings.TrimSpace(xmlData.Name),
		Author:      strings.TrimSpace(xmlData.Author),
		Version:     strings.TrimSpace(xmlData.Version),
		Description: strings.TrimSpace(xmlData.Description),
		Website:     strings.TrimSpace(xmlData.Website),
		ID:          strings.TrimSpace(xmlData.ID),
	}, nil
}

// ParseModuleConfig parses the ModuleConfig.xml file.
func (p *Parser) ParseModuleConfig() (*ModuleConfig, error) {
	configPath := p.findFile("ModuleConfig.xml")
	if configPath == "" {
		return nil, ErrNoModuleConfig
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read ModuleConfig.xml: %w", err)
	}

	var xmlData xmlConfig
	if err := decodeXML(data, &xmlData); err != nil {
		return nil, fmt.Errorf("parse ModuleConfig.xml: %w", err)
	}

	return convertConfig(&xmlData)
}

// findFile finds a file in the fomod directory (case-insensitive).
func (p *Parser) findFile(filename string) string {
	entries, err := os.ReadDir(p.fomodDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(entry.Name(), filename) {
			return filepath.Join(p.fomodDir, entry.Name())
		}
	}

	return ""
}

// decodeXML decodes XML data handling various encodings.
func decodeXML(data []byte, v interface{}) error {
	// Create a reader that can handle different character encodings
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidXML, err)
	}

	return nil
}

// convertConfig converts the XML config structure to the public API structure.
func convertConfig(xml *xmlConfig) (*ModuleConfig, error) {
	moduleName := strings.TrimSpace(xml.ModuleName.Value)
	if moduleName == "" {
		return nil, ErrMissingModuleName
	}

	config := &ModuleConfig{
		ModuleName: moduleName,
	}

	// Convert module image
	if xml.ModuleImage != nil {
		config.ModuleImage = convertHeaderImage(xml.ModuleImage)
	}

	// Convert module dependencies
	if xml.ModuleDependencies != nil {
		config.ModuleDependencies = convertDependency(xml.ModuleDependencies)
	}

	// Convert required install files
	if xml.RequiredInstallFiles != nil {
		config.RequiredInstallFiles = convertFileList(xml.RequiredInstallFiles)
	}

	// Convert install steps
	if xml.InstallSteps != nil {
		config.InstallSteps = convertInstallSteps(xml.InstallSteps)
	}

	// Convert conditional file installs
	if xml.ConditionalFileInstalls != nil {
		config.ConditionalFileInstalls = convertConditionalInstalls(xml.ConditionalFileInstalls)
	}

	return config, nil
}

func convertHeaderImage(xml *xmlHeaderImage) *HeaderImage {
	if xml == nil {
		return nil
	}

	img := &HeaderImage{
		Path:     xml.Path,
		ShowFade: parseBool(xml.ShowFade),
	}

	if xml.Height != "" {
		if height, err := strconv.Atoi(xml.Height); err == nil {
			img.Height = height
		}
	}

	return img
}

func convertInstallSteps(xml *xmlInstallSteps) []InstallStep {
	if xml == nil || len(xml.Steps) == 0 {
		return nil
	}

	steps := make([]InstallStep, 0, len(xml.Steps))
	for _, xmlStep := range xml.Steps {
		step := InstallStep{
			Name: xmlStep.Name,
		}

		if xmlStep.Visible != nil {
			step.Visible = convertDependency(xmlStep.Visible)
		}

		if xmlStep.OptionalFileGroups != nil {
			step.OptionGroups = convertGroups(xmlStep.OptionalFileGroups)
		}

		steps = append(steps, step)
	}

	return steps
}

func convertGroups(xml *xmlOptionalFileGroups) []OptionGroup {
	if xml == nil || len(xml.Groups) == 0 {
		return nil
	}

	groups := make([]OptionGroup, 0, len(xml.Groups))
	for _, xmlGroup := range xml.Groups {
		group := OptionGroup{
			Name: xmlGroup.Name,
			Type: GroupType(xmlGroup.Type),
		}

		if xmlGroup.Plugins != nil {
			group.Plugins = convertPlugins(xmlGroup.Plugins)
		}

		groups = append(groups, group)
	}

	return groups
}

func convertPlugins(xml *xmlPlugins) []Plugin {
	if xml == nil || len(xml.Plugins) == 0 {
		return nil
	}

	plugins := make([]Plugin, 0, len(xml.Plugins))
	for _, xmlPlugin := range xml.Plugins {
		plugin := Plugin{
			Name:        xmlPlugin.Name,
			Description: strings.TrimSpace(xmlPlugin.Description),
		}

		if xmlPlugin.Image != nil {
			plugin.Image = xmlPlugin.Image.Path
		}

		if xmlPlugin.Files != nil {
			plugin.Files = convertFileList(xmlPlugin.Files)
		}

		if xmlPlugin.ConditionFlags != nil {
			plugin.ConditionFlags = convertConditionFlags(xmlPlugin.ConditionFlags)
		}

		if xmlPlugin.TypeDescriptor != nil {
			plugin.TypeDescriptor = convertTypeDescriptor(xmlPlugin.TypeDescriptor)
		}

		plugins = append(plugins, plugin)
	}

	return plugins
}

func convertTypeDescriptor(xml *xmlTypeDescriptor) *TypeDescriptor {
	if xml == nil {
		return nil
	}

	td := &TypeDescriptor{}

	if xml.Type != nil {
		td.Type = PluginType(xml.Type.Name)
	}

	if xml.DependencyType != nil {
		td.DependencyType = convertDependencyPluginType(xml.DependencyType)
	}

	return td
}

func convertDependencyPluginType(xml *xmlDependencyPluginType) *DependencyPluginType {
	if xml == nil {
		return nil
	}

	dpt := &DependencyPluginType{}

	if xml.DefaultType != nil {
		dpt.DefaultType = PluginType(xml.DefaultType.Name)
	}

	if xml.Patterns != nil && len(xml.Patterns.Patterns) > 0 {
		dpt.Patterns = make([]DependencyPattern, 0, len(xml.Patterns.Patterns))
		for _, xmlPattern := range xml.Patterns.Patterns {
			pattern := DependencyPattern{}

			if xmlPattern.Dependencies != nil {
				pattern.Dependencies = convertDependency(xmlPattern.Dependencies)
			}

			if xmlPattern.Type != nil {
				pattern.Type = PluginType(xmlPattern.Type.Name)
			}

			dpt.Patterns = append(dpt.Patterns, pattern)
		}
	}

	return dpt
}

func convertConditionFlags(xml *xmlConditionFlags) []ConditionFlag {
	if xml == nil || len(xml.Flags) == 0 {
		return nil
	}

	flags := make([]ConditionFlag, 0, len(xml.Flags))
	for _, xmlFlag := range xml.Flags {
		flags = append(flags, ConditionFlag{
			Name:  xmlFlag.Name,
			Value: strings.TrimSpace(xmlFlag.Value),
		})
	}

	return flags
}

func convertFileList(xml *xmlFileList) *FileList {
	if xml == nil {
		return nil
	}

	fl := &FileList{}

	if len(xml.Files) > 0 {
		fl.Files = make([]FileInstall, 0, len(xml.Files))
		for _, xmlFile := range xml.Files {
			fl.Files = append(fl.Files, FileInstall{
				Source:          xmlFile.Source,
				Destination:     xmlFile.Destination,
				Priority:        parseInt(xmlFile.Priority),
				AlwaysInstall:   parseBool(xmlFile.AlwaysInstall),
				InstallIfUsable: parseBool(xmlFile.InstallIfUsable),
			})
		}
	}

	if len(xml.Folders) > 0 {
		fl.Folders = make([]FolderInstall, 0, len(xml.Folders))
		for _, xmlFolder := range xml.Folders {
			fl.Folders = append(fl.Folders, FolderInstall{
				Source:          xmlFolder.Source,
				Destination:     xmlFolder.Destination,
				Priority:        parseInt(xmlFolder.Priority),
				AlwaysInstall:   parseBool(xmlFolder.AlwaysInstall),
				InstallIfUsable: parseBool(xmlFolder.InstallIfUsable),
			})
		}
	}

	return fl
}

func convertDependency(xml *xmlCompositeDependency) *Dependency {
	if xml == nil {
		return nil
	}

	dep := &Dependency{}

	// Set operator if present
	if xml.Operator != "" {
		dep.Operator = DependencyOperator(xml.Operator)
	}

	// Convert file dependencies
	for _, xmlFD := range xml.FileDependencies {
		if dep.FileDependency == nil && len(xml.FileDependencies) == 1 &&
			len(xml.FlagDependencies) == 0 && len(xml.GameDependencies) == 0 &&
			len(xml.FommDependencies) == 0 && len(xml.Dependencies) == 0 {
			// Single file dependency - set directly
			dep.FileDependency = &FileDependency{
				File:  xmlFD.File,
				State: FileState(xmlFD.State),
			}
		} else {
			// Multiple dependencies - add as children
			dep.Children = append(dep.Children, Dependency{
				FileDependency: &FileDependency{
					File:  xmlFD.File,
					State: FileState(xmlFD.State),
				},
			})
		}
	}

	// Convert flag dependencies
	for _, xmlFlagD := range xml.FlagDependencies {
		if dep.FlagDependency == nil && len(xml.FileDependencies) == 0 &&
			len(xml.FlagDependencies) == 1 && len(xml.GameDependencies) == 0 &&
			len(xml.FommDependencies) == 0 && len(xml.Dependencies) == 0 {
			// Single flag dependency - set directly
			dep.FlagDependency = &FlagDependency{
				Flag:  xmlFlagD.Flag,
				Value: xmlFlagD.Value,
			}
		} else {
			// Multiple dependencies - add as children
			dep.Children = append(dep.Children, Dependency{
				FlagDependency: &FlagDependency{
					Flag:  xmlFlagD.Flag,
					Value: xmlFlagD.Value,
				},
			})
		}
	}

	// Convert game dependencies
	for _, xmlGameD := range xml.GameDependencies {
		dep.Children = append(dep.Children, Dependency{
			GameDependency: &VersionDependency{
				Version: xmlGameD.Version,
			},
		})
	}

	// Convert FOMM dependencies
	for _, xmlFommD := range xml.FommDependencies {
		dep.Children = append(dep.Children, Dependency{
			FommDependency: &VersionDependency{
				Version: xmlFommD.Version,
			},
		})
	}

	// Convert nested composite dependencies
	for _, xmlNestedDep := range xml.Dependencies {
		nestedDep := convertDependency(&xmlNestedDep)
		if nestedDep != nil {
			dep.Children = append(dep.Children, *nestedDep)
		}
	}

	return dep
}

func convertConditionalInstalls(xml *xmlConditionalFileInstalls) []ConditionalInstallItem {
	if xml == nil || len(xml.Patterns) == 0 {
		return nil
	}

	items := make([]ConditionalInstallItem, 0, len(xml.Patterns))
	for _, xmlPattern := range xml.Patterns {
		item := ConditionalInstallItem{}

		if xmlPattern.Dependencies != nil {
			item.Dependencies = convertDependency(xmlPattern.Dependencies)
		}

		if xmlPattern.Files != nil {
			item.Files = convertFileList(xmlPattern.Files)
		}

		items = append(items, item)
	}

	return items
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}

// ParseFromReader parses ModuleConfig.xml from an io.Reader.
// This is useful for testing or when the XML is available in memory.
func ParseModuleConfigFromReader(r io.Reader) (*ModuleConfig, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}

	var xmlData xmlConfig
	if err := decodeXML(data, &xmlData); err != nil {
		return nil, fmt.Errorf("parse ModuleConfig.xml: %w", err)
	}

	return convertConfig(&xmlData)
}

// ParseInfoFromReader parses info.xml from an io.Reader.
func ParseInfoFromReader(r io.Reader) (*Info, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}

	var xmlData xmlInfo
	if err := decodeXML(data, &xmlData); err != nil {
		return nil, fmt.Errorf("parse info.xml: %w", err)
	}

	return &Info{
		Name:        strings.TrimSpace(xmlData.Name),
		Author:      strings.TrimSpace(xmlData.Author),
		Version:     strings.TrimSpace(xmlData.Version),
		Description: strings.TrimSpace(xmlData.Description),
		Website:     strings.TrimSpace(xmlData.Website),
		ID:          strings.TrimSpace(xmlData.ID),
	}, nil
}
