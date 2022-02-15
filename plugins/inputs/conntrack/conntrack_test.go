//go:build linux
// +build linux

package conntrack

import (
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
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

	require.EqualError(t, err, "Conntrack input failed to collect metrics. "+
		"Is the conntrack kernel module loaded?")
}

func TestDefaultsUsed(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)
	tmpdir, err := os.MkdirTemp("", "tmp1")
	require.NoError(t, err)
	defer os.Remove(tmpdir)

	tmpFile, err := os.CreateTemp(tmpdir, "ip_conntrack_count")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	dfltDirs = []string{tmpdir}
	fname := path.Base(tmpFile.Name())
	dfltFiles = []string{fname}

	count := 1234321
	require.NoError(t, os.WriteFile(tmpFile.Name(), []byte(strconv.Itoa(count)), 0660))
	c := &Conntrack{}
	acc := &testutil.Accumulator{}

	require.NoError(t, c.Gather(acc))
	acc.AssertContainsFields(t, inputName, map[string]interface{}{
		fname: float64(count)})
}

func TestConfigsUsed(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)
	tmpdir, err := os.MkdirTemp("", "tmp1")
	require.NoError(t, err)
	defer os.Remove(tmpdir)

	cntFile, err := os.CreateTemp(tmpdir, "nf_conntrack_count")
	require.NoError(t, err)
	maxFile, err := os.CreateTemp(tmpdir, "nf_conntrack_max")
	require.NoError(t, err)
	defer os.Remove(cntFile.Name())
	defer os.Remove(maxFile.Name())

	dfltDirs = []string{tmpdir}
	cntFname := path.Base(cntFile.Name())
	maxFname := path.Base(maxFile.Name())
	dfltFiles = []string{cntFname, maxFname}

	count := 1234321
	max := 9999999
	require.NoError(t, os.WriteFile(cntFile.Name(), []byte(strconv.Itoa(count)), 0660))
	require.NoError(t, os.WriteFile(maxFile.Name(), []byte(strconv.Itoa(max)), 0660))
	c := &Conntrack{}
	acc := &testutil.Accumulator{}

	require.NoError(t, c.Gather(acc))

	fix := func(s string) string {
		return strings.Replace(s, "nf_", "ip_", 1)
	}

	acc.AssertContainsFields(t, inputName,
		map[string]interface{}{
			fix(cntFname): float64(count),
			fix(maxFname): float64(max),
		})
}
