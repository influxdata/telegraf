//go:build !custom || (migrations && (inputs || inputs.filecount))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_filecount" // register migration
