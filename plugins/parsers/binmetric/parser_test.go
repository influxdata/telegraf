package binmetric

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {

	binMetric, err := ioutil.ReadFile("metric.bin")
	require.NoError(t, err)

	parser := BinMetric{}
	metrics, err := parser.Parse(binMetric)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	require.Equal(t, "drone_status", metrics[0].Name())

	for key, value := range metrics[0].Fields() {
		fmt.Println(key, value)
	}
}
