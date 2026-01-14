package plugin

// PluginType represents the type of plugin file based on flags.
type PluginType string

const (
	// PluginTypeESM is an Elder Scrolls Master file.
	PluginTypeESM PluginType = "ESM"
	// PluginTypeESP is an Elder Scrolls Plugin file.
	PluginTypeESP PluginType = "ESP"
	// PluginTypeESL is an Elder Scrolls Light plugin file.
	PluginTypeESL PluginType = "ESL"
)

// PluginFlags contains the parsed flags from the plugin header.
type PluginFlags struct {
	// IsMaster indicates the plugin has the ESM flag set.
	IsMaster bool `json:"isMaster"`
	// IsLight indicates the plugin has the ESL/Light flag set.
	IsLight bool `json:"isLight"`
	// IsLocalized indicates the plugin uses localized strings.
	IsLocalized bool `json:"isLocalized"`
}

// Master represents a master file dependency.
type Master struct {
	// Filename is the name of the master file.
	Filename string `json:"filename"`
	// Size is the recorded size of the master file (may be 0).
	Size uint64 `json:"size,omitempty"`
}

// PluginHeader contains the parsed header information from a plugin file.
type PluginHeader struct {
	// Filename is the original filename of the plugin.
	Filename string `json:"filename"`
	// Type is the determined plugin type based on flags and extension.
	Type PluginType `json:"type"`
	// Flags contains the parsed plugin flags.
	Flags PluginFlags `json:"flags"`
	// Author is the author string from the CNAM subrecord.
	Author string `json:"author,omitempty"`
	// Description is the description from the SNAM subrecord.
	Description string `json:"description,omitempty"`
	// Masters is the list of master file dependencies in load order.
	Masters []Master `json:"masters"`
	// FormVersion is the form version from the header.
	FormVersion uint16 `json:"formVersion"`
	// NumRecords is the number of records in the file (if available).
	NumRecords uint32 `json:"numRecords,omitempty"`
}

// Record flag constants for the TES4 record.
const (
	// FlagMaster indicates the plugin is a master file (.esm behavior).
	FlagMaster uint32 = 0x00000001
	// FlagLocalized indicates the plugin uses localized strings.
	FlagLocalized uint32 = 0x00000080
	// FlagLight indicates the plugin is a light plugin (.esl behavior).
	// This flag was added in Skyrim Special Edition.
	FlagLight uint32 = 0x00000200
)

// Common TES4/5 record type signatures.
const (
	// SignatureTES4 is the header record signature for all plugin files.
	SignatureTES4 = "TES4"
	// SignatureHEDR is the header data subrecord.
	SignatureHEDR = "HEDR"
	// SignatureCNAM is the author subrecord.
	SignatureCNAM = "CNAM"
	// SignatureSNAM is the description subrecord.
	SignatureSNAM = "SNAM"
	// SignatureINTV is the internal version subrecord.
	SignatureINTV = "INTV"
	// SignatureMAST is the master file subrecord.
	SignatureMAST = "MAST"
	// SignatureDATA is the master file size subrecord.
	SignatureDATA = "DATA"
	// SignatureONAM is the overridden forms subrecord.
	SignatureONAM = "ONAM"
)
