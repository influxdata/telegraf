package redfish

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"net/http"
	//"strconv"
)

type Cpu struct {
	Name           string    `json:"Name"`
	ReadingCelsius int64     `json:"ReadingCelsius"`
	Status         CpuStatus `json:"Status"`
}
type Payload struct {
	Temperatures  []*Cpu   `json:",omitempty"`
	Fans          []*speed `json:",omitempty"`
	PowerSupplies []*Watt  `json:",omitempty"`
	Hostname      string   `json:",omitempty"`
	Voltages      []*volt  `json:",omitempty"`
	Location      *Address `json:",omitempty"`
}
type CpuStatus struct {
	State  string
	Health string
}
type speed struct {
	Name    string
	Reading int64
	Status  FansStatus
}
type FansStatus struct {
	State  string `json:"State"`
	Health string `json:"Health"`
}
type Watt struct {
	Name                 string       `json:",omitempty"`
	PowerInputWatts      float64      `json:",omitempty"`
	PowerCapacityWatts   float64      `json:",omitempty"`
	PowerOutputWatts     float64      `json:",omitempty"`
	LastPowerOutputWatts float64      `json:",omitempty"`
	Status               *PowerStatus `json:",omitempty"`
	LineInputVoltage     float64      `json:",omitempty"`
}
type PowerStatus struct {
	State  string
	Health string
}
type volt struct {
	Name         string
	ReadingVolts int64
	Status       VoltStatus
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
	Host              string `toml:"host"`
	BasicAuthUsername string `toml:"username"`
	BasicAuthPassword string `toml:"password"`
	Id                string `toml:"id"`
	client            http.Client
	tls.ClientConfig
	Timeout internal.Duration `toml:"timeout"`
}

func (r *Redfish) Description() string {
	return "Read CPU, Fans, Powersupply and Voltage metrics of Dell/HP hardware server through redfish APIs"
}

var redfishConfig = `
  ##Server  OOB-IP
  host = "192.0.0.1"

  ##Username,  Password   for   hardware   server
  username = "test"
  password = "test"

  ##Resource  Id   for   redfish   APIs
  id="System.Embedded.1"

  ##Optional TLS   Config, if not provided insecure skip verifies defaults to true
  #tls_ca = "/etc/telegraf/ca.pem"
  #tls_cert = "/etc/telegraf/cert.pem"
  #tls_key = "/etc/telegraf/key.pem"
  

  ## Amount   of   time   allowed   to   complete   the   HTTP   request
  # timeout = "5s"
`

func (r *Redfish) SampleConfig() string {
	return redfishConfig
}

func (r *Redfish) Init() error {
	tlsCfg, err := r.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	if (len(r.ClientConfig.TLSCA) == 0) || (len(r.ClientConfig.TLSCert) == 0 && len(r.ClientConfig.TLSKey) == 0) {
		var insecuretls tls.ClientConfig
		insecuretls.InsecureSkipVerify = true
		tlsCfg, err = insecuretls.TLSConfig()
		if err != nil {
			return err
		}
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

func (r *Redfish) Gather(acc telegraf.Accumulator) error {
	var url []string
	var payload Payload
	url = append(url, fmt.Sprint("https://", r.Host, "/redfish/v1/Chassis/", r.Id, "/Thermal"), fmt.Sprint("https://", r.Host, "/redfish/v1/Chassis/", r.Id, "/Power"), fmt.Sprint("https://", r.Host, "/redfish/v1/Systems/", r.Id), fmt.Sprint("https://", r.Host, "/redfish/v1/Chassis/", r.Id, "/"))

	if len(r.Host) == 0 || len(r.BasicAuthUsername) == 0 || len(r.BasicAuthPassword) == 0 {
		return fmt.Errorf("Did not provide IP or username and password")
	}
	if len(r.Id) == 0 {
		return fmt.Errorf("Did not provide all the ID of the resource")
	}

	for _, i := range url {
		req, err := http.NewRequest("GET", i, nil)
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
		if resp.StatusCode == 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("%v", err)
			}
			jsonErr := json.Unmarshal(body, &payload)
			if jsonErr != nil {
				return fmt.Errorf("error parsing input: %v", jsonErr)
			}

		} else {
			return fmt.Errorf("received status code %d (%s), expected 200",
				resp.StatusCode,
				http.StatusText(resp.StatusCode))
		}
	}
	if payload.Location != nil {
		for _, j := range payload.Temperatures {
			//  Tags
			tags := map[string]string{}
			tags["source_ip"] = r.Host
			tags["name"] = j.Name
			tags["source"] = payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			tags["datacenter"] = payload.Location.PostalAddress.DataCenter
			tags["room"] = payload.Location.PostalAddress.Room
			tags["rack"] = payload.Location.Placement.Rack
			tags["row"] = payload.Location.Placement.Row
			//  Fields
			fields := make(map[string]interface{})
			fields["temperature"] = j.ReadingCelsius
			acc.AddFields("redfish_thermal_temperatures", fields, tags)
		}
		for _, j := range payload.Fans {
			//  Tags
			tags := map[string]string{}
			tags["source_ip"] = r.Host
			tags["name"] = j.Name
			tags["source"] = payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			tags["datacenter"] = payload.Location.PostalAddress.DataCenter
			tags["room"] = payload.Location.PostalAddress.Room
			tags["rack"] = payload.Location.Placement.Rack
			tags["row"] = payload.Location.Placement.Row
			//  Fields
			fields := make(map[string]interface{})
			fields["fanspeed"] = j.Reading
			acc.AddFields("redfish_thermal_fans", fields, tags)
		}
		for _, j := range payload.PowerSupplies {
			//  Tags
			tags := map[string]string{}
			tags["source_ip"] = r.Host
			tags["name"] = j.Name //j.Name
			tags["source"] = payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			tags["datacenter"] = payload.Location.PostalAddress.DataCenter
			tags["room"] = payload.Location.PostalAddress.Room
			tags["rack"] = payload.Location.Placement.Rack
			tags["row"] = payload.Location.Placement.Row
			//  Fields
			fields := make(map[string]interface{})
			fields["power_input_watts"] = j.PowerInputWatts
			fields["power_output_watts"] = j.PowerOutputWatts
			fields["line_input_voltage"] = j.LineInputVoltage
			fields["last_power_output_watts"] = j.LastPowerOutputWatts
			fields["power_capacity_watts"] = j.PowerCapacityWatts
			acc.AddFields("redfish_power_powersupplies", fields, tags)
		}
		for _, j := range payload.Voltages {
			//  Tags
			tags := map[string]string{}
			tags["source_ip"] = r.Host
			tags["name"] = j.Name
			tags["source"] = payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			tags["datacenter"] = payload.Location.PostalAddress.DataCenter
			tags["room"] = payload.Location.PostalAddress.Room
			tags["rack"] = payload.Location.Placement.Rack
			tags["row"] = payload.Location.Placement.Row
			//  Fields
			fields := make(map[string]interface{})
			fields["voltage"] = j.ReadingVolts
			acc.AddFields("redfish_power_voltages", fields, tags)
		}
	} else {
		for _, j := range payload.Temperatures {
			//  Tags
			tags := map[string]string{}
			tags["source_ip"] = r.Host
			tags["name"] = j.Name
			tags["source"] = payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			//  Fields
			fields := make(map[string]interface{})
			fields["temperature"] = j.ReadingCelsius
			acc.AddFields("redfish_thermal_temperatures", fields, tags)
		}
		for _, j := range payload.Fans {
			//  Tags
			tags := map[string]string{}
			tags["source_ip"] = r.Host
			tags["name"] = j.Name
			tags["source"] = payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			//  Fields
			fields := make(map[string]interface{})
			fields["fanspeed"] = j.Reading
			acc.AddFields("redfish_thermal_fans", fields, tags)
		}
		for _, j := range payload.PowerSupplies {
			//  Tags
			tags := map[string]string{}
			tags["source_ip"] = r.Host
			tags["name"] = j.Name //j.Name
			tags["source"] = payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			//  Fields
			fields := make(map[string]interface{})
			/* if (j.PowerInputWatts != 0){
			        fields["power_input_watts"]  = j.PowerInputWatts
			        fields["power_output_watts"] = j.PowerOutputWatts
			}*/
			fields["line_input_voltage"] = j.LineInputVoltage
			fields["last_power_output_watts"] = j.LastPowerOutputWatts
			fields["power_capacity_watts"] = j.PowerCapacityWatts
			acc.AddFields("redfish_power_powersupplies", fields, tags)
		}
		for _, j := range payload.Voltages {
			//  Tags
			tags := map[string]string{}
			tags["source_ip"] = r.Host
			tags["name"] = j.Name
			tags["source"] = payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			//  Fields
			fields := make(map[string]interface{})
			fields["voltage"] = j.ReadingVolts
			acc.AddFields("redfish_power_voltages", fields, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("redfish", func() telegraf.Input { return &Redfish{} })
}
