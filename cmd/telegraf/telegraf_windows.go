//go:build windows

//go:generate ../../scripts/windows-gen-syso.sh $GOARCH

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/influxdata/telegraf/logger"
)

func getLockedMemoryLimit() uint64 {
	handle := windows.CurrentProcess()

	var min, max uintptr
	var flag uint32
	windows.GetProcessWorkingSetSizeEx(handle, &min, &max, &flag)

	return uint64(max)
}

func (t *Telegraf) Run() error {
	// Register the eventlog logging target for windows.
	if err := logger.RegisterEventLogger(t.serviceName); err != nil {
		return err
	}

	// Process the service commands
	if t.service != "" {
		fmt.Println("The use of --service is deprecated, please use the 'service' command instead!")
		switch t.service {
		case "install":
			cfg := &serviceConfig{
				displayName:  t.serviceDisplayName,
				restartDelay: t.serviceRestartDelay,
				autoRestart:  t.serviceAutoRestart,
				configs:      t.config,
				configDirs:   t.configDir,
				watchConfig:  t.watchConfig,
			}
			if err := installService(t.serviceName, cfg); err != nil {
				return err
			}
			fmt.Printf("Successfully installed service %q\n", t.serviceName)
		case "uninstall":
			if err := uninstallService(t.serviceName); err != nil {
				return err
			}
			fmt.Printf("Successfully uninstalled service %q\n", t.serviceName)
		case "start":
			if err := startService(t.serviceName); err != nil {
				return err
			}
			fmt.Printf("Successfully started service %q\n", t.serviceName)
		case "stop":
			if err := stopService(t.serviceName); err != nil {
				return err
			}
			fmt.Printf("Successfully stopped service %q\n", t.serviceName)
		case "status":
			status, err := queryService(t.serviceName)
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

// Handler for the Windows service framework
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
		log.Printf("E! Initializing the service logger failed: %s", err)
		return true, 1
	}
	defer svclog.Close()

	// Load the configuration file(s)
	cfg, err := t.loadConfiguration()
	if err != nil {
		if lerr := svclog.Error(100, err.Error()); lerr != nil {
			log.Printf("E! Logging error %q failed: %s", err, lerr)
		}
		return true, 2
	}
	t.cfg = cfg

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

	for {
		select {
		case err := <-loopErr:
			if err != nil {
				if lerr := svclog.Error(100, err.Error()); lerr != nil {
					log.Printf("E! Logging error %q failed: %s", err, lerr)
				}
				return true, 3
			}
			return false, 0
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
			default:
				msg := fmt.Sprintf("Unexpected control request #%d", c)
				if lerr := svclog.Error(100, msg); lerr != nil {
					log.Printf("E! Logging error %q failed: %s", msg, lerr)
				}
			}
		}
	}
}

type serviceConfig struct {
	displayName  string
	restartDelay string
	autoRestart  bool

	// Telegraf parameters
	configs     []string
	configDirs  []string
	watchConfig string
}

func installService(name string, cfg *serviceConfig) error {
	// Determine the executable to use in the service
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("determining executable failed: %w", err)
	}

	// Determine the program files directory name
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" { // Should never happen
		programFiles = "C:\\Program Files"
	}

	// Collect the command line arguments
	args := make([]string, 0, 2*(len(cfg.configs)+len(cfg.configDirs))+2)
	for _, fn := range cfg.configs {
		args = append(args, "--config", fn)
	}
	for _, dn := range cfg.configDirs {
		args = append(args, "--config-directory", dn)
	}
	if len(args) == 0 {
		args = append(args, "--config", filepath.Join(programFiles, "Telegraf", "telegraf.conf"))
	}
	if cfg.watchConfig != "" {
		args = append(args, "--watch-config", cfg.watchConfig)
	}
	// Pass the service name to the command line, to have a custom name when relaunching as a service
	args = append(args, "--service-name", name)

	// Create a configuration for the service
	svccfg := mgr.Config{
		DisplayName: cfg.displayName,
		Description: "Collects, processes and publishes data using a series of plugins.",
		StartType:   mgr.StartAutomatic,
		ServiceType: windows.SERVICE_WIN32_OWN_PROCESS,
	}

	// Connect to the service manager and try to install the service if it
	// doesn't exist. Fail on existing service and stop installation.
	svcmgr, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connecting to service manager failed: %w", err)
	}
	defer svcmgr.Disconnect()

	if service, err := svcmgr.OpenService(name); err == nil {
		service.Close()
		return fmt.Errorf("service %q is already installed", name)
	}

	service, err := svcmgr.CreateService(name, executable, svccfg, args...)
	if err != nil {
		return fmt.Errorf("creating service failed: %w", err)
	}
	defer service.Close()

	// Set the recovery strategy to restart with a fixed period of 10 seconds
	// and the user specified delay if requested
	if cfg.autoRestart {
		delay, err := time.ParseDuration(cfg.restartDelay)
		if err != nil {
			return fmt.Errorf("cannot parse restart delay %q: %w", cfg.restartDelay, err)
		}
		recovery := []mgr.RecoveryAction{{Type: mgr.ServiceRestart, Delay: delay}}
		if err := service.SetRecoveryActions(recovery, 10); err != nil {
			return err
		}
	}

	// Register the event as a source of eventlog events
	events := uint32(eventlog.Error | eventlog.Warning | eventlog.Info)
	if err := eventlog.InstallAsEventCreate(name, events); err != nil {
		//nolint:errcheck // Try to remove the service on best effort basis as we cannot handle any error here
		service.Delete()
		return fmt.Errorf("setting up eventlog source failed: %w", err)
	}

	return nil
}

func uninstallService(name string) error {
	// Connect to the service manager and try to open the service. In case the
	// service is not installed, return with the corresponding error.
	svcmgr, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connecting to service manager failed: %w", err)
	}
	defer svcmgr.Disconnect()

	service, err := svcmgr.OpenService(name)
	if err != nil {
		return fmt.Errorf("opening service failed: %w", err)
	}
	defer service.Close()

	// Uninstall the service and remove the eventlog source
	if err := service.Delete(); err != nil {
		return fmt.Errorf("uninstalling service failed: %w", err)
	}

	if err := eventlog.Remove(name); err != nil {
		return fmt.Errorf("removing eventlog source failed: %w", err)
	}

	return nil
}

func startService(name string) error {
	nameUTF16, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return fmt.Errorf("conversion of service name %q to UTF16 failed: %w", name, err)
	}

	// Open the service manager and service with the least privileges required to start the service
	mgrhandle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT|windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return fmt.Errorf("opening service manager failed: %w", err)
	}
	defer windows.CloseServiceHandle(mgrhandle)

	svchandle, err := windows.OpenService(mgrhandle, nameUTF16, windows.SERVICE_QUERY_STATUS|windows.SERVICE_START)
	if err != nil {
		return fmt.Errorf("opening service failed: %w", err)
	}
	service := &mgr.Service{Handle: svchandle, Name: name}
	defer service.Close()

	// Check if the service is actually stopped
	status, err := service.Query()
	if err != nil {
		return fmt.Errorf("querying service state failed: %w", err)
	}
	if status.State != svc.Stopped {
		return fmt.Errorf("service is not stopped but in state %q", stateDescription(status.State))
	}

	return service.Start()
}

func stopService(name string) error {
	nameUTF16, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return fmt.Errorf("conversion of service name %q to UTF16 failed: %w", name, err)
	}

	// Open the service manager and service with the least privileges required to start the service
	mgrhandle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT|windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return fmt.Errorf("opening service manager failed: %w", err)
	}
	defer windows.CloseServiceHandle(mgrhandle)

	svchandle, err := windows.OpenService(mgrhandle, nameUTF16, windows.SERVICE_QUERY_STATUS|windows.SERVICE_STOP)
	if err != nil {
		return fmt.Errorf("opening service failed: %w", err)
	}
	service := &mgr.Service{Handle: svchandle, Name: name}
	defer service.Close()

	// Stop the service and wait for it to finish
	status, err := service.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("stopping service failed: %w", err)
	}
	for status.State != svc.Stopped {
		// Wait for the hinted time, but clip it to prevent stalling operation
		wait := time.Duration(status.WaitHint) * time.Millisecond
		if wait < 100*time.Millisecond {
			wait = 100 * time.Millisecond
		} else if wait > 10*time.Second {
			wait = 10 * time.Second
		}
		time.Sleep(wait)

		status, err = service.Query()
		if err != nil {
			return fmt.Errorf("querying service state failed: %w", err)
		}
	}

	return nil
}

func queryService(name string) (string, error) {
	nameUTF16, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return "", fmt.Errorf("conversion of service name %q to UTF16 failed: %w", name, err)
	}

	// Open the service manager and service with the least privileges required to start the service
	mgrhandle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT|windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return "", fmt.Errorf("opening service manager failed: %w", err)
	}
	defer windows.CloseServiceHandle(mgrhandle)

	svchandle, err := windows.OpenService(mgrhandle, nameUTF16, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return "", fmt.Errorf("opening service failed: %w", err)
	}
	service := &mgr.Service{Handle: svchandle, Name: name}
	defer service.Close()

	// Query the service state and report it to the user
	status, err := service.Query()
	if err != nil {
		return "", fmt.Errorf("querying service state failed: %w", err)
	}

	return stateDescription(status.State), nil
}

func stateDescription(state svc.State) string {
	switch state {
	case svc.Stopped:
		return "stopped"
	case svc.StartPending:
		return "start pending"
	case svc.StopPending:
		return "stop pending"
	case svc.Running:
		return "running"
	case svc.ContinuePending:
		return "continue pending"
	case svc.PausePending:
		return "pause pending"
	case svc.Paused:
		return "paused"
	}
	return fmt.Sprintf("unknown %v", state)
}
