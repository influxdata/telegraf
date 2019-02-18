// +build linux

package diskio

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var nullDiskInfo = []byte(`
E:MY_PARAM_1=myval1
E:MY_PARAM_2=myval2
S:foo/bar/devlink
S:foo/bar/devlink1
`)

// setupNullDisk sets up fake udev info as if /dev/null were a disk.
func setupNullDisk(t *testing.T) func() error {
	td, err := ioutil.TempDir("", ".telegraf.TestDiskInfo")
	require.NoError(t, err)

	origUdevPath := udevPath

	cleanFunc := func() error {
		udevPath = origUdevPath
		return os.RemoveAll(td)
	}

	udevPath = td
	err = ioutil.WriteFile(td+"/b1:3", nullDiskInfo, 0644) // 1:3 is the 'null' device
	if err != nil {
		cleanFunc()
		t.Fatal(err)
	}

	return cleanFunc
}

func TestDiskInfo(t *testing.T) {
	clean := setupNullDisk(t)
	defer clean()

	s := &DiskIO{}
	di, err := s.diskInfo("null")
	require.NoError(t, err)
	assert.Equal(t, "myval1", di["MY_PARAM_1"])
	assert.Equal(t, "myval2", di["MY_PARAM_2"])
	assert.Equal(t, "/dev/foo/bar/devlink /dev/foo/bar/devlink1", di["DEVLINKS"])

	// test that data is cached
	err = clean()
	require.NoError(t, err)

	di, err = s.diskInfo("null")
	require.NoError(t, err)
	assert.Equal(t, "myval1", di["MY_PARAM_1"])
	assert.Equal(t, "myval2", di["MY_PARAM_2"])
	assert.Equal(t, "/dev/foo/bar/devlink /dev/foo/bar/devlink1", di["DEVLINKS"])

	// unfortunately we can't adjust mtime on /dev/null to test cache invalidation
}

// DiskIOStats.diskName isn't a linux specific function, but dependent
// functions are a no-op on non-Linux.
func TestDiskIOStats_diskName(t *testing.T) {
	defer setupNullDisk(t)()

	tests := []struct {
		templates []string
		expected  string
	}{
		{[]string{"$MY_PARAM_1"}, "myval1"},
		{[]string{"${MY_PARAM_1}"}, "myval1"},
		{[]string{"x$MY_PARAM_1"}, "xmyval1"},
		{[]string{"x${MY_PARAM_1}x"}, "xmyval1x"},
		{[]string{"$MISSING", "$MY_PARAM_1"}, "myval1"},
		{[]string{"$MY_PARAM_1", "$MY_PARAM_2"}, "myval1"},
		{[]string{"$MISSING"}, "null"},
		{[]string{"$MY_PARAM_1/$MY_PARAM_2"}, "myval1/myval2"},
		{[]string{"$MY_PARAM_2/$MISSING"}, "null"},
	}

	for _, tc := range tests {
		s := DiskIO{
			NameTemplates: tc.templates,
		}
		assert.Equal(t, tc.expected, s.diskName("null"), "Templates: %#v", tc.templates)
	}
}

// DiskIOStats.diskTags isn't a linux specific function, but dependent
// functions are a no-op on non-Linux.
func TestDiskIOStats_diskTags(t *testing.T) {
	defer setupNullDisk(t)()

	s := &DiskIO{
		DeviceTags: []string{"MY_PARAM_2"},
	}
	dt := s.diskTags("null")
	assert.Equal(t, map[string]string{"MY_PARAM_2": "myval2"}, dt)
}
