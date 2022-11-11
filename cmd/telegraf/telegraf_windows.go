//go:build windows

//go:generate ../../scripts/windows-gen-syso.sh $GOARCH

package main

import (
	"fmt"
	"os"

	"github.com/kardianos/service"
	"github.com/urfave/cli/v2"

	"github.com/influxdata/telegraf/logger"
)

func cliFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "service",
			Usage: "operate on the service (windows only)",
		},
		&cli.StringFlag{
			Name:  "service-name",
			Value: "telegraf",
			Usage: "service name (windows only)",
		},
		&cli.StringFlag{
			Name:  "service-display-name",
			Value: "Telegraf Data Collector Service",
			Usage: "service display name (windows only)",
		},
		&cli.StringFlag{
			Name:  "service-restart-delay",
			Value: "5m",
		},
		&cli.BoolFlag{
			Name:  "service-auto-restart",
			Usage: "auto restart service on failure (windows only)",
		},
		&cli.BoolFlag{
			Name:  "console",
			Usage: "run as console application (windows only)",
		},
	}
}

func (t *Telegraf) Run() error {
	// Register the eventlog logging target for windows.
	err := logger.RegisterEventLogger(t.serviceName)
	if err != nil {
		return err
	}

	if !t.windowsRunAsService() {
		stop = make(chan struct{})
		return t.reloadLoop()
	}

	return t.runAsWindowsService()
}

type program struct {
	*Telegraf
}

func (p *program) Start(s service.Service) error {
	go func() {
		stop = make(chan struct{})
		err := p.reloadLoop()
		if err != nil {
			fmt.Printf("E! %v\n", err)
		}
		close(stop)
	}()
	return nil
}

func (p *program) run(errChan chan error) {
	stop = make(chan struct{})
	err := p.reloadLoop()
	errChan <- err
	close(stop)
}

func (p *program) Stop(s service.Service) error {
	var empty struct{}
	stop <- empty // signal reloadLoop to finish (context cancel)
	<-stop        // wait for reloadLoop to finish and close channel
	return nil
}

func (t *Telegraf) runAsWindowsService() error {
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" { // Should never happen
		programFiles = "C:\\Program Files"
	}
	svcConfig := &service.Config{
		Name:        t.serviceName,
		DisplayName: t.serviceDisplayName,
		Description: "Collects data using a series of plugins and publishes it to " +
			"another series of plugins.",
		Arguments: []string{"--config", programFiles + "\\Telegraf\\telegraf.conf"},
	}

	prg := &program{
		Telegraf: t,
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return err
	}
	// Handle the --service flag here to prevent any issues with tooling that
	// may not have an interactive session, e.g. installing from Ansible.
	if t.service != "" {
		if len(t.config) > 0 {
			svcConfig.Arguments = []string{}
		}
		for _, fConfig := range t.config {
			svcConfig.Arguments = append(svcConfig.Arguments, "--config", fConfig)
		}

		for _, fConfigDirectory := range t.configDir {
			svcConfig.Arguments = append(svcConfig.Arguments, "--config-directory", fConfigDirectory)
		}

		//set servicename to service cmd line, to have a custom name after relaunch as a service
		svcConfig.Arguments = append(svcConfig.Arguments, "--service-name", t.serviceName)

		if t.serviceAutoRestart {
			svcConfig.Option = service.KeyValue{"OnFailure": "restart", "OnFailureDelayDuration": t.serviceRestartDelay}
		}

		err := service.Control(s, t.service)
		if err != nil {
			return err
		}
	} else {
		err = logger.SetupLogging(logger.LogConfig{LogTarget: logger.LogTargetEventlog})
		if err != nil {
			return err
		}

		err = s.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// Return true if Telegraf should create a Windows service.
func (t *Telegraf) windowsRunAsService() bool {
	if t.service != "" {
		return true
	}

	if t.console {
		return false
	}

	return !service.Interactive()
}
