//go:build windows
// +build windows

//go:generate goversioninfo -icon=../../assets/windows/tiger.ico

package main

import (
	"fmt"
	"os"

	"github.com/influxdata/telegraf/logger"
	"github.com/kardianos/service"
	"github.com/urfave/cli/v2"
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

func (a *Telegraf) Run() error {
	// Register the eventlog logging target for windows.
	err := logger.RegisterEventLogger(a.serviceName)
	if err != nil {
		return err
	}

	if !a.windowsRunAsService() {
		stop = make(chan struct{})
		return a.reloadLoop()
	}

	return a.runAsWindowsService()
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

func (a *Telegraf) runAsWindowsService() error {
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" { // Should never happen
		programFiles = "C:\\Program Files"
	}
	svcConfig := &service.Config{
		Name:        a.serviceName,
		DisplayName: a.serviceDisplayName,
		Description: "Collects data using a series of plugins and publishes it to " +
			"another series of plugins.",
		Arguments: []string{"--config", programFiles + "\\Telegraf\\telegraf.conf"},
	}

	prg := &program{
		Telegraf: a,
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return fmt.Errorf("E! " + err.Error())
	}
	// Handle the --service flag here to prevent any issues with tooling that
	// may not have an interactive session, e.g. installing from Ansible.
	if a.service != "" {
		if len(a.config) > 0 {
			svcConfig.Arguments = []string{}
		}
		for _, fConfig := range a.config {
			svcConfig.Arguments = append(svcConfig.Arguments, "--config", fConfig)
		}

		for _, fConfigDirectory := range a.configDir {
			svcConfig.Arguments = append(svcConfig.Arguments, "--config-directory", fConfigDirectory)
		}

		//set servicename to service cmd line, to have a custom name after relaunch as a service
		svcConfig.Arguments = append(svcConfig.Arguments, "--service-name", a.serviceName)

		if a.serviceAutoRestart {
			svcConfig.Option = service.KeyValue{"OnFailure": "restart", "OnFailureDelayDuration": a.serviceRestartDelay}
		}

		err := service.Control(s, a.service)
		if err != nil {
			return fmt.Errorf("E! " + err.Error())
		}
	} else {
		logger.SetupLogging(logger.LogConfig{LogTarget: logger.LogTargetEventlog})
		err = s.Run()
		if err != nil {
			return fmt.Errorf("E! " + err.Error())
		}
	}
	return nil
}

// Return true if Telegraf should create a Windows service.
func (a *Telegraf) windowsRunAsService() bool {
	if a.service != "" {
		return true
	}

	if a.console {
		return false
	}

	return !service.Interactive()
}
