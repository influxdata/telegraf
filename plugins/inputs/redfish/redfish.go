package redfish

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Cpu struct {
	Name                   string
	ReadingCelsius         int
	UpperThresholdCritical int
	UpperThresholdFatal    int
	Status                 CpuStatus
}
type Payload struct {
	Temperatures  []Cpu    `json:",omitempty"`
	Fans          []Speed  `json:",omitempty"`
	PowerSupplies []Watt   `json:",omitempty"`
	Hostname      string   `json:",omitempty"`
	Voltages      []Volt   `json:",omitempty"`
	Location      *Address `json:",omitempty"`
}
type CpuStatus struct {
	State  string
	Health string
}
type Speed struct {
	Name                   string
	Reading                int
	ReadingUnits           string
	UpperThresholdCritical int
	UpperThresholdFatal    int
	Status                 FansStatus
}
type FansStatus struct {
	State  string
	Health string
}
type Watt struct {
	Name                 string
	PowerInputWatts      float64
	PowerCapacityWatts   float64
	PowerOutputWatts     float64
	LastPowerOutputWatts float64
	Status               PowerStatus
	LineInputVoltage     float64
}
type PowerStatus struct {
	State  string
	Health string
}
type Volt struct {
	Name                   string
	ReadingVolts           float64
	UpperThresholdCritical float64
	UpperThresholdFatal    float64
	Status                 VoltStatus
}
type VoltStatus struct {
	State  string
	Health string
}
type Address struct {
	PostalAddress PostalAddress
	Placement     Placement
}
type PostalAddress struct {
	DataCenter string
	Room       string
}
type Placement struct {
	Rack string
	Row  string
}
type Redfish struct {
	Address           string `toml:"address"`
	BasicAuthUsername string `toml:"username"`
	BasicAuthPassword string `toml:"password"`
	ComputerSystemId  string `toml:"computer_system_id"`
	client            http.Client
	tls.ClientConfig
	Timeout internal.Duration `toml:"timeout"`
	payload Payload
}

func (r *Redfish) Description() string {
	return "Read CPU, Fans, Powersupply and Voltage metrics of hardware server through redfish APIs"
}

var redfishConfig = `
  ## Redfish API Base URL.
  address = "https://127.0.0.1:5000"

  ## Credentials for the Redfish API.
  username = "root"
  password = "password123456"

  ## System Id to collect data for in Redfish APIs.
  computer_system_id="System.Embedded.1"

  ## Amount of time allowed to complete the HTTP request
  # timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (r *Redfish) SampleConfig() string {
	return redfishConfig
}

func (r *Redfish) Init() error {
	if len(r.Address) == 0 || len(r.BasicAuthUsername) == 0 || len(r.BasicAuthPassword) == 0 {
		return fmt.Errorf("did not provide IP or username and password")
	}
	if len(r.ComputerSystemId) == 0 {
		return fmt.Errorf("did not provide the computer system ID of the resource")
	}
	tlsCfg, err := r.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	r.client = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
		Timeout: r.Timeout.Duration,
	}

	return nil
}

func (r *Redfish) GetData(url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(r.BasicAuthUsername, r.BasicAuthPassword)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(body, &r.payload)
		if err != nil {
			return fmt.Errorf("error parsing input: %v", err)
		}

	} else {
		return fmt.Errorf("received status code %d (%s), expected 200",
			resp.StatusCode,
			http.StatusText(resp.StatusCode))
	}
	return nil
}

func (r *Redfish) Gather(acc telegraf.Accumulator) error {
	var url []string
	url = append(url, fmt.Sprint(r.Address, "/redfish/v1/Chassis/", r.ComputerSystemId, "/Thermal"), fmt.Sprint(r.Address, "/redfish/v1/Chassis/", r.ComputerSystemId, "/Power"), fmt.Sprint(r.Address, "/redfish/v1/Systems/", r.ComputerSystemId), fmt.Sprint(r.Address, "/redfish/v1/Chassis/", r.ComputerSystemId, "/"))
	for _, i := range url {
		err := r.GetData(i)
		if err != nil {
			return err
		}
	}
	if r.payload.Location != nil {
		for _, j := range r.payload.Temperatures {
			//  Tags
			tags := map[string]string{}
			tags["address"] = r.Address
			tags["name"] = j.Name
			tags["source"] = r.payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			tags["datacenter"] = r.payload.Location.PostalAddress.DataCenter
			tags["room"] = r.payload.Location.PostalAddress.Room
			tags["rack"] = r.payload.Location.Placement.Rack
			tags["row"] = r.payload.Location.Placement.Row
			//  Fields
			fields := make(map[string]interface{})
			fields["reading_celsius"] = j.ReadingCelsius
			fields["upper_threshold_critical"] = j.UpperThresholdCritical
			fields["upper_threshold_fatal"] = j.UpperThresholdFatal
			acc.AddFields("redfish_thermal_temperatures", fields, tags)
		}
		for _, j := range r.payload.Fans {
			//  Tags
			tags := map[string]string{}
			fields := make(map[string]interface{})
			tags["address"] = r.Address
			tags["name"] = j.Name
			tags["source"] = r.payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			tags["datacenter"] = r.payload.Location.PostalAddress.DataCenter
			tags["room"] = r.payload.Location.PostalAddress.Room
			tags["rack"] = r.payload.Location.Placement.Rack
			tags["row"] = r.payload.Location.Placement.Row

			//  Fields
			if j.ReadingUnits == "RPM" {
				fields["upper_threshold_critical"] = j.UpperThresholdCritical
				fields["upper_threshold_fatal"] = j.UpperThresholdFatal
				fields["reading_rpm"] = j.Reading

			} else {
				fields["reading_percent"] = j.Reading
			}
			acc.AddFields("redfish_thermal_fans", fields, tags)
		}
		for _, j := range r.payload.PowerSupplies {
			//  Tags
			tags := map[string]string{}
			tags["address"] = r.Address
			tags["name"] = j.Name
			tags["source"] = r.payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			tags["datacenter"] = r.payload.Location.PostalAddress.DataCenter
			tags["room"] = r.payload.Location.PostalAddress.Room
			tags["rack"] = r.payload.Location.Placement.Rack
			tags["row"] = r.payload.Location.Placement.Row
			//  Fields
			fields := make(map[string]interface{})
			fields["power_input_watts"] = j.PowerInputWatts
			fields["power_output_watts"] = j.PowerOutputWatts
			fields["line_input_voltage"] = j.LineInputVoltage
			fields["last_power_output_watts"] = j.LastPowerOutputWatts
			fields["power_capacity_watts"] = j.PowerCapacityWatts
			acc.AddFields("redfish_power_powersupplies", fields, tags)
		}
		for _, j := range r.payload.Voltages {
			//  Tags
			tags := map[string]string{}
			tags["address"] = r.Address
			tags["name"] = j.Name
			tags["source"] = r.payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			tags["datacenter"] = r.payload.Location.PostalAddress.DataCenter
			tags["room"] = r.payload.Location.PostalAddress.Room
			tags["rack"] = r.payload.Location.Placement.Rack
			tags["row"] = r.payload.Location.Placement.Row
			//  Fields
			fields := make(map[string]interface{})
			fields["reading_volts"] = j.ReadingVolts
			fields["upper_threshold_critical"] = j.UpperThresholdCritical
			fields["upper_threshold_fatal"] = j.UpperThresholdFatal
			acc.AddFields("redfish_power_voltages", fields, tags)
		}
	} else {
		for _, j := range r.payload.Temperatures {
			//  Tags
			tags := map[string]string{}
			tags["address"] = r.Address
			tags["name"] = j.Name
			tags["source"] = r.payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			//  Fields
			fields := make(map[string]interface{})
			fields["reading_celsius"] = j.ReadingCelsius
			fields["upper_threshold_critical"] = j.UpperThresholdCritical
			fields["upper_threshold_fatal"] = j.UpperThresholdFatal
			acc.AddFields("redfish_thermal_temperatures", fields, tags)
		}
		for _, j := range r.payload.Fans {
			//  Tags
			tags := map[string]string{}
			fields := make(map[string]interface{})
			tags["address"] = r.Address
			tags["name"] = j.Name
			tags["source"] = r.payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			//  Fields
			if j.ReadingUnits == "RPM" {
				fields["upper_threshold_critical"] = j.UpperThresholdCritical
				fields["upper_threshold_fatal"] = j.UpperThresholdFatal
				fields["reading_rpm"] = j.Reading
			} else {
				fields["reading_percent"] = j.Reading
			}
			acc.AddFields("redfish_thermal_fans", fields, tags)
		}
		for _, j := range r.payload.PowerSupplies {
			//  Tags
			tags := map[string]string{}
			tags["address"] = r.Address
			tags["name"] = j.Name //j.Name
			tags["source"] = r.payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			//  Fields
			fields := make(map[string]interface{})
			fields["line_input_voltage"] = j.LineInputVoltage
			fields["last_power_output_watts"] = j.LastPowerOutputWatts
			fields["power_capacity_watts"] = j.PowerCapacityWatts
			acc.AddFields("redfish_power_powersupplies", fields, tags)
		}
		for _, j := range r.payload.Voltages {
			//  Tags
			tags := map[string]string{}
			tags["address"] = r.Address
			tags["name"] = j.Name
			tags["source"] = r.payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			//  Fields
			fields := make(map[string]interface{})
			fields["reading_volts"] = j.ReadingVolts
			fields["upper_threshold_critical"] = j.UpperThresholdCritical
			fields["upper_threshold_fatal"] = j.UpperThresholdFatal
			acc.AddFields("redfish_power_voltages", fields, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("redfish", func() telegraf.Input { return &Redfish{} })
}
