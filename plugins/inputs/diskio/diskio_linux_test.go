//go:build linux

package diskio

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiskInfo(t *testing.T) {
	plugin := &DiskIO{}
	require.NoError(t, plugin.Init())
	plugin.infoCache = map[string]diskInfoCache{
		"null": {
			modifiedAt:   0,
			udevDataPath: "testdata/udev.txt",
			sysBlockPath: "testdata",
			values:       map[string]string{},
		},
	}

	di, err := plugin.diskInfo("null")
	require.NoError(t, err)
	require.Equal(t, "myval1", di["MY_PARAM_1"])
	require.Equal(t, "myval2", di["MY_PARAM_2"])
	require.Equal(t, "/dev/foo/bar/devlink /dev/foo/bar/devlink1", di["DEVLINKS"])
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

	for i, tc := range tests {
		t.Run(fmt.Sprintf("template %d", i), func(t *testing.T) {
			plugin := &DiskIO{}
			plugin.NameTemplates = tc.templates
			require.NoError(t, plugin.Init())
			plugin.infoCache = map[string]diskInfoCache{
				"null": {
					modifiedAt:   0,
					udevDataPath: "testdata/udev.txt",
					sysBlockPath: "testdata",
					values:       map[string]string{},
				},
			}
			name, _ := plugin.diskName("null")
			require.Equal(t, tc.expected, name, "Templates: %#v", tc.templates)
		})
	}
}

// DiskIOStats.diskTags isn't a linux specific function, but dependent
// functions are a no-op on non-Linux.
func TestDiskIOStats_diskTags(t *testing.T) {
	plugin := &DiskIO{}
	plugin.DeviceTags = []string{"MY_PARAM_2"}
	require.NoError(t, plugin.Init())
	plugin.infoCache = map[string]diskInfoCache{
		"null": {
			modifiedAt:   0,
			udevDataPath: "testdata/udev.txt",
			sysBlockPath: "testdata",
			values:       map[string]string{},
		},
	}
	dt := plugin.diskTags("null")
	require.Equal(t, map[string]string{"MY_PARAM_2": "myval2"}, dt)
}

func TestDiskInfoHonorsHostDev(t *testing.T) {
	t.Setenv("HOST_DEV", filepath.Join("testdata", "hostfs", "dev"))

	plugin := &DiskIO{}
	require.NoError(t, plugin.Init())
	plugin.infoCache = map[string]diskInfoCache{
		"mockdev": {
			modifiedAt:   0,
			udevDataPath: "testdata/udev.txt",
			values:       map[string]string{},
		},
	}
	di, err := plugin.diskInfo("mockdev")
	require.NoError(t, err)
	require.Equal(t, "myval1", di["MY_PARAM_1"])
}

func TestDiskInfoHonorsHostPrefixFallbacksForDev(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "host root fallback",
			env: map[string]string{
				"HOST_ROOT":         filepath.Join("testdata", "hostfs"),
				"HOST_MOUNT_PREFIX": filepath.Join("testdata", "hostfs", "unused"),
			},
		},
		{
			name: "host mount prefix fallback",
			env: map[string]string{
				"HOST_MOUNT_PREFIX": filepath.Join("testdata", "hostfs"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for key, value := range tc.env {
				t.Setenv(key, value)
			}

			plugin := &DiskIO{}
			require.NoError(t, plugin.Init())
			plugin.infoCache = map[string]diskInfoCache{
				"mockdev": {
					modifiedAt:   0,
					udevDataPath: "testdata/udev.txt",
					values:       map[string]string{},
				},
			}

			di, err := plugin.diskInfo("mockdev")
			require.NoError(t, err)
			require.Equal(t, "myval1", di["MY_PARAM_1"])
		})
	}
}

func TestGetDeviceWWIDHonorsHostSys(t *testing.T) {
	t.Setenv("HOST_SYS", filepath.Join("testdata", "hostfs", "sys"))

	plugin := &DiskIO{}
	require.NoError(t, plugin.Init())
	require.Equal(t, "my-wwid", plugin.getDeviceWWID("sda"))
}
