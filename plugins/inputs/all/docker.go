//go:build !custom || inputs || inputs.docker

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/docker" // register plugin
