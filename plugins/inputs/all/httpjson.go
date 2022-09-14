//go:build !custom || inputs || inputs.httpjson

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/httpjson" // register plugin
