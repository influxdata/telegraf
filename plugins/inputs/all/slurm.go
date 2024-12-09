//go:build !custom || inputs || inputs.slurm

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/slurm" // register plugin
