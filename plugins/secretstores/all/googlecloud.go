//go:build !custom || secretstores || secretstores.googlecloud

package all

import _ "github.com/influxdata/telegraf/plugins/secretstores/googlecloud" // register plugin
