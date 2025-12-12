package psd

import (
	"fmt"
	"testing"
)

func TestLayerExport(t *testing.T) {
	err := Open("testdata/example.psd", func(p *PSD) error {
		tree := p.Tree()
		if tree == nil {
			return fmt.Errorf("tree is nil")
		}

		// Find a specific layer
		var testNode *Node
		for _, node := range tree.Descendants() {
			if node.Type == NodeTypeLayer && node.Name == "Matte" {
				testNode = node
				break
			}
		}

		if testNode == nil {
			return fmt.Errorf("could not find Matte layer")
		}

		t.Logf("Found layer: %s", testNode.Name)
		t.Logf("Layer dimensions: %dx%d", testNode.Width(), testNode.Height())

		if testNode.Layer == nil {
			return fmt.Errorf("testNode.Layer is nil")
		}

		t.Logf("Layer has %d channel(s)", len(testNode.Layer.channels))
		t.Logf("Layer ChannelInfo: %+v", testNode.Layer.ChannelInfo)

		// Try calling parseChannelData again to see if it helps
		// This should already have been called, but let's verify
		t.Logf("Attempting to manually call parseChannelData...")
		// Note: We can't call this because we don't have access to the file
		// The issue is that channels should already be populated

		// Check ChannelData map
		t.Logf("ChannelData map has %d entries", len(testNode.Layer.ChannelData))
		for id, data := range testNode.Layer.ChannelData {
			t.Logf("  Channel %d: %d bytes", id, len(data))
		}

		// Try to get image
		img, err := testNode.Layer.ToImage()
		if err != nil {
			return fmt.Errorf("ToImage failed: %w", err)
		}

		if img == nil {
			return fmt.Errorf("ToImage returned nil")
		}

		// Check if image has any non-zero pixels
		bounds := img.Bounds()
		hasColor := false
		for y := bounds.Min.Y; y < bounds.Max.Y && !hasColor; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				if r > 0 || g > 0 || b > 0 {
					hasColor = true
					t.Logf("Found color pixel at (%d,%d): R=%d, G=%d, B=%d", x, y, r>>8, g>>8, b>>8)
					break
				}
			}
		}

		if !hasColor {
			return fmt.Errorf("image is completely black")
		}

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}
