// +build windows

package main

import (
	"log"
	"os"
	"runtime"

	"github.com/influxdata/telegraf/logger"
	"github.com/kardianos/service"
)

func run(inputFilters, outputFilters, aggregatorFilters, processorFilters []string) {
	if runtime.GOOS == "windows" && windowsRunAsService() {
		runAsWindowsService(
			inputFilters,
			outputFilters,
			aggregatorFilters,
			processorFilters,
		)
	} else {
		stop = make(chan struct{})
		reloadLoop(
			inputFilters,
			outputFilters,
			aggregatorFilters,
			processorFilters,
		)
	}
}

type program struct {
	inputFilters      []string
	outputFilters     []string
	aggregatorFilters []string
	processorFilters  []string
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}
func (p *program) run() {
	stop = make(chan struct{})
	reloadLoop(
		p.inputFilters,
		p.outputFilters,
		p.aggregatorFilters,
		p.processorFilters,
	)
}
func (p *program) Stop(s service.Service) error {
	close(stop)
	return nil
}

func runAsWindowsService(inputFilters, outputFilters, aggregatorFilters, processorFilters []string) {
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" { // Should never happen
		programFiles = "C:\\Program Files"
	}
	svcConfig := &service.Config{
		Name:        *fServiceName,
		DisplayName: *fServiceDisplayName,
		Description: "Collects data using a series of plugins and publishes it to " +
			"another series of plugins.",
		Arguments: []string{"--config", programFiles + "\\Telegraf\\telegraf.conf"},
	}

	prg := &program{
		inputFilters:      inputFilters,
		outputFilters:     outputFilters,
		aggregatorFilters: aggregatorFilters,
		processorFilters:  processorFilters,
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal("E! " + err.Error())
	}
	// Handle the --service flag here to prevent any issues with tooling that
	// may not have an interactive session, e.g. installing from Ansible.
	if *fService != "" {
		if *fConfig != "" {
			svcConfig.Arguments = []string{"--config", *fConfig}
		}
		if *fConfigDirectory != "" {
			svcConfig.Arguments = append(svcConfig.Arguments, "--config-directory", *fConfigDirectory)
		}
		//set servicename to service cmd line, to have a custom name after relaunch as a service
		svcConfig.Arguments = append(svcConfig.Arguments, "--service-name", *fServiceName)

		err := service.Control(s, *fService)
		if err != nil {
			log.Fatal("E! " + err.Error())
		}
		os.Exit(0)
	} else {
		winlogger, err := s.Logger(nil)
		if err == nil {
			//When in service mode, register eventlog target andd setup default logging to eventlog
			logger.RegisterEventLogger(winlogger)
			logger.SetupLogging(logger.LogConfig{LogTarget: logger.LogTargetEventlog})
		}
		err = s.Run()

		if err != nil {
			log.Println("E! " + err.Error())
		}
	}
}

// Return true if Telegraf should create a Windows service.
func windowsRunAsService() bool {
	if *fService != "" {
		return true
	}

	if *fRunAsConsole {
		return false
	}

	return !service.Interactive()
}
