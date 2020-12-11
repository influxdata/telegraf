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
	modifiedAt   int64 // Unix Nano timestamp of the last modification of the device. This value is used to invalidate the cache
	udevDataPath string
	values       map[string]string
}

var udevPath = "/run/udev/data"

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
	udevDataPath := fmt.Sprintf("%s/b%d:%d", udevPath, major, minor)

	di := map[string]string{}

	s.infoCache[devName] = diskInfoCache{
		modifiedAt:   stat.Mtim.Nano(),
		udevDataPath: udevDataPath,
		values:       di,
	}

	f, err := os.Open(udevDataPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

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
		di[kv[0]] = kv[1]
	}

	if devlinks.Len() > 0 {
		di["DEVLINKS"] = devlinks.String()
	}

	return di, nil
}
