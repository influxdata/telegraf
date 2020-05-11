// Copyright 2020, Verizon
//Licensed under the terms of the MIT License. SEE LICENSE file in project root for terms.

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
	"strconv"
)

type Hostname struct {
	Hostname string `json:"HostName"`
}
type Cpu struct {
	Name        string    `json:"Name"`
	Temperature int64     `json:"ReadingCelsius"`
	Status      CpuStatus `json:"Status"`
}
type Temperatures struct {
	Temperatures []Cpu `json:"Temperatures"`
}
type CpuStatus struct {
	State  string `json:"State"`
	Health string `json:"Health"`
}
type Fans struct {
	Fans []speed `json:"Fans"`
}
type speed struct {
	Name   string     `json:"Name"`
	Speed  int64      `json:"Reading"`
	Status FansStatus `json:"Status"`
}
type FansStatus struct {
	State  string `json:"State"`
	Health string `json:"Health"`
}
type PowerSupplies struct {
	PowerSupplies []watt `json:"PowerSupplies"`
}
type PowerSupplieshp struct {
	PowerSupplieshp []watthp `json:"PowerSupplies"`
}
type watt struct {
	Name               string      `json:"Name"`
	PowerInputWatts    float64     `json:"PowerInputWatts"`
	PowerCapacityWatts float64     `json:"PowerCapacityWatts"`
	PowerOutputWatts   float64     `json:"PowerOutputWatts"`
	Status             PowerStatus `json:"Status"`
}
type watthp struct {
	Name                 string  `json:"Name"`
	MemberID             string  `json:"MemberId"`
	PowerCapacityWatts   float64 `json:"PowerCapacityWatts"`
	LastPowerOutputWatts float64 `json:"LastPowerOutputWatts"`
	LineInputVoltage     float64 `json:"LineInputVoltage"`
}
type PowerStatus struct {
	State  string `json:"State"`
	Health string `json:"Health"`
}
type Voltages struct {
	Voltages []volt `json:"Voltages"`
}
type volt struct {
	Name         string     `json:"Name"`
	ReadingVolts int64      `json:"ReadingVolts"`
	Status       VoltStatus `json:"Status"`
}
type VoltStatus struct {
	State  string `json:"State"`
	Health string `json:"Health"`
}
type Location struct {
	Location Address `json:"Location"`
}
type Address struct {
	PostalAddress PostalAddress `json:"PostalAddress"`
	Placement     Placement     `json:"Placement"`
}
type PostalAddress struct {
	DataCenter string `json:"Building"`
	Room       string `json:"Room"`
}
type Placement struct {
	Rack string `json:"Rack"`
	Row  string `json:"Row"`
}
type Redfish struct {
	Host              string `toml:"host"`
	BasicAuthUsername string `toml:"basicauthusername"`
	BasicAuthPassword string `toml:"basicauthpassword"`
	Id                string `toml:"id"`
	Server            string `toml:"server"`
	client            http.Client
	tls.ClientConfig
	Timeout     internal.Duration `toml:"timeout"`
	hostname    Hostname
	temperature Temperatures
	fan         Fans
	powerdell   PowerSupplies
	voltage     Voltages
	powerhp     PowerSupplieshp
	location    Location
}

func (r *Redfish) getMetrics() error {
	url := make(map[string]map[string]interface{})
	url["Thermal"] = make(map[string]interface{})
	url["Power"] = make(map[string]interface{})
	url["Hostname"] = make(map[string]interface{})
	url["Thermal"]["endpoint"] = fmt.Sprint(r.Host, "/redfish/v1/Chassis/", r.Id, "/Thermal")
	url["Thermal"]["pointer"] = &r.temperature
	url["Thermal"]["fanpointer"] = &r.fan
	url["Power"]["endpoint"] = fmt.Sprint(r.Host, "/redfish/v1/Chassis/", r.Id, "/Power")
	if r.Server == "dell" {
		url["Power"]["pointer"] = &r.powerdell
		url["Power"]["voltpointer"] = &r.voltage
	} else if r.Server == "hp" {
		url["Power"]["pointer"] = &r.powerhp
	}
	url["Hostname"]["endpoint"] = fmt.Sprint(r.Host, "/redfish/v1/Systems/", r.Id)
	url["Hostname"]["pointer"] = &r.hostname
	if r.Server == "dell" {
		url["Location"] = make(map[string]interface{})
		url["Location"]["endpoint"] = fmt.Sprint(r.Host, "/redfish/v1/Chassis/", r.Id, "/")
		url["Location"]["pointer"] = &r.location
	}

	for key, value := range url {
		req, err := http.NewRequest("GET", value["endpoint"].(string), nil)
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
				return err
			}
			jsonErr := json.Unmarshal(body, value["pointer"])
			if jsonErr != nil {
				return fmt.Errorf("error parsing input: %v", jsonErr)
			}
			if r.Server == "dell" && key == "Power" {
				jsonErr = json.Unmarshal(body, value["voltpointer"])
				if jsonErr != nil {
					return fmt.Errorf("error parsing input: %v", jsonErr)
				}

			} else if key == "Thermal" {
				jsonErr = json.Unmarshal(body, value["fanpointer"])
				if jsonErr != nil {
					return fmt.Errorf("error parsing input: %v", jsonErr)
				}
			}

		} else {
			return fmt.Errorf("received status code %d (%s), expected 200",
				resp.StatusCode,
				http.StatusText(resp.StatusCode))
		}
	}
	return nil
}

func (r *Redfish) Description() string {
	return "Read CPU, Fans, Powersupply and Voltage metrics of Dell/HP hardware server through redfish APIs"
}

var redfishConfig = `
## Server OOB-IP
host = "https://192.0.0.1"

## Username,Password for hardware server
basicauthusername = "test"
basicauthpassword = "test"
## Server Vendor(dell or hp)
server= "dell"
## Resource Id for redfish APIs
id="System.Embedded.1"
## Optional TLS Config
# tls_ca = "/etc/telegraf/ca.pem"
# tls_cert = "/etc/telegraf/cert.pem"
# tls_key = "/etc/telegraf/key.pem"
## Use TLS but skip chain & host verification
# insecure_skip_verify = false

## Amount of time allowed to complete the HTTP request
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

	if len(r.Host) == 0 || len(r.BasicAuthUsername) == 0 || len(r.BasicAuthPassword) == 0 {
		return fmt.Errorf("Did not provide IP or username and password")
	}
	if len(r.Server) == 0 || len(r.Id) == 0 {
		return fmt.Errorf("Did not provide all the mandatory fields in the configuration")
	}
	if !(r.Server == "dell" || r.Server == "hp") {
		return fmt.Errorf("Did not provide correct server information, supported server details are dell or hp")
	}
	err := r.getMetrics()
	if err != nil {
		return err
	}

	for i := 0; i < len(r.temperature.Temperatures); i++ {
		//  Tags
		tags := map[string]string{"oob_ip": r.Host, "name": r.temperature.Temperatures[i].Name, "hostname": r.hostname.Hostname}
		//  Fields
		fields := make(map[string]interface{})
		fields["temperature"] = strconv.FormatInt(r.temperature.Temperatures[i].Temperature, 10)
		fields["state"] = r.temperature.Temperatures[i].Status.State
		fields["health"] = r.temperature.Temperatures[i].Status.Health
		if r.Server == "dell" {
			fields["datacenter"] = r.location.Location.PostalAddress.DataCenter
			fields["room"] = r.location.Location.PostalAddress.Room
			fields["rack"] = r.location.Location.Placement.Rack
			fields["row"] = r.location.Location.Placement.Row
			acc.AddFields("cpu_temperature", fields, tags)
		}
		if r.Server == "hp" {
			acc.AddFields("cpu_temperature", fields, tags)
		}
	}
	for i := 0; i < len(r.fan.Fans); i++ {
		//  Tags
		tags := map[string]string{"oob_ip": r.Host, "name": r.fan.Fans[i].Name, "hostname": r.hostname.Hostname}
		//  Fields
		fields := make(map[string]interface{})
		fields["fanspeed"] = strconv.FormatInt(r.fan.Fans[i].Speed, 10)
		fields["state"] = r.fan.Fans[i].Status.State
		fields["health"] = r.fan.Fans[i].Status.Health
		if r.Server == "dell" {
			fields["datacenter"] = r.location.Location.PostalAddress.DataCenter
			fields["room"] = r.location.Location.PostalAddress.Room
			fields["rack"] = r.location.Location.Placement.Rack
			fields["row"] = r.location.Location.Placement.Row
			acc.AddFields("fans", fields, tags)
		}
		if r.Server == "hp" {
			acc.AddFields("fans", fields, tags)
		}
	}
	if r.Server == "dell" {
		for i := 0; i < len(r.powerdell.PowerSupplies); i++ {
			//  Tags
			tags := map[string]string{"oob_ip": r.Host, "name": r.powerdell.PowerSupplies[i].Name, "hostname": r.hostname.Hostname}
			//  Fields
			fields := make(map[string]interface{})
			fields["power_input_watts"] = strconv.FormatFloat(r.powerdell.PowerSupplies[i].PowerInputWatts, 'f', -1, 64)
			fields["power_capacity_watts"] = strconv.FormatFloat(r.powerdell.PowerSupplies[i].PowerCapacityWatts, 'f', -1, 64)
			fields["power_output_watts"] = strconv.FormatFloat(r.powerdell.PowerSupplies[i].PowerOutputWatts, 'f', -1, 64)
			fields["state"] = r.powerdell.PowerSupplies[i].Status.State
			fields["health"] = r.powerdell.PowerSupplies[i].Status.Health
			fields["datacenter"] = r.location.Location.PostalAddress.DataCenter
			fields["room"] = r.location.Location.PostalAddress.Room
			fields["rack"] = r.location.Location.Placement.Rack
			fields["row"] = r.location.Location.Placement.Row
			acc.AddFields("powersupply", fields, tags)
		}
		for i := 0; i < len(r.voltage.Voltages); i++ {
			//  Tags
			tags := map[string]string{"oob_ip": r.Host, "name": r.voltage.Voltages[i].Name, "hostname": r.hostname.Hostname}
			//  Fields
			fields := make(map[string]interface{})
			fields["voltage"] = strconv.FormatInt(r.voltage.Voltages[i].ReadingVolts, 10)
			fields["state"] = r.voltage.Voltages[i].Status.State
			fields["health"] = r.voltage.Voltages[i].Status.Health
			fields["datacenter"] = r.location.Location.PostalAddress.DataCenter
			fields["room"] = r.location.Location.PostalAddress.Room
			fields["rack"] = r.location.Location.Placement.Rack
			fields["row"] = r.location.Location.Placement.Row
			acc.AddFields("voltages", fields, tags)
		}

	}
	if r.Server == "hp" {
		for i := 0; i < len(r.powerhp.PowerSupplieshp); i++ {
			//  Tags
			tags := map[string]string{"oob_ip": r.Host, "name": r.powerhp.PowerSupplieshp[i].Name, "member_id": r.powerhp.PowerSupplieshp[i].MemberID, "hostname": r.hostname.Hostname}
			//  Fields
			fields := make(map[string]interface{})
			fields["line_input_voltage"] = strconv.FormatFloat(r.powerhp.PowerSupplieshp[i].LineInputVoltage, 'f', -1, 64)
			fields["power_capacity_watts"] = strconv.FormatFloat(r.powerhp.PowerSupplieshp[i].PowerCapacityWatts, 'f', -1, 64)
			fields["last_power_output_watts"] = strconv.FormatFloat(r.powerhp.PowerSupplieshp[i].LastPowerOutputWatts, 'f', -1, 64)
			acc.AddFields("powersupply", fields, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("redfish", func() telegraf.Input { return &Redfish{} })
}
