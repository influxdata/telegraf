//go:build !custom || (migrations && (outputs || outputs.wavefront))

package all

import _ "github.com/influxdata/telegraf/migrations/outputs_wavefront" // register migration
