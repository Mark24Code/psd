package psd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// TypeToolInfo contains text layer information
type TypeToolInfo struct {
	Version    uint16
	Transform  Transform
	TextData   map[string]interface{}
	WarpData   map[string]interface{}
	Left       int32
	Top        int32
	Right      int32
	Bottom     int32
	EngineData string
}

// Transform represents the transformation matrix
type Transform struct {
	XX float64
	XY float64
	YX float64
	YY float64
	TX float64
	TY float64
}

// Text returns the text content
func (t *TypeToolInfo) Text() string {
	if t.TextData == nil {
		return ""
	}

	// Try to get text from 'Txt ' key
	if txtValue, ok := t.TextData["Txt "].(string); ok {
		return txtValue
	}

	return ""
}

// Fonts returns the list of fonts (from engine data if available)
func (t *TypeToolInfo) Fonts() []string {
	// This would require full engine data parsing
	// For now, return empty array
	return []string{}
}

// Sizes returns the list of font sizes
func (t *TypeToolInfo) Sizes() []float64 {
	// This would require engine data parsing
	return []float64{}
}

// Colors returns the list of colors as [R, G, B, A] arrays
func (t *TypeToolInfo) Colors() [][]uint8 {
	// This would require engine data parsing
	// Return default black
	return [][]uint8{{0, 0, 0, 255}}
}

// ParseTypeTool parses TypeTool data from a layer info block
func ParseTypeTool(data []byte) (*TypeToolInfo, error) {
	reader := bytes.NewReader(data)
	info := &TypeToolInfo{}

	// Read version
	if err := binary.Read(reader, binary.BigEndian, &info.Version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	// Read transform matrix (6 doubles)
	if err := binary.Read(reader, binary.BigEndian, &info.Transform.XX); err != nil {
		return nil, fmt.Errorf("failed to read transform XX: %w", err)
	}
	if err := binary.Read(reader, binary.BigEndian, &info.Transform.XY); err != nil {
		return nil, fmt.Errorf("failed to read transform XY: %w", err)
	}
	if err := binary.Read(reader, binary.BigEndian, &info.Transform.YX); err != nil {
		return nil, fmt.Errorf("failed to read transform YX: %w", err)
	}
	if err := binary.Read(reader, binary.BigEndian, &info.Transform.YY); err != nil {
		return nil, fmt.Errorf("failed to read transform YY: %w", err)
	}
	if err := binary.Read(reader, binary.BigEndian, &info.Transform.TX); err != nil {
		return nil, fmt.Errorf("failed to read transform TX: %w", err)
	}
	if err := binary.Read(reader, binary.BigEndian, &info.Transform.TY); err != nil {
		return nil, fmt.Errorf("failed to read transform TY: %w", err)
	}

	// Read text version
	var textVersion uint16
	if err := binary.Read(reader, binary.BigEndian, &textVersion); err != nil {
		return nil, fmt.Errorf("failed to read text version: %w", err)
	}

	// Read descriptor version
	var descriptorVersion uint32
	if err := binary.Read(reader, binary.BigEndian, &descriptorVersion); err != nil {
		return nil, fmt.Errorf("failed to read descriptor version: %w", err)
	}

	// Parse text descriptor
	remaining := make([]byte, reader.Len())
	if _, err := io.ReadFull(reader, remaining); err != nil {
		return nil, fmt.Errorf("failed to read remaining data: %w", err)
	}

	// Create descriptor parser starting from text data
	textParser := NewDescriptorParser(remaining)
	textData, err := textParser.Parse()
	if err != nil {
		// If descriptor parsing fails, continue with empty data
		textData = make(map[string]interface{})
	}
	info.TextData = textData

	// Try to extract engine data string from TextData
	if engineDataBytes, ok := textData["EngineData"].([]byte); ok {
		info.EngineData = string(engineDataBytes)
	}

	// Note: Warp data parsing is skipped for now as it's after engine data
	// and we'd need to track position carefully

	// Bounds would be at the end if we could parse everything
	// For now, leave them as zero

	return info, nil
}

// HasTextContent checks if this TypeTool has actual text content
func (t *TypeToolInfo) HasTextContent() bool {
	return t.Text() != ""
}
