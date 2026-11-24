//go:build windows

// Command handling for configuration "service" command
package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/urfave/cli/v2"
	"golang.org/x/sys/windows"

	"github.com/influxdata/telegraf/cmd/telegraf/commands"
)

func getServiceActions(outputBuffer io.Writer) commands.ServiceActions {
	return commands.ServiceActions{
		Install: func(cCtx *cli.Context) error {
			cfg := &serviceConfig{
				displayName:  cCtx.String("display-name"),
				restartDelay: cCtx.String("restart-delay"),
				autoRestart:  cCtx.Bool("auto-restart"),

				configs:    cCtx.StringSlice("config"),
				configDirs: cCtx.StringSlice("config-directory"),
			}
			name := cCtx.String("service-name")
			if err := installService(name, cfg); err != nil {
				return err
			}
			fmt.Fprintf(outputBuffer, "Successfully installed service %q\n", name)
			return nil
		},
		Uninstall: func(cCtx *cli.Context) error {
			name := cCtx.String("service-name")
			if err := uninstallService(name); err != nil {
				return err
			}
			fmt.Fprintf(outputBuffer, "Successfully uninstalled service %q\n", name)
			return nil
		},
		Start: func(cCtx *cli.Context) error {
			name := cCtx.String("service-name")
			if err := startService(name); err != nil {
				return err
			}
			fmt.Fprintf(outputBuffer, "Successfully started service %q\n", name)
			return nil
		},
		Stop: func(cCtx *cli.Context) error {
			name := cCtx.String("service-name")
			if err := stopService(name); err != nil {
				if errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
					fmt.Fprintf(outputBuffer, "Service %q not started\n", name)
					return nil
				}
				return err
			}
			fmt.Fprintf(outputBuffer, "Successfully stopped service %q\n", name)
			return nil
		},
		Status: func(cCtx *cli.Context) error {
			name := cCtx.String("service-name")
			status, err := queryService(name)
			if err != nil {
				return err
			}
			fmt.Fprintf(outputBuffer, "Service %q is in %q state\n", name, status)
			return nil
		},
	}
}
