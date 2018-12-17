package diskio

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

type diskInfoCache struct {
	modifiedAt int64 // Unix Nano timestamp of the last modification of the device. This value is used to invalidate the cache
	values     map[string]string
}

func (s *DiskIO) diskInfo(devName string) (map[string]string, error) {
	var err error
	var stat unix.Stat_t

	path := "/dev/" + devName
	err = unix.Stat(path, &stat)
	if err != nil {
		return nil, err
	}

	if s.infoCache == nil {
		s.infoCache = map[string]diskInfoCache{}
	}
	ic, ok := s.infoCache[devName]

	if ok && stat.Mtim.Nano() == ic.modifiedAt {
		return ic.values, nil
	}

	major := unix.Major(uint64(stat.Rdev))
	minor := unix.Minor(uint64(stat.Rdev))
	udevV1 := fmt.Sprintf("/dev/.udev/db/block:%s", devName)     // Non-systemd
	udevV2 := fmt.Sprintf("/run/udev/data/b%d:%d", major, minor) // Systemd

	di := map[string]string{}

	s.infoCache[devName] = diskInfoCache{
		modifiedAt: stat.Mtim.Nano(),
		values:     di,
	}

	var udevPath *os.File
	f1, err1 := os.Open(udevV1)
	if err1 == nil {
		udevPath = f1
		defer f1.Close()
	}
	f2, err2 := os.Open(udevV2)
	if err2 == nil {
		udevPath = f2
		defer f2.Close()
	}

	scnr := bufio.NewScanner(udevPath)
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
		di[kv[0]] = kv[1]
	}

	if devlinks.Len() > 0 {
		di["DEVLINKS"] = devlinks.String()
	}

	return di, nil
}
