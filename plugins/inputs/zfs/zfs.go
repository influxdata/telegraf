//go:generate ../../../tools/readme_config_includer/generator
package zfs

import (
	_ "embed"

	"github.com/influxdata/telegraf"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embedd the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

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

func (*Zfs) SampleConfig() string {
	return sampleConfig
}
