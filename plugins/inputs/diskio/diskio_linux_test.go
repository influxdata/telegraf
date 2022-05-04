//go:build linux
// +build linux

package diskio

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var nullDiskInfo = []byte(`
E:MY_PARAM_1=myval1
E:MY_PARAM_2=myval2
S:foo/bar/devlink
S:foo/bar/devlink1
`)

// setupNullDisk sets up fake udev info as if /dev/null were a disk.
func setupNullDisk(t *testing.T, s *DiskIO, devName string) func() {
	td, err := os.CreateTemp("", ".telegraf.DiskInfoTest")
	require.NoError(t, err)

	if s.infoCache == nil {
		s.infoCache = make(map[string]diskInfoCache)
	}
	ic, ok := s.infoCache[devName]
	if !ok {
		// No previous calls for the device were done, easy to poison the cache
		s.infoCache[devName] = diskInfoCache{
			modifiedAt:   0,
			udevDataPath: td.Name(),
			values:       map[string]string{},
		}
	}
	origUdevPath := ic.udevDataPath

	cleanFunc := func() {
		ic.udevDataPath = origUdevPath
		//nolint:errcheck,revive // we cannot do anything if file cannot be removed
		os.Remove(td.Name())
	}

	ic.udevDataPath = td.Name()
	_, err = td.Write(nullDiskInfo)
	if err != nil {
		cleanFunc()
		t.Fatal(err)
	}

	return cleanFunc
}

func TestDiskInfo(t *testing.T) {
	s := &DiskIO{}
	clean := setupNullDisk(t, s, "null")
	defer clean()
	di, err := s.diskInfo("null")
	require.NoError(t, err)
	require.Equal(t, "myval1", di["MY_PARAM_1"])
	require.Equal(t, "myval2", di["MY_PARAM_2"])
	require.Equal(t, "/dev/foo/bar/devlink /dev/foo/bar/devlink1", di["DEVLINKS"])

	// test that data is cached
	clean()

	di, err = s.diskInfo("null")
	require.NoError(t, err)
	require.Equal(t, "myval1", di["MY_PARAM_1"])
	require.Equal(t, "myval2", di["MY_PARAM_2"])
	require.Equal(t, "/dev/foo/bar/devlink /dev/foo/bar/devlink1", di["DEVLINKS"])
	// unfortunately we can't adjust mtime on /dev/null to test cache invalidation
}

// DiskIOStats.diskName isn't a linux specific function, but dependent
// functions are a no-op on non-Linux.
func TestDiskIOStats_diskName(t *testing.T) {
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
		func() {
			s := DiskIO{
				NameTemplates: tc.templates,
			}
			defer setupNullDisk(t, &s, "null")() //nolint:revive // done on purpose, cleaning will be executed properly
			name, _ := s.diskName("null")
			require.Equal(t, tc.expected, name, "Templates: %#v", tc.templates)
		}()
	}
}

// DiskIOStats.diskTags isn't a linux specific function, but dependent
// functions are a no-op on non-Linux.
func TestDiskIOStats_diskTags(t *testing.T) {
	s := &DiskIO{
		DeviceTags: []string{"MY_PARAM_2"},
	}
	defer setupNullDisk(t, s, "null")() //nolint:revive // done on purpose, cleaning will be executed properly
	dt := s.diskTags("null")
	require.Equal(t, map[string]string{"MY_PARAM_2": "myval2"}, dt)
}
