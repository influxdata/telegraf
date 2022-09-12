//go:build !custom || inputs || inputs.jenkins

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/jenkins" // register plugin
