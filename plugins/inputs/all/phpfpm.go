//go:build !custom || inputs || inputs.phpfpm

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/phpfpm" // register plugin
