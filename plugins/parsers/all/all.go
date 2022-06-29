package all

import (
	//Blank imports for plugins to register themselves
	_ "github.com/influxdata/telegraf/plugins/parsers/collectd"
	_ "github.com/influxdata/telegraf/plugins/parsers/csv"
	_ "github.com/influxdata/telegraf/plugins/parsers/dropwizard"
	_ "github.com/influxdata/telegraf/plugins/parsers/form_urlencoded"
	_ "github.com/influxdata/telegraf/plugins/parsers/graphite"
	_ "github.com/influxdata/telegraf/plugins/parsers/json"
	_ "github.com/influxdata/telegraf/plugins/parsers/json_v2"
	_ "github.com/influxdata/telegraf/plugins/parsers/logfmt"
	_ "github.com/influxdata/telegraf/plugins/parsers/value"
	_ "github.com/influxdata/telegraf/plugins/parsers/wavefront"
	_ "github.com/influxdata/telegraf/plugins/parsers/xpath"
)
