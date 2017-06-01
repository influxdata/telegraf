package entropy

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

func restoreDflt(savedProc string) {
	dfltProc = savedProc
}

func TestNoFilesFound(t *testing.T) {
	defer restoreDflt(dfltProc)

	dfltProc = "baz.txt"
	e := &Entropy{}
	acc := &testutil.Accumulator{}
	err := e.Gather(acc)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not stat proc")
}

func TestDefaultsUsed(t *testing.T) {
	defer restoreDflt(dfltProc)
	tmpdir, err := ioutil.TempDir("", "tmp1")
	assert.NoError(t, err)
	defer os.Remove(tmpdir)

	tmpFile, err := ioutil.TempFile(tmpdir, "entropy_avail_test")
	assert.NoError(t, err)

	dfltProc = tmpFile.Name()

	avail := 2048
	ioutil.WriteFile(tmpFile.Name(), []byte(strconv.Itoa(avail)), 0660)
	e := &Entropy{}
	acc := &testutil.Accumulator{}

	e.Gather(acc)
	acc.AssertContainsFields(t, inputName, map[string]interface{}{"available": int(avail)})
}

func TestConfigsUsed(t *testing.T) {
	defer restoreDflt(dfltProc)
	tmpdir, err := ioutil.TempDir("", "tmp1")
	assert.NoError(t, err)
	defer os.Remove(tmpdir)

	tmpFile, err := ioutil.TempFile(tmpdir, "entropy_availa_test")
	assert.NoError(t, err)

	avail := 1234
	ioutil.WriteFile(tmpFile.Name(), []byte(strconv.Itoa(avail)), 0660)
	e := &Entropy{Proc: tmpFile.Name()}
	acc := &testutil.Accumulator{}

	e.Gather(acc)
	acc.AssertContainsFields(t, inputName, map[string]interface{}{"available": int(avail)})
}
