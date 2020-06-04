// +build windows

package win_eventlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeUTF16(t *testing.T) {
	input := "T e s t  S t r i n g "
	want := "Test String"
	got, _ := DecodeUTF16([]byte(input))
	assert.Equal(t, string(got), want)
}
