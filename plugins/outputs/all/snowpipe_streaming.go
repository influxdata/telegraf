//go:build !custom || outputs || outputs.snowpipe_streaming

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/snowpipe_streaming" // register plugin
