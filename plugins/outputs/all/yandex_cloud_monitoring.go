//go:build !custom || outputs || outputs.yandex_cloud_monitoring

package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/yandex_cloud_monitoring"
)
