//go:build !custom || outputs || outputs.nebius_cloud_monitoring

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/nebius_cloud_monitoring" // register plugin
