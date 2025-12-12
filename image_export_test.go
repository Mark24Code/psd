package psd

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageExporting(t *testing.T) {
	psd, err := New("testdata/pixel.psd")
	require.NoError(t, err)
	defer psd.Close()

	err = psd.Parse()
	require.NoError(t, err)

	img := psd.Image()
	assert.NotNil(t, img)
	assert.Equal(t, uint32(1), img.Width())
	assert.Equal(t, uint32(1), img.Height())

	pixelData := img.PixelData()
	assert.Equal(t, 1, len(pixelData))

	expectedColor := color.RGBA{R: 0, G: 100, B: 200, A: 255}
	assert.Equal(t, expectedColor, pixelData[0])

	png := img.ToPNG()
	assert.NotNil(t, png)
	assert.Equal(t, 1, png.Bounds().Dx())
	assert.Equal(t, 1, png.Bounds().Dy())

	pngColor := png.At(0, 0)
	r, g, b, a := pngColor.RGBA()
	// RGBA() returns values scaled to 16-bit, so we need to scale back
	assert.Equal(t, uint32(0), r>>8)
	assert.Equal(t, uint32(100), g>>8)
	assert.Equal(t, uint32(200), b>>8)
	assert.Equal(t, uint32(255), a>>8)
}
