package zfs

import (
	"github.com/influxdata/telegraf"
)

type Sysctl func(metric string) ([]string, error)
type Zpool func() ([]string, error)
type Zdataset func(properties []string) ([]string, error)

type Zfs struct {
	KstatPath      string
	KstatMetrics   []string
	PoolMetrics    bool
	DatasetMetrics bool
	sysctl         Sysctl          //nolint:varcheck,unused // False positive - this var is used for non-default build tag: freebsd
	zpool          Zpool           //nolint:varcheck,unused // False positive - this var is used for non-default build tag: freebsd
	zdataset       Zdataset        //nolint:varcheck,unused // False positive - this var is used for non-default build tag: freebsd
	Log            telegraf.Logger `toml:"-"`
}
