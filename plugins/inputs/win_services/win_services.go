// +build windows

package win_services

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type ServiceErr struct {
	Message string
	Service string
	Err     error
}

func (e *ServiceErr) Error() string {
	return fmt.Sprintf("%s: '%s': %v", e.Message, e.Service, e.Err)
}

func IsPermission(err error) bool {
	if err, ok := err.(*ServiceErr); ok {
		return os.IsPermission(err.Err)
	}
	return false
}

// WinService provides interface for svc.Service
type WinService interface {
	Close() error
	Config() (mgr.Config, error)
	Query() (svc.Status, error)
}

// ManagerProvider sets interface for acquiring manager instance, like mgr.Mgr
type ManagerProvider interface {
	Connect() (WinServiceManager, error)
}

// WinServiceManager provides interface for mgr.Mgr
type WinServiceManager interface {
	Disconnect() error
	OpenService(name string) (WinService, error)
	ListServices() ([]string, error)
}

// WinSvcMgr is wrapper for mgr.Mgr implementing WinServiceManager interface
type WinSvcMgr struct {
	realMgr *mgr.Mgr
}

func (m *WinSvcMgr) Disconnect() error {
	return m.realMgr.Disconnect()
}

func (m *WinSvcMgr) OpenService(name string) (WinService, error) {
	return m.realMgr.OpenService(name)
}
func (m *WinSvcMgr) ListServices() ([]string, error) {
	return m.realMgr.ListServices()
}

// MgProvider is an implementation of WinServiceManagerProvider interface returning WinSvcMgr
type MgProvider struct {
}

func (rmr *MgProvider) Connect() (WinServiceManager, error) {
	scmgr, err := mgr.Connect()
	if err != nil {
		return nil, err
	} else {
		return &WinSvcMgr{scmgr}, nil
	}
}

var sampleConfig = `
  ## Names of the services to monitor. Leave empty to monitor all the available services on the host
  ## Wildcards (*) are supported 
  service_names = [
    "LanmanServer",
    "TermService",
  ]

  ## set enable_regex to true to interpret service_names as regular expressions
  # enable_regex = false
`

var description = "Input plugin to report Windows services info."

//WinServices is an implementation if telegraf.Input interface, providing info about Windows Services
type WinServices struct {
	Log telegraf.Logger

	ServiceNames []string `toml:"service_names"`
	EnableRegex  bool     `toml:"enable_regex"`

	mgrProvider  ManagerProvider
	regexedNames []*regexp.Regexp
}

type ServiceInfo struct {
	ServiceName string
	DisplayName string
	State       int
	StartUpMode int
}

func (m *WinServices) Description() string {
	return description
}

func (m *WinServices) SampleConfig() string {
	return sampleConfig
}

// Init the plugin by parsing the service names
func (m *WinServices) Init() error {

	for _, service := range m.ServiceNames {
		if m.EnableRegex {
			re, err := regexp.Compile(service)

			if err != nil {
				return err
			}

			m.regexedNames = append(m.regexedNames, re)
		} else {
			service = strings.Replace(service, "*", ".*", -1)
			re, err := regexp.Compile("^" + service + "$") // exact match

			if err != nil {
				return err
			}

			m.regexedNames = append(m.regexedNames, re)
		}
	}

	return nil
}

func (m *WinServices) Gather(acc telegraf.Accumulator) error {
	scmgr, err := m.mgrProvider.Connect()
	if err != nil {
		return fmt.Errorf("Could not open service manager: %s", err)
	}
	defer scmgr.Disconnect()

	serviceNames, err := listServices(scmgr, m.regexedNames)
	if err != nil {
		return err
	}

	for _, srvName := range serviceNames {
		service, err := collectServiceInfo(scmgr, srvName)
		if err != nil {
			if IsPermission(err) {
				m.Log.Debug(err.Error())
			} else {
				m.Log.Error(err.Error())
			}
			continue
		}

		tags := map[string]string{
			"service_name": service.ServiceName,
		}
		//display name could be empty, but still valid service
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
func listServices(scmgr WinServiceManager, regexedNames []*regexp.Regexp) ([]string, error) {

	names, err := scmgr.ListServices()
	if err != nil {
		return nil, fmt.Errorf("Could not list services: %s", err)
	}

	if len(regexedNames) == 0 { // nothing selected, return everything
		return names, err
	}

	var serviceNames []string

	for _, re := range regexedNames {
		for _, name := range names {
			if re.MatchString(name) {
				serviceNames = append(serviceNames, name)
			}
		}
	}

	return serviceNames, nil
}

// collectServiceInfo gathers info about a service.
func collectServiceInfo(scmgr WinServiceManager, serviceName string) (*ServiceInfo, error) {
	srv, err := scmgr.OpenService(serviceName)
	if err != nil {
		return nil, &ServiceErr{
			Message: "could not open service",
			Service: serviceName,
			Err:     err,
		}
	}
	defer srv.Close()

	srvStatus, err := srv.Query()
	if err != nil {
		return nil, &ServiceErr{
			Message: "could not query service",
			Service: serviceName,
			Err:     err,
		}
	}

	srvCfg, err := srv.Config()
	if err != nil {
		return nil, &ServiceErr{
			Message: "could not get config of service",
			Service: serviceName,
			Err:     err,
		}
	}

	serviceInfo := &ServiceInfo{
		ServiceName: serviceName,
		DisplayName: srvCfg.DisplayName,
		StartUpMode: int(srvCfg.StartType),
		State:       int(srvStatus.State),
	}
	return serviceInfo, nil
}

func init() {
	inputs.Add("win_services", func() telegraf.Input {
		return &WinServices{
			mgrProvider: &MgProvider{},
		}
	})
}
