//go:build windows
// +build windows

//go:generate goversioninfo -icon=../../assets/windows/tiger.ico

package main

import (
	"log"
	"os"
	"runtime"

	"github.com/influxdata/telegraf/logger"
	"github.com/kardianos/service"
)

func run(inputFilters, outputFilters []string) {
	// Register the eventlog logging target for windows.
	logger.RegisterEventLogger(*fServiceName)

	if runtime.GOOS == "windows" && windowsRunAsService() {
		runAsWindowsService(
			inputFilters,
			outputFilters,
		)
	} else {
		stop = make(chan struct{})
		reloadLoop(
			inputFilters,
			outputFilters,
		)
	}
}

type program struct {
	inputFilters  []string
	outputFilters []string
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
	)
	close(stop)
}
func (p *program) Stop(s service.Service) error {
	var empty struct{}
	stop <- empty // signal reloadLoop to finish (context cancel)
	<-stop        // wait for reloadLoop to finish and close channel
	return nil
}

func runAsWindowsService(inputFilters, outputFilters []string) {
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
		inputFilters:  inputFilters,
		outputFilters: outputFilters,
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal("E! " + err.Error())
	}
	// Handle the --service flag here to prevent any issues with tooling that
	// may not have an interactive session, e.g. installing from Ansible.
	if *fService != "" {
		if len(fConfigs) > 0 {
			svcConfig.Arguments = []string{}
		}
		for _, fConfig := range fConfigs {
			svcConfig.Arguments = append(svcConfig.Arguments, "--config", fConfig)
		}

		for _, fConfigDirectory := range fConfigDirs {
			svcConfig.Arguments = append(svcConfig.Arguments, "--config-directory", fConfigDirectory)
		}

		//set servicename to service cmd line, to have a custom name after relaunch as a service
		svcConfig.Arguments = append(svcConfig.Arguments, "--service-name", *fServiceName)

		if *fServiceAutoRestart {
			svcConfig.Option = service.KeyValue{"OnFailure": "restart", "OnFailureDelayDuration": *fServiceRestartDelay}
		}

		err := service.Control(s, *fService)
		if err != nil {
			log.Fatal("E! " + err.Error())
		}
		os.Exit(0)
	} else {
		logger.SetupLogging(logger.LogConfig{LogTarget: logger.LogTargetEventlog})
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
