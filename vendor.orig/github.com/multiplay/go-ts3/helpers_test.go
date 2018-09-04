package ts3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeDecode(t *testing.T) {
	str := "\\/ |\a\b\f\n\r\t\v"
	assert.Equal(t, `\\\/\s\p\a\b\f\n\r\t\v`, encoder.Replace(str))
	assert.Equal(t, str, Decode(encoder.Replace(str)))
}

type testResp struct {
	Response string
	ID       int
	Valid    bool
}

func TestDecodeResponse(t *testing.T) {
	r := &testResp{}
	expected := &testResp{
		Response: "test",
		ID:       1,
		Valid:    false,
	}
	assert.NoError(t, DecodeResponse([]string{"response=test id=1 valid"}, r))
	assert.Equal(t, expected, r)

	assert.Error(t, DecodeResponse([]string{"line1", "line2"}, r))
}
