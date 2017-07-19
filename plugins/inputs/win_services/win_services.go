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

type Win_Services struct {
	ServiceNames []string `toml:"service_names"`
}

type ServiceInfo struct {
	ServiceName string
	DisplayName string
	State       int
	StartUpMode int
	Error       error
}

var ServiceStatesMap = map[int]string{
	0x00000001: "stopped",
	0x00000002: "start_pending",
	0x00000003: "stop_pending",
	0x00000004: "running",
	0x00000005: "continue_pending",
	0x00000006: "pause_pending",
	0x00000007: "paused",
}

var ServiceStartupModeMap = map[int]string{
	0x00000000: "boot_start",
	0x00000001: "system_start",
	0x00000002: "auto_start",
	0x00000003: "demand_start",
	0x00000004: "disabled",
}

func (m *Win_Services) Description() string {
	return description
}

func (m *Win_Services) SampleConfig() string {
	return sampleConfig
}

func (m *Win_Services) Gather(acc telegraf.Accumulator) error {

	serviceInfos, err := listServices(m.ServiceNames)

	if err != nil {
		return err
	}

	for _, service := range serviceInfos {
		if service.Error == nil {
			fields := make(map[string]interface{})
			tags := make(map[string]string)

			tags["display_name"] = service.DisplayName
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

func collectServiceInfo(scmgr *mgr.Mgr, serviceName string) (serviceInfo ServiceInfo) {

	serviceInfo.ServiceName = serviceName
	srv, err := scmgr.OpenService(serviceName)
	if err != nil {
		serviceInfo.Error = fmt.Errorf("Could not open service '%s': %s", serviceName, err)
		return
	}
	defer srv.Close()

	//While getting service info there could a theoretically a lot of errors on different places.
	//However in reality if there is a problem with a service then usually openService fails and if it passes, other calls will most probably be ok
	//So, following error checking is just for sake
	srvStatus, err := srv.Query()
	if err == nil {
		state := int(srvStatus.State)
		if !checkState(state) {
			serviceInfo.Error = fmt.Errorf("Uknown state of Service %s: %d", serviceName, state)
			//finish collecting info on first found error
			return
		}
		serviceInfo.State = state
	} else {
		serviceInfo.Error = fmt.Errorf("Could not query service '%s': %s", serviceName, err)
		//finish collecting info on first found error
		return
	}

	srvCfg, err := srv.Config()
	if err == nil {
		startupMode := int(srvCfg.StartType)
		if !checkStartupMode(startupMode) {
			serviceInfo.Error = fmt.Errorf("Uknown startup mode of Service %s: %d", serviceName, startupMode)
			//finish collecting info on first found error
			return
		}
		serviceInfo.DisplayName = srvCfg.DisplayName
		serviceInfo.StartUpMode = startupMode
	} else {
		serviceInfo.Error = fmt.Errorf("Could not get config of service '%s': %s", serviceName, err)
	}
	return
}

//returns true of state is in valid range
func checkState(state int) bool {
	_, ok := ServiceStatesMap[state]
	return ok
}

//returns true of startup mode is in valid range
func checkStartupMode(startupMode int) bool {
	_, ok := ServiceStartupModeMap[startupMode]
	return ok
}

func init() {
	inputs.Add("win_services", func() telegraf.Input { return &Win_Services{} })
}
