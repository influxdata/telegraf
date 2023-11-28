//go:build !custom || secretstores || secretstores.systemd

package all

import _ "github.com/influxdata/telegraf/plugins/secretstores/systemd" // register plugin
