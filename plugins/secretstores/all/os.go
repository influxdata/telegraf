//go:build !custom || secretstores || secretstores.os

package all

import _ "github.com/influxdata/telegraf/plugins/secretstores/os" // register plugin
