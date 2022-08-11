//go:build !custom || inputs || inputs.salesforce

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/salesforce" // register plugin
