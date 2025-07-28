package iftop

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParsing0ByteReading(t *testing.T) {
	value := ByteReading("0B")
	assert.Equal(t, float64(0), value.ToFloat64())
}

func TestParsing1ByteReading(t *testing.T) {
	value := ByteReading("1B")
	assert.Equal(t, float64(1), value.ToFloat64())
}

func TestParsing1KBReading(t *testing.T) {
	value := ByteReading("1KB")
	assert.Equal(t, float64(1024), value.ToFloat64())
}

func TestParsing1_06KBReading(t *testing.T) {
	value := ByteReading("1.06KB")
	assert.Equal(t, float64(1024*1.06), value.ToFloat64())
}

func TestParsing3_8MKBReading(t *testing.T) {
	value := ByteReading("3.8MB")
	assert.Equal(t, float64(3.8*1024*1024), value.ToFloat64())
}

func TestParsing43_6GBReading(t *testing.T) {
	value := ByteReading("43.6GB")
	assert.Equal(t, float64(43.6*1024*1024*1024), value.ToFloat64())
}
