package tail

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTailFromBeginning(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	tt := NewTail()
	tt.FromBeginning = true
	tt.Files = []string{tmpfile.Name()}
	p, _ := parsers.NewInfluxParser()
	tt.SetParser(p)
	defer tt.Stop()
	defer tmpfile.Close()

	acc := testutil.Accumulator{}
	require.NoError(t, tt.Start(&acc))

	_, err = tmpfile.WriteString("cpu,mytag=foo usage_idle=100\n")
	require.NoError(t, err)
	require.NoError(t, tt.Gather(&acc))
	// arbitrary sleep to wait for message to show up
	time.Sleep(time.Millisecond * 250)

	acc.AssertContainsTaggedFields(t, "cpu",
		map[string]interface{}{
			"usage_idle": float64(100),
		},
		map[string]string{
			"mytag": "foo",
		})
}

func TestTailFromEnd(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.WriteString("cpu,mytag=foo usage_idle=100\n")
	require.NoError(t, err)

	tt := NewTail()
	tt.Files = []string{tmpfile.Name()}
	p, _ := parsers.NewInfluxParser()
	tt.SetParser(p)
	defer tt.Stop()
	defer tmpfile.Close()

	acc := testutil.Accumulator{}
	require.NoError(t, tt.Start(&acc))
	time.Sleep(time.Millisecond * 100)

	_, err = tmpfile.WriteString("cpu,othertag=foo usage_idle=100\n")
	require.NoError(t, err)
	require.NoError(t, tt.Gather(&acc))
	time.Sleep(time.Millisecond * 50)

	acc.AssertContainsTaggedFields(t, "cpu",
		map[string]interface{}{
			"usage_idle": float64(100),
		},
		map[string]string{
			"othertag": "foo",
		})
	assert.Len(t, acc.Metrics, 1)
}

func TestTailBadLine(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	tt := NewTail()
	tt.FromBeginning = true
	tt.Files = []string{tmpfile.Name()}
	p, _ := parsers.NewInfluxParser()
	tt.SetParser(p)
	defer tt.Stop()
	defer tmpfile.Close()

	acc := testutil.Accumulator{}
	require.NoError(t, tt.Start(&acc))

	_, err = tmpfile.WriteString("cpu mytag= foo usage_idle= 100\n")
	require.NoError(t, err)
	require.NoError(t, tt.Gather(&acc))
	time.Sleep(time.Millisecond * 50)

	assert.Len(t, acc.Metrics, 0)
}
