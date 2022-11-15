package syslog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFraming(t *testing.T) {
	var f1 Framing
	err := f1.UnmarshalTOML([]byte(`"non-transparent"`))
	require.NoError(t, err)
	require.Equal(t, NonTransparent, f1)

	var f2 Framing
	err = f2.UnmarshalTOML([]byte(`non-transparent`))
	require.NoError(t, err)
	require.Equal(t, NonTransparent, f2)

	var f3 Framing
	err = f3.UnmarshalTOML([]byte(`'non-transparent'`))
	require.NoError(t, err)
	require.Equal(t, NonTransparent, f3)

	var f4 Framing
	err = f4.UnmarshalTOML([]byte(`"octet-counting"`))
	require.NoError(t, err)
	require.Equal(t, OctetCounting, f4)

	var f5 Framing
	err = f5.UnmarshalTOML([]byte(`octet-counting`))
	require.NoError(t, err)
	require.Equal(t, OctetCounting, f5)

	var f6 Framing
	err = f6.UnmarshalTOML([]byte(`'octet-counting'`))
	require.NoError(t, err)
	require.Equal(t, OctetCounting, f6)

	var f7 Framing
	err = f7.UnmarshalTOML([]byte(`nope`))
	require.Error(t, err)
	require.Equal(t, Framing(-1), f7)
}
