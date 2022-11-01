//go:build !custom || inputs || inputs.libvirt

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/libvirt" // register plugin
