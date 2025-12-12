package psd

import (
	"fmt"
	"image"
	"image/color"
)

// Image represents the flattened preview image
type Image struct {
	file      *File
	header    *Header
	width     uint32
	height    uint32
	pixelData []color.RGBA
	parsed    bool
}

// Parse parses the image data
func (img *Image) Parse() error {
	if img.parsed {
		return nil
	}

	img.width = img.header.Width()
	img.height = img.header.Height()

	// Read compression method
	compression, err := img.file.ReadUint16()
	if err != nil {
		return fmt.Errorf("failed to read compression: %w", err)
	}

	totalPixels := int(img.width * img.height)
	img.pixelData = make([]color.RGBA, totalPixels)

	switch compression {
	case 0: // Raw data
		if err := img.parseRaw(); err != nil {
			return err
		}
	case 1: // RLE compression
		if err := img.parseRLE(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported compression method: %d", compression)
	}

	img.parsed = true
	return nil
}

func (img *Image) parseRaw() error {
	channels := int(img.header.Channels)
	totalPixels := int(img.width * img.height)

	// Read channel data
	channelData := make([][]byte, channels)
	for i := 0; i < channels; i++ {
		channelData[i] = make([]byte, totalPixels)
		if _, err := img.file.Read(channelData[i]); err != nil {
			return fmt.Errorf("failed to read channel %d: %w", i, err)
		}
	}

	// Convert to RGBA
	for i := 0; i < totalPixels; i++ {
		if img.header.IsRGB() && channels >= 3 {
			img.pixelData[i] = color.RGBA{
				R: channelData[0][i],
				G: channelData[1][i],
				B: channelData[2][i],
				A: 255,
			}
		} else if channels == 1 {
			// Grayscale
			gray := channelData[0][i]
			img.pixelData[i] = color.RGBA{R: gray, G: gray, B: gray, A: 255}
		}
	}

	return nil
}

func (img *Image) parseRLE() error {
	channels := int(img.header.Channels)
	height := int(img.height)
	width := int(img.width)

	// Read byte counts for each scanline
	totalScanlines := channels * height
	byteCounts := make([]uint16, totalScanlines)
	for i := 0; i < totalScanlines; i++ {
		count, err := img.file.ReadUint16()
		if err != nil {
			return fmt.Errorf("failed to read byte count: %w", err)
		}
		byteCounts[i] = count
	}

	// Decode RLE data for each channel
	channelData := make([][]byte, channels)
	for ch := 0; ch < channels; ch++ {
		channelData[ch] = make([]byte, width*height)

		pos := 0
		for row := 0; row < height; row++ {
			scanlineIdx := ch*height + row
			byteCount := int(byteCounts[scanlineIdx])

			if byteCount == 0 {
				pos += width
				continue
			}

			endPos := pos + width
			scanlineData := make([]byte, byteCount)
			if _, err := img.file.Read(scanlineData); err != nil {
				return fmt.Errorf("failed to read RLE scanline: %w", err)
			}

			// Decode RLE
			dataIdx := 0
			for pos < endPos && dataIdx < len(scanlineData) {
				length := int(scanlineData[dataIdx])
				dataIdx++

				if length < 128 {
					// Copy next length+1 bytes literally
					length++
					for i := 0; i < length && pos < endPos && dataIdx < len(scanlineData); i++ {
						channelData[ch][pos] = scanlineData[dataIdx]
						pos++
						dataIdx++
					}
				} else if length > 128 {
					// Repeat next byte (257-length) times
					length = 257 - length
					if dataIdx < len(scanlineData) {
						val := scanlineData[dataIdx]
						dataIdx++
						for i := 0; i < length && pos < endPos; i++ {
							channelData[ch][pos] = val
							pos++
						}
					}
				}
				// length == 128 is a no-op
			}
		}
	}

	// Convert to RGBA
	totalPixels := width * height
	for i := 0; i < totalPixels; i++ {
		if img.header.IsRGB() && channels >= 3 {
			img.pixelData[i] = color.RGBA{
				R: channelData[0][i],
				G: channelData[1][i],
				B: channelData[2][i],
				A: 255,
			}
		} else if channels == 1 {
			gray := channelData[0][i]
			img.pixelData[i] = color.RGBA{R: gray, G: gray, B: gray, A: 255}
		}
	}

	return nil
}

// Width returns the image width
func (img *Image) Width() uint32 {
	if !img.parsed {
		img.width = img.header.Width()
	}
	return img.width
}

// Height returns the image height
func (img *Image) Height() uint32 {
	if !img.parsed {
		img.height = img.header.Height()
	}
	return img.height
}

// PixelData returns the raw pixel data
func (img *Image) PixelData() []color.RGBA {
	if !img.parsed {
		img.Parse()
	}
	return img.pixelData
}

// ToPNG converts the image to a Go image.Image
func (img *Image) ToPNG() *image.RGBA {
	if !img.parsed {
		img.Parse()
	}

	bounds := image.Rect(0, 0, int(img.width), int(img.height))
	rgba := image.NewRGBA(bounds)

	for y := 0; y < int(img.height); y++ {
		for x := 0; x < int(img.width); x++ {
			idx := y*int(img.width) + x
			rgba.Set(x, y, img.pixelData[idx])
		}
	}

	return rgba
}
