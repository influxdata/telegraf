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
	values       map[string]string
}

func (d *DiskIO) diskInfo(devName string) (map[string]string, error) {
	var err error
	var stat unix.Stat_t

	path := "/dev/" + devName
	err = unix.Stat(path, &stat)
	if err != nil {
		return nil, err
	}

	if d.infoCache == nil {
		d.infoCache = map[string]diskInfoCache{}
	}
	ic, ok := d.infoCache[devName]

	if ok && stat.Mtim.Nano() == ic.modifiedAt {
		return ic.values, nil
	}

	var udevDataPath string
	if ok && len(ic.udevDataPath) > 0 {
		// We can reuse the udev data path from a "previous" entry.
		// This allows us to also "poison" it during test scenarios
		udevDataPath = ic.udevDataPath
	} else {
		major := unix.Major(uint64(stat.Rdev)) //nolint:unconvert // Conversion needed for some architectures
		minor := unix.Minor(uint64(stat.Rdev)) //nolint:unconvert // Conversion needed for some architectures
		udevDataPath = fmt.Sprintf("/run/udev/data/b%d:%d", major, minor)

		_, err := os.Stat(udevDataPath)
		if err != nil {
			// This path failed, try the fallback .udev style (non-systemd)
			udevDataPath = fmt.Sprintf("/dev/.udev/db/block:%s", devName)
			_, err := os.Stat(udevDataPath)
			if err != nil {
				// Giving up, cannot retrieve disk info
				return nil, err
			}
		}
	}
	// Final open of the confirmed (or the previously detected/used) udev file
	f, err := os.Open(udevDataPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	di := map[string]string{}

	d.infoCache[devName] = diskInfoCache{
		modifiedAt:   stat.Mtim.Nano(),
		udevDataPath: udevDataPath,
		values:       di,
	}

	scnr := bufio.NewScanner(f)
	var devlinks bytes.Buffer
	for scnr.Scan() {
		l := scnr.Text()
		if len(l) < 4 {
			continue
		}
		if l[:2] == "S:" {
			if devlinks.Len() > 0 {
				//nolint:errcheck,revive // this will never fail
				devlinks.WriteString(" ")
			}
			//nolint:errcheck,revive // this will never fail
			devlinks.WriteString("/dev/")
			//nolint:errcheck,revive // this will never fail
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
		di[kv[0]] = kv[1]
	}

	if devlinks.Len() > 0 {
		di["DEVLINKS"] = devlinks.String()
	}

	return di, nil
}

func resolveName(name string) string {
	resolved, err := filepath.EvalSymlinks(name)
	if err == nil {
		return resolved
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return name
	}
	// Try to prepend "/dev"
	resolved, err = filepath.EvalSymlinks(filepath.Join("/dev", name))
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
