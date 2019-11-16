package protodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_embedded(t *testing.T) {
	s, ok := GetServByPort("tcp", 443)
	assert.True(t, ok)
	assert.Equal(t, s, "https")
}

func Test_etcServices(t *testing.T) {
	snm := make(map[string]string)
	assert.NoError(t, populateServiceNameMapFromEtcServices(snm))
	assert.NotEqual(t, 0, len(snm))
}
