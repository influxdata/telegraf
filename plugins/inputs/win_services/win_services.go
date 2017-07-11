// +build windows

package win_services

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/sys/windows/svc/mgr"
    "strconv"
    "errors"
)

var sampleConfig = `
  ## This plugin returns by default service state and startup mode
  ## See the README file for more examples.
  ## Uncomment examples below or write your own as you see fit. If the system
  ## being polled for data does not have the Object at startup of the Telegraf
  ## agent, it will not be gathered.
  ## Settings:

  # Names of services to monitor. Empty for all
  Services = [
    "Server"
  ]
  # CustomTagName=Group
  # CustomTagValue=alpha
`

var description = "Input plugin to report Windows services info: name, display name, state, startup mode"

type Win_Services struct {
	Services    []string
	CustomTagName	string
	CustomTagValue  string
}

type ServiceInfo struct {
	ServiceName		string
	DisplayName		string
	State      		int
	StartUpMode     int
    Error           error
}


func (m *Win_Services) Description() string {
	return description
}

func (m *Win_Services) SampleConfig() string {
	return sampleConfig
}


func (m *Win_Services) Gather(acc telegraf.Accumulator) error {

    serviceInfos, err := listServices(m.Services)

    if err != nil {
        return err
    }

    for _, service := range serviceInfos {
        fields := make(map[string]interface{})
        tags := make(map[string]string)
        if service.Error == nil {
            fields["displayname"] = service.DisplayName
            tags["state"] = strconv.Itoa(service.State)
            tags["startupMode"] = strconv.Itoa(service.StartUpMode)
        } else {
            fields["service"] = service.ServiceName
            tags["error"] = service.Error.Error()

        }
        acc.AddFields(service.ServiceName, fields, tags)
    }

	return nil
}

func listServices(userServices []string) ([]ServiceInfo, error) {

    scmgr, err := mgr.Connect()
    if err != nil {
        return nil, errors.New("Could not open service manager: " + err.Error());
    }
    defer scmgr.Disconnect()

    var serviceNames []string
    if len(userServices) == 0 {
        //Listing service names from system
        serviceNames, err = scmgr.ListServices()
        if err != nil {
            return nil, errors.New("Could not list services: " + err.Error());
        }
    } else {
        serviceNames = userServices
    }
    serviceInfos := make([]ServiceInfo, len(serviceNames))

    for i, srvName := range serviceNames {
        serviceInfos[i].ServiceName = srvName
        srv, err := scmgr.OpenService(srvName)
        if err != nil {
            serviceInfos[i].Error = errors.New("Could not open service: " + err.Error());
            continue
        }
        srvStatus, err := srv.Query()
        if err == nil {
            serviceInfos[i].State = int(srvStatus.State)
        } else {
            serviceInfos[i].Error = errors.New("Could not query service: " + err.Error());
        }

        srvCfg, err := srv.Config()
        if err == nil {
            serviceInfos[i].DisplayName = srvCfg.DisplayName
            serviceInfos[i].StartUpMode = int(srvCfg.StartType)
        } else {
            serviceInfos[i].Error = errors.New("Could not get service config: " + err.Error());
        }
        srv.Close()
    }
    return serviceInfos, nil
}

func init() {
	inputs.Add("win_services", func() telegraf.Input { return &Win_Services{} })
}
