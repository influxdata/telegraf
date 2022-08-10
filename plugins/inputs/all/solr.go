//go:build !custom || inputs || inputs.solr

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/solr"
)
