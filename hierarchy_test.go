package psd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTree(t *testing.T) {
	psd, err := New("testdata/example.psd")
	require.NoError(t, err)
	defer psd.Close()

	err = psd.Parse()
	require.NoError(t, err)

	tree := psd.Tree()
	assert.NotNil(t, tree)

	hash := tree.ToHash()
	assert.NotNil(t, hash)
	assert.Contains(t, hash, "children")

	children, ok := hash["children"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 3, len(children))
}

func TestAncestry(t *testing.T) {
	psd, err := New("testdata/example.psd")
	require.NoError(t, err)
	defer psd.Close()

	err = psd.Parse()
	require.NoError(t, err)

	tree := psd.Tree()

	// Test root
	assert.True(t, tree.IsRoot())
	assert.Equal(t, tree, tree.Root())
	assert.Equal(t, tree, tree.Children[len(tree.Children)-1].Root())

	// Test descendants
	descendants := tree.Descendants()
	assert.Equal(t, 12, len(descendants))

	descendantLayers := tree.DescendantLayers()
	assert.Equal(t, 9, len(descendantLayers))

	descendantGroups := tree.DescendantGroups()
	assert.Equal(t, 3, len(descendantGroups))

	assert.NotEqual(t, tree, descendants[0])

	// Test subtree
	subtree := tree.Subtree()
	assert.Equal(t, 13, len(subtree))

	subtreeLayers := tree.SubtreeLayers()
	assert.Equal(t, 9, len(subtreeLayers))

	subtreeGroups := tree.SubtreeGroups()
	assert.Equal(t, 3, len(subtreeGroups))

	assert.Equal(t, tree, subtree[0])

	// Test children
	assert.True(t, tree.HasChildren())
	assert.False(t, tree.IsChildless())
	assert.False(t, descendantLayers[0].HasChildren())
	assert.True(t, descendantLayers[0].IsChildless())

	// Test siblings
	firstChild := tree.Children[0]
	siblings := firstChild.Siblings()
	assert.Equal(t, tree.Children, siblings)
	assert.Contains(t, siblings, firstChild)
	assert.True(t, firstChild.HasSiblings())
	assert.False(t, firstChild.IsOnlyChild())

	// Test depth
	assert.Equal(t, 0, tree.Depth())
	lastDescendantLayer := descendantLayers[len(descendantLayers)-1]
	assert.Equal(t, 2, lastDescendantLayer.Depth())
	assert.Equal(t, 1, tree.Children[0].Depth())

	// Test path
	nodes := tree.ChildrenAtPath("Version A/Matte")
	assert.Equal(t, 1, len(nodes))
	node := nodes[0]

	path := node.Path()
	assert.Equal(t, "Version A/Matte", path)

	pathArray := node.Path(true)
	assert.Equal(t, []string{"Version A", "Matte"}, pathArray)
}

func TestSearching(t *testing.T) {
	psd, err := New("testdata/example.psd")
	require.NoError(t, err)
	defer psd.Close()

	err = psd.Parse()
	require.NoError(t, err)

	tree := psd.Tree()

	// Find node at path
	nodes := tree.ChildrenAtPath("Version A/Matte")
	assert.Equal(t, 1, len(nodes))
	assert.Equal(t, NodeTypeLayer, nodes[0].Type)

	// Ignore leading slashes
	nodes2 := tree.ChildrenAtPath("/Version A/Matte")
	assert.Equal(t, 1, len(nodes2))

	// Return empty array when not found
	notFound := tree.ChildrenAtPath("NOPE")
	assert.Equal(t, 0, len(notFound))
}

func TestEmptyLayer(t *testing.T) {
	psd, err := New("testdata/empty-layer.psd")
	require.NoError(t, err)
	defer psd.Close()

	err = psd.Parse()
	require.NoError(t, err)

	tree := psd.Tree()

	emptyLayer := tree.ChildrenAtPath("group/empty layer")
	assert.Equal(t, 1, len(emptyLayer))
	assert.True(t, emptyLayer[0].IsEmpty())

	assert.False(t, tree.Children[0].IsEmpty())

	// Test group size calculation
	group := tree.Children[0]
	assert.Equal(t, int32(100), group.Width())
	assert.Equal(t, int32(100), group.Height())
	assert.Equal(t, int32(450), group.Left)
	assert.Equal(t, int32(450), group.Top)
}
