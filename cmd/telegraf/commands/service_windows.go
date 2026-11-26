//go:build windows

// Command handling for configuration "service" command
package commands

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
	"golang.org/x/sys/windows"

	"github.com/influxdata/telegraf/cmd/telegraf/commands/winservice"
)

func getServiceActions() serviceActions {
	return serviceActions{
		Install:   serviceInstallAction,
		Uninstall: serviceUninstallAction,
		Start:     serviceStartAction,
		Stop:      serviceStopAction,
		Status:    serviceStatusAction,
	}
}

func serviceInstallAction(cCtx *cli.Context) error {
	cfg := &winservice.Config{
		DisplayName:  cCtx.String("display-name"),
		RestartDelay: cCtx.String("restart-delay"),
		AutoRestart:  cCtx.Bool("auto-restart"),

		Configs:    cCtx.StringSlice("config"),
		ConfigDirs: cCtx.StringSlice("config-directory"),
	}
	name := cCtx.String("service-name")
	if err := winservice.Install(name, cfg); err != nil {
		return err
	}
	fmt.Fprintf(cCtx.App.Writer, "Successfully installed service %q\n", name)
	return nil
}

func serviceUninstallAction(cCtx *cli.Context) error {
	name := cCtx.String("service-name")
	if err := winservice.Uninstall(name); err != nil {
		return err
	}
	fmt.Fprintf(cCtx.App.Writer, "Successfully uninstalled service %q\n", name)
	return nil
}

func serviceStartAction(cCtx *cli.Context) error {
	name := cCtx.String("service-name")
	if err := winservice.Start(name); err != nil {
		return err
	}
	fmt.Fprintf(cCtx.App.Writer, "Successfully started service %q\n", name)
	return nil
}

func serviceStopAction(cCtx *cli.Context) error {
	name := cCtx.String("service-name")
	if err := winservice.Stop(name); err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
			fmt.Fprintf(cCtx.App.Writer, "Service %q not started\n", name)
			return nil
		}
		return err
	}
	fmt.Fprintf(cCtx.App.Writer, "Successfully stopped service %q\n", name)
	return nil
}

func serviceStatusAction(cCtx *cli.Context) error {
	name := cCtx.String("service-name")
	status, err := winservice.Query(name)
	if err != nil {
		return err
	}
	fmt.Fprintf(cCtx.App.Writer, "Service %q is in %q state\n", name, status)
	return nil
}
