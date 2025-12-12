package psd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlendModes(t *testing.T) {
	psd, err := New("testdata/blendmodes.psd")
	require.NoError(t, err)
	defer psd.Close()

	err = psd.Parse()
	require.NoError(t, err)

	layers := psd.Layers()
	for _, layer := range layers {
		blendMode := layer.BlendMode()
		assert.Equal(t, layer.Name, blendMode.Mode, "Layer %s should have blend mode %s", layer.Name, layer.Name)
	}
}
