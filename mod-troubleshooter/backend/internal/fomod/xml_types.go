package fomod

import "encoding/xml"

// XML structures for parsing info.xml

type xmlInfo struct {
	XMLName     xml.Name `xml:"fomod"`
	Name        string   `xml:"Name"`
	Author      string   `xml:"Author"`
	Version     string   `xml:"Version"`
	Description string   `xml:"Description"`
	Website     string   `xml:"Website"`
	ID          string   `xml:"Id"`
}

// XML structures for parsing ModuleConfig.xml

type xmlConfig struct {
	XMLName                 xml.Name                    `xml:"config"`
	ModuleName              xmlModuleName               `xml:"moduleName"`
	ModuleImage             *xmlHeaderImage             `xml:"moduleImage"`
	ModuleDependencies      *xmlCompositeDependency     `xml:"moduleDependencies"`
	RequiredInstallFiles    *xmlFileList                `xml:"requiredInstallFiles"`
	InstallSteps            *xmlInstallSteps            `xml:"installSteps"`
	ConditionalFileInstalls *xmlConditionalFileInstalls `xml:"conditionalFileInstalls"`
}

type xmlModuleName struct {
	Value    string `xml:",chardata"`
	Position string `xml:"position,attr"`
	Colour   string `xml:"colour,attr"`
}

type xmlHeaderImage struct {
	Path     string `xml:"path,attr"`
	ShowFade string `xml:"showFade,attr"`
	Height   string `xml:"height,attr"`
}

type xmlInstallSteps struct {
	Order string           `xml:"order,attr"`
	Steps []xmlInstallStep `xml:"installStep"`
}

type xmlInstallStep struct {
	Name               string                  `xml:"name,attr"`
	Visible            *xmlCompositeDependency `xml:"visible>dependencies"`
	OptionalFileGroups *xmlOptionalFileGroups  `xml:"optionalFileGroups"`
}

type xmlOptionalFileGroups struct {
	Order  string     `xml:"order,attr"`
	Groups []xmlGroup `xml:"group"`
}

type xmlGroup struct {
	Name    string      `xml:"name,attr"`
	Type    string      `xml:"type,attr"`
	Plugins *xmlPlugins `xml:"plugins"`
}

type xmlPlugins struct {
	Order   string      `xml:"order,attr"`
	Plugins []xmlPlugin `xml:"plugin"`
}

type xmlPlugin struct {
	Name           string              `xml:"name,attr"`
	Description    string              `xml:"description"`
	Image          *xmlImage           `xml:"image"`
	Files          *xmlFileList        `xml:"files"`
	ConditionFlags *xmlConditionFlags  `xml:"conditionFlags"`
	TypeDescriptor *xmlTypeDescriptor  `xml:"typeDescriptor"`
}

type xmlImage struct {
	Path string `xml:"path,attr"`
}

type xmlTypeDescriptor struct {
	Type           *xmlPluginType           `xml:"type"`
	DependencyType *xmlDependencyPluginType `xml:"dependencyType"`
}

type xmlPluginType struct {
	Name string `xml:"name,attr"`
}

type xmlDependencyPluginType struct {
	DefaultType *xmlPluginType      `xml:"defaultType"`
	Patterns    *xmlPatternList     `xml:"patterns"`
}

type xmlPatternList struct {
	Patterns []xmlPattern `xml:"pattern"`
}

type xmlPattern struct {
	Dependencies *xmlCompositeDependency `xml:"dependencies"`
	Type         *xmlPluginType          `xml:"type"`
}

type xmlConditionFlags struct {
	Flags []xmlSetConditionFlag `xml:"flag"`
}

type xmlSetConditionFlag struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type xmlFileList struct {
	Files   []xmlFile   `xml:"file"`
	Folders []xmlFolder `xml:"folder"`
}

type xmlFile struct {
	Source          string `xml:"source,attr"`
	Destination     string `xml:"destination,attr"`
	Priority        string `xml:"priority,attr"`
	AlwaysInstall   string `xml:"alwaysInstall,attr"`
	InstallIfUsable string `xml:"installIfUsable,attr"`
}

type xmlFolder struct {
	Source          string `xml:"source,attr"`
	Destination     string `xml:"destination,attr"`
	Priority        string `xml:"priority,attr"`
	AlwaysInstall   string `xml:"alwaysInstall,attr"`
	InstallIfUsable string `xml:"installIfUsable,attr"`
}

type xmlCompositeDependency struct {
	Operator         string                    `xml:"operator,attr"`
	FileDependencies []xmlFileDependency       `xml:"fileDependency"`
	FlagDependencies []xmlFlagDependency       `xml:"flagDependency"`
	GameDependencies []xmlVersionDependency    `xml:"gameDependency"`
	FommDependencies []xmlVersionDependency    `xml:"fommDependency"`
	Dependencies     []xmlCompositeDependency  `xml:"dependencies"`
}

type xmlFileDependency struct {
	File  string `xml:"file,attr"`
	State string `xml:"state,attr"`
}

type xmlFlagDependency struct {
	Flag  string `xml:"flag,attr"`
	Value string `xml:"value,attr"`
}

type xmlVersionDependency struct {
	Version string `xml:"version,attr"`
}

type xmlConditionalFileInstalls struct {
	Patterns []xmlConditionalPattern `xml:"patterns>pattern"`
}

type xmlConditionalPattern struct {
	Dependencies *xmlCompositeDependency `xml:"dependencies"`
	Files        *xmlFileList            `xml:"files"`
}
