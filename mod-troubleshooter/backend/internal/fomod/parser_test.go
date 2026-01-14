package fomod

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseInfo(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		want    *Info
		wantErr bool
	}{
		{
			name: "complete info.xml",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<fomod>
  <Name>Test Mod</Name>
  <Author>Test Author</Author>
  <Version>1.0.0</Version>
  <Description>A test mod description</Description>
  <Website>https://example.com</Website>
  <Id>12345</Id>
</fomod>`,
			want: &Info{
				Name:        "Test Mod",
				Author:      "Test Author",
				Version:     "1.0.0",
				Description: "A test mod description",
				Website:     "https://example.com",
				ID:          "12345",
			},
		},
		{
			name: "partial info.xml",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<fomod>
  <Name>Minimal Mod</Name>
  <Author>Someone</Author>
</fomod>`,
			want: &Info{
				Name:   "Minimal Mod",
				Author: "Someone",
			},
		},
		{
			name: "empty info.xml",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<fomod>
</fomod>`,
			want: &Info{},
		},
		{
			name:    "invalid xml",
			xml:     `not valid xml`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInfoFromReader(strings.NewReader(tt.xml))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInfoFromReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Author != tt.want.Author {
				t.Errorf("Author = %q, want %q", got.Author, tt.want.Author)
			}
			if got.Version != tt.want.Version {
				t.Errorf("Version = %q, want %q", got.Version, tt.want.Version)
			}
			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if got.Website != tt.want.Website {
				t.Errorf("Website = %q, want %q", got.Website, tt.want.Website)
			}
			if got.ID != tt.want.ID {
				t.Errorf("ID = %q, want %q", got.ID, tt.want.ID)
			}
		})
	}
}

func TestParseModuleConfig(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		check   func(*testing.T, *ModuleConfig)
		wantErr bool
	}{
		{
			name: "basic module config",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
</config>`,
			check: func(t *testing.T, c *ModuleConfig) {
				if c.ModuleName != "Test Module" {
					t.Errorf("ModuleName = %q, want %q", c.ModuleName, "Test Module")
				}
			},
		},
		{
			name: "install steps with groups",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
  <installSteps order="Explicit">
    <installStep name="Step 1">
      <optionalFileGroups order="Explicit">
        <group name="Options" type="SelectExactlyOne">
          <plugins order="Explicit">
            <plugin name="Option A">
              <description>Description A</description>
              <typeDescriptor>
                <type name="Recommended"/>
              </typeDescriptor>
            </plugin>
            <plugin name="Option B">
              <description>Description B</description>
              <typeDescriptor>
                <type name="Optional"/>
              </typeDescriptor>
            </plugin>
          </plugins>
        </group>
      </optionalFileGroups>
    </installStep>
  </installSteps>
</config>`,
			check: func(t *testing.T, c *ModuleConfig) {
				if len(c.InstallSteps) != 1 {
					t.Fatalf("InstallSteps count = %d, want 1", len(c.InstallSteps))
				}

				step := c.InstallSteps[0]
				if step.Name != "Step 1" {
					t.Errorf("Step name = %q, want %q", step.Name, "Step 1")
				}

				if len(step.OptionGroups) != 1 {
					t.Fatalf("OptionGroups count = %d, want 1", len(step.OptionGroups))
				}

				group := step.OptionGroups[0]
				if group.Name != "Options" {
					t.Errorf("Group name = %q, want %q", group.Name, "Options")
				}
				if group.Type != GroupSelectExactlyOne {
					t.Errorf("Group type = %q, want %q", group.Type, GroupSelectExactlyOne)
				}

				if len(group.Plugins) != 2 {
					t.Fatalf("Plugins count = %d, want 2", len(group.Plugins))
				}

				pluginA := group.Plugins[0]
				if pluginA.Name != "Option A" {
					t.Errorf("Plugin A name = %q, want %q", pluginA.Name, "Option A")
				}
				if pluginA.Description != "Description A" {
					t.Errorf("Plugin A description = %q, want %q", pluginA.Description, "Description A")
				}
				if pluginA.TypeDescriptor == nil || pluginA.TypeDescriptor.Type != PluginRecommended {
					t.Errorf("Plugin A type = %v, want %q", pluginA.TypeDescriptor, PluginRecommended)
				}

				pluginB := group.Plugins[1]
				if pluginB.TypeDescriptor == nil || pluginB.TypeDescriptor.Type != PluginOptional {
					t.Errorf("Plugin B type = %v, want %q", pluginB.TypeDescriptor, PluginOptional)
				}
			},
		},
		{
			name: "condition flags",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
  <installSteps order="Explicit">
    <installStep name="Step 1">
      <optionalFileGroups order="Explicit">
        <group name="Options" type="SelectExactlyOne">
          <plugins order="Explicit">
            <plugin name="Option A">
              <description>Sets a flag</description>
              <conditionFlags>
                <flag name="optionA">On</flag>
              </conditionFlags>
              <typeDescriptor>
                <type name="Optional"/>
              </typeDescriptor>
            </plugin>
          </plugins>
        </group>
      </optionalFileGroups>
    </installStep>
  </installSteps>
</config>`,
			check: func(t *testing.T, c *ModuleConfig) {
				plugin := c.InstallSteps[0].OptionGroups[0].Plugins[0]
				if len(plugin.ConditionFlags) != 1 {
					t.Fatalf("ConditionFlags count = %d, want 1", len(plugin.ConditionFlags))
				}
				flag := plugin.ConditionFlags[0]
				if flag.Name != "optionA" {
					t.Errorf("Flag name = %q, want %q", flag.Name, "optionA")
				}
				if flag.Value != "On" {
					t.Errorf("Flag value = %q, want %q", flag.Value, "On")
				}
			},
		},
		{
			name: "dependency type descriptor",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
  <installSteps order="Explicit">
    <installStep name="Step 1">
      <optionalFileGroups order="Explicit">
        <group name="Options" type="SelectAny">
          <plugins order="Explicit">
            <plugin name="Conditional Plugin">
              <description>Conditional type</description>
              <typeDescriptor>
                <dependencyType>
                  <defaultType name="NotUsable"/>
                  <patterns>
                    <pattern>
                      <dependencies>
                        <flagDependency flag="someFlag" value="On"/>
                      </dependencies>
                      <type name="Required"/>
                    </pattern>
                  </patterns>
                </dependencyType>
              </typeDescriptor>
            </plugin>
          </plugins>
        </group>
      </optionalFileGroups>
    </installStep>
  </installSteps>
</config>`,
			check: func(t *testing.T, c *ModuleConfig) {
				plugin := c.InstallSteps[0].OptionGroups[0].Plugins[0]
				if plugin.TypeDescriptor == nil {
					t.Fatal("TypeDescriptor is nil")
				}
				if plugin.TypeDescriptor.DependencyType == nil {
					t.Fatal("DependencyType is nil")
				}

				dt := plugin.TypeDescriptor.DependencyType
				if dt.DefaultType != PluginNotUsable {
					t.Errorf("DefaultType = %q, want %q", dt.DefaultType, PluginNotUsable)
				}
				if len(dt.Patterns) != 1 {
					t.Fatalf("Patterns count = %d, want 1", len(dt.Patterns))
				}

				pattern := dt.Patterns[0]
				if pattern.Type != PluginRequired {
					t.Errorf("Pattern type = %q, want %q", pattern.Type, PluginRequired)
				}
				if pattern.Dependencies == nil || pattern.Dependencies.FlagDependency == nil {
					t.Fatal("Pattern dependencies or FlagDependency is nil")
				}
				if pattern.Dependencies.FlagDependency.Flag != "someFlag" {
					t.Errorf("Flag dependency flag = %q, want %q", pattern.Dependencies.FlagDependency.Flag, "someFlag")
				}
				if pattern.Dependencies.FlagDependency.Value != "On" {
					t.Errorf("Flag dependency value = %q, want %q", pattern.Dependencies.FlagDependency.Value, "On")
				}
			},
		},
		{
			name: "required install files",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
  <requiredInstallFiles>
    <file source="required.esp" destination="Data"/>
    <folder source="textures" destination="Data/textures" priority="1"/>
  </requiredInstallFiles>
</config>`,
			check: func(t *testing.T, c *ModuleConfig) {
				if c.RequiredInstallFiles == nil {
					t.Fatal("RequiredInstallFiles is nil")
				}
				if len(c.RequiredInstallFiles.Files) != 1 {
					t.Fatalf("Files count = %d, want 1", len(c.RequiredInstallFiles.Files))
				}
				if len(c.RequiredInstallFiles.Folders) != 1 {
					t.Fatalf("Folders count = %d, want 1", len(c.RequiredInstallFiles.Folders))
				}

				file := c.RequiredInstallFiles.Files[0]
				if file.Source != "required.esp" {
					t.Errorf("File source = %q, want %q", file.Source, "required.esp")
				}
				if file.Destination != "Data" {
					t.Errorf("File destination = %q, want %q", file.Destination, "Data")
				}

				folder := c.RequiredInstallFiles.Folders[0]
				if folder.Source != "textures" {
					t.Errorf("Folder source = %q, want %q", folder.Source, "textures")
				}
				if folder.Priority != 1 {
					t.Errorf("Folder priority = %d, want 1", folder.Priority)
				}
			},
		},
		{
			name: "plugin with files",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
  <installSteps order="Explicit">
    <installStep name="Step 1">
      <optionalFileGroups order="Explicit">
        <group name="Options" type="SelectExactlyOne">
          <plugins order="Explicit">
            <plugin name="Option A">
              <description>Has files</description>
              <files>
                <file source="optionA.esp"/>
                <folder source="optionA_textures" destination="Data/textures"/>
              </files>
              <typeDescriptor>
                <type name="Optional"/>
              </typeDescriptor>
            </plugin>
          </plugins>
        </group>
      </optionalFileGroups>
    </installStep>
  </installSteps>
</config>`,
			check: func(t *testing.T, c *ModuleConfig) {
				plugin := c.InstallSteps[0].OptionGroups[0].Plugins[0]
				if plugin.Files == nil {
					t.Fatal("Plugin files is nil")
				}
				if len(plugin.Files.Files) != 1 {
					t.Fatalf("Plugin files count = %d, want 1", len(plugin.Files.Files))
				}
				if len(plugin.Files.Folders) != 1 {
					t.Fatalf("Plugin folders count = %d, want 1", len(plugin.Files.Folders))
				}

				file := plugin.Files.Files[0]
				if file.Source != "optionA.esp" {
					t.Errorf("File source = %q, want %q", file.Source, "optionA.esp")
				}
			},
		},
		{
			name: "all group types",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
  <installSteps order="Explicit">
    <installStep name="Step 1">
      <optionalFileGroups order="Explicit">
        <group name="G1" type="SelectExactlyOne"><plugins/></group>
        <group name="G2" type="SelectAtMostOne"><plugins/></group>
        <group name="G3" type="SelectAtLeastOne"><plugins/></group>
        <group name="G4" type="SelectAny"><plugins/></group>
        <group name="G5" type="SelectAll"><plugins/></group>
      </optionalFileGroups>
    </installStep>
  </installSteps>
</config>`,
			check: func(t *testing.T, c *ModuleConfig) {
				groups := c.InstallSteps[0].OptionGroups
				if len(groups) != 5 {
					t.Fatalf("Groups count = %d, want 5", len(groups))
				}

				expectedTypes := []GroupType{
					GroupSelectExactlyOne,
					GroupSelectAtMostOne,
					GroupSelectAtLeastOne,
					GroupSelectAny,
					GroupSelectAll,
				}

				for i, expected := range expectedTypes {
					if groups[i].Type != expected {
						t.Errorf("Group %d type = %q, want %q", i, groups[i].Type, expected)
					}
				}
			},
		},
		{
			name: "step visibility with composite dependency",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
  <installSteps order="Explicit">
    <installStep name="Conditional Step">
      <visible>
        <dependencies operator="And">
          <flagDependency flag="flag1" value="On"/>
          <flagDependency flag="flag2" value="On"/>
        </dependencies>
      </visible>
      <optionalFileGroups/>
    </installStep>
  </installSteps>
</config>`,
			check: func(t *testing.T, c *ModuleConfig) {
				step := c.InstallSteps[0]
				if step.Visible == nil {
					t.Fatal("Step visible is nil")
				}
				if step.Visible.Operator != DependencyOperatorAnd {
					t.Errorf("Visible operator = %q, want %q", step.Visible.Operator, DependencyOperatorAnd)
				}
				if len(step.Visible.Children) != 2 {
					t.Fatalf("Visible children count = %d, want 2", len(step.Visible.Children))
				}
			},
		},
		{
			name:    "missing module name",
			xml:     `<?xml version="1.0" encoding="UTF-8"?><config><moduleName></moduleName></config>`,
			wantErr: true,
		},
		{
			name:    "invalid xml",
			xml:     `not valid xml at all`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseModuleConfigFromReader(strings.NewReader(tt.xml))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseModuleConfigFromReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestNewParser(t *testing.T) {
	// Create a temp directory structure
	tmpDir := t.TempDir()

	t.Run("finds fomod directory", func(t *testing.T) {
		// Create fomod subdirectory
		fomodDir := filepath.Join(tmpDir, "fomod")
		if err := os.MkdirAll(fomodDir, 0755); err != nil {
			t.Fatal(err)
		}

		parser, err := NewParser(tmpDir)
		if err != nil {
			t.Fatalf("NewParser() error = %v", err)
		}
		if parser == nil {
			t.Fatal("NewParser() returned nil")
		}
	})

	t.Run("finds FOMOD directory (uppercase)", func(t *testing.T) {
		tmpDir2 := t.TempDir()
		fomodDir := filepath.Join(tmpDir2, "FOMOD")
		if err := os.MkdirAll(fomodDir, 0755); err != nil {
			t.Fatal(err)
		}

		parser, err := NewParser(tmpDir2)
		if err != nil {
			t.Fatalf("NewParser() error = %v", err)
		}
		if parser == nil {
			t.Fatal("NewParser() returned nil")
		}
	})

	t.Run("returns error when fomod directory not found", func(t *testing.T) {
		tmpDir3 := t.TempDir()
		_, err := NewParser(tmpDir3)
		if err == nil {
			t.Fatal("NewParser() should return error when fomod not found")
		}
		if err != ErrNoFomodDir {
			t.Errorf("NewParser() error = %v, want %v", err, ErrNoFomodDir)
		}
	})
}

func TestParser_Parse(t *testing.T) {
	// Create a full test structure
	tmpDir := t.TempDir()
	fomodDir := filepath.Join(tmpDir, "fomod")
	if err := os.MkdirAll(fomodDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write info.xml
	infoContent := `<?xml version="1.0" encoding="UTF-8"?>
<fomod>
  <Name>Test Mod</Name>
  <Author>Test Author</Author>
</fomod>`
	if err := os.WriteFile(filepath.Join(fomodDir, "info.xml"), []byte(infoContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write ModuleConfig.xml
	configContent := `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
  <installSteps order="Explicit">
    <installStep name="Choose Options">
      <optionalFileGroups order="Explicit">
        <group name="Main Options" type="SelectExactlyOne">
          <plugins order="Explicit">
            <plugin name="Option 1">
              <description>First option</description>
              <typeDescriptor>
                <type name="Recommended"/>
              </typeDescriptor>
            </plugin>
          </plugins>
        </group>
      </optionalFileGroups>
    </installStep>
  </installSteps>
</config>`
	if err := os.WriteFile(filepath.Join(fomodDir, "ModuleConfig.xml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(tmpDir)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	data, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check info
	if data.Info == nil {
		t.Fatal("Info is nil")
	}
	if data.Info.Name != "Test Mod" {
		t.Errorf("Info.Name = %q, want %q", data.Info.Name, "Test Mod")
	}
	if data.Info.Author != "Test Author" {
		t.Errorf("Info.Author = %q, want %q", data.Info.Author, "Test Author")
	}

	// Check config
	if data.Config == nil {
		t.Fatal("Config is nil")
	}
	if data.Config.ModuleName != "Test Module" {
		t.Errorf("Config.ModuleName = %q, want %q", data.Config.ModuleName, "Test Module")
	}
	if len(data.Config.InstallSteps) != 1 {
		t.Fatalf("InstallSteps count = %d, want 1", len(data.Config.InstallSteps))
	}
}

func TestParser_ParseWithoutInfo(t *testing.T) {
	tmpDir := t.TempDir()
	fomodDir := filepath.Join(tmpDir, "fomod")
	if err := os.MkdirAll(fomodDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Only write ModuleConfig.xml (no info.xml)
	configContent := `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
</config>`
	if err := os.WriteFile(filepath.Join(fomodDir, "ModuleConfig.xml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(tmpDir)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	data, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Info should be nil when info.xml doesn't exist
	if data.Info != nil {
		t.Error("Info should be nil when info.xml doesn't exist")
	}

	// Config should still be parsed
	if data.Config == nil {
		t.Fatal("Config is nil")
	}
	if data.Config.ModuleName != "Test Module" {
		t.Errorf("Config.ModuleName = %q, want %q", data.Config.ModuleName, "Test Module")
	}
}

func TestParser_CaseInsensitiveFilenames(t *testing.T) {
	tmpDir := t.TempDir()
	fomodDir := filepath.Join(tmpDir, "FoMoD") // Mixed case
	if err := os.MkdirAll(fomodDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write MODULECONFIG.XML (uppercase)
	configContent := `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
</config>`
	if err := os.WriteFile(filepath.Join(fomodDir, "MODULECONFIG.XML"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write INFO.XML (uppercase)
	infoContent := `<?xml version="1.0" encoding="UTF-8"?>
<fomod>
  <Name>Test</Name>
</fomod>`
	if err := os.WriteFile(filepath.Join(fomodDir, "INFO.XML"), []byte(infoContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := NewParser(tmpDir)
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	data, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if data.Config == nil || data.Config.ModuleName != "Test Module" {
		t.Error("Failed to parse uppercase ModuleConfig.xml")
	}
	if data.Info == nil || data.Info.Name != "Test" {
		t.Error("Failed to parse uppercase info.xml")
	}
}

func TestConditionalFileInstalls(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<config xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://qconsulting.ca/fo3/ModConfig5.0.xsd">
  <moduleName>Test Module</moduleName>
  <conditionalFileInstalls>
    <patterns>
      <pattern>
        <dependencies>
          <flagDependency flag="optionA" value="On"/>
        </dependencies>
        <files>
          <file source="optionA.esp"/>
        </files>
      </pattern>
      <pattern>
        <dependencies operator="And">
          <flagDependency flag="optionB" value="On"/>
          <flagDependency flag="optionC" value="On"/>
        </dependencies>
        <files>
          <folder source="bc_textures"/>
        </files>
      </pattern>
    </patterns>
  </conditionalFileInstalls>
</config>`

	config, err := ParseModuleConfigFromReader(strings.NewReader(xml))
	if err != nil {
		t.Fatalf("ParseModuleConfigFromReader() error = %v", err)
	}

	if len(config.ConditionalFileInstalls) != 2 {
		t.Fatalf("ConditionalFileInstalls count = %d, want 2", len(config.ConditionalFileInstalls))
	}

	// Check first pattern
	pattern1 := config.ConditionalFileInstalls[0]
	if pattern1.Dependencies == nil || pattern1.Dependencies.FlagDependency == nil {
		t.Fatal("Pattern 1 dependencies are nil")
	}
	if pattern1.Dependencies.FlagDependency.Flag != "optionA" {
		t.Errorf("Pattern 1 flag = %q, want %q", pattern1.Dependencies.FlagDependency.Flag, "optionA")
	}
	if pattern1.Files == nil || len(pattern1.Files.Files) != 1 {
		t.Fatal("Pattern 1 files are incorrect")
	}

	// Check second pattern (composite dependency)
	pattern2 := config.ConditionalFileInstalls[1]
	if pattern2.Dependencies == nil || pattern2.Dependencies.Operator != DependencyOperatorAnd {
		t.Errorf("Pattern 2 operator = %q, want %q", pattern2.Dependencies.Operator, DependencyOperatorAnd)
	}
	if len(pattern2.Dependencies.Children) != 2 {
		t.Fatalf("Pattern 2 children count = %d, want 2", len(pattern2.Dependencies.Children))
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("parseBool", func(t *testing.T) {
		tests := []struct {
			input string
			want  bool
		}{
			{"true", true},
			{"True", true},
			{"TRUE", true},
			{"1", true},
			{"yes", true},
			{"Yes", true},
			{"false", false},
			{"False", false},
			{"0", false},
			{"no", false},
			{"", false},
			{"  true  ", true},
		}

		for _, tt := range tests {
			if got := parseBool(tt.input); got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	})

	t.Run("parseInt", func(t *testing.T) {
		tests := []struct {
			input string
			want  int
		}{
			{"0", 0},
			{"1", 1},
			{"42", 42},
			{"-1", -1},
			{"", 0},
			{"  5  ", 5},
			{"invalid", 0},
		}

		for _, tt := range tests {
			if got := parseInt(tt.input); got != tt.want {
				t.Errorf("parseInt(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	})
}
