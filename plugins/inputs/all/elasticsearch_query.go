//go:build !custom || inputs || inputs.elasticsearch_query

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/elasticsearch_query" // register plugin
