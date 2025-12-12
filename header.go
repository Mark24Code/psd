package psd

import (
	"fmt"
)

// Header represents the PSD file header
type Header struct {
	file     *File
	Sig      string
	Version  uint16
	Channels uint16
	Rows     uint32
	Cols     uint32
	Depth    uint16
	Mode     uint16
}

// Color modes
const (
	ColorModeBitmap           = 0
	ColorModeGrayscale        = 1
	ColorModeIndexedColor     = 2
	ColorModeRGBColor         = 3
	ColorModeCMYKColor        = 4
	ColorModeHSLColor         = 5
	ColorModeHSBColor         = 6
	ColorModeMultichannel     = 7
	ColorModeDuotone          = 8
	ColorModeLabColor         = 9
	ColorModeGray16           = 10
	ColorModeRGB48            = 11
	ColorModeLab48            = 12
	ColorModeCMYK64           = 13
	ColorModeDeepMultichannel = 14
	ColorModeDuotone16        = 15
)

var colorModeNames = []string{
	"Bitmap",
	"GrayScale",
	"IndexedColor",
	"RGBColor",
	"CMYKColor",
	"HSLColor",
	"HSBColor",
	"Multichannel",
	"Duotone",
	"LabColor",
	"Gray16",
	"RGB48",
	"Lab48",
	"CMYK64",
	"DeepMultichannel",
	"Duotone16",
}

// Width returns the width of the document
func (h *Header) Width() uint32 {
	return h.Cols
}

// Height returns the height of the document
func (h *Header) Height() uint32 {
	return h.Rows
}

// ModeName returns the human-readable color mode name
func (h *Header) ModeName() string {
	if int(h.Mode) < len(colorModeNames) {
		return colorModeNames[h.Mode]
	}
	return fmt.Sprintf("Unknown(%d)", h.Mode)
}

// IsBig returns true if this is a PSB (large document format)
func (h *Header) IsBig() bool {
	return h.Version == 2
}

// IsRGB returns true if the color mode is RGB
func (h *Header) IsRGB() bool {
	return h.Mode == ColorModeRGBColor
}

// IsCMYK returns true if the color mode is CMYK
func (h *Header) IsCMYK() bool {
	return h.Mode == ColorModeCMYKColor
}

// Parse parses the header section
func (h *Header) Parse() error {
	// Read signature (4 bytes)
	sig, err := h.file.ReadString(4)
	if err != nil {
		return fmt.Errorf("failed to read signature: %w", err)
	}
	if sig != "8BPS" {
		return fmt.Errorf("invalid PSD signature: %s", sig)
	}
	h.Sig = sig

	// Read version (2 bytes)
	version, err := h.file.ReadUint16()
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}
	if version != 1 && version != 2 {
		return fmt.Errorf("unsupported PSD version: %d", version)
	}
	h.Version = version

	// Skip reserved bytes (6 bytes)
	if err := h.file.Skip(6); err != nil {
		return fmt.Errorf("failed to skip reserved bytes: %w", err)
	}

	// Read number of channels (2 bytes)
	channels, err := h.file.ReadUint16()
	if err != nil {
		return fmt.Errorf("failed to read channels: %w", err)
	}
	h.Channels = channels

	// Read height (4 bytes)
	rows, err := h.file.ReadUint32()
	if err != nil {
		return fmt.Errorf("failed to read rows: %w", err)
	}
	h.Rows = rows

	// Read width (4 bytes)
	cols, err := h.file.ReadUint32()
	if err != nil {
		return fmt.Errorf("failed to read cols: %w", err)
	}
	h.Cols = cols

	// Read depth (2 bytes)
	depth, err := h.file.ReadUint16()
	if err != nil {
		return fmt.Errorf("failed to read depth: %w", err)
	}
	h.Depth = depth

	// Read color mode (2 bytes)
	mode, err := h.file.ReadUint16()
	if err != nil {
		return fmt.Errorf("failed to read mode: %w", err)
	}
	h.Mode = mode

	// Read and skip color mode data
	colorDataLen, err := h.file.ReadUint32()
	if err != nil {
		return fmt.Errorf("failed to read color data length: %w", err)
	}
	if colorDataLen > 0 {
		if err := h.file.Skip(int64(colorDataLen)); err != nil {
			return fmt.Errorf("failed to skip color data: %w", err)
		}
	}

	return nil
}
