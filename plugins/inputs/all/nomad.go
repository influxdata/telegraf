//go:build !custom || inputs || inputs.nomad

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/nomad" // register plugin
