package diskio

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

type diskInfoCache struct {
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
	if ok {
		return ic.values, nil
	}

	major := stat.Rdev >> 8 & 0xff
	minor := stat.Rdev & 0xff
	udevDataPath := fmt.Sprintf("%s/b%d:%d", udevPath, major, minor)

	di := map[string]string{}

	s.infoCache[devName] = diskInfoCache{
		udevDataPath: udevDataPath,
		values:       di,
	}

	f, err := os.Open(udevDataPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scnr := bufio.NewScanner(f)
	for scnr.Scan() {
		l := scnr.Text()
		if len(l) < 4 || l[:2] != "E:" {
			continue
		}
		kv := strings.SplitN(l[2:], "=", 2)
		if len(kv) < 2 {
			continue
		}
		di[kv[0]] = kv[1]
	}

	return di, nil
}
