//go:build !custom || inputs || inputs.jenkins_builds

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/jenkins_builds" // register plugin
