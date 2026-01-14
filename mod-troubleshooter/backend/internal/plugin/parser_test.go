package plugin

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
)

// createTestPlugin creates a minimal valid plugin file in memory for testing.
func createTestPlugin(t *testing.T, opts testPluginOptions) []byte {
	t.Helper()

	var buf bytes.Buffer

	// Build the TES4 record data (subrecords)
	var recordData bytes.Buffer

	// HEDR subrecord (12 bytes: version float, numRecords uint32, nextObjectID uint32)
	writeSubrecord(&recordData, SignatureHEDR, []byte{
		0x9A, 0x99, 0xD9, 0x3F, // version 1.7 as float32
		byte(opts.numRecords), byte(opts.numRecords >> 8), byte(opts.numRecords >> 16), byte(opts.numRecords >> 24),
		0x01, 0x00, 0x00, 0x00, // nextObjectID
	})

	// CNAM subrecord (author)
	if opts.author != "" {
		writeSubrecord(&recordData, SignatureCNAM, append([]byte(opts.author), 0))
	}

	// SNAM subrecord (description)
	if opts.description != "" {
		writeSubrecord(&recordData, SignatureSNAM, append([]byte(opts.description), 0))
	}

	// MAST and DATA subrecords for masters
	for _, master := range opts.masters {
		writeSubrecord(&recordData, SignatureMAST, append([]byte(master.Filename), 0))
		// DATA subrecord (8 bytes for size)
		var sizeData [8]byte
		binary.LittleEndian.PutUint64(sizeData[:], master.Size)
		writeSubrecord(&recordData, SignatureDATA, sizeData[:])
	}

	recordBytes := recordData.Bytes()

	// TES4 record header (24 bytes)
	// Type (4 bytes)
	buf.WriteString(SignatureTES4)
	// Data size (4 bytes)
	binary.Write(&buf, binary.LittleEndian, uint32(len(recordBytes)))
	// Flags (4 bytes)
	binary.Write(&buf, binary.LittleEndian, opts.flags)
	// Form ID (4 bytes)
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	// Timestamp (4 bytes)
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	// Form version (2 bytes)
	binary.Write(&buf, binary.LittleEndian, uint16(44)) // Skyrim SE form version
	// Unknown (2 bytes)
	binary.Write(&buf, binary.LittleEndian, uint16(0))

	// Write record data
	buf.Write(recordBytes)

	return buf.Bytes()
}

type testPluginOptions struct {
	flags       uint32
	numRecords  uint32
	author      string
	description string
	masters     []Master
}

func writeSubrecord(buf *bytes.Buffer, signature string, data []byte) {
	buf.WriteString(signature)
	binary.Write(buf, binary.LittleEndian, uint16(len(data)))
	buf.Write(data)
}

func TestParser_Parse_ESP(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	data := createTestPlugin(t, testPluginOptions{
		flags:       0,
		numRecords:  100,
		author:      "Test Author",
		description: "Test Description",
		masters:     []Master{{Filename: "Skyrim.esm", Size: 12345}},
	})

	header, err := parser.Parse(ctx, bytes.NewReader(data), "test.esp")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if header.Type != PluginTypeESP {
		t.Errorf("expected type ESP, got %s", header.Type)
	}

	if header.Author != "Test Author" {
		t.Errorf("expected author 'Test Author', got '%s'", header.Author)
	}

	if header.Description != "Test Description" {
		t.Errorf("expected description 'Test Description', got '%s'", header.Description)
	}

	if len(header.Masters) != 1 {
		t.Fatalf("expected 1 master, got %d", len(header.Masters))
	}

	if header.Masters[0].Filename != "Skyrim.esm" {
		t.Errorf("expected master 'Skyrim.esm', got '%s'", header.Masters[0].Filename)
	}

	if header.NumRecords != 100 {
		t.Errorf("expected 100 records, got %d", header.NumRecords)
	}
}

func TestParser_Parse_ESM(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	data := createTestPlugin(t, testPluginOptions{
		flags: FlagMaster,
	})

	header, err := parser.Parse(ctx, bytes.NewReader(data), "test.esm")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if header.Type != PluginTypeESM {
		t.Errorf("expected type ESM, got %s", header.Type)
	}

	if !header.Flags.IsMaster {
		t.Error("expected IsMaster flag to be true")
	}
}

func TestParser_Parse_ESL(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	// ESL with both Master and Light flags
	data := createTestPlugin(t, testPluginOptions{
		flags: FlagMaster | FlagLight,
	})

	header, err := parser.Parse(ctx, bytes.NewReader(data), "test.esl")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if header.Type != PluginTypeESL {
		t.Errorf("expected type ESL, got %s", header.Type)
	}

	if !header.Flags.IsLight {
		t.Error("expected IsLight flag to be true")
	}
}

func TestParser_Parse_LocalizedPlugin(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	data := createTestPlugin(t, testPluginOptions{
		flags: FlagLocalized,
	})

	header, err := parser.Parse(ctx, bytes.NewReader(data), "test.esp")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !header.Flags.IsLocalized {
		t.Error("expected IsLocalized flag to be true")
	}
}

func TestParser_Parse_MultipleMasters(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	masters := []Master{
		{Filename: "Skyrim.esm", Size: 100000},
		{Filename: "Update.esm", Size: 200000},
		{Filename: "Dawnguard.esm", Size: 300000},
	}

	data := createTestPlugin(t, testPluginOptions{
		masters: masters,
	})

	header, err := parser.Parse(ctx, bytes.NewReader(data), "test.esp")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(header.Masters) != 3 {
		t.Fatalf("expected 3 masters, got %d", len(header.Masters))
	}

	for i, m := range masters {
		if header.Masters[i].Filename != m.Filename {
			t.Errorf("master %d: expected filename '%s', got '%s'", i, m.Filename, header.Masters[i].Filename)
		}
		if header.Masters[i].Size != m.Size {
			t.Errorf("master %d: expected size %d, got %d", i, m.Size, header.Masters[i].Size)
		}
	}
}

func TestParser_Parse_NoMasters(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	data := createTestPlugin(t, testPluginOptions{
		flags: FlagMaster, // Base game ESM has no masters
	})

	header, err := parser.Parse(ctx, bytes.NewReader(data), "Skyrim.esm")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(header.Masters) != 0 {
		t.Errorf("expected 0 masters, got %d", len(header.Masters))
	}
}

func TestParser_Parse_InvalidSignature(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	// Create invalid data with wrong signature
	data := []byte("XXXX" + string(make([]byte, 20)))

	_, err := parser.Parse(ctx, bytes.NewReader(data), "test.esp")
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestParser_Parse_TruncatedFile(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	// Only 10 bytes, not enough for header
	data := []byte("TES4" + string(make([]byte, 6)))

	_, err := parser.Parse(ctx, bytes.NewReader(data), "test.esp")
	if err == nil {
		t.Error("expected error for truncated file")
	}
}

func TestParser_Parse_ContextCancellation(t *testing.T) {
	parser := NewParser()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	data := createTestPlugin(t, testPluginOptions{})

	_, err := parser.Parse(ctx, bytes.NewReader(data), "test.esp")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestParser_DeterminePluginType(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		flags    PluginFlags
		filename string
		expected PluginType
	}{
		{
			name:     "ESP by extension",
			flags:    PluginFlags{},
			filename: "mod.esp",
			expected: PluginTypeESP,
		},
		{
			name:     "ESM by extension",
			flags:    PluginFlags{},
			filename: "mod.esm",
			expected: PluginTypeESM,
		},
		{
			name:     "ESL by extension",
			flags:    PluginFlags{},
			filename: "mod.esl",
			expected: PluginTypeESL,
		},
		{
			name:     "ESM by flag overrides ESP extension",
			flags:    PluginFlags{IsMaster: true},
			filename: "mod.esp",
			expected: PluginTypeESM,
		},
		{
			name:     "ESL by flag overrides ESP extension",
			flags:    PluginFlags{IsLight: true},
			filename: "mod.esp",
			expected: PluginTypeESL,
		},
		{
			name:     "ESL flag takes precedence over ESM flag",
			flags:    PluginFlags{IsMaster: true, IsLight: true},
			filename: "mod.esm",
			expected: PluginTypeESL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.determinePluginType(tt.flags, tt.filename)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestIsPluginFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"mod.esp", true},
		{"mod.esm", true},
		{"mod.esl", true},
		{"MOD.ESP", true},
		{"Skyrim.ESM", true},
		{"mod.bsa", false},
		{"mod.txt", false},
		{"mod", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := IsPluginFile(tt.filename)
			if result != tt.expected {
				t.Errorf("IsPluginFile(%q) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}
