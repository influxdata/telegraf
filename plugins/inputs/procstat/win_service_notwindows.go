//go:build !windows
// +build !windows

package procstat

import (
	"fmt"
)

func queryPidWithWinServiceName(_ string) (uint32, error) {
	return 0, fmt.Errorf("os not support win_service option")
}
