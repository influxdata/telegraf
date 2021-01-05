package main

// can make telegraf ~82% smaller

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf/builder/builder"
)

var (
	pluginsFlag *string = flag.String("plugins", "", "inputs.cpu,processors.starlark,outputs.influxdb_v2")
	goArchFlag  *string = flag.String("GOARCH", "amd64", "GOARCH, eg amd64")
	goOSFlag    *string = flag.String("GOOS", "darwin", "GOOS, eg darwin")
	// buildTagFlag *string = flag.String("sha", "v1.17.0", "telegraf tag or SHA to build, eg v1.17.0")
)

func main() {
	flag.Parse()
	if *pluginsFlag == "" {
		flag.Usage()
		os.Exit(1)
	}
	plugins := strings.Split(*pluginsFlag, ",")

	build := &builder.Build{
		Plugins: plugins,
		GOOS:    *goOSFlag,
		GOARCH:  *goArchFlag,
	}
	start := time.Now()
	if err := build.Compile(); err != nil {
		panic(err)
	}
	log.Printf("Built %d plugins in %s to %s\n", len(build.Plugins), time.Since(start), "tmp/telegraf")
}
