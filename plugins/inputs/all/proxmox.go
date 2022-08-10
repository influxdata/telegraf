//go:build !custom || inputs || inputs.proxmox

package all

import (
	_ "github.com/influxdata/telegraf/plugins/inputs/proxmox"
)
