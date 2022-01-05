package all

import (
	//Blank imports for plugins to register themselves
	_ "github.com/influxdata/telegraf/plugins/outputs/opentsdb"
	_ "github.com/influxdata/telegraf/plugins/outputs/wavefront"
)
