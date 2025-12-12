package psd

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// RendererOptions contains options for rendering
type RendererOptions struct {
	ExcludeTextLayers bool     // Exclude text layers from rendering
	ExcludeTypes      []string // Exclude specific node types
}

// Renderer handles rendering nodes to images
type Renderer struct {
	node    *Node
	canvas  *image.RGBA
	options RendererOptions
}

// NewRenderer creates a new renderer for the given node
func NewRenderer(node *Node) *Renderer {
	return NewRendererWithOptions(node, RendererOptions{})
}

// NewRendererWithOptions creates a new renderer with options
func NewRendererWithOptions(node *Node, options RendererOptions) *Renderer {
	width := int(node.Width())
	height := int(node.Height())

	// Create canvas with proper bounds
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))

	return &Renderer{
		node:    node,
		canvas:  canvas,
		options: options,
	}
}

// Render renders the node and all its children to an image
func (r *Renderer) Render() (*image.RGBA, error) {
	// Clear canvas with transparent background
	for y := 0; y < r.canvas.Bounds().Dy(); y++ {
		for x := 0; x < r.canvas.Bounds().Dx(); x++ {
			r.canvas.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	// Render the node
	if err := r.renderNode(r.node, 0, 0); err != nil {
		return nil, err
	}

	return r.canvas, nil
}

// renderNode recursively renders a node and its children
func (r *Renderer) renderNode(node *Node, offsetX, offsetY int32) error {
	if !node.Visible {
		return nil
	}

	// Apply filters
	if r.shouldExcludeNode(node) {
		return nil
	}

	if node.Type == NodeTypeLayer {
		// Render layer
		if node.Layer != nil {
			return r.renderLayer(node.Layer, offsetX, offsetY)
		}
	} else if node.Type == NodeTypeGroup || node.Type == NodeTypeRoot {
		// Render children in reverse order (bottom to top)
		for i := len(node.Children) - 1; i >= 0; i-- {
			child := node.Children[i]
			if err := r.renderNode(child, offsetX, offsetY); err != nil {
				return err
			}
		}
	}

	return nil
}

// shouldExcludeNode checks if a node should be excluded based on options
func (r *Renderer) shouldExcludeNode(node *Node) bool {
	// Check if text layers should be excluded
	if r.options.ExcludeTextLayers && node.IsTextLayer() {
		return true
	}

	// Check if node type is in exclusion list
	for _, excludeType := range r.options.ExcludeTypes {
		if node.Type == excludeType {
			return true
		}
	}

	return false
}

// renderLayer renders a single layer to the canvas
// This matches Ruby's Blender.compose! method (blender.rb:18-42)
func (r *Renderer) renderLayer(layer *Layer, offsetX, offsetY int32) error {
	// Skip if layer has no image data
	if len(layer.channels) == 0 {
		return nil
	}

	// Get layer image
	layerImg, err := layer.ToImage()
	if err != nil {
		return fmt.Errorf("failed to get layer image: %w", err)
	}

	if layerImg == nil {
		return nil
	}

	// Calculate position on canvas
	// The renderer's canvas starts at node's top-left corner (0,0)
	// Layer positions are relative to the PSD document
	// We need to adjust layer position relative to the node being rendered
	canvasX := int(layer.Left - r.node.Left + offsetX)
	canvasY := int(layer.Top - r.node.Top + offsetY)

	// Get layer bounds
	layerBounds := layerImg.Bounds()

	// Calculate opacity using Ruby's formula:
	// calculated_opacity = opacity * fill_opacity / 255
	// This matches Ruby's Blender.calculated_opacity (blender.rb:50)
	calculatedOpacity := uint8((uint32(layer.Opacity) * uint32(layer.FillOpacity())) / 255)

	// Get mask data if present
	// This matches Ruby's Canvas.apply_masks (canvas.rb:52-55)
	var maskData []byte
	if layer.Mask != nil && !layer.Mask.IsEmpty() {
		if ch, exists := layer.channels[-2]; exists {
			maskData = ch.Data
		}
	}

	// Composite layer onto canvas pixel by pixel
	// This matches Ruby's Blender.compose! loop (blender.rb:30-41)
	for y := layerBounds.Min.Y; y < layerBounds.Max.Y; y++ {
		for x := layerBounds.Min.X; x < layerBounds.Max.X; x++ {
			// Calculate destination position
			dstX := canvasX + x
			dstY := canvasY + y

			// Check if within canvas bounds
			// This matches Ruby's: next if base_x < 0 || base_y < 0 || ...
			if dstX < 0 || dstY < 0 || dstX >= r.canvas.Bounds().Dx() || dstY >= r.canvas.Bounds().Dy() {
				continue
			}

			// Get source color
			srcColor := layerImg.At(x, y)

			// Apply mask if present (matches Ruby's Mask.apply! in mask.rb:23-47)
			if maskData != nil {
				maskWidth := int(layer.Mask.Width())
				maskHeight := int(layer.Mask.Height())

				// Calculate mask coordinates relative to layer
				// This matches Ruby's: mask_x = x + @left - @mask.left
				maskX := x - int(layer.Mask.Left-layer.Left)
				maskY := y - int(layer.Mask.Top-layer.Top)

				// Apply mask to alpha
				r, g, b, a := srcColor.RGBA()
				if maskX < 0 || maskX >= maskWidth || maskY < 0 || maskY >= maskHeight {
					// Outside mask bounds = fully transparent
					// This matches Ruby's: color[3] = 0
					a = 0
				} else {
					maskIdx := maskY*maskWidth + maskX
					if maskIdx < len(maskData) {
						// Apply mask value to alpha
						// This matches Ruby's: color[3] = color[3] * @mask_data[@mask_width * mask_y + mask_x] / 255
						a = (a >> 8) * uint32(maskData[maskIdx]) / 255
					}
				}
				srcColor = color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a)}
			}

			// Get destination color
			dstColor := r.canvas.At(dstX, dstY)

			// Get blend function based on layer's blend mode
			// This matches Ruby's: Compose.send(fg.node.blending_mode, ...)
			blendFunc := GetBlendFunc(layer.BlendModeKey)
			blended := blendFunc(srcColor, dstColor, calculatedOpacity)

			r.canvas.Set(dstX, dstY, blended)
		}
	}

	return nil
}

// ToPNG renders the node to a PNG image
func (n *Node) ToPNG() (*image.RGBA, error) {
	renderer := NewRenderer(n)
	return renderer.Render()
}

// SaveAsPNG renders the node and saves it as a PNG file
func (n *Node) SaveAsPNG(filename string) error {
	img, err := n.ToPNG()
	if err != nil {
		return fmt.Errorf("failed to render node: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}
