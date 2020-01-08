// +build linux

package diskio

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

// Using lsblk to grab a block device which as a LABEL value
// Also retrieves the DiskInfo using udevadm info
func GetSuitableBlockDevice() (string, map[string]string, error) {
	var o bytes.Buffer
	c := exec.Command("lsblk", "-n", "-o", "KNAME,MAJ:MIN,LABEL,FSTYPE")
	c.Stdout = &o
	err := c.Run()
	if err != nil {
		return "", nil, err
	}
	devices := strings.Split(o.String(), "\n")
	for _, dev := range devices {
		devinfo := strings.Fields(dev)
		if len(devinfo) >= 3 {
			// Fetching the data via udevadm
			var b bytes.Buffer
			o := exec.Command("udevadm", "info", "-q", "property", "-n", devinfo[0])
			o.Stdout = &b
			err := o.Run()
			if err != nil {
				return "", nil, err
			}
			// Breaking the result of udevadm into key/value pairs
			di := make(map[string]string)
			for _, line := range strings.Split(b.String(), "\n") {
				info := strings.Split(line, "=")
				if len(info) == 2 {
					di[info[0]] = info[1]
				}
			}
			return devinfo[0], di, nil
		}
	}
	return "", nil, os.ErrNotExist
}

func TestDiskInfo(t *testing.T) {
	devname, devinfo, err := GetSuitableBlockDevice()
	assert.NoError(t, err)

	// Test the code now
	s := &DiskIO{}
	di, err := s.diskInfo(devname)
	require.NoError(t, err)
	assert.Equal(t, devinfo["ID_FS_LABEL"], di["ID_FS_LABEL"])
	assert.Equal(t, devinfo["ID_PART_ENTRY_UUID"], di["ID_PART_ENTRY_UUID"])

	// test that data is cached
	before := s.infoCache[devname].modifiedAt

	// resetting cache
	s.infoCache[devname] = diskInfoCache{}

	// fetching disk info again should yield same modifiedAt timestamp
	di, err = s.diskInfo(devname)
	assert.NoError(t, err)
	assert.Equal(t, before, s.infoCache[devname].modifiedAt, "Unexpected difference in modified timestamp")
}

// DiskIOStats.diskName isn't a linux specific function, but dependent
// functions are a no-op on non-Linux.
func TestDiskIOStats_diskName(t *testing.T) {
	devname, devinfo, err := GetSuitableBlockDevice()
	assert.NoError(t, err)

	tests := []struct {
		templates []string
		expected  string
	}{
		{[]string{"$ID_FS_LABEL"}, devinfo["ID_FS_LABEL"]},
		{[]string{"${ID_FS_LABEL}"}, devinfo["ID_FS_LABEL"]},
		{[]string{"x$ID_FS_LABEL"}, fmt.Sprintf("x%s", devinfo["ID_FS_LABEL"])},
		{[]string{"x${ID_FS_LABEL}x"}, fmt.Sprintf("x%sx", devinfo["ID_FS_LABEL"])},
		{[]string{"$MISSING", "$ID_FS_LABEL"}, devinfo["ID_FS_LABEL"]},
		{[]string{"$ID_FS_LABEL", "$MY_PARAM_2"}, devinfo["ID_FS_LABEL"]},
		{[]string{"$MISSING"}, devname},
		{[]string{"$ID_BUS/$ID_FS_LABEL"}, fmt.Sprintf("%s/%s", devinfo["ID_BUS"], devinfo["ID_FS_LABEL"])},
		{[]string{"$MY_PARAM_2/$MISSING"}, devname},
	}

	for _, tc := range tests {
		s := DiskIO{
			NameTemplates: tc.templates,
		}
		name, _ := s.diskName(devname)
		assert.Equal(t, tc.expected, name, "Templates: %#v", tc.templates)
	}
}

// DiskIOStats.diskTags isn't a linux specific function, but dependent
// functions are a no-op on non-Linux.
func TestDiskIOStats_diskTags(t *testing.T) {
	devname, devinfo, err := GetSuitableBlockDevice()
	assert.NoError(t, err)

	s := &DiskIO{
		DeviceTags: []string{"ID_FS_LABEL"},
	}
	dt := s.diskTags(devname)
	assert.Equal(t, map[string]string{"ID_FS_LABEL": devinfo["ID_FS_LABEL"]}, dt)
}
