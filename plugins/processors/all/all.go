package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/clone"
	_ "github.com/influxdata/telegraf/plugins/processors/converter"
	_ "github.com/influxdata/telegraf/plugins/processors/date"
	_ "github.com/influxdata/telegraf/plugins/processors/enum"
	_ "github.com/influxdata/telegraf/plugins/processors/override"
	_ "github.com/influxdata/telegraf/plugins/processors/parser"
	_ "github.com/influxdata/telegraf/plugins/processors/pivot"
	_ "github.com/influxdata/telegraf/plugins/processors/printer"
	_ "github.com/influxdata/telegraf/plugins/processors/regex"
	_ "github.com/influxdata/telegraf/plugins/processors/rename"
	_ "github.com/influxdata/telegraf/plugins/processors/strings"
	_ "github.com/influxdata/telegraf/plugins/processors/tag_limit"
	_ "github.com/influxdata/telegraf/plugins/processors/template"
	_ "github.com/influxdata/telegraf/plugins/processors/topk"
	_ "github.com/influxdata/telegraf/plugins/processors/unpivot"
)
