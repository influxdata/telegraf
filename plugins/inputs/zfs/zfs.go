//go:generate ../../../tools/readme_config_includer/generator
package zfs

import (
	_ "embed"

	"github.com/influxdata/telegraf"
)

//go:embed sample.conf
var sampleConfig string

type Sysctl func(metric string) ([]string, error)
type Zpool func() ([]string, error)
type Zdataset func(properties []string) ([]string, error)
type Uname func() (string, error)

type Zfs struct {
	KstatPath      string
	KstatMetrics   []string
	PoolMetrics    bool
	DatasetMetrics bool
	Log            telegraf.Logger `toml:"-"`

	sysctl   Sysctl   //nolint:unused // False positive - this var is used for non-default build tag: freebsd
	zpool    Zpool    //nolint:unused // False positive - this var is used for non-default build tag: freebsd
	zdataset Zdataset //nolint:unused // False positive - this var is used for non-default build tag: freebsd
	uname    Uname    //nolint:unused // False positive - this var is used for non-default build tag: freebsd
	version  int64    //nolint:unused // False positive - this var is used for non-default build tag: freebsd
}

func (*Zfs) SampleConfig() string {
	return sampleConfig
}
