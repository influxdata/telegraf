package redfish

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
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

func Severity(critical, fatal, val int) string {
	severity := "NA"
	if (critical != 0) || (fatal != 0) {
		if (val >= fatal) && (fatal != 0) {
			severity = "Fatal"
		} else if (val >= critical) && (critical != 0) {
			severity = "Critical"
		} else {
			severity = "OK"
		}
	}
	return severity
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
			tags["severity"] = Severity(j.UpperThresholdCritical, j.UpperThresholdFatal, j.ReadingCelsius)
			//  Fields
			fields := make(map[string]interface{})
			fields["temperature"] = j.ReadingCelsius
			fields["upper_threshold_critical"] = j.UpperThresholdCritical
			fields["upper_threshold_fatal"] = j.UpperThresholdFatal
			acc.AddFields("redfish_thermal_temperatures", fields, tags)
		}
		for _, j := range payload.Fans {
			//  Tags
			tags := map[string]string{}
			fields := make(map[string]interface{})
			tags["source_ip"] = r.Host
			tags["name"] = j.Name
			tags["source"] = payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			tags["datacenter"] = payload.Location.PostalAddress.DataCenter
			tags["room"] = payload.Location.PostalAddress.Room
			tags["rack"] = payload.Location.Placement.Rack
			tags["row"] = payload.Location.Placement.Row
			if j.ReadingUnits == "RPM" {
				tags["severity"] = Severity(j.UpperThresholdCritical, j.UpperThresholdFatal, j.Reading)
				fields["upper_threshold_critical"] = j.UpperThresholdCritical
				fields["upper_threshold_fatal"] = j.UpperThresholdFatal

			}

			//  Fields
			fields["fanspeed"] = j.Reading
			acc.AddFields("redfish_thermal_fans", fields, tags)
		}
		for _, j := range payload.PowerSupplies {
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
			fields["power_input_watts"] = math.Round(j.PowerInputWatts*100) / 100
			fields["power_output_watts"] = math.Round(j.PowerOutputWatts*100) / 100
			fields["line_input_voltage"] = math.Round(j.LineInputVoltage*100) / 100
			fields["last_power_output_watts"] = math.Round(j.LastPowerOutputWatts*100) / 100
			fields["power_capacity_watts"] = math.Round(j.PowerCapacityWatts*100) / 100
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
			tags["severity"] = Severity(int(math.Round(j.UpperThresholdCritical)), int(math.Round(j.UpperThresholdFatal)), int(math.Round(j.ReadingVolts)))
			//  Fields
			fields := make(map[string]interface{})
			fields["voltage"] = math.Round(j.ReadingVolts*100) / 100
			fields["upper_threshold_critical"] = math.Round(j.UpperThresholdCritical*100) / 100
			fields["upper_threshold_fatal"] = math.Round(j.UpperThresholdFatal*100) / 100
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
			tags["severity"] = Severity(j.UpperThresholdCritical, j.UpperThresholdFatal, j.ReadingCelsius)
			//  Fields
			fields := make(map[string]interface{})
			fields["temperature"] = j.ReadingCelsius
			fields["upper_threshold_critical"] = j.UpperThresholdCritical
			fields["upper_threshold_fatal"] = j.UpperThresholdFatal
			acc.AddFields("redfish_thermal_temperatures", fields, tags)
		}
		for _, j := range payload.Fans {
			//  Tags
			tags := map[string]string{}
			fields := make(map[string]interface{})
			tags["source_ip"] = r.Host
			tags["name"] = j.Name
			tags["source"] = payload.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			if j.ReadingUnits == "RPM" {
				tags["upper_threshold_critical"] = strconv.Itoa(j.UpperThresholdCritical)
				fields["upper_threshold_fatal"] = strconv.Itoa(j.UpperThresholdFatal)
				fields["severity"] = Severity(j.UpperThresholdCritical, j.UpperThresholdFatal, j.Reading)
			}
			//  Fields
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
			fields["line_input_voltage"] = math.Round(j.LineInputVoltage*100) / 100
			fields["last_power_output_watts"] = math.Round(j.LastPowerOutputWatts*100) / 100
			fields["power_capacity_watts"] = math.Round(j.PowerCapacityWatts*100) / 100
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
			tags["severity"] = Severity(int(math.Round(j.UpperThresholdCritical)), int(math.Round(j.UpperThresholdFatal)), int(math.Round(j.ReadingVolts)))
			//  Fields
			fields := make(map[string]interface{})
			fields["voltage"] = math.Round(j.ReadingVolts*100) / 100
			fields["upper_threshold_critical"] = math.Round(j.UpperThresholdCritical*100) / 100
			fields["upper_threshold_fatal"] = math.Round(j.UpperThresholdFatal*100) / 100
			acc.AddFields("redfish_power_voltages", fields, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("redfish", func() telegraf.Input { return &Redfish{} })
}
