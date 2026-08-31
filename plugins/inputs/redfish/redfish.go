//go:generate ../../../tools/readme_config_includer/generator
package redfish

import (
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"
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
	Log              telegraf.Logger `toml:"-"`

	tagSet map[string]bool
	client http.Client
	tls.ClientConfig
	gf *gofish.Service
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
		// References are resolved against the configured address, so an empty
		// one would request the device's web root instead of a Redfish resource
		if link.Ref == "" {
			r.Log.Warn("Skipping chassis without reference")
			continue
		}

		chassis, err := r.getChassis(link.Ref)
		if err != nil {
			return err
		}

		for _, metric := range r.IncludeMetrics {
			var err error
			switch metric {
			case "thermal":
				if chassis.Thermal.Ref == "" {
					r.Log.Warnf("Skipping thermal data of chassis %q without reference", link.Ref)
					continue
				}
				err = r.gatherThermal(acc, address, system, chassis)
			case "power":
				if chassis.Power.Ref == "" {
					r.Log.Warnf("Skipping power data of chassis %q without reference", link.Ref)
					continue
				}
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
