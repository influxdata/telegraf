package rfc5424

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleUTF8DecimalConversion(t *testing.T) {
	slice := []uint8{49, 48, 49}
	res := unsafeUTF8DecimalCodePointsToInt(slice)
	assert.Equal(t, 101, res)
}

func TestNumberStartingWithZero(t *testing.T) {
	slice := []uint8{48, 48, 50}
	res := unsafeUTF8DecimalCodePointsToInt(slice)
	assert.Equal(t, 2, res)
}

func TestCharsNotInRange(t *testing.T) {
	point := 10
	slice := []uint8{uint8(point)} // Line Feed (LF)
	res := unsafeUTF8DecimalCodePointsToInt(slice)
	assert.Equal(t, res, -(48 - point))
}

func TestAllDigits(t *testing.T) {
	slice := []uint8{49, 50, 51, 52, 53, 54, 55, 56, 57, 48}
	res := unsafeUTF8DecimalCodePointsToInt(slice)
	assert.Equal(t, 1234567890, res)
}
