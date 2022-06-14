package all

import (
	//Blank imports for plugins to register themselves
	_ "github.com/influxdata/telegraf/plugins/parsers/csv"
	_ "github.com/influxdata/telegraf/plugins/parsers/xpath"
)
