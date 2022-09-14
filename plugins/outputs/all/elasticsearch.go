//go:build !custom || outputs || outputs.elasticsearch

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/elasticsearch" // register plugin
