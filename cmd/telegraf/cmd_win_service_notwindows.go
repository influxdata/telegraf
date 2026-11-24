//go:build !windows

package main

import (
	"io"

	"github.com/influxdata/telegraf/cmd/telegraf/commands"
)

func getServiceActions(io.Writer) commands.ServiceActions { return commands.ServiceActions{} }
