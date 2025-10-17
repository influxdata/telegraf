//go:build !custom || secretstores || secretstores.vault

package all

import _ "github.com/influxdata/telegraf/plugins/secretstores/vault" // register plugin
