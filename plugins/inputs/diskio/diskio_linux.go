package diskio

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

type diskInfoCache struct {
	modifiedAt   int64 // Unix Nano timestamp of the last modification of the device. This value is used to invalidate the cache
	udevDataPath string
	sysBlockPath string
	values       map[string]string
}

func (d *DiskIO) diskInfo(devName string) (map[string]string, error) {
	// Check if the device exists
	path := "/dev/" + devName
	var stat unix.Stat_t
	if err := unix.Stat(path, &stat); err != nil {
		return nil, err
	}

	// Check if we already got a cached and valid entry
	ic, ok := d.infoCache[devName]
	if ok && stat.Mtim.Nano() == ic.modifiedAt {
		return ic.values, nil
	}

	// Determine udev properties
	var udevDataPath string
	if ok && len(ic.udevDataPath) > 0 {
		// We can reuse the udev data path from a "previous" entry.
		// This allows us to also "poison" it during test scenarios
		udevDataPath = ic.udevDataPath
	} else {
		major := unix.Major(uint64(stat.Rdev)) //nolint:unconvert // Conversion needed for some architectures
		minor := unix.Minor(uint64(stat.Rdev)) //nolint:unconvert // Conversion needed for some architectures
		udevDataPath = fmt.Sprintf("/run/udev/data/b%d:%d", major, minor)
		if _, err := os.Stat(udevDataPath); err != nil {
			// This path failed, try the fallback .udev style (non-systemd)
			udevDataPath = "/dev/.udev/db/block:" + devName
			if _, err := os.Stat(udevDataPath); err != nil {
				// Giving up, cannot retrieve disk info
				return nil, err
			}
		}
	}

	info, err := readUdevData(udevDataPath)
	if err != nil {
		return nil, err
	}

	// Read additional (optional) device properties
	var sysBlockPath string
	if ok && len(ic.sysBlockPath) > 0 {
		// We can reuse the /sys block path from a "previous" entry.
		// This allows us to also "poison" it during test scenarios
		sysBlockPath = ic.sysBlockPath
	} else {
		sysBlockPath = "/sys/class/block/" + devName
	}

	devInfo, err := readDevData(sysBlockPath)
	if err == nil {
		for k, v := range devInfo {
			info[k] = v
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	d.infoCache[devName] = diskInfoCache{
		modifiedAt:   stat.Mtim.Nano(),
		udevDataPath: udevDataPath,
		values:       info,
	}

	return info, nil
}

func readUdevData(path string) (map[string]string, error) {
	// Final open of the confirmed (or the previously detected/used) udev file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info := make(map[string]string)
	scnr := bufio.NewScanner(f)
	var devlinks bytes.Buffer
	for scnr.Scan() {
		l := scnr.Text()
		if len(l) < 4 {
			continue
		}
		if l[:2] == "S:" {
			if devlinks.Len() > 0 {
				devlinks.WriteString(" ")
			}
			devlinks.WriteString("/dev/")
			devlinks.WriteString(l[2:])
			continue
		}
		if l[:2] != "E:" {
			continue
		}
		kv := strings.SplitN(l[2:], "=", 2)
		if len(kv) < 2 {
			continue
		}
		info[kv[0]] = kv[1]
	}

	if devlinks.Len() > 0 {
		info["DEVLINKS"] = devlinks.String()
	}

	return info, nil
}

func readDevData(path string) (map[string]string, error) {
	// Open the file and read line-wise
	f, err := os.Open(filepath.Join(path, "uevent"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read DEVNAME and DEVTYPE
	info := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "DEV") {
			continue
		}

		k, v, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		info[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	if d, found := info["DEVNAME"]; found && !strings.HasPrefix(d, "/dev") {
		info["DEVNAME"] = "/dev/" + d
	}

	// Find the DEVPATH property
	if devlnk, err := filepath.EvalSymlinks(filepath.Join(path, "device")); err == nil {
		devlnk = filepath.Join(devlnk, filepath.Base(path))
		devlnk = strings.TrimPrefix(devlnk, "/sys")
		info["DEVPATH"] = devlnk
	}

	return info, nil
}

func resolveName(name string) string {
	resolved, err := filepath.EvalSymlinks(name)
	if err == nil {
		return resolved
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return name
	}
	// Try to prepend "/dev"
	resolved, err = filepath.EvalSymlinks("/dev/" + name)
	if err != nil {
		return name
	}

	return resolved
}

func getDeviceWWID(name string) string {
	path := fmt.Sprintf("/sys/block/%s/wwid", filepath.Base(name))
	buf, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(string(buf), "\n")
}
