package multifile

import (
	"os"
	"path"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileTypes(t *testing.T) {
	wd, _ := os.Getwd()

	m := MultiFile{
		BaseDir:   path.Join(wd, `testdata`),
		FailEarly: true,
		Files: []File{
			{Name: `bool.txt`, Dest: `examplebool`, Conversion: `bool`},
			{Name: `float.txt`, Dest: `examplefloat`, Conversion: `float`},
			{Name: `int.txt`, Dest: `examplefloatX`, Conversion: `float(3)`},
			{Name: `int.txt`, Dest: `exampleint`, Conversion: `int`},
			{Name: `string.txt`, Dest: `examplestring`},
			{Name: `tag.txt`, Dest: `exampletag`, Conversion: `tag`},
			{Name: `int.txt`, Conversion: `int`},
		},
	}

	var acc testutil.Accumulator

	err := m.Gather(&acc)

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"exampletag": "test"}, acc.Metrics[0].Tags)
	assert.Equal(t, map[string]interface{}{
		"examplebool":   true,
		"examplestring": "hello world",
		"exampleint":    int64(123456),
		"int.txt":       int64(123456),
		"examplefloat":  123.456,
		"examplefloatX": 123.456,
	}, acc.Metrics[0].Fields)
}

func FailEarly(failEarly bool, t *testing.T) error {
	wd, _ := os.Getwd()

	m := MultiFile{
		BaseDir:   path.Join(wd, `testdata`),
		FailEarly: failEarly,
		Files: []File{
			{Name: `int.txt`, Dest: `exampleint`, Conversion: `int`},
			{Name: `int.txt`, Dest: `exampleerror`, Conversion: `bool`},
		},
	}

	var acc testutil.Accumulator

	err := m.Gather(&acc)

	if err == nil {
		assert.Equal(t, map[string]interface{}{
			"exampleint": int64(123456),
		}, acc.Metrics[0].Fields)
	}

	return err
}

func TestFailEarly(t *testing.T) {
	err := FailEarly(false, t)
	require.NoError(t, err)
	err = FailEarly(true, t)
	require.Error(t, err)
}
