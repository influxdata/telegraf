// +build windows

package win_services

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

//WinService provides interface for svc.Service
type WinService interface {
	Close() error
	Config() (mgr.Config, error)
	Query() (svc.Status, error)
}

//WinServiceManagerProvider sets interface for acquiring manager instance, like mgr.Mgr
type WinServiceManagerProvider interface {
	Connect() (WinServiceManager, error)
}

//WinServiceManager provides interface for mgr.Mgr
type WinServiceManager interface {
	Disconnect() error
	OpenService(name string) (WinService, error)
	ListServices() ([]string, error)
}

//WinSvcMgr is wrapper for mgr.Mgr implementing WinServiceManager interface
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

//MgProvider is an implementation of WinServiceManagerProvider interface returning WinSvcMgr
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
  service_names = [
    "LanmanServer",
    "TermService",
  ]
`

var description = "Input plugin to report Windows services info."

//WinServices is an implementation if telegraf.Input interface, providing info about Windows Services
type WinServices struct {
	ServiceNames []string `toml:"service_names"`
	mgrProvider  WinServiceManagerProvider
}

type ServiceInfo struct {
	ServiceName string
	DisplayName string
	State       int
	StartUpMode int
	Error       error
}

func (m *WinServices) Description() string {
	return description
}

func (m *WinServices) SampleConfig() string {
	return sampleConfig
}

func (m *WinServices) Gather(acc telegraf.Accumulator) error {

	serviceInfos, err := listServices(m.mgrProvider, m.ServiceNames)

	if err != nil {
		return err
	}

	for _, service := range serviceInfos {
		if service.Error == nil {
			fields := make(map[string]interface{})
			tags := make(map[string]string)

			//display name could be empty, but still valid service
			if len(service.DisplayName) > 0 {
				tags["display_name"] = service.DisplayName
			}
			tags["service_name"] = service.ServiceName

			fields["state"] = service.State
			fields["startup_mode"] = service.StartUpMode

			acc.AddFields("win_services", fields, tags)
		} else {
			acc.AddError(service.Error)
		}
	}

	return nil
}

//listServices gathers info about given services. If userServices is empty, it return info about all services on current Windows host.  Any a critical error is returned.
func listServices(mgrProv WinServiceManagerProvider, userServices []string) ([]ServiceInfo, error) {
	scmgr, err := mgrProv.Connect()
	if err != nil {
		return nil, fmt.Errorf("Could not open service manager: %s", err)
	}
	defer scmgr.Disconnect()

	var serviceNames []string
	if len(userServices) == 0 {
		//Listing service names from system
		serviceNames, err = scmgr.ListServices()
		if err != nil {
			return nil, fmt.Errorf("Could not list services: %s", err)
		}
	} else {
		serviceNames = userServices
	}
	serviceInfos := make([]ServiceInfo, len(serviceNames))

	for i, srvName := range serviceNames {
		serviceInfos[i] = collectServiceInfo(scmgr, srvName)
	}

	return serviceInfos, nil
}

//collectServiceInfo gathers info about a  service from WindowsAPI
func collectServiceInfo(scmgr WinServiceManager, serviceName string) (serviceInfo ServiceInfo) {

	serviceInfo.ServiceName = serviceName
	srv, err := scmgr.OpenService(serviceName)
	if err != nil {
		serviceInfo.Error = fmt.Errorf("Could not open service '%s': %s", serviceName, err)
		return
	}
	defer srv.Close()

	srvStatus, err := srv.Query()
	if err == nil {
		serviceInfo.State = int(srvStatus.State)
	} else {
		serviceInfo.Error = fmt.Errorf("Could not query service '%s': %s", serviceName, err)
		//finish collecting info on first found error
		return
	}

	srvCfg, err := srv.Config()
	if err == nil {
		serviceInfo.DisplayName = srvCfg.DisplayName
		serviceInfo.StartUpMode = int(srvCfg.StartType)
	} else {
		serviceInfo.Error = fmt.Errorf("Could not get config of service '%s': %s", serviceName, err)
	}
	return
}

func init() {
	inputs.Add("win_services", func() telegraf.Input { return &WinServices{mgrProvider: &MgProvider{}} })
}
