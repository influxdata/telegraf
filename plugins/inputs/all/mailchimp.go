//go:build !custom || inputs || inputs.mailchimp

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/mailchimp" // register plugin
