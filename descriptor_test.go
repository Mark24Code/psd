package psd

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDescriptorParser_ParseBoolean(t *testing.T) {
	// Create a simple boolean descriptor
	buf := new(bytes.Buffer)

	// Write class
	writeUnicodeString(buf, "TestClass")
	writeString(buf, "Test")

	// Write 1 item
	binary.Write(buf, binary.BigEndian, uint32(1))

	// Write key
	writeString(buf, "bool")

	// Write type and value
	buf.WriteString("bool")
	buf.WriteByte(1) // true

	parser := NewDescriptorParser(buf.Bytes())
	result, err := parser.Parse()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, true, result["bool"])
}

func TestDescriptorParser_ParseInt(t *testing.T) {
	buf := new(bytes.Buffer)

	// Write class
	writeUnicodeString(buf, "Test")
	writeString(buf, "Test")

	// Write 1 item
	binary.Write(buf, binary.BigEndian, uint32(1))

	// Write key
	writeString(buf, "num")

	// Write type and value
	buf.WriteString("long")
	binary.Write(buf, binary.BigEndian, int32(42))

	parser := NewDescriptorParser(buf.Bytes())
	result, err := parser.Parse()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(42), result["num"])
}

func TestDescriptorParser_ParseDouble(t *testing.T) {
	buf := new(bytes.Buffer)

	// Write class
	writeUnicodeString(buf, "Test")
	writeString(buf, "Test")

	// Write 1 item
	binary.Write(buf, binary.BigEndian, uint32(1))

	// Write key
	writeString(buf, "val")

	// Write type and value
	buf.WriteString("doub")
	binary.Write(buf, binary.BigEndian, float64(3.14))

	parser := NewDescriptorParser(buf.Bytes())
	result, err := parser.Parse()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.InDelta(t, 3.14, result["val"], 0.001)
}

func TestDescriptorParser_ParseText(t *testing.T) {
	buf := new(bytes.Buffer)

	// Write class
	writeUnicodeString(buf, "Test")
	writeString(buf, "Test")

	// Write 1 item
	binary.Write(buf, binary.BigEndian, uint32(1))

	// Write key
	writeString(buf, "text")

	// Write type and value
	buf.WriteString("TEXT")
	writeUnicodeString(buf, "Hello World")

	parser := NewDescriptorParser(buf.Bytes())
	result, err := parser.Parse()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Hello World", result["text"])
}

func TestDescriptorParser_ParseEnum(t *testing.T) {
	buf := new(bytes.Buffer)

	// Write class
	writeUnicodeString(buf, "Test")
	writeString(buf, "Test")

	// Write 1 item
	binary.Write(buf, binary.BigEndian, uint32(1))

	// Write key
	writeString(buf, "mode")

	// Write type and enum
	buf.WriteString("enum")
	writeString(buf, "Type")
	writeString(buf, "Val ")

	parser := NewDescriptorParser(buf.Bytes())
	result, err := parser.Parse()

	assert.NoError(t, err)
	assert.NotNil(t, result)

	enum := result["mode"].(map[string]interface{})
	assert.Equal(t, "Type", enum["type"])
	assert.Equal(t, "Val ", enum["value"])
}

func TestDescriptorParser_ParseList(t *testing.T) {
	buf := new(bytes.Buffer)

	// Write class
	writeUnicodeString(buf, "Test")
	writeString(buf, "Test")

	// Write 1 item
	binary.Write(buf, binary.BigEndian, uint32(1))

	// Write key
	writeString(buf, "list")

	// Write type
	buf.WriteString("VlLs")

	// Write list count
	binary.Write(buf, binary.BigEndian, uint32(3))

	// Write 3 integers
	buf.WriteString("long")
	binary.Write(buf, binary.BigEndian, int32(1))
	buf.WriteString("long")
	binary.Write(buf, binary.BigEndian, int32(2))
	buf.WriteString("long")
	binary.Write(buf, binary.BigEndian, int32(3))

	parser := NewDescriptorParser(buf.Bytes())
	result, err := parser.Parse()

	assert.NoError(t, err)
	assert.NotNil(t, result)

	list := result["list"].([]interface{})
	assert.Len(t, list, 3)
	assert.Equal(t, int32(1), list[0])
	assert.Equal(t, int32(2), list[1])
	assert.Equal(t, int32(3), list[2])
}

// Helper functions for test data generation
func writeUnicodeString(buf *bytes.Buffer, s string) {
	runes := []rune(s)
	binary.Write(buf, binary.BigEndian, uint32(len(runes)))
	for _, r := range runes {
		binary.Write(buf, binary.BigEndian, uint16(r))
	}
}

func writeString(buf *bytes.Buffer, s string) {
	if len(s) == 4 {
		// 4-byte code
		binary.Write(buf, binary.BigEndian, uint32(0))
		buf.WriteString(s)
	} else {
		binary.Write(buf, binary.BigEndian, uint32(len(s)))
		buf.WriteString(s)
	}
}
