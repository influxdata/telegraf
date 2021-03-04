// +build linux

package kernel

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

func TestFullProcFile(t *testing.T) {
	tmpfile := makeFakeStatFile([]byte(statFileFull))
	tmpfile2 := makeFakeStatFile([]byte(entropyStatFileFull))
	defer os.Remove(tmpfile)
	defer os.Remove(tmpfile2)

	k := Kernel{
		statFile:        tmpfile,
		entropyStatFile: tmpfile2,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.NoError(t, err)

	fields := map[string]interface{}{
		"boot_time":        int64(1457505775),
		"context_switches": int64(2626618),
		"disk_pages_in":    int64(5741),
		"disk_pages_out":   int64(1808),
		"interrupts":       int64(1472736),
		"processes_forked": int64(10673),
		"entropy_avail":    int64(1024),
	}
	acc.AssertContainsFields(t, "kernel", fields)
}

func TestPartialProcFile(t *testing.T) {
	tmpfile := makeFakeStatFile([]byte(statFilePartial))
	tmpfile2 := makeFakeStatFile([]byte(entropyStatFilePartial))
	defer os.Remove(tmpfile)
	defer os.Remove(tmpfile2)

	k := Kernel{
		statFile:        tmpfile,
		entropyStatFile: tmpfile2,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.NoError(t, err)

	fields := map[string]interface{}{
		"boot_time":        int64(1457505775),
		"context_switches": int64(2626618),
		"disk_pages_in":    int64(5741),
		"disk_pages_out":   int64(1808),
		"interrupts":       int64(1472736),
		"entropy_avail":    int64(1024),
	}
	acc.AssertContainsFields(t, "kernel", fields)
}

func TestInvalidProcFile1(t *testing.T) {
	tmpfile := makeFakeStatFile([]byte(statFileInvalid))
	tmpfile2 := makeFakeStatFile([]byte(entropyStatFileInvalid))
	defer os.Remove(tmpfile)
	defer os.Remove(tmpfile2)

	k := Kernel{
		statFile:        tmpfile,
		entropyStatFile: tmpfile2,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.Error(t, err)
}

func TestInvalidProcFile2(t *testing.T) {
	tmpfile := makeFakeStatFile([]byte(statFileInvalid2))
	defer os.Remove(tmpfile)

	k := Kernel{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.Error(t, err)
}

func TestNoProcFile(t *testing.T) {
	tmpfile := makeFakeStatFile([]byte(statFileInvalid2))
	os.Remove(tmpfile)

	k := Kernel{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

const statFileFull = `cpu  6796 252 5655 10444977 175 0 101 0 0 0
cpu0 6796 252 5655 10444977 175 0 101 0 0 0
intr 1472736 57 10 0 0 0 0 0 0 0 0 0 0 156 0 0 0 0 0 0 111551 42541 12356 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 2626618
btime 1457505775
processes 10673
procs_running 2
procs_blocked 0
softirq 1031662 0 649485 20946 111071 11620 0 1 0 994 237545
page 5741 1808
swap 1 0
entropy_avail 1024
`

const statFilePartial = `cpu  6796 252 5655 10444977 175 0 101 0 0 0
cpu0 6796 252 5655 10444977 175 0 101 0 0 0
intr 1472736 57 10 0 0 0 0 0 0 0 0 0 0 156 0 0 0 0 0 0 111551 42541 12356 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 2626618
btime 1457505775
procs_running 2
procs_blocked 0
softirq 1031662 0 649485 20946 111071 11620 0 1 0 994 237545
page 5741 1808
`

// missing btime measurement
const statFileInvalid = `cpu  6796 252 5655 10444977 175 0 101 0 0 0
cpu0 6796 252 5655 10444977 175 0 101 0 0 0
intr 1472736 57 10 0 0 0 0 0 0 0 0 0 0 156 0 0 0 0 0 0 111551 42541 12356 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 2626618
btime
processes 10673
procs_running 2
procs_blocked 0
softirq 1031662 0 649485 20946 111071 11620 0 1 0 994 237545
page 5741 1808
swap 1 0
entropy_avail 1024
`

// missing second page measurement
const statFileInvalid2 = `cpu  6796 252 5655 10444977 175 0 101 0 0 0
cpu0 6796 252 5655 10444977 175 0 101 0 0 0
intr 1472736 57 10 0 0 0 0 0 0 0 0 0 0 156 0 0 0 0 0 0 111551 42541 12356 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 2626618
processes 10673
procs_running 2
page 5741
procs_blocked 0
softirq 1031662 0 649485 20946 111071 11620 0 1 0 994 237545
entropy_avail 1024 2048
`

const entropyStatFileFull = `1024`

const entropyStatFilePartial = `1024`

const entropyStatFileInvalid = ``

func makeFakeStatFile(content []byte) string {
	tmpfile, err := ioutil.TempFile("", "kernel_test")
	if err != nil {
		panic(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		panic(err)
	}
	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	return tmpfile.Name()
}
