//go:build !custom || inputs || inputs.slurm_v0041

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/slurm_v0041" // register plugin
