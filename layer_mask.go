package psd

import (
	"fmt"
)

// LayerMask represents the layer and mask information section
type LayerMask struct {
	file   *File
	header *Header
	Layers []*Layer
	tree   *Node
}

// Parse parses the layer and mask section
func (lm *LayerMask) Parse() error {
	// Read layer and mask information section length
	length, err := lm.file.ReadUint32()
	if err != nil {
		return fmt.Errorf("failed to read layer mask length: %w", err)
	}

	if length == 0 {
		lm.Layers = []*Layer{}
		return nil
	}

	// Mark start position
	startPos, err := lm.file.Tell()
	if err != nil {
		return err
	}
	endPos := startPos + int64(length)

	// Parse layer info
	if err := lm.parseLayerInfo(); err != nil {
		return fmt.Errorf("failed to parse layer info: %w", err)
	}

	// Skip to end of section if needed
	currentPos, err := lm.file.Tell()
	if err != nil {
		return err
	}
	if currentPos < endPos {
		if err := lm.file.Skip(endPos - currentPos); err != nil {
			return err
		}
	}

	// Build tree structure
	lm.buildTree()

	return nil
}

func (lm *LayerMask) parseLayerInfo() error {
	// Read layer info section length
	length, err := lm.file.ReadUint32()
	if err != nil {
		return err
	}

	if length == 0 {
		lm.Layers = []*Layer{}
		return nil
	}

	// Read layer count
	layerCount, err := lm.file.ReadInt16()
	if err != nil {
		return err
	}

	// Negative layer count means first alpha channel contains transparency data
	if layerCount < 0 {
		layerCount = -layerCount
	}

	lm.Layers = make([]*Layer, layerCount)

	// Parse layer records
	for i := int16(0); i < layerCount; i++ {
		layer := &Layer{
			file:   lm.file,
			header: lm.header,
		}
		if err := layer.parseRecord(); err != nil {
			return fmt.Errorf("failed to parse layer %d: %w", i, err)
		}
		lm.Layers[i] = layer
	}

	// Parse layer channel image data
	for _, layer := range lm.Layers {
		if err := layer.parseChannelData(); err != nil {
			return fmt.Errorf("failed to parse channel data for layer %s: %w", layer.Name, err)
		}
	}

	// Reverse layers array to match Ruby's order (top to bottom instead of bottom to top)
	for i, j := 0, len(lm.Layers)-1; i < j; i, j = i+1, j-1 {
		lm.Layers[i], lm.Layers[j] = lm.Layers[j], lm.Layers[i]
	}

	return nil
}

func (lm *LayerMask) buildTree() {
	root := &Node{
		Type:     NodeTypeRoot,
		Name:     "Root",
		Children: []*Node{},
		Left:     0,
		Top:      0,
		Right:    int32(lm.header.Width()),
		Bottom:   int32(lm.header.Height()),
		Visible:  true,
		Opacity:  255,
	}

	stack := []*Node{root}

	// Build hierarchy from layers (forward iteration like Ruby)
	for _, layer := range lm.Layers {
		if layer.IsFolder() {
			if layer.IsFolderEnd() {
				// This is a folder end marker - pop the current group and add to parent
				if len(stack) > 1 {
					currentGroup := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					parent := stack[len(stack)-1]
					parent.Children = append(parent.Children, currentGroup)
				}
			} else {
				// This is a folder start marker - create new group and push to stack
				node := &Node{
					Type:      NodeTypeGroup,
					Name:      layer.Name,
					Layer:     layer,
					Children:  []*Node{},
					Visible:   layer.Visible(),
					Opacity:   layer.Opacity,
					BlendMode: layer.blendModeString(),
					Left:      layer.Left,
					Top:       layer.Top,
					Right:     layer.Right,
					Bottom:    layer.Bottom,
				}
				node.Parent = stack[len(stack)-1]
				stack = append(stack, node)
			}
		} else {
			// Regular layer - add to current parent
			node := &Node{
				Type:      NodeTypeLayer,
				Name:      layer.Name,
				Layer:     layer,
				Children:  []*Node{},
				Visible:   layer.Visible(),
				Opacity:   layer.Opacity,
				BlendMode: layer.blendModeString(),
				Left:      layer.Left,
				Top:       layer.Top,
				Right:     layer.Right,
				Bottom:    layer.Bottom,
			}
			parent := stack[len(stack)-1]
			node.Parent = parent
			parent.Children = append(parent.Children, node)
		}
	}

	lm.tree = root

	// Update dimensions for all group nodes
	lm.tree.UpdateDimensions()
}

// Tree returns the layer tree
func (lm *LayerMask) Tree() *Node {
	return lm.tree
}
