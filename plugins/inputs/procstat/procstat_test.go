package procstat

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestGather(t *testing.T) {
	var acc testutil.Accumulator
	pid := os.Getpid()
	file, err := ioutil.TempFile(os.TempDir(), "telegraf")
	require.NoError(t, err)
	file.Write([]byte(strconv.Itoa(pid)))
	file.Close()
	defer os.Remove(file.Name())
	p := Procstat{
		PidFile: file.Name(),
		Prefix:  "foo",
		tagmap:  make(map[int32]map[string]string),
	}
	p.Gather(&acc)
	assert.True(t, acc.HasFloatField("procstat", "foo_cpu_time_user"))
	assert.True(t, acc.HasUIntField("procstat", "foo_memory_vms"))
}
