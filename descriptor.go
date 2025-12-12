package psd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Descriptor represents a PSD descriptor structure
// Descriptors are complex data structures used in modern PSD features
type Descriptor struct {
	Class string
	Data  map[string]interface{}
}

// DescriptorParser parses descriptor data from PSD files
type DescriptorParser struct {
	reader *bytes.Reader
}

// NewDescriptorParser creates a new descriptor parser
func NewDescriptorParser(data []byte) *DescriptorParser {
	return &DescriptorParser{
		reader: bytes.NewReader(data),
	}
}

// Parse parses a descriptor and returns the result as a map
func (d *DescriptorParser) Parse() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Parse class
	class, err := d.parseClass()
	if err != nil {
		return nil, fmt.Errorf("failed to parse class: %w", err)
	}
	result["class"] = class

	// Read number of items
	var numItems uint32
	if err := binary.Read(d.reader, binary.BigEndian, &numItems); err != nil {
		return nil, fmt.Errorf("failed to read num items: %w", err)
	}

	// Parse each item
	for i := uint32(0); i < numItems; i++ {
		key, value, err := d.parseKeyItem()
		if err != nil {
			return nil, fmt.Errorf("failed to parse key item %d: %w", i, err)
		}
		result[key] = value
	}

	return result, nil
}

// parseClass parses a class structure
func (d *DescriptorParser) parseClass() (map[string]interface{}, error) {
	class := make(map[string]interface{})

	// Parse name (Unicode string)
	name, err := d.readUnicodeString()
	if err != nil {
		return nil, fmt.Errorf("failed to read class name: %w", err)
	}
	class["name"] = name

	// Parse ID
	id, err := d.parseID()
	if err != nil {
		return nil, fmt.Errorf("failed to read class ID: %w", err)
	}
	class["id"] = id

	return class, nil
}

// parseID parses an ID (length-prefixed string or 4-byte code)
func (d *DescriptorParser) parseID() (string, error) {
	var length uint32
	if err := binary.Read(d.reader, binary.BigEndian, &length); err != nil {
		return "", err
	}

	if length == 0 {
		// 4-byte code
		buf := make([]byte, 4)
		if _, err := io.ReadFull(d.reader, buf); err != nil {
			return "", err
		}
		return string(buf), nil
	}

	// Variable length string
	buf := make([]byte, length)
	if _, err := io.ReadFull(d.reader, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

// parseKeyItem parses a key-value pair
func (d *DescriptorParser) parseKeyItem() (string, interface{}, error) {
	// Parse key ID
	key, err := d.parseID()
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse key: %w", err)
	}

	// Parse value
	value, err := d.parseItem("")
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse value for key %s: %w", key, err)
	}

	return key, value, nil
}

// parseItem parses a value of any type
func (d *DescriptorParser) parseItem(itemType string) (interface{}, error) {
	// Read type if not provided
	if itemType == "" {
		typeBytes := make([]byte, 4)
		if _, err := io.ReadFull(d.reader, typeBytes); err != nil {
			return nil, err
		}
		itemType = string(typeBytes)
	}

	switch itemType {
	case "bool":
		return d.parseBoolean()
	case "type", "GlbC":
		return d.parseClass()
	case "Objc", "GlbO":
		// Nested descriptor
		return d.Parse()
	case "doub":
		return d.parseDouble()
	case "enum":
		return d.parseEnum()
	case "alis":
		return d.parseAlias()
	case "long":
		return d.parseInt()
	case "comp":
		return d.parseLargeInt()
	case "VlLs":
		return d.parseList()
	case "ObAr":
		return d.parseObjectArray()
	case "tdta":
		return d.parseRawData()
	case "obj ":
		return d.parseReference()
	case "TEXT":
		return d.readUnicodeString()
	case "UntF":
		return d.parseUnitDouble()
	case "UnFl":
		return d.parseUnitFloat()
	default:
		return nil, fmt.Errorf("unknown descriptor type: %s", itemType)
	}
}

// Basic type parsers
func (d *DescriptorParser) parseBoolean() (bool, error) {
	var value byte
	if err := binary.Read(d.reader, binary.BigEndian, &value); err != nil {
		return false, err
	}
	return value != 0, nil
}

func (d *DescriptorParser) parseDouble() (float64, error) {
	var value float64
	if err := binary.Read(d.reader, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (d *DescriptorParser) parseInt() (int32, error) {
	var value int32
	if err := binary.Read(d.reader, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return value, nil
}

func (d *DescriptorParser) parseLargeInt() (int64, error) {
	var value int64
	if err := binary.Read(d.reader, binary.BigEndian, &value); err != nil {
		return 0, err
	}
	return value, nil
}

// parseEnum parses an enumerated value
func (d *DescriptorParser) parseEnum() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	typeID, err := d.parseID()
	if err != nil {
		return nil, fmt.Errorf("failed to parse enum type: %w", err)
	}
	result["type"] = typeID

	valueID, err := d.parseID()
	if err != nil {
		return nil, fmt.Errorf("failed to parse enum value: %w", err)
	}
	result["value"] = valueID

	return result, nil
}

// parseAlias parses an alias (length-prefixed data)
func (d *DescriptorParser) parseAlias() ([]byte, error) {
	var length uint32
	if err := binary.Read(d.reader, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(d.reader, data); err != nil {
		return nil, err
	}

	return data, nil
}

// parseList parses a list of items
func (d *DescriptorParser) parseList() ([]interface{}, error) {
	var count uint32
	if err := binary.Read(d.reader, binary.BigEndian, &count); err != nil {
		return nil, err
	}

	items := make([]interface{}, count)
	for i := uint32(0); i < count; i++ {
		value, err := d.parseItem("")
		if err != nil {
			return nil, fmt.Errorf("failed to parse list item %d: %w", i, err)
		}
		items[i] = value
	}

	return items, nil
}

// parseObjectArray parses an object array (not fully implemented in Ruby version)
func (d *DescriptorParser) parseObjectArray() (interface{}, error) {
	// This is not fully implemented in psd.rb either
	// Return nil for now to match Ruby behavior
	return nil, fmt.Errorf("object array parsing not implemented")
}

// parseRawData parses raw binary data
func (d *DescriptorParser) parseRawData() ([]byte, error) {
	var length uint32
	if err := binary.Read(d.reader, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(d.reader, data); err != nil {
		return nil, err
	}

	return data, nil
}

// parseReference parses a reference
func (d *DescriptorParser) parseReference() ([]map[string]interface{}, error) {
	var numItems uint32
	if err := binary.Read(d.reader, binary.BigEndian, &numItems); err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, numItems)
	for i := uint32(0); i < numItems; i++ {
		typeBytes := make([]byte, 4)
		if _, err := io.ReadFull(d.reader, typeBytes); err != nil {
			return nil, err
		}
		refType := string(typeBytes)

		var value interface{}
		var err error

		switch refType {
		case "prop":
			value, err = d.parseProperty()
		case "Clss":
			value, err = d.parseClass()
		case "Enmr":
			value, err = d.parseEnumReference()
		case "Idnt":
			value, err = d.parseInt()
		case "indx":
			value, err = d.parseInt()
		case "name":
			value, err = d.readUnicodeString()
		case "rele":
			value, err = d.parseInt()
		default:
			return nil, fmt.Errorf("unknown reference type: %s", refType)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to parse reference item %d: %w", i, err)
		}

		items[i] = map[string]interface{}{
			"type":  refType,
			"value": value,
		}
	}

	return items, nil
}

// parseProperty parses a property
func (d *DescriptorParser) parseProperty() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	class, err := d.parseClass()
	if err != nil {
		return nil, fmt.Errorf("failed to parse property class: %w", err)
	}
	result["class"] = class

	id, err := d.parseID()
	if err != nil {
		return nil, fmt.Errorf("failed to parse property ID: %w", err)
	}
	result["id"] = id

	return result, nil
}

// parseEnumReference parses an enum reference
func (d *DescriptorParser) parseEnumReference() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	class, err := d.parseClass()
	if err != nil {
		return nil, fmt.Errorf("failed to parse enum class: %w", err)
	}
	result["class"] = class

	typeID, err := d.parseID()
	if err != nil {
		return nil, fmt.Errorf("failed to parse enum type: %w", err)
	}
	result["type"] = typeID

	valueID, err := d.parseID()
	if err != nil {
		return nil, fmt.Errorf("failed to parse enum value: %w", err)
	}
	result["value"] = valueID

	return result, nil
}

// Unit types
var unitTypes = map[string]string{
	"#Ang": "Angle",
	"#Rsl": "Density",
	"#Rlt": "Distance",
	"#Nne": "None",
	"#Prc": "Percent",
	"#Pxl": "Pixels",
	"#Mlm": "Millimeters",
	"#Pnt": "Points",
}

// parseUnitDouble parses a unit double value
func (d *DescriptorParser) parseUnitDouble() (map[string]interface{}, error) {
	unitIDBytes := make([]byte, 4)
	if _, err := io.ReadFull(d.reader, unitIDBytes); err != nil {
		return nil, err
	}
	unitID := string(unitIDBytes)

	var value float64
	if err := binary.Read(d.reader, binary.BigEndian, &value); err != nil {
		return nil, err
	}

	unit := unitTypes[unitID]
	if unit == "" {
		unit = "Unknown"
	}

	return map[string]interface{}{
		"id":    unitID,
		"unit":  unit,
		"value": value,
	}, nil
}

// parseUnitFloat parses a unit float value
func (d *DescriptorParser) parseUnitFloat() (map[string]interface{}, error) {
	unitIDBytes := make([]byte, 4)
	if _, err := io.ReadFull(d.reader, unitIDBytes); err != nil {
		return nil, err
	}
	unitID := string(unitIDBytes)

	var value float32
	if err := binary.Read(d.reader, binary.BigEndian, &value); err != nil {
		return nil, err
	}

	unit := unitTypes[unitID]
	if unit == "" {
		unit = "Unknown"
	}

	return map[string]interface{}{
		"id":    unitID,
		"unit":  unit,
		"value": value,
	}, nil
}

// readUnicodeString reads a UTF-16 string
func (d *DescriptorParser) readUnicodeString() (string, error) {
	// Read length (number of UTF-16 characters, not bytes)
	var length uint32
	if err := binary.Read(d.reader, binary.BigEndian, &length); err != nil {
		return "", err
	}

	if length == 0 {
		return "", nil
	}

	// Read UTF-16 big-endian data
	data := make([]byte, length*2)
	if _, err := io.ReadFull(d.reader, data); err != nil {
		return "", err
	}

	// Convert UTF-16 BE to UTF-8
	runes := make([]rune, length)
	for i := uint32(0); i < length; i++ {
		runes[i] = rune(binary.BigEndian.Uint16(data[i*2:]))
	}

	return string(runes), nil
}
