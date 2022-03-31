package redfish

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Redfish struct {
	Address          string          `toml:"address"`
	Username         string          `toml:"username"`
	Password         string          `toml:"password"`
	ComputerSystemID string          `toml:"computer_system_id"`
	Timeout          config.Duration `toml:"timeout"`

	client http.Client
	tls.ClientConfig
	baseURL *url.URL
}

type System struct {
	Hostname string `json:"hostname"`
	Links    struct {
		Chassis []struct {
			Ref string `json:"@odata.id"`
		}
	}
}

type Chassis struct {
	Location *Location
	Power    struct {
		Ref string `json:"@odata.id"`
	}
	Thermal struct {
		Ref string `json:"@odata.id"`
	}
}

type Power struct {
	PowerSupplies []struct {
		Name                 string
		MemberID             string
		PowerInputWatts      *float64
		PowerCapacityWatts   *float64
		PowerOutputWatts     *float64
		LastPowerOutputWatts *float64
		Status               Status
		LineInputVoltage     *float64
	}
	Voltages []struct {
		Name                   string
		MemberID               string
		ReadingVolts           *float64
		UpperThresholdCritical *float64
		UpperThresholdFatal    *float64
		LowerThresholdCritical *float64
		LowerThresholdFatal    *float64
		Status                 Status
	}
}

type Thermal struct {
	Fans []struct {
		Name                   string
		MemberID               string
		Reading                *int64
		ReadingUnits           *string
		UpperThresholdCritical *int64
		UpperThresholdFatal    *int64
		LowerThresholdCritical *int64
		LowerThresholdFatal    *int64
		Status                 Status
	}
	Temperatures []struct {
		Name                   string
		MemberID               string
		ReadingCelsius         *float64
		UpperThresholdCritical *float64
		UpperThresholdFatal    *float64
		LowerThresholdCritical *float64
		LowerThresholdFatal    *float64
		Status                 Status
	}
}

type Location struct {
	PostalAddress struct {
		DataCenter string
		Room       string
	}
	Placement struct {
		Rack string
		Row  string
	}
}

type Status struct {
	State  string
	Health string
}

func (r *Redfish) Init() error {
	if r.Address == "" {
		return fmt.Errorf("did not provide IP")
	}

	if r.Username == "" && r.Password == "" {
		return fmt.Errorf("did not provide username and password")
	}

	if r.ComputerSystemID == "" {
		return fmt.Errorf("did not provide the computer system ID of the resource")
	}

	var err error
	r.baseURL, err = url.Parse(r.Address)
	if err != nil {
		return err
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
		Timeout: time.Duration(r.Timeout),
	}

	return nil
}

func (r *Redfish) getData(address string, payload interface{}) error {
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(r.Username, r.Password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OData-Version", "4.0")
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("received status code %d (%s) for address %s, expected 200",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			r.Address)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		return fmt.Errorf("error parsing input: %v", err)
	}

	return nil
}

func (r *Redfish) getComputerSystem(id string) (*System, error) {
	loc := r.baseURL.ResolveReference(&url.URL{Path: path.Join("/redfish/v1/Systems/", id)})
	system := &System{}
	err := r.getData(loc.String(), system)
	if err != nil {
		return nil, err
	}
	return system, nil
}

func (r *Redfish) getChassis(ref string) (*Chassis, error) {
	loc := r.baseURL.ResolveReference(&url.URL{Path: ref})
	chassis := &Chassis{}
	err := r.getData(loc.String(), chassis)
	if err != nil {
		return nil, err
	}
	return chassis, nil
}

func (r *Redfish) getPower(ref string) (*Power, error) {
	loc := r.baseURL.ResolveReference(&url.URL{Path: ref})
	power := &Power{}
	err := r.getData(loc.String(), power)
	if err != nil {
		return nil, err
	}
	return power, nil
}

func (r *Redfish) getThermal(ref string) (*Thermal, error) {
	loc := r.baseURL.ResolveReference(&url.URL{Path: ref})
	thermal := &Thermal{}
	err := r.getData(loc.String(), thermal)
	if err != nil {
		return nil, err
	}
	return thermal, nil
}

func (r *Redfish) Gather(acc telegraf.Accumulator) error {
	address, _, err := net.SplitHostPort(r.baseURL.Host)
	if err != nil {
		address = r.baseURL.Host
	}

	system, err := r.getComputerSystem(r.ComputerSystemID)
	if err != nil {
		return err
	}

	for _, link := range system.Links.Chassis {
		chassis, err := r.getChassis(link.Ref)
		if err != nil {
			return err
		}

		thermal, err := r.getThermal(chassis.Thermal.Ref)
		if err != nil {
			return err
		}

		for _, j := range thermal.Temperatures {
			tags := map[string]string{}
			tags["member_id"] = j.MemberID
			tags["address"] = address
			tags["name"] = j.Name
			tags["source"] = system.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			if chassis.Location != nil {
				tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
				tags["room"] = chassis.Location.PostalAddress.Room
				tags["rack"] = chassis.Location.Placement.Rack
				tags["row"] = chassis.Location.Placement.Row
			}

			fields := make(map[string]interface{})
			fields["reading_celsius"] = j.ReadingCelsius
			fields["upper_threshold_critical"] = j.UpperThresholdCritical
			fields["upper_threshold_fatal"] = j.UpperThresholdFatal
			fields["lower_threshold_critical"] = j.LowerThresholdCritical
			fields["lower_threshold_fatal"] = j.LowerThresholdFatal
			acc.AddFields("redfish_thermal_temperatures", fields, tags)
		}

		for _, j := range thermal.Fans {
			tags := map[string]string{}
			fields := make(map[string]interface{})
			tags["member_id"] = j.MemberID
			tags["address"] = address
			tags["name"] = j.Name
			tags["source"] = system.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			if chassis.Location != nil {
				tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
				tags["room"] = chassis.Location.PostalAddress.Room
				tags["rack"] = chassis.Location.Placement.Rack
				tags["row"] = chassis.Location.Placement.Row
			}

			if j.ReadingUnits != nil && *j.ReadingUnits == "RPM" {
				fields["upper_threshold_critical"] = j.UpperThresholdCritical
				fields["upper_threshold_fatal"] = j.UpperThresholdFatal
				fields["lower_threshold_critical"] = j.LowerThresholdCritical
				fields["lower_threshold_fatal"] = j.LowerThresholdFatal
				fields["reading_rpm"] = j.Reading
			} else {
				fields["reading_percent"] = j.Reading
			}
			acc.AddFields("redfish_thermal_fans", fields, tags)
		}

		power, err := r.getPower(chassis.Power.Ref)
		if err != nil {
			return err
		}

		for _, j := range power.PowerSupplies {
			tags := map[string]string{}
			tags["member_id"] = j.MemberID
			tags["address"] = address
			tags["name"] = j.Name
			tags["source"] = system.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			if chassis.Location != nil {
				tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
				tags["room"] = chassis.Location.PostalAddress.Room
				tags["rack"] = chassis.Location.Placement.Rack
				tags["row"] = chassis.Location.Placement.Row
			}

			fields := make(map[string]interface{})
			fields["power_input_watts"] = j.PowerInputWatts
			fields["power_output_watts"] = j.PowerOutputWatts
			fields["line_input_voltage"] = j.LineInputVoltage
			fields["last_power_output_watts"] = j.LastPowerOutputWatts
			fields["power_capacity_watts"] = j.PowerCapacityWatts
			acc.AddFields("redfish_power_powersupplies", fields, tags)
		}

		for _, j := range power.Voltages {
			tags := map[string]string{}
			tags["member_id"] = j.MemberID
			tags["address"] = address
			tags["name"] = j.Name
			tags["source"] = system.Hostname
			tags["state"] = j.Status.State
			tags["health"] = j.Status.Health
			if chassis.Location != nil {
				tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
				tags["room"] = chassis.Location.PostalAddress.Room
				tags["rack"] = chassis.Location.Placement.Rack
				tags["row"] = chassis.Location.Placement.Row
			}

			fields := make(map[string]interface{})
			fields["reading_volts"] = j.ReadingVolts
			fields["upper_threshold_critical"] = j.UpperThresholdCritical
			fields["upper_threshold_fatal"] = j.UpperThresholdFatal
			fields["lower_threshold_critical"] = j.LowerThresholdCritical
			fields["lower_threshold_fatal"] = j.LowerThresholdFatal
			acc.AddFields("redfish_power_voltages", fields, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("redfish", func() telegraf.Input {
		return &Redfish{}
	})
}
