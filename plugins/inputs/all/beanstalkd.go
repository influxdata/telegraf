//go:build !custom || inputs || inputs.beanstalkd

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/beanstalkd"
)
