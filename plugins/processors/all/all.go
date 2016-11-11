package all

import (
	_ "github.com/influxdata/telegraf/plugins/processors/aws/ec2"
	_ "github.com/influxdata/telegraf/plugins/processors/aws/elb"
	_ "github.com/influxdata/telegraf/plugins/processors/aws/rds"
	_ "github.com/influxdata/telegraf/plugins/processors/aws/sqs"
	_ "github.com/influxdata/telegraf/plugins/processors/printer"
)
