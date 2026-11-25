package winservice

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

type Config struct {
	DisplayName  string
	RestartDelay string
	AutoRestart  bool

	// Telegraf parameters
	Configs     []string
	ConfigDirs  []string
	WatchConfig string
}

func Install(name string, cfg *Config) error {
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
	args := make([]string, 0, 2*(len(cfg.Configs)+len(cfg.ConfigDirs))+2)
	for _, fn := range cfg.Configs {
		args = append(args, "--config", fn)
	}
	for _, dn := range cfg.ConfigDirs {
		args = append(args, "--config-directory", dn)
	}
	if len(args) == 0 {
		args = append(args, "--config", filepath.Join(programFiles, "Telegraf", "telegraf.conf"))
	}
	if cfg.WatchConfig != "" {
		args = append(args, "--watch-config", cfg.WatchConfig)
	}
	// Pass the service name to the command line, to have a custom name when relaunching as a service
	args = append(args, "--service-name", name)

	// Create a configuration for the service
	svccfg := mgr.Config{
		DisplayName: cfg.DisplayName,
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
	if cfg.AutoRestart {
		delay, err := time.ParseDuration(cfg.RestartDelay)
		if err != nil {
			return fmt.Errorf("cannot parse restart delay %q: %w", cfg.RestartDelay, err)
		}
		recovery := []mgr.RecoveryAction{{Type: mgr.ServiceRestart, Delay: delay}}
		if err := service.SetRecoveryActions(recovery, 10); err != nil {
			return err
		}
	}

	// Register the event as a source of eventlog events
	events := uint32(eventlog.Error | eventlog.Warning | eventlog.Info)
	if err := eventlog.InstallAsEventCreate(name, events); err != nil && !strings.Contains(err.Error(), "key already exists") {
		//nolint:errcheck // Try to remove the service on best effort basis as we cannot handle any error here
		service.Delete()
		return fmt.Errorf("setting up eventlog source failed: %w", err)
	}

	return nil
}

func Uninstall(name string) error {
	// Connect to the service manager and try to open the service. In case the
	// service is not installed, return with the corresponding error.
	svcmgr, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connecting to service manager failed: %w", err)
	}
	defer svcmgr.Disconnect()

	service, err := svcmgr.OpenService(name)
	if err != nil {
		if !errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return fmt.Errorf("opening service failed: %w", err)
		}
	} else {
		defer service.Close()

		// Uninstall the service
		if err := service.Delete(); err != nil {
			return fmt.Errorf("uninstalling service failed: %w", err)
		}
	}

	// Remove the eventlog source if there is any
	if err := eventlog.Remove(name); err != nil && !errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) {
		return fmt.Errorf("removing eventlog source failed: %w", err)
	}

	return nil
}

func Start(name string) error {
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

func Stop(name string) error {
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

func Query(name string) (string, error) {
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
