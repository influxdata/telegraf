//go:build !custom || outputs || outputs.opensearch

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/opensearch" // register plugin
