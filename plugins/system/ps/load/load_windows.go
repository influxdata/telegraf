// +build windows

package load

import (
	common "github.com/koksan83/telegraf/plugins/system/ps/common"
)

func LoadAvg() (*LoadAvgStat, error) {
	ret := LoadAvgStat{}

	return &ret, common.NotImplementedError
}
