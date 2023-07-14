//go:build !custom || secretstores || secretstores.oauth2

package all

import _ "github.com/influxdata/telegraf/plugins/secretstores/oauth2" // register plugin
