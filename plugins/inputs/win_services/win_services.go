// +build windows

package win_services

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/sys/windows/svc/mgr"
    "fmt"
)

var sampleConfig = `
  ## Name of services to monitor. Set empty to monitor all the available services on the host
  service_names = [
    "LanmanServer",
    "TermService",
  ]
`

var description = "Input plugin to report Windows services info: service name, display name, state, startup mode"

type Win_Services struct {
	ServiceNames []string `toml:"service_names"`
}

type ServiceInfo struct {
	ServiceName		string
	DisplayName		string
	State      		int
	StartUpMode     int
    Error           error
}

var  ServiceStatesMap = map [int]string{
    0x00000001: "service_stopped",
    0x00000002: "service_start_pending",
    0x00000003: "service_stop_pending",
    0x00000004: "service_running",
    0x00000005: "service_continue_pending",
    0x00000006: "service_pause_pending",
    0x00000007: "service_paused",
}

var  ServiceStartupModeMap = map [int]string{
    0x00000000: "service_boot_start",
    0x00000001: "service_system_start",
    0x00000002: "service_auto_start",
    0x00000003: "service_demand_start",
    0x00000004: "service_disabled",

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

            fields["display_name"] = service.DisplayName
            tags["service_name"] = service.ServiceName
            tags["state"] = ServiceStatesMap[service.State]
            tags["startup_mode"] = ServiceStartupModeMap[service.StartUpMode]

            acc.AddFields("win_services", fields, tags)
        } else {
            acc.AddError(err)
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
        serviceInfos[i].ServiceName = srvName
        srv, err := scmgr.OpenService(srvName)
        if err != nil {
            serviceInfos[i].Error = fmt.Errorf("Could not open service '%s': %s",srvName, err)
            continue
        }
        srvStatus, err := srv.Query()
        if err == nil {
            serviceInfos[i].State = int(srvStatus.State)
        } else {
            serviceInfos[i].Error = fmt.Errorf("Could not query service '%s': %s",srvName, err)
        }

        srvCfg, err := srv.Config()
        if err == nil {
            serviceInfos[i].DisplayName = srvCfg.DisplayName
            serviceInfos[i].StartUpMode = int(srvCfg.StartType)
        } else {
            serviceInfos[i].Error = fmt.Errorf("Could not get config of service '%s': %s", srvName, err)
        }
        srv.Close()
    }
    return serviceInfos, nil
}

func init() {
	inputs.Add("win_services", func() telegraf.Input { return &Win_Services{} })
}
