//go:build windows

//go:generate ../../scripts/windows-gen-syso.sh $GOARCH

package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/winservice"
)

const eventID = 100

func getLockedMemoryLimit() uint64 {
	handle := windows.CurrentProcess()

	var low, high uintptr
	var flag uint32
	windows.GetProcessWorkingSetSizeEx(handle, &low, &high, &flag)

	return uint64(high)
}

func (t *Telegraf) Run() error {
	// Process the service commands
	if t.service != "" {
		fmt.Println("The use of --service is deprecated, please use the 'service' command instead!")
		switch t.service {
		case "install":
			cfg := &winservice.Config{
				DisplayName:  t.serviceDisplayName,
				RestartDelay: t.serviceRestartDelay,
				AutoRestart:  t.serviceAutoRestart,
				Configs:      t.config,
				ConfigDirs:   t.configDir,
				WatchConfig:  t.watchConfig,
			}
			if err := winservice.Install(t.serviceName, cfg); err != nil {
				return err
			}
			fmt.Printf("Successfully installed service %q\n", t.serviceName)
		case "uninstall":
			if err := winservice.Uninstall(t.serviceName); err != nil {
				return err
			}
			fmt.Printf("Successfully uninstalled service %q\n", t.serviceName)
		case "start":
			if err := winservice.Start(t.serviceName); err != nil {
				return err
			}
			fmt.Printf("Successfully started service %q\n", t.serviceName)
		case "stop":
			if err := winservice.Stop(t.serviceName); err != nil {
				return err
			}
			fmt.Printf("Successfully stopped service %q\n", t.serviceName)
		case "status":
			status, err := winservice.Query(t.serviceName)
			if err != nil {
				return err
			}
			fmt.Printf("Service %q is in %q state\n", t.serviceName, status)
		default:
			return fmt.Errorf("invalid service command %q", t.service)
		}
		return nil
	}

	// Determine if Telegraf is started as a Windows service.
	isWinService, err := svc.IsWindowsService()
	if err != nil {
		return fmt.Errorf("cannot determine if run as Windows service: %w", err)
	}
	if !t.console && isWinService {
		return svc.Run(t.serviceName, t)
	}

	// Load the configuration file(s)
	cfg, err := t.loadConfiguration()
	if err != nil {
		return err
	}
	t.cfg = cfg

	stop = make(chan struct{})
	defer close(stop)
	return t.reloadLoop()
}

// Execute is the handler for the Windows service framework
func (t *Telegraf) Execute(_ []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	// Mark the status as startup pending until we are fully started
	const accepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	defer func() {
		changes <- svc.Status{State: svc.Stopped}
	}()

	// Create a eventlog logger for  all service related things
	svclog, err := eventlog.Open(t.serviceName)
	if err != nil {
		return true, 1
	}
	defer svclog.Close()

	//nolint:errcheck // We have no way to route the error to the user so ignore it
	svclog.Info(eventID, fmt.Sprintf("Starting Telegraf %s...", internal.Version))

	// Load the configuration file(s)
	cfg, err := t.loadConfiguration()
	if err != nil {
		//nolint:errcheck // We have no way to route the error to the user so ignore it
		svclog.Error(eventID, err.Error())
		return true, 2
	}
	t.cfg = cfg

	//nolint:errcheck // We have no way to route the error to the user so ignore it
	svclog.Info(eventID, "Finished loading configurations, starting data collection...")

	// Actually start the processing loop in the background to be able to
	// react to service change requests
	loopErr := make(chan error)
	stop = make(chan struct{})
	defer close(loopErr)
	defer close(stop)
	go func() {
		loopErr <- t.reloadLoop()
	}()
	changes <- svc.Status{State: svc.Running, Accepts: accepted}

	//nolint:errcheck // We have no way to route the error to the user so ignore it
	svclog.Info(eventID, "Telegraf is running...")

	for {
		select {
		case err := <-loopErr:
			if err != nil {
				log.Printf("E! %s", err)
				//nolint:errcheck // We have no way to route the error to the user so ignore it
				svclog.Error(eventID, err.Error())
				return true, 3
			}
			//nolint:errcheck // We have no way to route the error to the user so ignore it
			svclog.Info(eventID, "Stopped Telegraf")
			return false, windows.NO_ERROR
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				var empty struct{}
				stop <- empty // signal reloadLoop to finish (context cancel)

				//nolint:errcheck // We have no way to route the error to the user so ignore it
				svclog.Info(eventID, "Stopping Telegraf...")
			default:
				//nolint:errcheck // We have no way to route the error to the user so ignore it
				svclog.Error(eventID, fmt.Sprintf("Unexpected control request #%d", c))
			}
		}
	}
}
