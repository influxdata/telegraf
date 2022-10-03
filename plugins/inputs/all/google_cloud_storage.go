//go:build !custom || inputs || inputs.google_cloud_storage

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/google_cloud_storage" // register plugin
