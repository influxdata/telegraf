//go:build !custom || inputs || inputs.seaweedfs

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/seaweedfs" // register plugin
