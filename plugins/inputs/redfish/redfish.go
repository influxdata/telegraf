//go:generate ../../../tools/readme_config_includer/generator
package redfish

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	// tag sets used for including redfish OData link parent data
	tagSetChassisLocation = "chassis.location"
	tagSetChassis         = "chassis"
)

type Redfish struct {
	Address          string          `toml:"address"`
	Username         config.Secret   `toml:"username"`
	Password         config.Secret   `toml:"password"`
	ComputerSystemID string          `toml:"computer_system_id"`
	IncludeMetrics   []string        `toml:"include_metrics"`
	IncludeTagSets   []string        `toml:"include_tag_sets"`
	Workarounds      []string        `toml:"workarounds"`
	Timeout          config.Duration `toml:"timeout"`

	tagSet map[string]bool
	client http.Client
	tls.ClientConfig
	baseURL *url.URL
}

type system struct {
	Hostname string `json:"hostname"`
	Links    struct {
		Chassis []struct {
			Ref string `json:"@odata.id"`
		}
	}
}

type chassis struct {
	ChassisType  string
	Location     *location
	Manufacturer string
	Model        string
	PartNumber   string
	Power        struct {
		Ref string `json:"@odata.id"`
	}
	PowerState   string
	SKU          string
	SerialNumber string
	Status       status
	Thermal      struct {
		Ref string `json:"@odata.id"`
	}
}

type location struct {
	PostalAddress struct {
		DataCenter string
		Room       string
	}
	Placement struct {
		Rack string
		Row  string
	}
}

type status struct {
	State  string
	Health string
}

func (*Redfish) SampleConfig() string {
	return sampleConfig
}

func (r *Redfish) Init() error {
	if r.Address == "" {
		return errors.New("did not provide IP")
	}

	if r.Username.Empty() && r.Password.Empty() {
		return errors.New("did not provide username and password")
	}

	if r.ComputerSystemID == "" {
		return errors.New("did not provide the computer system ID of the resource")
	}

	if len(r.IncludeMetrics) == 0 {
		return errors.New("no metrics specified to collect")
	}
	for _, metric := range r.IncludeMetrics {
		switch metric {
		case "thermal", "power":
		default:
			return fmt.Errorf("unknown metric requested: %s", metric)
		}
	}

	for _, workaround := range r.Workarounds {
		switch workaround {
		case "ilo4-thermal":
		default:
			return fmt.Errorf("unknown workaround requested: %s", workaround)
		}
	}
	r.tagSet = make(map[string]bool, len(r.IncludeTagSets))
	for _, setLabel := range r.IncludeTagSets {
		r.tagSet[setLabel] = true
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

		for _, metric := range r.IncludeMetrics {
			var err error
			switch metric {
			case "thermal":
				err = r.gatherThermal(acc, address, system, chassis)
			case "power":
				err = r.gatherPower(acc, address, system, chassis)
			default:
				return fmt.Errorf("unknown metric requested: %s", metric)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Redfish) getData(address string, payload interface{}) error {
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return err
	}

	username, err := r.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	user := username.String()
	username.Destroy()

	password, err := r.Password.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %w", err)
	}
	pass := password.String()
	password.Destroy()

	req.SetBasicAuth(user, pass)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OData-Version", "4.0")

	// workaround for iLO4 thermal data
	if slices.Contains(r.Workarounds, "ilo4-thermal") && strings.Contains(address, "/Thermal") {
		req.Header.Del("OData-Version")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("received status code %d (%s) for address %s, expected 200",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			address)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		return fmt.Errorf("error parsing input from %s (Content-Type %q): %w", address, resp.Header.Get("Content-Type"), err)
	}

	return nil
}

func (r *Redfish) getComputerSystem(id string) (*system, error) {
	loc := r.baseURL.ResolveReference(&url.URL{Path: path.Join("/redfish/v1/Systems/", id)})
	system := &system{}
	err := r.getData(loc.String(), system)
	if err != nil {
		return nil, err
	}
	return system, nil
}

func (r *Redfish) getChassis(ref string) (*chassis, error) {
	loc := r.baseURL.ResolveReference(&url.URL{Path: ref})
	chassis := &chassis{}
	err := r.getData(loc.String(), chassis)
	if err != nil {
		return nil, err
	}
	return chassis, nil
}

func setChassisTags(chassis *chassis, tags map[string]string) {
	tags["chassis_chassistype"] = chassis.ChassisType
	tags["chassis_manufacturer"] = chassis.Manufacturer
	tags["chassis_model"] = chassis.Model
	tags["chassis_partnumber"] = chassis.PartNumber
	tags["chassis_powerstate"] = chassis.PowerState
	tags["chassis_sku"] = chassis.SKU
	tags["chassis_serialnumber"] = chassis.SerialNumber
	tags["chassis_state"] = chassis.Status.State
	tags["chassis_health"] = chassis.Status.Health
}

func init() {
	inputs.Add("redfish", func() telegraf.Input {
		return &Redfish{
			// default tag set of chassis.location required for backwards compatibility
			IncludeTagSets: []string{tagSetChassisLocation},
			IncludeMetrics: []string{"power", "thermal"},
		}
	})
}
