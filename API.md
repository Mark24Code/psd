# PSD Library API Documentation

## Overview

A high-performance Go implementation of PSD (Photoshop Document) file parser. This library provides complete parsing of PSD files including layers, groups, resources, and image data, with rendering capabilities.

## Installation

```go
import psd "github.com/layervault/psd.rb/go"
```

Or install via go get:
```bash
go get github.com/layervault/psd.rb/go
```

## Quick Start

```go
package main

import (
    "fmt"
    psd "github.com/layervault/psd.rb/go"
)

func main() {
    // Open and parse a PSD file
    err := psd.Open("file.psd", func(p *psd.PSD) error {
        // Access header information
        h := p.Header()
        fmt.Printf("Size: %dx%d\n", h.Width(), h.Height())
        fmt.Printf("Color Mode: %s\n", h.ModeName())

        // Access layers
        for _, layer := range p.Layers() {
            fmt.Printf("Layer: %s\n", layer.Name)
        }

        return nil
    })

    if err != nil {
        panic(err)
    }
}
```

## Core Types

### PSD

The main entry point for working with PSD files.

#### Functions

**`New(filename string) (*PSD, error)`**

Creates a new PSD instance from a file path. The file is opened but not parsed.

```go
psd, err := psd.New("file.psd")
if err != nil {
    return err
}
defer psd.Close()
```

**`Open(filename string, fn func(*PSD) error) error`**

Opens a PSD file, parses it, and executes the provided function. The file is automatically closed after the function returns.

```go
err := psd.Open("file.psd", func(p *psd.PSD) error {
    // Work with the parsed PSD
    return nil
})
```

#### Methods

**`Parse() error`**

Parses all sections of the PSD file (header, resources, layer mask, image data).

**`Close() error`**

Closes the underlying file.

**`Header() *Header`**

Returns the PSD header information. Lazily parses if not already parsed.

**`Layers() []*Layer`**

Returns all layers in the document (flattened list, top to bottom).

**`Tree() *Node`**

Returns the layer tree structure with groups and hierarchy.

**`LayerComps() []LayerComp`**

Returns all layer compositions defined in the document.

**`Slices() (*SlicesResource, error)`**

Returns slice information from the document.

**`Guides() (*GuidesResource, error)`**

Returns guide information from the document.

**`Image() *Image`**

Returns the flattened preview image.

---

### Header

Represents the PSD file header containing document metadata.

#### Fields

- `Version uint16` - PSD format version (1 for PSD, 2 for PSB)
- `Channels uint16` - Number of color channels (1-56)
- `Rows uint32` - Height in pixels
- `Cols uint32` - Width in pixels
- `Depth uint16` - Bits per channel (1, 8, 16, or 32)
- `Mode uint16` - Color mode (RGB, CMYK, etc.)

#### Methods

**`Width() uint32`**

Returns the document width in pixels.

**`Height() uint32`**

Returns the document height in pixels.

**`ModeName() string`**

Returns the human-readable color mode name (e.g., "RGBColor", "CMYKColor").

**`IsRGB() bool`**

Returns true if the color mode is RGB.

**`IsCMYK() bool`**

Returns true if the color mode is CMYK.

**`IsBig() bool`**

Returns true if this is a PSB (large document format).

---

### Layer

Represents a single layer in the PSD document.

#### Fields

- `Name string` - Layer name
- `Top, Left, Bottom, Right int32` - Layer bounds
- `Opacity uint8` - Layer opacity (0-255)
- `BlendModeKey string` - Blend mode key
- `Flags uint8` - Layer flags
- `Channels uint16` - Number of channels
- `ChannelData map[int16][]byte` - Raw channel pixel data

#### Methods

**`Width() int32`**

Returns the layer width.

**`Height() int32`**

Returns the layer height.

**`Visible() bool`**

Returns whether the layer is visible.

**`IsFolder() bool`**

Returns whether this layer is a folder/group.

**`IsFolderEnd() bool`**

Returns whether this is a folder end marker.

**`BlendMode() *BlendMode`**

Returns blend mode information including mode name, opacity, and visibility.

**`ToImage() (*image.RGBA, error)`**

Converts the layer to an RGBA image.

---

### Node

Represents a node in the layer tree structure (root, group, or layer).

#### Fields

- `Type string` - Node type ("root", "group", or "layer")
- `Name string` - Node name
- `Layer *Layer` - Associated layer (nil for root and groups without layers)
- `Parent *Node` - Parent node
- `Children []*Node` - Child nodes
- `Visible bool` - Visibility
- `Opacity uint8` - Opacity
- `BlendMode string` - Blend mode
- `Left, Top, Right, Bottom int32` - Bounding box

#### Methods

**Tree Traversal:**

- `Root() *Node` - Returns the root node
- `IsRoot() bool` - Returns whether this is the root node
- `HasChildren() bool` - Returns whether this node has children
- `Descendants() []*Node` - Returns all descendant nodes
- `Subtree() []*Node` - Returns all nodes in the subtree (including this node)
- `Siblings() []*Node` - Returns all siblings including this node
- `Depth() int` - Returns the depth of this node in the tree

**Filtering:**

- `DescendantLayers() []*Node` - Returns all descendant layer nodes
- `DescendantGroups() []*Node` - Returns all descendant group nodes
- `SubtreeLayers() []*Node` - Returns all layer nodes in the subtree
- `SubtreeGroups() []*Node` - Returns all group nodes in the subtree

**Searching:**

- `ChildrenAtPath(path interface{}) []*Node` - Finds nodes at the given path
- `Path(asArray ...bool) interface{}` - Returns the path to this node

**Dimensions:**

- `Width() int32` - Returns the node width
- `Height() int32` - Returns the node height
- `IsEmpty() bool` - Returns whether the node is empty (zero size)

**Export:**

- `ToPNG() (*image.RGBA, error)` - Renders the node to a PNG image
- `SaveAsPNG(filename string) error` - Renders and saves the node as a PNG file

**Other:**

- `ToHash() map[string]interface{}` - Converts the node tree to a hash/map structure
- `IsVisible() bool` - Returns whether the node is visible

---

### Image

Represents the flattened preview image of the document.

#### Methods

**`Width() uint32`**

Returns the image width.

**`Height() uint32`**

Returns the image height.

**`PixelData() []color.RGBA`**

Returns the raw pixel data as a slice of RGBA colors.

**`ToPNG() *image.RGBA`**

Converts the image to a Go image.RGBA.

---

### Resources

#### SlicesResource

Represents slice information (Resource ID 1050).

**Fields:**
- `Version int32` - Slice format version
- `Bounds Rectangle` - Slice bounding rectangle
- `Name string` - Slice group name
- `Slices []Slice` - Individual slices

#### GuidesResource

Represents guide information (Resource ID 1032).

**Fields:**
- `Guides []Guide` - List of guides

**Guide Fields:**
- `Position int32` - Guide position in pixels
- `IsHorizontal bool` - Whether the guide is horizontal

#### LayerComp

Represents a layer composition.

**Fields:**
- `ID int` - Layer comp ID
- `Name string` - Layer comp name

---

## Examples

### Example 1: Parse and Extract Information

```go
package main

import (
    "fmt"
    psd "github.com/layervault/psd.rb/go"
)

func main() {
    err := psd.Open("design.psd", func(p *psd.PSD) error {
        // Get document info
        h := p.Header()
        fmt.Printf("Document: %dx%d pixels\n", h.Width(), h.Height())
        fmt.Printf("Color Mode: %s\n", h.ModeName())
        fmt.Printf("Bit Depth: %d\n", h.Depth)
        fmt.Printf("Channels: %d\n", h.Channels)

        // List all layers
        fmt.Printf("\nLayers (%d):\n", len(p.Layers()))
        for i, layer := range p.Layers() {
            visible := "hidden"
            if layer.Visible() {
                visible = "visible"
            }
            fmt.Printf("  %d. %s (%s, opacity: %d%%)\n",
                i+1, layer.Name, visible, layer.Opacity*100/255)
        }

        return nil
    })

    if err != nil {
        panic(err)
    }
}
```

### Example 2: Export Layers as PNG

```go
package main

import (
    "fmt"
    psd "github.com/layervault/psd.rb/go"
)

func main() {
    psd.Open("design.psd", func(p *psd.PSD) error {
        // Export full document
        img := p.Image()
        img.ToPNG().SavePNG("output.png")

        // Export individual layers
        for _, layer := range p.Layers() {
            if layer.IsFolder() {
                continue
            }

            layerImg, err := layer.ToImage()
            if err != nil {
                continue
            }

            filename := fmt.Sprintf("%s.png", layer.Name)
            // Save layer image...
            _ = layerImg
            _ = filename
        }

        return nil
    })
}
```

### Example 3: Traverse Layer Tree

```go
package main

import (
    "fmt"
    "strings"
    psd "github.com/layervault/psd.rb/go"
)

func main() {
    psd.Open("design.psd", func(p *psd.PSD) error {
        tree := p.Tree()

        // Print tree structure
        printNode(tree, 0)

        // Find specific layer by path
        nodes := tree.ChildrenAtPath("Folder/Sublayer")
        if len(nodes) > 0 {
            fmt.Printf("\nFound node at path: %s\n", nodes[0].Name)
        }

        // Get all visible layers
        for _, node := range tree.DescendantLayers() {
            if node.Visible {
                fmt.Printf("Visible layer: %s\n", node.Name)
            }
        }

        return nil
    })
}

func printNode(node *psd.Node, depth int) {
    indent := strings.Repeat("  ", depth)
    fmt.Printf("%s%s [%s]\n", indent, node.Name, node.Type)

    for _, child := range node.Children {
        printNode(child, depth+1)
    }
}
```

### Example 4: Access Resources (Guides, Slices)

```go
package main

import (
    "fmt"
    psd "github.com/layervault/psd.rb/go"
)

func main() {
    psd.Open("design.psd", func(p *psd.PSD) error {
        // Get guides
        guides, err := p.Guides()
        if err == nil && len(guides.Guides) > 0 {
            fmt.Printf("Guides (%d):\n", len(guides.Guides))
            for _, guide := range guides.Guides {
                orientation := "vertical"
                if guide.IsHorizontal {
                    orientation = "horizontal"
                }
                fmt.Printf("  - %s at %dpx\n", orientation, guide.Position)
            }
        }

        // Get slices
        slices, err := p.Slices()
        if err == nil && len(slices.Slices) > 0 {
            fmt.Printf("\nSlices (%d):\n", len(slices.Slices))
            for _, slice := range slices.Slices {
                fmt.Printf("  - %s\n", slice.Name)
            }
        }

        return nil
    })
}
```

### Example 5: Render Node to PNG

```go
package main

import (
    psd "github.com/layervault/psd.rb/go"
)

func main() {
    psd.Open("design.psd", func(p *psd.PSD) error {
        tree := p.Tree()

        // Render specific folder/group
        nodes := tree.ChildrenAtPath("Background")
        if len(nodes) > 0 {
            nodes[0].SaveAsPNG("background.png")
        }

        // Render all top-level groups
        for i, child := range tree.Children {
            filename := fmt.Sprintf("group_%d_%s.png", i, child.Name)
            child.SaveAsPNG(filename)
        }

        return nil
    })
}
```

## Color Modes

Supported color modes (use with `Header.Mode`):

- `ColorModeRGBColor` (3) - RGB Color
- `ColorModeCMYKColor` (4) - CMYK Color
- `ColorModeGrayscale` (1) - Grayscale
- `ColorModeBitmap` (0) - Bitmap
- `ColorModeIndexedColor` (2) - Indexed Color
- `ColorModeLabColor` (9) - Lab Color
- `ColorModeDuotone` (8) - Duotone
- `ColorModeMultichannel` (7) - Multichannel

## Blend Modes

Supported blend modes (returned by `Layer.BlendMode().Mode`):

- normal, dissolve
- multiply, screen, overlay
- soft_light, hard_light
- color_dodge, color_burn, linear_burn, linear_dodge
- darken, lighten
- difference, exclusion
- hue, saturation, color, luminosity
- vivid_light, linear_light, pin_light, hard_mix
- lighter_color, darker_color
- subtract, divide

## Performance Characteristics

- **Fast Parsing**: Optimized binary parsing with minimal allocations
- **Lazy Loading**: Sections are only parsed when accessed
- **Memory Efficient**: Streams data from disk, doesn't load entire file into memory
- **Concurrent Safe**: Can parse multiple PSD files in parallel (each PSD instance is not thread-safe)

## Limitations

- **Rendering**: Only basic normal blend mode is fully supported in rendering engine
- **Text Layers**: Text content is not extracted (engine data parsing not implemented)
- **Layer Styles**: Layer effects (drop shadow, stroke, etc.) are not applied during rendering
- **Adjustment Layers**: Adjustment layer effects are not applied
- **Smart Objects**: Smart object contents are not extracted
- **PSB Files**: Large document format (PSB) is partially supported

## Error Handling

All file operations and parsing return errors that should be checked:

```go
psd, err := psd.New("file.psd")
if err != nil {
    // Handle file open error
    return err
}
defer psd.Close()

err = psd.Parse()
if err != nil {
    // Handle parsing error
    return err
}
```

## Thread Safety

- Each `PSD` instance is **not** thread-safe
- You can safely parse multiple PSD files concurrently using separate goroutines
- Don't share a single `PSD` instance across goroutines without synchronization

## Dependencies

- **Runtime**: None (uses only Go standard library)
- **Testing**: `github.com/stretchr/testify` (for tests only)

## Requirements

- Go 1.21 or later

## License

MIT License - See LICENSE file for details

## Contributing

Contributions are welcome! Please ensure:
- All tests pass (`go test -v`)
- Code is formatted (`gofmt`)
- No vet warnings (`go vet`)
- Public APIs have godoc comments

## Support

For issues, questions, or contributions, please visit:
https://github.com/layervault/psd.rb
