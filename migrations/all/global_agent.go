//go:build !custom || migrations

package all

import _ "github.com/influxdata/telegraf/migrations/global_agent" // register migration
