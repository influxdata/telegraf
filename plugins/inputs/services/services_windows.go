package services

import (
	"log"

	"github.com/StackExchange/wmi"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Services is a telegraf plugin to gather services status from systemd and windows services
type Services struct {
	wmiQuery wmiQuery
}

type win32service struct {
	ExitCode  int
	Name      string
	ProcessID int
	StartMode string
	State     string
	Status    string
}

type wmiQuery func(query string, dst interface{}, connectServerArgs ...interface{}) error

const measurement = "services"

// Description returns a short description of the plugin
func (services *Services) Description() string {
	return "Gather service status for systemd units and windows services"
}

// SampleConfig returns sample configuration options.
func (services *Services) SampleConfig() string {
	return `
  ## no options
`
}

// Gather parses wmi outputs and adds counters to the Accumulator
func (services *Services) Gather(acc telegraf.Accumulator) error {
	var dst []win32service
	// win32_service is documented at https://docs.microsoft.com/en-us/windows/desktop/cimwin32prov/win32-service
	q := "select ExitCode, Name, ProcessId, StartMode, State, Status from Win32_Service where startmode = 'auto'"
	err := services.wmiQuery(q, &dst)
	if err != nil {
		return err
	}

	for _, service := range dst {
		tags := map[string]string{
			"name": service.Name,
		}
		// translate win32_service.status to nagios-style simple status
		var status int
		switch service.Status {
		case "Ok":
			status = 0 // ok
		case "Degraded", "Pred Fail":
			status = 1 // warning
		case "Error", "Starting", "Stopping", "Service", "Stressed", "NonRecover":
			status = 2 // error
		case "Unknown", "No Contact", "Lost Comm":
			fallthrough
		default:
			status = 3 // unknown
		}
		fields := map[string]interface{}{
			"state":  service.State,
			"status": status,
		}
		acc.AddCounter(measurement, fields, tags)
	}

	return nil
}

func init() {
	// Shim to avoid memory leak in reopening SWbemServices
	// https://github.com/StackExchange/wmi/issues/23
	// https://github.com/StackExchange/wmi/issues/27
	// https://github.com/martinlindhe/wmi_exporter/issues/77
	s, err := wmi.InitializeSWbemServices(wmi.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}
	wmi.DefaultClient.SWbemServicesClient = s

	inputs.Add("services", func() telegraf.Input {
		return &Services{
			wmiQuery: wmi.Query,
		}
	})
}
