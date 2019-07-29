// +build dragonfly linux netbsd openbsd solaris

package postfix

import (
	"syscall"
	"time"
)

func statCTime(sys interface{}) time.Time {
	stat, ok := sys.(*syscall.Stat_t)
	if !ok {
		return time.Time{}
	}
	return time.Unix(stat.Ctim.Unix())
}
