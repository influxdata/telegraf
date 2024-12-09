//go:build !custom || parsers || parsers.form_urlencoded

package all

import _ "github.com/influxdata/telegraf/plugins/parsers/form_urlencoded" // register plugin
