package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	// TODO: import your plugins
	// _ "github.com/my_github_user/my_plugin_repo/plugins/inputs/mypluginname"

	"github.com/influxdata/telegraf/plugins/common/shim"
)

var pollInterval = flag.Duration("poll_interval", 1*time.Second, "how often to send metrics")
var pollIntervalDisabled = flag.Bool("poll_interval_disabled", false, "set to true to disable polling. You want to use this when you are sending metrics on your own schedule")
var configFile = flag.String("config", "", "path to the config file for this plugin")
var err error

// This is designed to be simple; Just change the import above, and you're good.
//
// However, if you want to do all your config in code, you can like so:
//
// // initialize your plugin with any settings you want
//
//	myInput := &mypluginname.MyPlugin{
//		DefaultSettingHere: 3,
//	}
//
// shim := shim.New()
//
// shim.AddInput(myInput)
//
// // now the shim.Run() call as below. Note the shim is only intended to run a single plugin.
func main() {
	// parse command line options
	flag.Parse()
	if *pollIntervalDisabled {
		*pollInterval = shim.PollIntervalDisabled
	}

	// create the shim. This is what will run your plugins.
	shimLayer := shim.New()

	// If no config is specified, all imported plugins are loaded.
	// otherwise, follow what the config asks for.
	// Check for settings from a config toml file,
	// (or just use whatever plugins were imported above)
	if err = shimLayer.LoadConfig(configFile); err != nil {
		fmt.Fprintf(os.Stderr, "Err loading input: %s\n", err)
		os.Exit(1)
	}

	// run a single plugin until stdin closes, or we receive a termination signal
	if err = shimLayer.Run(*pollInterval); err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s\n", err)
		os.Exit(1)
	}
}
