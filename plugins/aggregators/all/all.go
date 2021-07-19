package all

import (
	//Blank imports for plugins to register themselves
	_ "github.com/influxdata/telegraf/plugins/aggregators/basicstats"
	_ "github.com/influxdata/telegraf/plugins/aggregators/derivative"
	_ "github.com/influxdata/telegraf/plugins/aggregators/final"
	_ "github.com/influxdata/telegraf/plugins/aggregators/histogram"
	_ "github.com/influxdata/telegraf/plugins/aggregators/merge"
	_ "github.com/influxdata/telegraf/plugins/aggregators/minmax"
	_ "github.com/influxdata/telegraf/plugins/aggregators/quantile"
	_ "github.com/influxdata/telegraf/plugins/aggregators/starlark"
	_ "github.com/influxdata/telegraf/plugins/aggregators/valuecounter"
)
