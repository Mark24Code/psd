package psd

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Resource represents a single image resource
type Resource struct {
	Type string
	ID   uint16
	Name string
	Data []byte
}

// ResourceSection represents the image resources section
type ResourceSection struct {
	file      *File
	Resources map[uint16]*Resource
}

// Rectangle represents a bounding box
type Rectangle struct {
	Top    int32
	Left   int32
	Bottom int32
	Right  int32
}

// Slice represents a slice in the PSD
type Slice struct {
	ID                int32
	GroupID           int32
	Origin            int32
	AssociatedLayerID int32
	Name              string
	Type              int32
	Bounds            Rectangle
	URL               string
	Target            string
	Message           string
	Alt               string
	CellTextIsHTML    bool
	CellText          string
	HorizontalAlign   int32
	VerticalAlign     int32
}

// SlicesResource represents the slices resource (ID 1050)
type SlicesResource struct {
	Version int32
	Bounds  Rectangle
	Name    string
	Slices  []Slice
}

// Guide represents a guide in the PSD
type Guide struct {
	Position     int32
	IsHorizontal bool
}

// GuidesResource represents the guides resource (ID 1032)
type GuidesResource struct {
	Guides []Guide
}

// Parse parses the resources section
func (r *ResourceSection) Parse() error {
	// Read resources length
	length, err := r.file.ReadUint32()
	if err != nil {
		return fmt.Errorf("failed to read resources length: %w", err)
	}

	if length == 0 {
		r.Resources = make(map[uint16]*Resource)
		return nil
	}

	// Mark start position
	startPos, err := r.file.Tell()
	if err != nil {
		return err
	}

	r.Resources = make(map[uint16]*Resource)
	endPos := startPos + int64(length)

	// Parse resources
	for {
		currentPos, err := r.file.Tell()
		if err != nil {
			return err
		}
		if currentPos >= endPos {
			break
		}

		resource, err := r.parseResource()
		if err != nil {
			return fmt.Errorf("failed to parse resource: %w", err)
		}

		r.Resources[resource.ID] = resource
	}

	return nil
}

func (r *ResourceSection) parseResource() (*Resource, error) {
	resource := &Resource{}

	// Read type signature (4 bytes) - should be "8BIM"
	resourceType, err := r.file.ReadString(4)
	if err != nil {
		return nil, err
	}
	resource.Type = resourceType

	// Read resource ID (2 bytes)
	id, err := r.file.ReadUint16()
	if err != nil {
		return nil, err
	}
	resource.ID = id

	// Read Pascal string for name
	nameLen, err := r.file.ReadByte()
	if err != nil {
		return nil, err
	}

	if nameLen > 0 {
		name, err := r.file.ReadString(int(nameLen))
		if err != nil {
			return nil, err
		}
		resource.Name = name
	}

	// Pascal string is padded to even size
	if (nameLen+1)%2 != 0 {
		if err := r.file.Skip(1); err != nil {
			return nil, err
		}
	}

	// Read resource data size
	dataSize, err := r.file.ReadUint32()
	if err != nil {
		return nil, err
	}

	// Read resource data
	if dataSize > 0 {
		data := make([]byte, dataSize)
		if _, err := r.file.Read(data); err != nil {
			return nil, err
		}
		resource.Data = data

		// Data is padded to even size
		if dataSize%2 != 0 {
			if err := r.file.Skip(1); err != nil {
				return nil, err
			}
		}
	}

	return resource, nil
}

// ParseSlices parses the slices resource (ID 1050)
func (r *ResourceSection) ParseSlices() (*SlicesResource, error) {
	resource, exists := r.Resources[1050]
	if !exists || len(resource.Data) == 0 {
		// Return default slice for files without slices
		return &SlicesResource{
			Version: 6,
			Slices:  []Slice{{ID: 0}},
		}, nil
	}

	reader := bytes.NewReader(resource.Data)
	result := &SlicesResource{}

	// Read version
	if err := binary.Read(reader, binary.BigEndian, &result.Version); err != nil {
		return nil, err
	}

	if result.Version == 6 {
		// Parse version 6 (legacy format)
		if err := binary.Read(reader, binary.BigEndian, &result.Bounds); err != nil {
			return nil, err
		}

		// Read name (Unicode string)
		var nameLen uint32
		if err := binary.Read(reader, binary.BigEndian, &nameLen); err != nil {
			return nil, err
		}
		if nameLen > 0 {
			nameBytes := make([]byte, nameLen*2) // Unicode is 2 bytes per char
			if _, err := reader.Read(nameBytes); err != nil {
				return nil, err
			}
			result.Name = decodeUnicodeString(nameBytes)
		}

		// Read slice count
		var sliceCount int32
		if err := binary.Read(reader, binary.BigEndian, &sliceCount); err != nil {
			return nil, err
		}

		result.Slices = make([]Slice, sliceCount)
		for i := int32(0); i < sliceCount; i++ {
			slice := &result.Slices[i]

			binary.Read(reader, binary.BigEndian, &slice.ID)
			binary.Read(reader, binary.BigEndian, &slice.GroupID)
			binary.Read(reader, binary.BigEndian, &slice.Origin)

			if slice.Origin == 1 {
				binary.Read(reader, binary.BigEndian, &slice.AssociatedLayerID)
			}

			// Read name
			var nameLen uint32
			binary.Read(reader, binary.BigEndian, &nameLen)
			if nameLen > 0 {
				nameBytes := make([]byte, nameLen*2)
				reader.Read(nameBytes)
				slice.Name = decodeUnicodeString(nameBytes)
			}

			binary.Read(reader, binary.BigEndian, &slice.Type)
			binary.Read(reader, binary.BigEndian, &slice.Bounds)

			// Read URL, target, message, alt (Unicode strings)
			slice.URL = readUnicodeStringFromReader(reader)
			slice.Target = readUnicodeStringFromReader(reader)
			slice.Message = readUnicodeStringFromReader(reader)
			slice.Alt = readUnicodeStringFromReader(reader)

			var htmlFlag byte
			binary.Read(reader, binary.BigEndian, &htmlFlag)
			slice.CellTextIsHTML = htmlFlag != 0

			slice.CellText = readUnicodeStringFromReader(reader)

			binary.Read(reader, binary.BigEndian, &slice.HorizontalAlign)
			binary.Read(reader, binary.BigEndian, &slice.VerticalAlign)

			// Skip color (ARGB)
			reader.Seek(4, 1)
		}
	} else {
		// Version 7/8 uses descriptor format - simplified parsing
		// For now, return empty slices
		result.Slices = []Slice{}
	}

	return result, nil
}

// ParseGuides parses the guides resource (ID 1032)
func (r *ResourceSection) ParseGuides() (*GuidesResource, error) {
	resource, exists := r.Resources[1032]
	if !exists || len(resource.Data) == 0 {
		return &GuidesResource{Guides: []Guide{}}, nil
	}

	reader := bytes.NewReader(resource.Data)
	result := &GuidesResource{}

	// Skip version (4 bytes) and grid info (8 bytes)
	reader.Seek(12, 1)

	// Read guide count
	var guideCount uint32
	if err := binary.Read(reader, binary.BigEndian, &guideCount); err != nil {
		return nil, err
	}

	result.Guides = make([]Guide, guideCount)
	for i := uint32(0); i < guideCount; i++ {
		var position int32
		var direction byte

		binary.Read(reader, binary.BigEndian, &position)
		binary.Read(reader, binary.BigEndian, &direction)

		result.Guides[i] = Guide{
			Position:     position,
			IsHorizontal: direction == 0,
		}
	}

	return result, nil
}

// LayerComps returns layer comps from resources
func (r *ResourceSection) LayerComps() []LayerComp {
	// Resource ID 1065 contains layer comps
	// This is a simplified implementation
	// Full implementation would need to parse the descriptor data
	return []LayerComp{}
}

// LayerComp represents a layer comp
type LayerComp struct {
	ID   int
	Name string
}

// Helper functions for Unicode string handling
func decodeUnicodeString(data []byte) string {
	runes := make([]rune, len(data)/2)
	for i := 0; i < len(data)/2; i++ {
		runes[i] = rune(binary.BigEndian.Uint16(data[i*2:]))
	}
	return string(runes)
}

func readUnicodeStringFromReader(reader *bytes.Reader) string {
	var length uint32
	binary.Read(reader, binary.BigEndian, &length)
	if length == 0 {
		return ""
	}
	data := make([]byte, length*2)
	reader.Read(data)
	return decodeUnicodeString(data)
}
