//go:generate ../../../tools/readme_config_includer/generator
//go:build windows

package win_services

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type WinServices struct {
	ServiceNames         []string `toml:"service_names"`
	ServiceNamesExcluded []string `toml:"excluded_service_names"`

	Log telegraf.Logger `toml:"-"`

	mgrProvider    managerProvider
	servicesFilter filter.Filter
}

// winService provides interface for svc.Service
type winService interface {
	Close() error
	Config() (mgr.Config, error)
	Query() (svc.Status, error)
}

// managerProvider sets interface for acquiring manager instance, like mgr.Mgr
type managerProvider interface {
	connect() (winServiceManager, error)
}

// winServiceManager provides interface for mgr.Mgr
type winServiceManager interface {
	disconnect() error
	openService(name string) (winService, error)
	listServices() ([]string, error)
}

type serviceInfo struct {
	ServiceName string
	DisplayName string
	State       int
	StartUpMode int
}

func (*WinServices) SampleConfig() string {
	return sampleConfig
}

func (m *WinServices) Init() error {
	// For case insensitive comparison (see issue #8796) we need to transform the services
	// to lowercase
	servicesInclude := make([]string, 0, len(m.ServiceNames))
	for _, s := range m.ServiceNames {
		servicesInclude = append(servicesInclude, strings.ToLower(s))
	}
	servicesExclude := make([]string, 0, len(m.ServiceNamesExcluded))
	for _, s := range m.ServiceNamesExcluded {
		servicesExclude = append(servicesExclude, strings.ToLower(s))
	}

	f, err := filter.NewIncludeExcludeFilter(servicesInclude, servicesExclude)
	if err != nil {
		return err
	}
	m.servicesFilter = f

	return nil
}

func (m *WinServices) Gather(acc telegraf.Accumulator) error {
	scmgr, err := m.mgrProvider.connect()
	if err != nil {
		return fmt.Errorf("could not open service manager: %w", err)
	}
	defer scmgr.disconnect()

	serviceNames, err := m.listServices(scmgr)
	if err != nil {
		return err
	}

	for _, srvName := range serviceNames {
		service, err := collectServiceInfo(scmgr, srvName)
		if err != nil {
			if isPermission(err) {
				m.Log.Debug(err.Error())
			} else {
				m.Log.Error(err.Error())
			}
			continue
		}

		tags := map[string]string{
			"service_name": service.ServiceName,
		}
		// display name could be empty, but still valid service
		if len(service.DisplayName) > 0 {
			tags["display_name"] = service.DisplayName
		}

		fields := map[string]interface{}{
			"state":        service.State,
			"startup_mode": service.StartUpMode,
		}
		acc.AddFields("win_services", fields, tags)
	}

	return nil
}

// listServices returns a list of services to gather.
func (m *WinServices) listServices(scmgr winServiceManager) ([]string, error) {
	names, err := scmgr.listServices()
	if err != nil {
		return nil, fmt.Errorf("could not list services: %w", err)
	}

	var services []string
	for _, name := range names {
		// Compare case-insensitive. Use lowercase as we already converted the filter to use it.
		n := strings.ToLower(name)
		if m.servicesFilter.Match(n) {
			services = append(services, name)
		}
	}

	return services, nil
}

func isPermission(err error) bool {
	var serviceErr *serviceError
	if errors.As(err, &serviceErr) {
		return errors.Is(serviceErr, fs.ErrPermission)
	}
	return false
}

// collectServiceInfo gathers info about a service.
func collectServiceInfo(scmgr winServiceManager, serviceName string) (*serviceInfo, error) {
	srv, err := scmgr.openService(serviceName)
	if err != nil {
		return nil, &serviceError{
			message: "could not open service",
			service: serviceName,
			err:     err,
		}
	}
	defer srv.Close()

	srvStatus, err := srv.Query()
	if err != nil {
		return nil, &serviceError{
			message: "could not query service",
			service: serviceName,
			err:     err,
		}
	}

	srvCfg, err := srv.Config()
	if err != nil {
		return nil, &serviceError{
			message: "could not get config of service",
			service: serviceName,
			err:     err,
		}
	}

	serviceInfo := &serviceInfo{
		ServiceName: serviceName,
		DisplayName: srvCfg.DisplayName,
		StartUpMode: int(srvCfg.StartType),
		State:       int(srvStatus.State),
	}
	return serviceInfo, nil
}

type serviceError struct {
	message string
	service string
	err     error
}

func (e *serviceError) Error() string {
	return fmt.Sprintf("%s: %q: %v", e.message, e.service, e.err)
}

// winSvcMgr is wrapper for mgr.Mgr implementing winServiceManager interface
type winSvcMgr struct {
	realMgr *mgr.Mgr
}

func (m *winSvcMgr) disconnect() error {
	return m.realMgr.Disconnect()
}

func (m *winSvcMgr) openService(name string) (winService, error) {
	serviceName, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, fmt.Errorf("cannot convert service name %q: %w", name, err)
	}
	h, err := windows.OpenService(m.realMgr.Handle, serviceName, windows.GENERIC_READ)
	if err != nil {
		return nil, err
	}
	return &mgr.Service{Name: name, Handle: h}, nil
}

func (m *winSvcMgr) listServices() ([]string, error) {
	return m.realMgr.ListServices()
}

// mgProvider is an implementation of WinServiceManagerProvider interface returning winSvcMgr
type mgProvider struct {
}

func (*mgProvider) connect() (winServiceManager, error) {
	h, err := windows.OpenSCManager(nil, nil, windows.GENERIC_READ)
	if err != nil {
		return nil, err
	}
	scmgr := &mgr.Mgr{Handle: h}
	return &winSvcMgr{scmgr}, nil
}

func init() {
	inputs.Add("win_services", func() telegraf.Input {
		return &WinServices{
			mgrProvider: &mgProvider{},
		}
	})
}
