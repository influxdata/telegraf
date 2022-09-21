package main

import (
	"log"

	"github.com/influxdata/telegraf/cmd/telegraf/app"
)

func main() {
	pprof := app.NewPprofServer()
	runner := app.NewRunner(app.WithPProfServer(pprof))
	err := runner.RunApp()
	if err != nil {
		log.Fatalf("E! %s", err)
	}
}
