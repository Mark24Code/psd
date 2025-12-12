package psd

import (
	"fmt"
	"image"
	"image/color"
	"strings"
)

// Layer represents a single layer in the PSD
type Layer struct {
	file   *File
	header *Header

	// Layer record fields
	Top    int32
	Left   int32
	Bottom int32
	Right  int32

	Channels     uint16
	ChannelInfo  []ChannelInfo
	BlendModeKey string
	Opacity      uint8
	Clipping     uint8
	Flags        uint8
	Name         string

	// Additional layer information
	LayerInfo map[string][]byte

	// Parsed layer info
	TypeTool *TypeToolInfo

	// Channel image data
	channels    map[int16]*ChannelImage
	ChannelData map[int16][]byte
}

// ChannelImage represents decoded channel image data
type ChannelImage struct {
	ID          int16
	Data        []byte
	Compression uint16
}

// ChannelInfo represents channel information in the layer record
type ChannelInfo struct {
	ID     int16
	Length uint32
}

// parseRecord parses the layer record (not the channel image data)
func (l *Layer) parseRecord() error {
	// Read layer rectangle
	top, err := l.file.ReadInt32()
	if err != nil {
		return err
	}
	l.Top = top

	left, err := l.file.ReadInt32()
	if err != nil {
		return err
	}
	l.Left = left

	bottom, err := l.file.ReadInt32()
	if err != nil {
		return err
	}
	l.Bottom = bottom

	right, err := l.file.ReadInt32()
	if err != nil {
		return err
	}
	l.Right = right

	// Read number of channels
	channels, err := l.file.ReadUint16()
	if err != nil {
		return err
	}
	l.Channels = channels

	// Read channel information
	l.ChannelInfo = make([]ChannelInfo, channels)
	for i := uint16(0); i < channels; i++ {
		channelID, err := l.file.ReadInt16()
		if err != nil {
			return err
		}

		channelLength, err := l.file.ReadUint32()
		if err != nil {
			return err
		}

		l.ChannelInfo[i] = ChannelInfo{
			ID:     channelID,
			Length: channelLength,
		}
	}

	// Read blend mode signature (should be "8BIM")
	sig, err := l.file.ReadString(4)
	if err != nil {
		return err
	}
	if sig != "8BIM" {
		return fmt.Errorf("invalid blend mode signature: %s", sig)
	}

	// Read blend mode key
	blendMode, err := l.file.ReadString(4)
	if err != nil {
		return err
	}
	l.BlendModeKey = blendMode

	// Read opacity
	opacity, err := l.file.ReadByte()
	if err != nil {
		return err
	}
	l.Opacity = opacity

	// Read clipping
	clipping, err := l.file.ReadByte()
	if err != nil {
		return err
	}
	l.Clipping = clipping

	// Read flags
	flags, err := l.file.ReadByte()
	if err != nil {
		return err
	}
	l.Flags = flags

	// Skip filler
	if err := l.file.Skip(1); err != nil {
		return err
	}

	// Read extra data length
	extraLen, err := l.file.ReadUint32()
	if err != nil {
		return err
	}

	if extraLen > 0 {
		extraStart, err := l.file.Tell()
		if err != nil {
			return err
		}

		// Parse layer mask data
		if err := l.parseLayerMaskData(); err != nil {
			return err
		}

		// Parse layer blending ranges
		if err := l.parseBlendingRanges(); err != nil {
			return err
		}

		// Read layer name (Pascal string)
		if err := l.parseLayerName(); err != nil {
			return err
		}

		// Parse additional layer information
		currentPos, err := l.file.Tell()
		if err != nil {
			return err
		}
		remainingExtra := int64(extraLen) - (currentPos - extraStart)

		if remainingExtra > 0 {
			if err := l.parseAdditionalLayerInfo(remainingExtra); err != nil {
				return err
			}
		}
	}

	// Enhance layer with parsed info (e.g., Unicode name from 'luni')
	l.EnhanceLayerWithParsedInfo()

	return nil
}

func (l *Layer) parseLayerMaskData() error {
	length, err := l.file.ReadUint32()
	if err != nil {
		return err
	}

	if length > 0 {
		if err := l.file.Skip(int64(length)); err != nil {
			return err
		}
	}

	return nil
}

func (l *Layer) parseBlendingRanges() error {
	length, err := l.file.ReadUint32()
	if err != nil {
		return err
	}

	if length > 0 {
		if err := l.file.Skip(int64(length)); err != nil {
			return err
		}
	}

	return nil
}

func (l *Layer) parseLayerName() error {
	nameLen, err := l.file.ReadByte()
	if err != nil {
		return err
	}

	if nameLen > 0 {
		name, err := l.file.ReadString(int(nameLen))
		if err != nil {
			return err
		}
		l.Name = name
	}

	// Pascal string padding to multiple of 4
	padSize := (4 - ((nameLen + 1) % 4)) % 4
	if padSize > 0 {
		if err := l.file.Skip(int64(padSize)); err != nil {
			return err
		}
	}

	return nil
}

func (l *Layer) parseAdditionalLayerInfo(length int64) error {
	l.LayerInfo = make(map[string][]byte)

	endPos, err := l.file.Tell()
	if err != nil {
		return err
	}
	endPos += length

	for {
		currentPos, err := l.file.Tell()
		if err != nil {
			return err
		}
		if currentPos >= endPos {
			break
		}

		// Read signature
		sig, err := l.file.ReadString(4)
		if err != nil {
			break
		}
		if sig != "8BIM" && sig != "8B64" {
			break
		}

		// Read key
		key, err := l.file.ReadString(4)
		if err != nil {
			break
		}

		// Read length
		dataLen, err := l.file.ReadUint32()
		if err != nil {
			break
		}

		// Read data
		if dataLen > 0 {
			data := make([]byte, dataLen)
			if _, err := l.file.Read(data); err != nil {
				break
			}
			l.LayerInfo[key] = data

			// Parse TypeTool if this is a text layer
			if key == "TySh" {
				if typeTool, err := ParseTypeTool(data); err == nil {
					l.TypeTool = typeTool
				}
			}

			// Padding to multiple of 4
			if dataLen%4 != 0 {
				padSize := 4 - (dataLen % 4)
				l.file.Skip(int64(padSize))
			}
		}
	}

	return nil
}

func (l *Layer) parseChannelData() error {
	l.channels = make(map[int16]*ChannelImage)
	l.ChannelData = make(map[int16][]byte)

	// Debug logging for specific layer
	isColorFillLayer := l.Name == "颜色填充 2"
	if isColorFillLayer {
		fmt.Printf("\n=== DEBUG: Parsing channels for layer '%s' ===\n", l.Name)
		fmt.Printf("Layer Opacity: %d\n", l.Opacity)
		fmt.Printf("Number of channels: %d\n", len(l.ChannelInfo))
		for i, ch := range l.ChannelInfo {
			fmt.Printf("  Channel[%d]: ID=%d, Length=%d\n", i, ch.ID, ch.Length)
		}
	}

	for _, chanInfo := range l.ChannelInfo {
		// Record file position at start of this channel
		startPos, err := l.file.Tell()
		if err != nil {
			return fmt.Errorf("failed to get file position for channel %d: %w", chanInfo.ID, err)
		}

		// If channel has no data (length <= 2 means only compression header or nothing),
		// we still need to read/skip the bytes to keep file pointer aligned
		if chanInfo.Length <= 2 {
			if chanInfo.Length > 0 {
				// Seek past the channel data (usually just the 2-byte compression header)
				if err := l.file.Skip(int64(chanInfo.Length)); err != nil {
					return fmt.Errorf("failed to skip channel %d: %w", chanInfo.ID, err)
				}
			}
			continue
		}

		// Read compression method
		compression, err := l.file.ReadUint16()
		if err != nil {
			return fmt.Errorf("failed to read compression for channel %d (length=%d): %w", chanInfo.ID, chanInfo.Length, err)
		}

		dataLength := chanInfo.Length - 2

		switch compression {
		case 0: // Raw data
			data := make([]byte, dataLength)
			if _, err := l.file.Read(data); err != nil {
				return fmt.Errorf("failed to read raw data for channel %d: %w", chanInfo.ID, err)
			}
			l.ChannelData[chanInfo.ID] = data
			l.channels[chanInfo.ID] = &ChannelImage{
				ID:          chanInfo.ID,
				Data:        data,
				Compression: compression,
			}

		case 1: // RLE compression
			// Read RLE compressed data
			compressedData := make([]byte, dataLength)
			if _, err := l.file.Read(compressedData); err != nil {
				return fmt.Errorf("failed to read RLE data for channel %d: %w", chanInfo.ID, err)
			}

			// Decompress RLE
			decompressed, err := l.decompressRLE(compressedData, chanInfo.ID)
			if err != nil {
				return fmt.Errorf("failed to decompress RLE for channel %d: %w", chanInfo.ID, err)
			}

			l.ChannelData[chanInfo.ID] = decompressed
			l.channels[chanInfo.ID] = &ChannelImage{
				ID:          chanInfo.ID,
				Data:        decompressed,
				Compression: compression,
			}

			// Debug logging for color fill layer
			if isColorFillLayer {
				fmt.Printf("\nChannel %d (RLE decompressed): %d bytes\n", chanInfo.ID, len(decompressed))
				if len(decompressed) > 0 {
					// Print first 20 bytes
					end := 20
					if len(decompressed) < end {
						end = len(decompressed)
					}
					fmt.Printf("  First %d bytes: %v\n", end, decompressed[:end])

					// Check if all bytes are the same
					allSame := true
					if len(decompressed) > 1 {
						first := decompressed[0]
						for _, b := range decompressed[1:min(1000, len(decompressed))] {
							if b != first {
								allSame = false
								break
							}
						}
						if allSame {
							fmt.Printf("  First 1000 bytes are all: %d\n", first)
						}
					}
				}
			}

		default:
			// Skip unknown compression
			if err := l.file.Skip(int64(dataLength)); err != nil {
				return fmt.Errorf("failed to skip unknown compression %d for channel %d: %w", compression, chanInfo.ID, err)
			}
		}

		// Verify we read the correct number of bytes
		// This is CRITICAL - if we didn't read exactly chanInfo.Length bytes,
		// we need to seek to the correct position for the next channel
		finishPos, err := l.file.Tell()
		if err != nil {
			return fmt.Errorf("failed to get file position after channel %d: %w", chanInfo.ID, err)
		}

		expectedPos := startPos + int64(chanInfo.Length)

		if finishPos != expectedPos {
			_, err := l.file.Seek(expectedPos, 0)
			if err != nil {
				return fmt.Errorf("failed to seek to correct position after channel %d: %w", chanInfo.ID, err)
			}
		}
	}

	return nil
}

// Width returns the width of the layer
func (l *Layer) Width() int32 {
	return l.Right - l.Left
}

// Height returns the height of the layer
func (l *Layer) Height() int32 {
	return l.Bottom - l.Top
}

// Visible returns whether the layer is visible
func (l *Layer) Visible() bool {
	return l.Flags&0x02 == 0
}

// IsFolder returns whether this layer is a folder/group
func (l *Layer) IsFolder() bool {
	_, exists := l.LayerInfo["lsct"]
	if !exists {
		_, exists = l.LayerInfo["lsdk"]
	}
	return exists
}

// IsFolderEnd returns whether this is a folder end marker
func (l *Layer) IsFolderEnd() bool {
	if !l.IsFolder() {
		return false
	}

	// Check section divider type
	data, exists := l.LayerInfo["lsct"]
	if !exists {
		data, exists = l.LayerInfo["lsdk"]
	}

	if exists && len(data) >= 4 {
		// Type 3 = bounding section divider (folder end)
		return data[3] == 3
	}

	return false
}

// NodeType returns the node type for this layer
func (l *Layer) NodeType() string {
	if l.IsFolder() {
		return NodeTypeGroup
	}
	return NodeTypeLayer
}

// BlendMode returns the blend mode
func (l *Layer) BlendMode() *BlendMode {
	return &BlendMode{
		Mode:              l.blendModeString(),
		Opacity:           l.Opacity,
		OpacityPercentage: int(float64(l.Opacity) / 255.0 * 100),
		Visible:           l.Visible(),
	}
}

func (l *Layer) blendModeString() string {
	modes := map[string]string{
		"norm": "normal",
		"dark": "darken",
		"lite": "lighten",
		"hue ": "hue",
		"sat ": "saturation",
		"colr": "color",
		"lum ": "luminosity",
		"mul ": "multiply",
		"scrn": "screen",
		"diss": "dissolve",
		"over": "overlay",
		"hLit": "hard_light",
		"sLit": "soft_light",
		"diff": "difference",
		"smud": "exclusion",
		"div ": "color_dodge",
		"idiv": "color_burn",
		"lbrn": "linear_burn",
		"lddg": "linear_dodge",
		"vLit": "vivid_light",
		"lLit": "linear_light",
		"pLit": "pin_light",
		"hMix": "hard_mix",
		"lgCl": "lighter_color",
		"dkCl": "darker_color",
		"fsub": "subtract",
		"fdiv": "divide",
	}

	if mode, exists := modes[l.BlendModeKey]; exists {
		return mode
	}

	return strings.TrimSpace(l.BlendModeKey)
}

// BlendMode represents layer blend mode information
type BlendMode struct {
	Mode              string
	Opacity           uint8
	OpacityPercentage int
	Visible           bool
}

// decompressRLE decompresses RLE compressed channel data
func (l *Layer) decompressRLE(compressedData []byte, channelID int16) ([]byte, error) {
	width := int(l.Width())
	height := int(l.Height())

	if width == 0 || height == 0 {
		return []byte{}, nil
	}

	// The first part contains byte counts for each scanline
	byteCounts := make([]uint16, height)
	offset := 0

	for i := 0; i < height && offset+1 < len(compressedData); i++ {
		byteCounts[i] = uint16(compressedData[offset])<<8 | uint16(compressedData[offset+1])
		offset += 2
	}

	// Decompress the RLE data
	result := make([]byte, width*height)
	pos := 0

	for row := 0; row < height; row++ {
		byteCount := int(byteCounts[row])
		if byteCount == 0 {
			// Empty scanline
			pos += width
			continue
		}

		endPos := pos + width
		scanlineEnd := offset + byteCount
		if scanlineEnd > len(compressedData) {
			scanlineEnd = len(compressedData)
		}

		// Decode RLE for this scanline
		for offset < scanlineEnd && pos < endPos {
			if offset >= len(compressedData) {
				break
			}

			length := int(compressedData[offset])
			offset++

			if length < 128 {
				// Copy next length+1 bytes literally
				length++
				for i := 0; i < length && pos < endPos && offset < len(compressedData); i++ {
					result[pos] = compressedData[offset]
					pos++
					offset++
				}
			} else if length > 128 {
				// Repeat next byte (257-length) times
				length = 257 - length
				if offset < len(compressedData) {
					val := compressedData[offset]
					offset++
					for i := 0; i < length && pos < endPos; i++ {
						result[pos] = val
						pos++
					}
				}
			}
			// length == 128 is a no-op
		}

		// If we didn't fill the scanline, skip to next scanline start
		if pos < endPos {
			pos = endPos
		}
	}

	return result, nil
}

// ToImage converts the layer to an image.RGBA
func (l *Layer) ToImage() (*image.RGBA, error) {
	width := int(l.Width())
	height := int(l.Height())

	if width == 0 || height == 0 {
		return nil, nil
	}

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Channel IDs: -1 = transparency, 0 = red, 1 = green, 2 = blue
	var rData, gData, bData, aData []byte

	// Get channel data
	if ch, exists := l.channels[-1]; exists {
		aData = ch.Data
	}
	if ch, exists := l.channels[0]; exists {
		rData = ch.Data
	}
	if ch, exists := l.channels[1]; exists {
		gData = ch.Data
	}
	if ch, exists := l.channels[2]; exists {
		bData = ch.Data
	}

	// Fill image with pixel data
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x

			var r, g, b, a uint8 = 0, 0, 0, 255

			if rData != nil && idx < len(rData) {
				r = rData[idx]
			}
			if gData != nil && idx < len(gData) {
				g = gData[idx]
			}
			if bData != nil && idx < len(bData) {
				b = bData[idx]
			}
			if aData != nil && idx < len(aData) {
				a = aData[idx]
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	return img, nil
}
