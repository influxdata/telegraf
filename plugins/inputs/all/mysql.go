//go:build !custom || inputs || inputs.mysql

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/mysql" // register plugin
