//go:build !custom || secretstores || secretstores.jose

package all

import _ "github.com/influxdata/telegraf/plugins/secretstores/jose" // register plugin
