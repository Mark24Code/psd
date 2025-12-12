package psd

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// PSD represents a Photoshop document
type PSD struct {
	file      *File
	header    *Header
	resources *ResourceSection
	layerMask *LayerMask
	image     *Image
	parsed    bool
}

// New creates a new PSD instance from a file path
func New(filename string) (*PSD, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	file := &File{
		file:   f,
		reader: f,
	}

	return &PSD{
		file:   file,
		parsed: false,
	}, nil
}

// Open opens a PSD file, parses it, and executes the provided function
func Open(filename string, fn func(*PSD) error) error {
	psd, err := New(filename)
	if err != nil {
		return err
	}
	defer psd.Close()

	if err := psd.Parse(); err != nil {
		return err
	}

	return fn(psd)
}

// Close closes the underlying file
func (p *PSD) Close() error {
	if p.file != nil && p.file.file != nil {
		return p.file.file.Close()
	}
	return nil
}

// Parse parses all sections of the PSD file
func (p *PSD) Parse() error {
	if err := p.parseHeader(); err != nil {
		return fmt.Errorf("failed to parse header: %w", err)
	}

	if err := p.parseResources(); err != nil {
		return fmt.Errorf("failed to parse resources: %w", err)
	}

	if err := p.parseLayerMask(); err != nil {
		return fmt.Errorf("failed to parse layer mask: %w", err)
	}

	if err := p.parseImage(); err != nil {
		return fmt.Errorf("failed to parse image: %w", err)
	}

	p.parsed = true
	return nil
}

// Parsed returns whether the PSD has been parsed
func (p *PSD) Parsed() bool {
	return p.parsed
}

// Header returns the PSD header
func (p *PSD) Header() *Header {
	if p.header == nil {
		p.parseHeader()
	}
	return p.header
}

// Resources returns the resource section
func (p *PSD) Resources() *ResourceSection {
	if p.resources == nil {
		p.parseResources()
	}
	return p.resources
}

// LayerMask returns the layer mask section
func (p *PSD) LayerMask() *LayerMask {
	if p.layerMask == nil {
		p.parseLayerMask()
	}
	return p.layerMask
}

// Image returns the flattened image
func (p *PSD) Image() *Image {
	if p.image == nil {
		p.parseImage()
	}
	return p.image
}

// Layers returns all layers
func (p *PSD) Layers() []*Layer {
	if p.layerMask == nil {
		p.parseLayerMask()
	}
	return p.layerMask.Layers
}

// Tree returns the layer tree structure
func (p *PSD) Tree() *Node {
	if p.layerMask == nil {
		p.parseLayerMask()
	}
	return p.layerMask.Tree()
}

// LayerComps returns all layer comps
func (p *PSD) LayerComps() []LayerComp {
	if p.resources == nil {
		p.parseResources()
	}
	return p.resources.LayerComps()
}

// Slices returns all slices
func (p *PSD) Slices() (*SlicesResource, error) {
	if p.resources == nil {
		if err := p.parseResources(); err != nil {
			return nil, err
		}
	}
	return p.resources.ParseSlices()
}

// Guides returns all guides
func (p *PSD) Guides() (*GuidesResource, error) {
	if p.resources == nil {
		if err := p.parseResources(); err != nil {
			return nil, err
		}
	}
	return p.resources.ParseGuides()
}

func (p *PSD) parseHeader() error {
	if p.header != nil {
		return nil
	}

	header := &Header{file: p.file}
	if err := header.Parse(); err != nil {
		return err
	}

	p.header = header
	return nil
}

func (p *PSD) parseResources() error {
	if p.resources != nil {
		return nil
	}

	if p.header == nil {
		if err := p.parseHeader(); err != nil {
			return err
		}
	}

	resources := &ResourceSection{file: p.file}
	if err := resources.Parse(); err != nil {
		return err
	}

	p.resources = resources
	return nil
}

func (p *PSD) parseLayerMask() error {
	if p.layerMask != nil {
		return nil
	}

	if p.header == nil {
		if err := p.parseHeader(); err != nil {
			return err
		}
	}

	if p.resources == nil {
		if err := p.parseResources(); err != nil {
			return err
		}
	}

	layerMask := &LayerMask{file: p.file, header: p.header}
	if err := layerMask.Parse(); err != nil {
		return err
	}

	p.layerMask = layerMask
	return nil
}

func (p *PSD) parseImage() error {
	if p.image != nil {
		return nil
	}

	if p.header == nil {
		if err := p.parseHeader(); err != nil {
			return err
		}
	}

	if p.resources == nil {
		if err := p.parseResources(); err != nil {
			return err
		}
	}

	if p.layerMask == nil {
		if err := p.parseLayerMask(); err != nil {
			return err
		}
	}

	image := &Image{file: p.file, header: p.header}
	if err := image.Parse(); err != nil {
		return err
	}

	p.image = image
	return nil
}

// File represents a PSD file with convenience methods for reading binary data
type File struct {
	file   *os.File
	reader io.Reader
}

// Read reads bytes from the file
func (f *File) Read(p []byte) (n int, error error) {
	return io.ReadFull(f.reader, p)
}

// Seek seeks to a position in the file
func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

// Tell returns the current position in the file
func (f *File) Tell() (int64, error) {
	return f.file.Seek(0, io.SeekCurrent)
}

// ReadString reads a string of specified length
func (f *File) ReadString(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := f.Read(buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

// ReadByte reads a single byte
func (f *File) ReadByte() (byte, error) {
	buf := make([]byte, 1)
	if _, err := f.Read(buf); err != nil {
		return 0, err
	}
	return buf[0], nil
}

// ReadUint16 reads a 16-bit unsigned integer (big endian)
func (f *File) ReadUint16() (uint16, error) {
	buf := make([]byte, 2)
	if _, err := f.Read(buf); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buf), nil
}

// ReadInt16 reads a 16-bit signed integer (big endian)
func (f *File) ReadInt16() (int16, error) {
	buf := make([]byte, 2)
	if _, err := f.Read(buf); err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(buf)), nil
}

// ReadUint32 reads a 32-bit unsigned integer (big endian)
func (f *File) ReadUint32() (uint32, error) {
	buf := make([]byte, 4)
	if _, err := f.Read(buf); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf), nil
}

// ReadInt32 reads a 32-bit signed integer (big endian)
func (f *File) ReadInt32() (int32, error) {
	buf := make([]byte, 4)
	if _, err := f.Read(buf); err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(buf)), nil
}

// ReadUint64 reads a 64-bit unsigned integer (big endian)
func (f *File) ReadUint64() (uint64, error) {
	buf := make([]byte, 8)
	if _, err := f.Read(buf); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buf), nil
}

// Skip skips n bytes
func (f *File) Skip(n int64) error {
	_, err := f.Seek(n, io.SeekCurrent)
	return err
}
