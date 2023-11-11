//go:build !custom || inputs || inputs.qbittorrent

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/qbittorrent" // register plugin
