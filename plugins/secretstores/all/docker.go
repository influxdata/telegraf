//go:build !custom || secretstores || secretstores.docker

package all

import _ "github.com/influxdata/telegraf/plugins/secretstores/docker" // register plugin
