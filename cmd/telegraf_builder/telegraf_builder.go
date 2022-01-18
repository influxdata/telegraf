package main

import (
	"log" //nolint:revive

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/influxdata/telegraf/plugins/aggregators/all"
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	_ "github.com/influxdata/telegraf/plugins/outputs/all"
	_ "github.com/influxdata/telegraf/plugins/processors/all"
	"github.com/influxdata/telegraf/ui/sampleconfig_ui"
)

func main() {
	h := sampleconfig_ui.NewSampleConfigUI()
	if err := tea.NewProgram(h).Start(); err != nil {
		log.Fatalf("E! %s", err)
	}
}
