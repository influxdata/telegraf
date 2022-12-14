//go:build !custom || secretstores || secretstores.http

package all

import _ "github.com/influxdata/telegraf/plugins/secretstores/http" // register plugin
