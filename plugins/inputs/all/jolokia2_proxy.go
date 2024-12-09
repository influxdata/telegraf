//go:build !custom || inputs || inputs.jolokia2_proxy

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/jolokia2_proxy" // register plugin
