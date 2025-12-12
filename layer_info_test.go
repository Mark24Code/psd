package psd

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUnicodeName(t *testing.T) {
	buf := new(bytes.Buffer)

	// Write "Test Layer" in UTF-16
	testName := "Test Layer"
	runes := []rune(testName)
	binary.Write(buf, binary.BigEndian, uint32(len(runes)))
	for _, r := range runes {
		binary.Write(buf, binary.BigEndian, uint16(r))
	}

	reader := bytes.NewReader(buf.Bytes())
	name := parseUnicodeName(reader)

	assert.Equal(t, testName, name)
}

func TestParseLayerID(t *testing.T) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, int32(42))

	reader := bytes.NewReader(buf.Bytes())
	id := parseLayerID(reader)

	assert.Equal(t, int32(42), id)
}

func TestParseFillOpacity(t *testing.T) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint8(128))

	reader := bytes.NewReader(buf.Bytes())
	opacity := parseFillOpacity(reader)

	assert.Equal(t, uint8(128), opacity)
}

func TestParseSectionDivider(t *testing.T) {
	tests := []struct {
		name     string
		dataFunc func() []byte
		expected SectionDividerType
	}{
		{
			name: "open folder",
			dataFunc: func() []byte {
				buf := new(bytes.Buffer)
				binary.Write(buf, binary.BigEndian, int32(1))
				return buf.Bytes()
			},
			expected: SectionDividerOpenFolder,
		},
		{
			name: "closed folder",
			dataFunc: func() []byte {
				buf := new(bytes.Buffer)
				binary.Write(buf, binary.BigEndian, int32(2))
				return buf.Bytes()
			},
			expected: SectionDividerClosedFolder,
		},
		{
			name: "bounding section",
			dataFunc: func() []byte {
				buf := new(bytes.Buffer)
				binary.Write(buf, binary.BigEndian, int32(3))
				return buf.Bytes()
			},
			expected: SectionDividerBoundingStart,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.dataFunc()
			reader := bytes.NewReader(data)
			info := parseSectionDivider(reader)

			assert.NotNil(t, info)
			assert.Equal(t, tt.expected, info.Type)
		})
	}
}

func TestParseSectionDividerWithBlendMode(t *testing.T) {
	buf := new(bytes.Buffer)

	// Write type
	binary.Write(buf, binary.BigEndian, int32(1)) // Open folder

	// Write signature
	buf.WriteString("8BIM")

	// Write blend mode
	buf.WriteString("norm")

	reader := bytes.NewReader(buf.Bytes())
	info := parseSectionDivider(reader)

	assert.NotNil(t, info)
	assert.Equal(t, SectionDividerOpenFolder, info.Type)
	assert.Equal(t, "norm", info.BlendMode)
}

func TestParseVectorMask(t *testing.T) {
	buf := new(bytes.Buffer)

	// Write version
	binary.Write(buf, binary.BigEndian, uint32(3))

	// Write flags (inverted)
	binary.Write(buf, binary.BigEndian, uint32(1))

	// Write some path data
	buf.Write([]byte{0x01, 0x02, 0x03, 0x04})

	reader := bytes.NewReader(buf.Bytes())
	info := parseVectorMask(reader)

	assert.NotNil(t, info)
	assert.Equal(t, uint32(3), info.Version)
	assert.True(t, info.HasMask)
	assert.True(t, info.IsInverted)
	assert.NotEmpty(t, info.PathData)
}

func TestLayerGetParsedInfo(t *testing.T) {
	layer := &Layer{
		LayerInfo: make(map[string][]byte),
	}

	// Add unicode name
	buf := new(bytes.Buffer)
	testName := "测试图层" // Chinese characters
	runes := []rune(testName)
	binary.Write(buf, binary.BigEndian, uint32(len(runes)))
	for _, r := range runes {
		binary.Write(buf, binary.BigEndian, uint16(r))
	}
	layer.LayerInfo["luni"] = buf.Bytes()

	// Test unicode name
	unicodeName := layer.GetUnicodeName()
	assert.Equal(t, testName, unicodeName)

	// Add layer ID
	idBuf := new(bytes.Buffer)
	binary.Write(idBuf, binary.BigEndian, int32(999))
	layer.LayerInfo["lyid"] = idBuf.Bytes()

	// Test layer ID
	layerID := layer.GetLayerID()
	assert.Equal(t, int32(999), layerID)

	// Add fill opacity
	opacityBuf := new(bytes.Buffer)
	binary.Write(opacityBuf, binary.BigEndian, uint8(200))
	layer.LayerInfo["iOpa"] = opacityBuf.Bytes()

	// Test fill opacity
	fillOpacity := layer.GetFillOpacity()
	assert.Equal(t, uint8(200), fillOpacity)
}

func TestSectionDividerTypeString(t *testing.T) {
	tests := []struct {
		divType  SectionDividerType
		expected string
	}{
		{SectionDividerOther, "other"},
		{SectionDividerOpenFolder, "open folder"},
		{SectionDividerClosedFolder, "closed folder"},
		{SectionDividerBoundingStart, "bounding section divider"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.divType.String())
		})
	}
}
