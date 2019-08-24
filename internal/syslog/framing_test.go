package syslog

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFraming(t *testing.T) {
	var f1 Framing
	f1.UnmarshalTOML([]byte(`"non-transparent"`))
	assert.Equal(t, NonTransparent, f1)

	var f2 Framing
	f2.UnmarshalTOML([]byte(`non-transparent`))
	assert.Equal(t, NonTransparent, f2)

	var f3 Framing
	f3.UnmarshalTOML([]byte(`'non-transparent'`))
	assert.Equal(t, NonTransparent, f3)

	var f4 Framing
	f4.UnmarshalTOML([]byte(`"octet-counting"`))
	assert.Equal(t, OctetCounting, f4)

	var f5 Framing
	f5.UnmarshalTOML([]byte(`octet-counting`))
	assert.Equal(t, OctetCounting, f5)

	var f6 Framing
	f6.UnmarshalTOML([]byte(`'octet-counting'`))
	assert.Equal(t, OctetCounting, f6)

	var f7 Framing
	err := f7.UnmarshalTOML([]byte(`nope`))
	assert.Equal(t, Framing(-1), f7)
	assert.Error(t, err)
}
