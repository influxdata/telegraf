//go:build !custom || inputs || inputs.slurm_v0038

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/slurm_v0038" // register plugin
