package opcua

import (
	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetupWorkarounds(t *testing.T) {
	o := OpcUAClient{
		Config: &OpcUAClientConfig{
			Workarounds: OpcUAWorkarounds{
				AdditionalValidStatusCodes: []string{"0xC0", "0x00AA0000"},
			},
		},
	}

	err := o.setupWorkarounds()
	require.NoError(t, err)

	require.Len(t, o.codes, 3)
	require.Equal(t, o.codes[0], ua.StatusCode(0))
	require.Equal(t, o.codes[1], ua.StatusCode(192))
	require.Equal(t, o.codes[2], ua.StatusCode(11141120))
}

func TestCheckStatusCode(t *testing.T) {
	var o OpcUAClient
	o.codes = []ua.StatusCode{ua.StatusCode(0), ua.StatusCode(192), ua.StatusCode(11141120)}
	require.Equal(t, o.StatusCodeOK(ua.StatusCode(192)), true)
}
