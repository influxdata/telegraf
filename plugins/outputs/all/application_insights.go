//go:build !custom || outputs || outputs.application_insights

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/application_insights" // register plugin
