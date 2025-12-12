package psd

import (
	"fmt"
	"strings"
)

// Node types
const (
	NodeTypeRoot  = "root"
	NodeTypeGroup = "group"
	NodeTypeLayer = "layer"
)

// Node represents a node in the layer tree
type Node struct {
	Type      string
	Name      string
	Layer     *Layer
	Parent    *Node
	Children  []*Node
	Visible   bool
	Opacity   uint8
	BlendMode string
	Left      int32
	Top       int32
	Right     int32
	Bottom    int32
}

// Root returns the root node of the tree
func (n *Node) Root() *Node {
	current := n
	for current.Parent != nil {
		current = current.Parent
	}
	return current
}

// IsRoot returns whether this is the root node
func (n *Node) IsRoot() bool {
	return n.Type == NodeTypeRoot
}

// HasChildren returns whether this node has children
func (n *Node) HasChildren() bool {
	return len(n.Children) > 0
}

// IsChildless returns whether this node has no children
func (n *Node) IsChildless() bool {
	return !n.HasChildren()
}

// Descendants returns all descendant nodes (not including this node)
func (n *Node) Descendants() []*Node {
	var result []*Node
	for _, child := range n.Children {
		result = append(result, child)
		result = append(result, child.Descendants()...)
	}
	return result
}

// DescendantLayers returns all descendant layer nodes
func (n *Node) DescendantLayers() []*Node {
	var result []*Node
	for _, node := range n.Descendants() {
		if node.Type == NodeTypeLayer {
			result = append(result, node)
		}
	}
	return result
}

// DescendantGroups returns all descendant group nodes
func (n *Node) DescendantGroups() []*Node {
	var result []*Node
	for _, node := range n.Descendants() {
		if node.Type == NodeTypeGroup {
			result = append(result, node)
		}
	}
	return result
}

// Subtree returns all nodes in the subtree (including this node)
func (n *Node) Subtree() []*Node {
	result := []*Node{n}
	result = append(result, n.Descendants()...)
	return result
}

// SubtreeLayers returns all layer nodes in the subtree
func (n *Node) SubtreeLayers() []*Node {
	result := []*Node{}
	if n.Type == NodeTypeLayer {
		result = append(result, n)
	}
	result = append(result, n.DescendantLayers()...)
	return result
}

// SubtreeGroups returns all group nodes in the subtree
func (n *Node) SubtreeGroups() []*Node {
	result := []*Node{}
	if n.Type == NodeTypeGroup {
		result = append(result, n)
	}
	result = append(result, n.DescendantGroups()...)
	return result
}

// Siblings returns all siblings including this node
func (n *Node) Siblings() []*Node {
	if n.Parent == nil {
		return []*Node{n}
	}
	return n.Parent.Children
}

// HasSiblings returns whether this node has siblings
func (n *Node) HasSiblings() bool {
	return len(n.Siblings()) > 1
}

// IsOnlyChild returns whether this node is an only child
func (n *Node) IsOnlyChild() bool {
	return !n.HasSiblings()
}

// Depth returns the depth of this node in the tree (root is 0)
func (n *Node) Depth() int {
	depth := 0
	current := n
	for current.Parent != nil {
		depth++
		current = current.Parent
	}
	return depth
}

// Path returns the path to this node
func (n *Node) Path(asArray ...bool) interface{} {
	parts := []string{}
	current := n
	for current.Parent != nil {
		parts = append([]string{current.Name}, parts...)
		current = current.Parent
	}

	if len(asArray) > 0 && asArray[0] {
		return parts
	}

	return strings.Join(parts, "/")
}

// ChildrenAtPath finds nodes at the given path
func (n *Node) ChildrenAtPath(path interface{}) []*Node {
	var parts []string

	switch p := path.(type) {
	case string:
		// Remove leading slash
		p = strings.TrimPrefix(p, "/")
		if p == "" {
			return []*Node{}
		}
		parts = strings.Split(p, "/")
	case []string:
		parts = p
	default:
		return []*Node{}
	}

	return n.findAtPath(parts)
}

func (n *Node) findAtPath(parts []string) []*Node {
	if len(parts) == 0 {
		return []*Node{n}
	}

	target := parts[0]
	remaining := parts[1:]

	var results []*Node
	for _, child := range n.Children {
		if child.Name == target {
			if len(remaining) == 0 {
				results = append(results, child)
			} else {
				results = append(results, child.findAtPath(remaining)...)
			}
		}
	}

	return results
}

// FilterByComp filters the tree by layer comp
func (n *Node) FilterByComp(compName string) (*Node, error) {
	// This would require parsing layer comp data from resources
	// For now, return error indicating not implemented
	return nil, fmt.Errorf("layer comp not found")
}

// ToHash converts the node tree to a hash/map structure
func (n *Node) ToHash() map[string]interface{} {
	result := map[string]interface{}{
		"type":          n.Type,
		"name":          n.Name,
		"visible":       n.Visible,
		"opacity":       float64(n.Opacity) / 255.0,
		"blending_mode": n.BlendMode,
		"left":          n.Left,
		"top":           n.Top,
		"right":         n.Right,
		"bottom":        n.Bottom,
		"width":         n.Width(),
		"height":        n.Height(),
	}

	if len(n.Children) > 0 {
		children := make([]map[string]interface{}, len(n.Children))
		for i, child := range n.Children {
			children[i] = child.ToHash()
		}
		result["children"] = children
	}

	return result
}

// Width returns the width of the node
func (n *Node) Width() int32 {
	return n.Right - n.Left
}

// Height returns the height of the node
func (n *Node) Height() int32 {
	return n.Bottom - n.Top
}

// IsEmpty returns whether this node is empty (zero size)
func (n *Node) IsEmpty() bool {
	return n.Width() == 0 || n.Height() == 0
}

// IsVisible returns whether the node is visible
func (n *Node) IsVisible() bool {
	return n.Visible
}

// FillOpacity returns the fill opacity (default 255 for now)
func (n *Node) FillOpacity() uint8 {
	// TODO: Parse from layer info
	return 255
}

// UpdateDimensions recursively updates the dimensions of group nodes based on their children
func (n *Node) UpdateDimensions() {
	// Layer nodes don't need dimension updates
	if n.Type == NodeTypeLayer {
		return
	}

	// Recursively update all children first
	for _, child := range n.Children {
		child.UpdateDimensions()
	}

	// Root node dimensions are fixed
	if n.Type == NodeTypeRoot {
		return
	}

	// Calculate group bounding box from non-empty children
	nonEmptyChildren := []*Node{}
	for _, child := range n.Children {
		if !child.IsEmpty() {
			nonEmptyChildren = append(nonEmptyChildren, child)
		}
	}

	// If all children are empty, set dimensions to zero
	if len(nonEmptyChildren) == 0 {
		n.Left, n.Top, n.Right, n.Bottom = 0, 0, 0, 0
		return
	}

	// Calculate min/max bounds from non-empty children
	n.Left = minInt32FromNodes(nonEmptyChildren, func(node *Node) int32 { return node.Left })
	n.Top = minInt32FromNodes(nonEmptyChildren, func(node *Node) int32 { return node.Top })
	n.Right = maxInt32FromNodes(nonEmptyChildren, func(node *Node) int32 { return node.Right })
	n.Bottom = maxInt32FromNodes(nonEmptyChildren, func(node *Node) int32 { return node.Bottom })
}

// Helper function to find minimum int32 value from nodes
func minInt32FromNodes(nodes []*Node, selector func(*Node) int32) int32 {
	if len(nodes) == 0 {
		return 0
	}
	min := selector(nodes[0])
	for _, node := range nodes[1:] {
		val := selector(node)
		if val < min {
			min = val
		}
	}
	return min
}

// Helper function to find maximum int32 value from nodes
func maxInt32FromNodes(nodes []*Node, selector func(*Node) int32) int32 {
	if len(nodes) == 0 {
		return 0
	}
	max := selector(nodes[0])
	for _, node := range nodes[1:] {
		val := selector(node)
		if val > max {
			max = val
		}
	}
	return max
}
