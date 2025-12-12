package psd

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// LayerInfoType represents different types of layer information
type LayerInfoType string

const (
	LayerInfoUnicodeName     LayerInfoType = "luni" // Unicode layer name
	LayerInfoLayerID         LayerInfoType = "lyid" // Layer ID
	LayerInfoFillOpacity     LayerInfoType = "iOpa" // Fill opacity
	LayerInfoSectionDivider  LayerInfoType = "lsct" // Layer section divider
	LayerInfoSectionDivider2 LayerInfoType = "lsdk" // Layer section divider (older)
	LayerInfoVectorMask      LayerInfoType = "vmsk" // Vector mask
	LayerInfoVectorMask2     LayerInfoType = "vsms" // Vector mask (Photoshop 6.0)
)

// ParsedLayerInfo holds parsed layer information
type ParsedLayerInfo struct {
	UnicodeName    string
	LayerID        int32
	FillOpacity    uint8
	SectionType    int32
	HasVectorMask  bool
	VectorMaskData []byte
}

// parseLayerInfo parses specific layer info based on key
func parseLayerInfo(key string, data []byte) interface{} {
	reader := bytes.NewReader(data)

	switch key {
	case "luni":
		return parseUnicodeName(reader)
	case "lyid":
		return parseLayerID(reader)
	case "iOpa":
		return parseFillOpacity(reader)
	case "lsct", "lsdk":
		return parseSectionDivider(reader)
	case "vmsk", "vsms":
		return parseVectorMask(reader)
	default:
		return nil
	}
}

// parseUnicodeName parses Unicode layer name
func parseUnicodeName(reader *bytes.Reader) string {
	// Read length (number of UTF-16 characters)
	var length uint32
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return ""
	}

	if length == 0 {
		return ""
	}

	// Read UTF-16 data
	data := make([]byte, length*2)
	if _, err := reader.Read(data); err != nil {
		return ""
	}

	// Convert UTF-16 BE to UTF-8
	runes := make([]rune, length)
	for i := uint32(0); i < length; i++ {
		runes[i] = rune(binary.BigEndian.Uint16(data[i*2:]))
	}

	return string(runes)
}

// parseLayerID parses layer ID
func parseLayerID(reader *bytes.Reader) int32 {
	var id int32
	if err := binary.Read(reader, binary.BigEndian, &id); err != nil {
		return 0
	}
	return id
}

// parseFillOpacity parses fill opacity
func parseFillOpacity(reader *bytes.Reader) uint8 {
	var opacity uint8
	if err := binary.Read(reader, binary.BigEndian, &opacity); err != nil {
		return 255
	}
	return opacity
}

// SectionDividerType represents layer section divider types
type SectionDividerType int32

const (
	SectionDividerOther         SectionDividerType = 0
	SectionDividerOpenFolder    SectionDividerType = 1
	SectionDividerClosedFolder  SectionDividerType = 2
	SectionDividerBoundingStart SectionDividerType = 3 // Folder end marker
)

// SectionDividerInfo contains section divider information
type SectionDividerInfo struct {
	Type      SectionDividerType
	BlendMode string
	SubType   int32
}

// parseSectionDivider parses layer section divider info
func parseSectionDivider(reader *bytes.Reader) *SectionDividerInfo {
	info := &SectionDividerInfo{}

	// Read type (4 bytes)
	var sectionType int32
	if err := binary.Read(reader, binary.BigEndian, &sectionType); err != nil {
		return info
	}
	info.Type = SectionDividerType(sectionType)

	// If there's more data, read signature and blend mode
	if reader.Len() >= 8 {
		// Read signature (should be "8BIM")
		sig := make([]byte, 4)
		reader.Read(sig)

		// Read blend mode key
		blendKey := make([]byte, 4)
		reader.Read(blendKey)
		info.BlendMode = string(blendKey)
	}

	// If there's even more data, read sub type
	if reader.Len() >= 4 {
		binary.Read(reader, binary.BigEndian, &info.SubType)
	}

	return info
}

// VectorMaskInfo contains vector mask information
type VectorMaskInfo struct {
	Version    uint32
	Flags      uint32
	PathData   []byte
	HasMask    bool
	IsInverted bool
}

// parseVectorMask parses vector mask data
func parseVectorMask(reader *bytes.Reader) *VectorMaskInfo {
	info := &VectorMaskInfo{
		HasMask: true,
	}

	// Read version
	if err := binary.Read(reader, binary.BigEndian, &info.Version); err != nil {
		return info
	}

	// Read flags
	if err := binary.Read(reader, binary.BigEndian, &info.Flags); err != nil {
		return info
	}

	// Check if inverted
	info.IsInverted = (info.Flags & 0x01) != 0

	// Store remaining data as path data
	info.PathData = make([]byte, reader.Len())
	reader.Read(info.PathData)

	return info
}

// EnhanceLayerWithParsedInfo updates layer with parsed info
func (l *Layer) EnhanceLayerWithParsedInfo() {
	if l.LayerInfo == nil {
		return
	}

	// Parse unicode name if available
	if data, ok := l.LayerInfo["luni"]; ok {
		if unicodeName := parseLayerInfo("luni", data); unicodeName != nil {
			if name, ok := unicodeName.(string); ok && name != "" {
				l.Name = name // Override with unicode name
			}
		}
	}

	// Parse layer ID
	if data, ok := l.LayerInfo["lyid"]; ok {
		if layerID := parseLayerInfo("lyid", data); layerID != nil {
			// Store in a new field if needed
			_ = layerID
		}
	}

	// Parse fill opacity
	if data, ok := l.LayerInfo["iOpa"]; ok {
		if fillOpacity := parseLayerInfo("iOpa", data); fillOpacity != nil {
			// Store in a new field if needed
			_ = fillOpacity
		}
	}
}

// GetParsedInfo returns parsed layer info for a specific key
func (l *Layer) GetParsedInfo(key string) interface{} {
	if l.LayerInfo == nil {
		return nil
	}

	data, ok := l.LayerInfo[key]
	if !ok {
		return nil
	}

	return parseLayerInfo(key, data)
}

// GetUnicodeName returns the unicode name if available
func (l *Layer) GetUnicodeName() string {
	if info := l.GetParsedInfo("luni"); info != nil {
		if name, ok := info.(string); ok {
			return name
		}
	}
	return l.Name
}

// GetLayerID returns the layer ID
func (l *Layer) GetLayerID() int32 {
	if info := l.GetParsedInfo("lyid"); info != nil {
		if id, ok := info.(int32); ok {
			return id
		}
	}
	return 0
}

// GetFillOpacity returns the fill opacity
func (l *Layer) GetFillOpacity() uint8 {
	if info := l.GetParsedInfo("iOpa"); info != nil {
		if opacity, ok := info.(uint8); ok {
			return opacity
		}
	}
	return 255 // Default full opacity
}

// GetSectionDivider returns section divider info
func (l *Layer) GetSectionDivider() *SectionDividerInfo {
	// Try lsct first, then lsdk
	if info := l.GetParsedInfo("lsct"); info != nil {
		if divider, ok := info.(*SectionDividerInfo); ok {
			return divider
		}
	}
	if info := l.GetParsedInfo("lsdk"); info != nil {
		if divider, ok := info.(*SectionDividerInfo); ok {
			return divider
		}
	}
	return nil
}

// GetVectorMask returns vector mask info
func (l *Layer) GetVectorMask() *VectorMaskInfo {
	// Try vmsk first, then vsms
	if info := l.GetParsedInfo("vmsk"); info != nil {
		if mask, ok := info.(*VectorMaskInfo); ok {
			return mask
		}
	}
	if info := l.GetParsedInfo("vsms"); info != nil {
		if mask, ok := info.(*VectorMaskInfo); ok {
			return mask
		}
	}
	return nil
}

// HasVectorMask checks if layer has a vector mask
func (l *Layer) HasVectorMask() bool {
	return l.GetVectorMask() != nil
}

// IsFolderOpen checks if this is an open folder
func (l *Layer) IsFolderOpen() bool {
	divider := l.GetSectionDivider()
	if divider == nil {
		return false
	}
	return divider.Type == SectionDividerOpenFolder
}

// IsFolderClosed checks if this is a closed folder
func (l *Layer) IsFolderClosed() bool {
	divider := l.GetSectionDivider()
	if divider == nil {
		return false
	}
	return divider.Type == SectionDividerClosedFolder
}

// String returns a string representation of SectionDividerType
func (s SectionDividerType) String() string {
	switch s {
	case SectionDividerOther:
		return "other"
	case SectionDividerOpenFolder:
		return "open folder"
	case SectionDividerClosedFolder:
		return "closed folder"
	case SectionDividerBoundingStart:
		return "bounding section divider"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}
