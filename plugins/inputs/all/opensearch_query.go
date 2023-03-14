//go:build !custom || inputs || inputs.opensearch_query

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/opensearch_query" // register plugin
