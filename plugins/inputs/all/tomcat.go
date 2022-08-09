//go:build all || inputs || inputs.tomcat

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/tomcat"
)
