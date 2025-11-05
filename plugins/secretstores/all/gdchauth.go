//go:build !custom || secretstores || secretstores.gdchauth

package all

import _ "github.com/influxdata/telegraf/plugins/secretstores/gdchauth" // register plugin
