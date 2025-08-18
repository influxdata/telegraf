//go:build !custom || (migrations && (inputs || inputs.docker))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_docker" // register migration
