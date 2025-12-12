package psd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	psd, err := New("testdata/example.psd")
	require.NoError(t, err)
	defer psd.Close()

	err = psd.Parse()
	require.NoError(t, err)
	assert.True(t, psd.Parsed())
}

func TestHeader(t *testing.T) {
	psd, err := New("testdata/example.psd")
	require.NoError(t, err)
	defer psd.Close()

	err = psd.Parse()
	require.NoError(t, err)

	header := psd.Header()
	assert.NotNil(t, header)
	assert.Equal(t, uint16(1), header.Version)
	assert.Equal(t, uint16(3), header.Channels)
	assert.Equal(t, uint32(900), header.Width())
	assert.Equal(t, uint32(600), header.Height())
	assert.Equal(t, uint16(3), header.Mode)
	assert.Equal(t, "RGBColor", header.ModeName())
}

func TestResources(t *testing.T) {
	psd, err := New("testdata/example.psd")
	require.NoError(t, err)
	defer psd.Close()

	err = psd.Parse()
	require.NoError(t, err)

	resources := psd.Resources()
	assert.NotNil(t, resources)
	assert.NotNil(t, resources.Resources)
	assert.Greater(t, len(resources.Resources), 0)

	for _, r := range resources.Resources {
		assert.Equal(t, "8BIM", r.Type)
		assert.NotEqual(t, uint16(0), r.ID)
	}
}

func TestLayerMask(t *testing.T) {
	psd, err := New("testdata/example.psd")
	require.NoError(t, err)
	defer psd.Close()

	err = psd.Parse()
	require.NoError(t, err)

	layerMask := psd.LayerMask()
	assert.NotNil(t, layerMask)
	assert.Greater(t, len(layerMask.Layers), 0)
}

func TestLayers(t *testing.T) {
	psd, err := New("testdata/example.psd")
	require.NoError(t, err)
	defer psd.Close()

	err = psd.Parse()
	require.NoError(t, err)

	layers := psd.Layers()
	assert.Equal(t, 15, len(layers))

	// Test first layer
	firstLayer := layers[0]
	assert.Equal(t, "Version C", firstLayer.Name)

	// Test folder detection
	assert.True(t, firstLayer.IsFolder())

	// Find a non-folder layer
	var matteLayer *Layer
	for _, layer := range layers {
		if layer.Name == "Matte" {
			matteLayer = layer
			break
		}
	}
	assert.NotNil(t, matteLayer)
	assert.False(t, matteLayer.IsFolder())

	// Test visibility
	assert.False(t, firstLayer.Visible())

	var versionA *Layer
	for _, layer := range layers {
		if layer.Name == "Version A" {
			versionA = layer
			break
		}
	}
	assert.NotNil(t, versionA)
	assert.True(t, versionA.Visible())

	// Test dimensions
	var logoGlyph *Layer
	for _, layer := range layers {
		if layer.Name == "Logo_Glyph" {
			logoGlyph = layer
			break
		}
	}
	assert.NotNil(t, logoGlyph)
	assert.Equal(t, int32(142), logoGlyph.Width())
	assert.Equal(t, int32(179), logoGlyph.Height())
	assert.Equal(t, int32(379), logoGlyph.Left)
	assert.Equal(t, int32(210), logoGlyph.Top)

	// Test blend mode
	blendMode := versionA.BlendMode()
	assert.NotNil(t, blendMode)
	assert.Equal(t, "normal", blendMode.Mode)
	assert.Equal(t, uint8(255), blendMode.Opacity)
	assert.Equal(t, 100, blendMode.OpacityPercentage)
	assert.True(t, blendMode.Visible)
}
