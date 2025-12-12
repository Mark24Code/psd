package psd

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// Renderer handles rendering nodes to images
type Renderer struct {
	node   *Node
	canvas *image.RGBA
}

// NewRenderer creates a new renderer for the given node
func NewRenderer(node *Node) *Renderer {
	width := int(node.Width())
	height := int(node.Height())

	// Create canvas with proper bounds
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))

	return &Renderer{
		node:   node,
		canvas: canvas,
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

// renderLayer renders a single layer to the canvas
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

	// Composite layer onto canvas using normal blend mode
	for y := layerBounds.Min.Y; y < layerBounds.Max.Y; y++ {
		for x := layerBounds.Min.X; x < layerBounds.Max.X; x++ {
			// Calculate destination position
			dstX := canvasX + x
			dstY := canvasY + y

			// Check if within canvas bounds
			if dstX < 0 || dstY < 0 || dstX >= r.canvas.Bounds().Dx() || dstY >= r.canvas.Bounds().Dy() {
				continue
			}

			// Get source and destination colors
			srcColor := layerImg.At(x, y)
			dstColor := r.canvas.At(dstX, dstY)

			// Blend using normal blend mode with layer opacity
			blended := r.blendNormal(srcColor, dstColor, layer.Opacity)

			r.canvas.Set(dstX, dstY, blended)
		}
	}

	return nil
}

// blendNormal performs normal blend mode (source over)
func (r *Renderer) blendNormal(src, dst color.Color, opacity uint8) color.RGBA {
	sr, sg, sb, sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	// Apply layer opacity
	alpha := uint32(opacity) * sa / 255 / 257

	if alpha == 0 {
		return color.RGBA{uint8(dr >> 8), uint8(dg >> 8), uint8(db >> 8), uint8(da >> 8)}
	}

	if alpha == 255 && da == 0 {
		return color.RGBA{uint8(sr >> 8), uint8(sg >> 8), uint8(sb >> 8), uint8(alpha)}
	}

	// Alpha compositing: C = (Cs * As + Cd * Ad * (1 - As)) / Ao
	// where Ao = As + Ad * (1 - As)
	outAlpha := alpha + (da*(255-alpha))/255

	if outAlpha == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	// Convert from 16-bit to 8-bit color space
	sr8, sg8, sb8 := sr>>8, sg>>8, sb>>8
	dr8, dg8, db8 := dr>>8, dg>>8, db>>8

	// Blend colors
	outRed := (sr8*alpha + dr8*da*(255-alpha)/255) / outAlpha
	outGreen := (sg8*alpha + dg8*da*(255-alpha)/255) / outAlpha
	outBlue := (sb8*alpha + db8*da*(255-alpha)/255) / outAlpha

	return color.RGBA{
		uint8(outRed),
		uint8(outGreen),
		uint8(outBlue),
		uint8(outAlpha),
	}
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
