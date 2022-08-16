package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var categories = []string{
	"aggregators",
	"inputs",
	"outputs",
	"parsers",
	"processors",
}

const description = `
This is a tool build Telegraf with a custom set of plugins. The plugins are
select according to the specified Telegraf configuration files. This allows
to shrink the binary size by only selecting the plugins you really need.
A more detailed documentation is available at
http://github.com/influxdata/telegraf/tools/custom_builder/README.md
`

const examples = `
The following command with customize Telegraf to fit the configuration found
at the default locations

  custom_builder --config /etc/telegraf/telegraf.conf --config-dir /etc/telegraf/telegraf.d

You can the --config and --config-dir multiple times

  custom_builder --config global.conf --config myinputs.conf --config myoutputs.conf

or use one or more remote address(es) to load the config

  custom_builder --config global.conf --config http://myserver/plugins.conf

Combinations of local and remote config as well as config directories are
possible.
`

func usage() {
	_, _ = fmt.Fprint(flag.CommandLine.Output(), description)
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "")
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "  custom_builder [flags]")
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "")
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
	flag.PrintDefaults()
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "")
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "Examples:")
	_, _ = fmt.Fprint(flag.CommandLine.Output(), examples)
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "")
}

func main() {
	var dryrun, showtags, quiet bool
	var configFiles, configDirs []string

	flag.Func("config",
		"Import plugins from configuration file (can be used multiple times)",
		func(s string) error {
			configFiles = append(configFiles, s)
			return nil
		},
	)
	flag.Func("config-dir",
		"Import plugins from configs in the given directory (can be used multiple times)",
		func(s string) error {
			configDirs = append(configDirs, s)
			return nil
		},
	)
	flag.BoolVar(&dryrun, "dry-run", false, "Skip the actual building step")
	flag.BoolVar(&quiet, "quiet", false, "Print fewer log messages")
	flag.BoolVar(&showtags, "tags", false, "Show build-tags used")

	flag.Usage = usage
	flag.Parse()

	// Check configuration options
	if len(configFiles) == 0 && len(configDirs) == 0 {
		log.Fatalln("No configuration specified!")
	}

	// Import the plugin list from Telegraf configuration files
	log.Println("Importing configuration file(s)...")
	cfg, nfiles, err := ImportConfigurations(configFiles, configDirs)
	if err != nil {
		log.Fatalf("Importing configuration(s) failed: %v", err)
	}
	if !quiet {
		log.Printf("Found %d configuration files...", nfiles)
	}

	// Check if we do have a config
	if nfiles == 0 {
		log.Fatalln("No configuration files loaded!")
	}

	// Collect all available plugins
	packages := packageCollection{}
	if err := packages.CollectAvailable(); err != nil {
		log.Fatalf("Collecting plugins failed: %v", err)
	}

	// Process the plugin list with the given config. This will
	// only keep the plugins that adhere to the filtering criteria.
	enabled, err := cfg.Filter(packages)
	if err != nil {
		log.Fatalf("Filtering plugins failed: %v", err)
	}
	if !quiet {
		enabled.Print()
	}

	// Extract the build-tags
	tagset := enabled.ExtractTags()
	if len(tagset) == 0 {
		log.Fatalln("Nothing selected!")
	}
	tags := "custom," + strings.Join(tagset, ",")
	if showtags {
		fmt.Printf("Build tags: %s\n", tags)
	}

	if !dryrun {
		// Perform the build
		var out bytes.Buffer
		makeCmd := exec.Command("make", "clean", "all")
		makeCmd.Env = append(os.Environ(), "BUILDTAGS="+tags)
		makeCmd.Stdout = &out
		makeCmd.Stderr = &out

		if !quiet {
			log.Println("Running build...")
		}
		if err := makeCmd.Run(); err != nil {
			fmt.Println(out.String())
			log.Fatalf("Running make failed: %v", err)
		}
		if !quiet {
			fmt.Println(out.String())
		}
	} else if !quiet {
		log.Println("DRY-RUN: Skipping build.")
	}
}
