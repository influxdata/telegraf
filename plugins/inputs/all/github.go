//go:build !custom || inputs || inputs.github

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/github"
)
