// +build windows

package load

import (
	common "github.com/influxdb/tivan/plugins/system/ps/common"
)

func LoadAvg() (*LoadAvgStat, error) {
	ret := LoadAvgStat{}

	return &ret, common.NotImplementedError
}
