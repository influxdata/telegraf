//go:generate ../../../tools/readme_config_includer/generator
package zfs

import (
	_ "embed"

	"github.com/influxdata/telegraf"
)

//go:embed sample.conf
var sampleConfig string

type Zfs struct {
	KstatPath      string          `toml:"kstatPath"`
	KstatMetrics   []string        `toml:"kstatMetrics"`
	PoolMetrics    bool            `toml:"poolMetrics"`
	DatasetMetrics bool            `toml:"datasetMetrics"`
	Log            telegraf.Logger `toml:"-"`

	sysctl   sysctlF   //nolint:unused // False positive - this var is used for non-default build tag: freebsd
	zpool    zpoolF    //nolint:unused // False positive - this var is used for non-default build tag: freebsd
	zdataset zdatasetF //nolint:unused // False positive - this var is used for non-default build tag: freebsd
	uname    unameF    //nolint:unused // False positive - this var is used for non-default build tag: freebsd
}

type sysctlF func(metric string) ([]string, error)         //nolint:unused // False positive - this var is used for non-default build tag: freebsd
type zpoolF func() ([]string, error)                       //nolint:unused // False positive - this var is used for non-default build tag: freebsd
type zdatasetF func(properties []string) ([]string, error) //nolint:unused // False positive - this var is used for non-default build tag: freebsd
type unameF func() (string, error)                         //nolint:unused // False positive - this var is used for non-default build tag: freebsd

func (*Zfs) SampleConfig() string {
	return sampleConfig
}
