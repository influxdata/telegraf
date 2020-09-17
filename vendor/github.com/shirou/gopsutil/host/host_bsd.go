// +build darwin freebsd openbsd

package host

import (
	"context"
	"sync/atomic"
	"time"

	"golang.org/x/sys/unix"
)

// cachedBootTime must be accessed via atomic.Load/StoreUint64
var cachedBootTime uint64

func BootTime() (uint64, error) {
	return BootTimeWithContext(context.Background())
}

func BootTimeWithContext(ctx context.Context) (uint64, error) {
	t := atomic.LoadUint64(&cachedBootTime)
	if t != 0 {
		return t, nil
	}
	tv, err := unix.SysctlTimeval("kern.boottime")
	if err != nil {
		return 0, err
	}

	atomic.StoreUint64(&cachedBootTime, uint64(tv.Sec))

	return uint64(tv.Sec), nil
}

func uptime(boot uint64) uint64 {
	return uint64(time.Now().Unix()) - boot
}

func Uptime() (uint64, error) {
	return UptimeWithContext(context.Background())
}

func UptimeWithContext(ctx context.Context) (uint64, error) {
	boot, err := BootTimeWithContext(ctx)
	if err != nil {
		return 0, err
	}
	return uptime(boot), nil
}
