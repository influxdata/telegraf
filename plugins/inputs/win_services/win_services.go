// +build windows

package win_services

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/sys/windows/svc/mgr"
)

var sampleConfig = `
  ## Names of the services to monitor. Leave empty to monitor all the available services on the host
  service_names = [
    "LanmanServer",
    "TermService",
  ]
`

var description = "Input plugin to report Windows services info."

type WinServices struct {
	ServiceNames []string `toml:"service_names"`
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

	serviceInfos, err := listServices(m.ServiceNames)

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
func listServices(userServices []string) ([]ServiceInfo, error) {
	scmgr, err := mgr.Connect()
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
func collectServiceInfo(scmgr *mgr.Mgr, serviceName string) (serviceInfo ServiceInfo) {

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
	inputs.Add("win_services", func() telegraf.Input { return &WinServices{} })
}
