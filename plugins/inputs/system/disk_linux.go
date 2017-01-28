package system

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
)

type diskInfoCache struct {
	stat   syscall.Stat_t
	values map[string]string
}

var udevPath = "/run/udev/data"

func (s *DiskIOStats) diskInfo(devName string) (map[string]string, error) {
	fi, err := os.Stat("/dev/" + devName)
	if err != nil {
		return nil, err
	}
	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, nil
	}

	if s.infoCache == nil {
		s.infoCache = map[string]diskInfoCache{}
	}
	ic, ok := s.infoCache[devName]
	if ok {
		return ic.values, nil
	} else {
		ic = diskInfoCache{
			stat:   *stat,
			values: map[string]string{},
		}
		s.infoCache[devName] = ic
	}
	di := ic.values

	major := stat.Rdev >> 8 & 0xff
	minor := stat.Rdev & 0xff

	f, err := os.Open(fmt.Sprintf("%s/b%d:%d", udevPath, major, minor))
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
