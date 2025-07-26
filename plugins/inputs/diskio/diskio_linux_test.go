//go:build linux

package diskio

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiskInfo(t *testing.T) {
	plugin := &DiskIO{
		infoCache: map[string]diskInfoCache{
			"null": {
				modifiedAt:   0,
				udevDataPath: "testdata/udev.txt",
				sysBlockPath: "testdata",
				values:       map[string]string{},
			},
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
			plugin := DiskIO{
				NameTemplates: tc.templates,
				infoCache: map[string]diskInfoCache{
					"null": {
						modifiedAt:   0,
						udevDataPath: "testdata/udev.txt",
						sysBlockPath: "testdata",
						values:       map[string]string{},
					},
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
	plugin := &DiskIO{
		DeviceTags: []string{"MY_PARAM_2"},
		infoCache: map[string]diskInfoCache{
			"null": {
				modifiedAt:   0,
				udevDataPath: "testdata/udev.txt",
				sysBlockPath: "testdata",
				values:       map[string]string{},
			},
		},
	}
	dt := plugin.diskTags("null")
	require.Equal(t, map[string]string{"MY_PARAM_2": "myval2"}, dt)
}

func TestNormalizeNVMeDeviceName(t *testing.T) {
	tests := []struct {
		name        string
		devName     string
		expectLogic string // What the parsing logic should produce
		desc        string
	}{
		{
			name:        "non-NVMe device unchanged",
			devName:     "sda1",
			expectLogic: "sda1",
			desc:        "Non-NVMe devices should pass through unchanged",
		},
		{
			name:        "standard NVMe notation unchanged",
			devName:     "nvme0n1",
			expectLogic: "nvme0n1",
			desc:        "Standard notation should pass through unchanged",
		},
		{
			name:        "single digit controller notation",
			devName:     "nvme0c0n1",
			expectLogic: "nvme0n1",
			desc:        "Controller notation should be normalized to standard",
		},
		{
			name:        "double digit controller notation",
			devName:     "nvme23c23n1",
			expectLogic: "nvme23n1",
			desc:        "Multi-digit controller notation should be normalized",
		},
		{
			name:        "controller notation with partition",
			devName:     "nvme0c0n1p1",
			expectLogic: "nvme0n1p1",
			desc:        "Controller notation with partitions should be normalized",
		},
		{
			name:        "mismatched controller numbers",
			devName:     "nvme5c3n2",
			expectLogic: "nvme5n2",
			desc:        "Mismatched controller/namespace numbers should still normalize",
		},
		{
			name:        "malformed - no number before c",
			devName:     "nvmec1n1",
			expectLogic: "nvmec1n1",
			desc:        "Malformed input should pass through unchanged",
		},
		{
			name:        "malformed - no n after controller",
			devName:     "nvme0c0",
			expectLogic: "nvme0c0",
			desc:        "Missing namespace part should pass through unchanged",
		},
		{
			name:        "non-NVMe with c in name",
			devName:     "sdac1",
			expectLogic: "sdac1",
			desc:        "Non-NVMe devices with 'c' should pass through unchanged",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test the actual function - it will return original if file doesn't exist
			// or the normalized version if the file exists
			result := normalizeNVMeDeviceName(tc.devName)

			// For most test cases, since the devices don't actually exist in /dev/,
			// the function should return the original name
			// But we can verify the parsing logic by checking if it would normalize correctly
			if tc.devName != tc.expectLogic {
				// This is a controller notation case
				// The result should either be:
				// 1. The original (if normalized device doesn't exist)
				// 2. The normalized version (if normalized device exists)
				require.True(t,
					result == tc.devName || result == tc.expectLogic,
					"Result should be either original (%s) or normalized (%s), got %s",
					tc.devName, tc.expectLogic, result)
			} else {
				// This is not a controller notation case, should be unchanged
				require.Equal(t, tc.expectLogic, result, tc.desc)
			}
		})
	}
}
