// +build linux

package conntrack

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func restoreDflts(savedFiles, savedDirs []string) {
	dfltFiles = savedFiles
	dfltDirs = savedDirs
}

func TestNoFilesFound(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)

	dfltFiles = []string{"baz.txt"}
	dfltDirs = []string{"./foo/bar"}
	c := &Conntrack{}
	acc := &testutil.Accumulator{}
	err := c.Gather(acc)

	assert.EqualError(t, err, "Conntrack input failed to collect metrics. "+
		"Is the conntrack kernel module loaded?")
}

func TestDefaultsUsed(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)
	tmpdir, err := ioutil.TempDir("", "tmp1")
	assert.NoError(t, err)
	defer os.Remove(tmpdir)

	tmpFile, err := ioutil.TempFile(tmpdir, "ip_conntrack_count")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	dfltDirs = []string{tmpdir}
	fname := path.Base(tmpFile.Name())
	dfltFiles = []string{fname}

	count := 1234321
	ioutil.WriteFile(tmpFile.Name(), []byte(strconv.Itoa(count)), 0660)
	c := &Conntrack{}
	acc := &testutil.Accumulator{}

	c.Gather(acc)
	acc.AssertContainsFields(t, inputName, map[string]interface{}{
		fname: float64(count)})
}

func TestConfigsUsed(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)
	tmpdir, err := ioutil.TempDir("", "tmp1")
	assert.NoError(t, err)
	defer os.Remove(tmpdir)

	cntFile, err := ioutil.TempFile(tmpdir, "nf_conntrack_count")
	maxFile, err := ioutil.TempFile(tmpdir, "nf_conntrack_max")
	assert.NoError(t, err)
	defer os.Remove(cntFile.Name())
	defer os.Remove(maxFile.Name())

	dfltDirs = []string{tmpdir}
	cntFname := path.Base(cntFile.Name())
	maxFname := path.Base(maxFile.Name())
	dfltFiles = []string{cntFname, maxFname}

	count := 1234321
	max := 9999999
	ioutil.WriteFile(cntFile.Name(), []byte(strconv.Itoa(count)), 0660)
	ioutil.WriteFile(maxFile.Name(), []byte(strconv.Itoa(max)), 0660)
	c := &Conntrack{}
	acc := &testutil.Accumulator{}

	c.Gather(acc)

	fix := func(s string) string {
		return strings.Replace(s, "nf_", "ip_", 1)
	}

	acc.AssertContainsFields(t, inputName,
		map[string]interface{}{
			fix(cntFname): float64(count),
			fix(maxFname): float64(max),
		})
}
