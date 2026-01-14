package plugin

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Common errors returned by the parser.
var (
	ErrInvalidPlugin    = errors.New("invalid plugin file")
	ErrNotPlugin        = errors.New("file is not a valid plugin")
	ErrTruncatedFile    = errors.New("plugin file is truncated")
	ErrUnsupportedGame  = errors.New("unsupported game version")
	ErrInvalidSignature = errors.New("invalid record signature")
)

// Parser reads and parses plugin file headers.
type Parser struct{}

// NewParser creates a new plugin parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile parses a plugin file from disk and returns its header information.
func (p *Parser) ParseFile(ctx context.Context, filePath string) (*PluginHeader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open plugin file: %w", err)
	}
	defer file.Close()

	filename := filepath.Base(filePath)
	return p.Parse(ctx, file, filename)
}

// Parse reads and parses a plugin header from the given reader.
// The filename is used for determining the plugin type if flags are ambiguous.
func (p *Parser) Parse(ctx context.Context, r io.Reader, filename string) (*PluginHeader, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	header := &PluginHeader{
		Filename: filename,
		Masters:  []Master{},
	}

	// Read the TES4 record header (24 bytes for Skyrim+, 20 bytes for older games)
	recordHeader, err := p.readRecordHeader(r)
	if err != nil {
		return nil, err
	}

	// Verify this is a TES4 header record
	if recordHeader.signature != SignatureTES4 {
		return nil, fmt.Errorf("%w: expected TES4, got %s", ErrInvalidSignature, recordHeader.signature)
	}

	// Parse flags
	header.Flags = PluginFlags{
		IsMaster:    recordHeader.flags&FlagMaster != 0,
		IsLight:     recordHeader.flags&FlagLight != 0,
		IsLocalized: recordHeader.flags&FlagLocalized != 0,
	}

	// Determine plugin type based on flags and extension
	header.Type = p.determinePluginType(header.Flags, filename)

	// Read the record data
	recordData := make([]byte, recordHeader.dataSize)
	if _, err := io.ReadFull(r, recordData); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTruncatedFile, err)
	}

	// Parse subrecords from the record data
	if err := p.parseSubrecords(recordData, header); err != nil {
		return nil, err
	}

	return header, nil
}

// recordHeader represents the header portion of a record.
type recordHeader struct {
	signature string
	dataSize  uint32
	flags     uint32
	formID    uint32
	timestamp uint32 // or version control info
	formVersion uint16
	unknown   uint16
}

// readRecordHeader reads the fixed-size record header.
func (p *Parser) readRecordHeader(r io.Reader) (*recordHeader, error) {
	// Record header layout (Skyrim+):
	// - 4 bytes: Type (signature)
	// - 4 bytes: Data size
	// - 4 bytes: Flags
	// - 4 bytes: Form ID
	// - 4 bytes: Timestamp/VC info
	// - 2 bytes: Form version
	// - 2 bytes: Unknown

	var buf [24]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, fmt.Errorf("%w: %v", ErrTruncatedFile, err)
		}
		return nil, fmt.Errorf("read record header: %w", err)
	}

	signature := string(buf[0:4])

	// Validate signature contains printable ASCII
	for _, c := range signature {
		if c < 32 || c > 126 {
			return nil, fmt.Errorf("%w: invalid characters in signature", ErrNotPlugin)
		}
	}

	return &recordHeader{
		signature:   signature,
		dataSize:    binary.LittleEndian.Uint32(buf[4:8]),
		flags:       binary.LittleEndian.Uint32(buf[8:12]),
		formID:      binary.LittleEndian.Uint32(buf[12:16]),
		timestamp:   binary.LittleEndian.Uint32(buf[16:20]),
		formVersion: binary.LittleEndian.Uint16(buf[20:22]),
		unknown:     binary.LittleEndian.Uint16(buf[22:24]),
	}, nil
}

// parseSubrecords parses the subrecords from the TES4 record data.
func (p *Parser) parseSubrecords(data []byte, header *PluginHeader) error {
	reader := bytes.NewReader(data)

	// Track current master for pairing with DATA subrecord
	var currentMaster *Master

	for reader.Len() > 0 {
		// Subrecord header: 4 bytes type + 2 bytes size
		var subHeader [6]byte
		if _, err := io.ReadFull(reader, subHeader[:]); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("read subrecord header: %w", err)
		}

		subType := string(subHeader[0:4])
		subSize := binary.LittleEndian.Uint16(subHeader[4:6])

		// Read subrecord data
		subData := make([]byte, subSize)
		if _, err := io.ReadFull(reader, subData); err != nil {
			return fmt.Errorf("read subrecord %s data: %w", subType, err)
		}

		switch subType {
		case SignatureHEDR:
			// HEDR is 12 bytes: float32 version, uint32 numRecords, uint32 nextObjectID
			if len(subData) >= 12 {
				header.NumRecords = binary.LittleEndian.Uint32(subData[4:8])
			}

		case SignatureCNAM:
			// Author string (null-terminated)
			header.Author = p.readNullString(subData)

		case SignatureSNAM:
			// Description string (null-terminated)
			header.Description = p.readNullString(subData)

		case SignatureMAST:
			// Master filename (null-terminated)
			masterName := p.readNullString(subData)
			if masterName != "" {
				currentMaster = &Master{Filename: masterName}
				header.Masters = append(header.Masters, *currentMaster)
			}

		case SignatureDATA:
			// Master file size (8 bytes, paired with preceding MAST)
			if len(subData) >= 8 && len(header.Masters) > 0 {
				size := binary.LittleEndian.Uint64(subData[0:8])
				header.Masters[len(header.Masters)-1].Size = size
			}
		}
	}

	return nil
}

// readNullString reads a null-terminated string from data.
func (p *Parser) readNullString(data []byte) string {
	// Find null terminator
	for i, b := range data {
		if b == 0 {
			return string(data[:i])
		}
	}
	// No null terminator, return entire string
	return string(data)
}

// determinePluginType determines the plugin type based on flags and file extension.
func (p *Parser) determinePluginType(flags PluginFlags, filename string) PluginType {
	ext := strings.ToLower(filepath.Ext(filename))

	// ESL flag takes precedence
	if flags.IsLight {
		return PluginTypeESL
	}

	// Check for ESM flag
	if flags.IsMaster {
		return PluginTypeESM
	}

	// Fall back to extension-based detection
	switch ext {
	case ".esm":
		return PluginTypeESM
	case ".esl":
		return PluginTypeESL
	default:
		return PluginTypeESP
	}
}

// IsPluginFile checks if the given filename has a plugin extension.
func IsPluginFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".esp", ".esm", ".esl":
		return true
	default:
		return false
	}
}
