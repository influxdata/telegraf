//go:build !custom || inputs || inputs.marklogic

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/marklogic"
)
