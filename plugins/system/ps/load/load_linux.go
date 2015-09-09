// +build linux

package load

import (
	"io/ioutil"
	"strconv"
	"strings"
)

func LoadAvg() (*LoadAvgStat, error) {
	filename := "/rootfs/proc/loadavg"
	line, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	values := strings.Fields(string(line))

	load1, err := strconv.ParseFloat(values[0], 64)
	if err != nil {
		return nil, err
	}
	load5, err := strconv.ParseFloat(values[1], 64)
	if err != nil {
		return nil, err
	}
	load15, err := strconv.ParseFloat(values[2], 64)
	if err != nil {
		return nil, err
	}

	ret := &LoadAvgStat{
		Load1:  load1,
		Load5:  load5,
		Load15: load15,
	}

	return ret, nil
}
