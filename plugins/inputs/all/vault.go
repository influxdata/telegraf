//go:build !custom || inputs || inputs.vault

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/vault" // register plugin
