package psd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	psd, err := New("testdata/example.psd")
	require.NoError(t, err)
	assert.NotNil(t, psd)
	assert.False(t, psd.Parsed())

	defer psd.Close()
}

func TestNewBadFilename(t *testing.T) {
	_, err := New("")
	assert.Error(t, err)
}

func TestOpen(t *testing.T) {
	var parsed bool
	err := Open("testdata/example.psd", func(psd *PSD) error {
		parsed = psd.Parsed()
		return nil
	})
	require.NoError(t, err)
	assert.True(t, parsed)
}
