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

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"

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
	Log              telegraf.Logger `toml:"-"`

	tagSet map[string]bool
	client http.Client
	tls.ClientConfig
	gf   *gofish.Service
	host string
}

type datacenterTag struct {
	Location struct {
		PostalAddress struct {
			DataCenter string
		}
	}
}

func (*Redfish) SampleConfig() string {
	return sampleConfig
}

func (r *Redfish) Init() error {
	// Check all config values

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

	if len(r.IncludeTagSets) > 0 {
		r.Log.Warn("The tag Datacenter (Chassis.Location.PostalAddress.DataCenter) is not part of DMTF's standard an will be removed in a future version")
	}

	r.tagSet = make(map[string]bool, len(r.IncludeTagSets))
	for _, setLabel := range r.IncludeTagSets {
		r.tagSet[setLabel] = true
	}

	// Remove all the fluff form the endpoint address to just have the hostname or IP
	redfishURL, err := url.Parse(r.Address)
	if err != nil {
		r.host = redfishURL.Host
	} else {
		r.host, _, err = net.SplitHostPort(redfishURL.Host)
	}

	if err != nil {
		r.host = redfishURL.Host
	}

	return nil
}

func (r *Redfish) Start(acc telegraf.Accumulator) error {
	var err error
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

	gofishConfig := gofish.ClientConfig{
		Endpoint:  r.Address,
		Username:  user,
		Password:  pass,
		BasicAuth: true,
	}
	c, err := gofish.Connect(gofishConfig)
	if err != nil {
		return fmt.Errorf("error parsing input from %s. This is likely due to the BMC response being in text/html: %w",
			r.Address+"/redfish/v1/Systems/"+r.ComputerSystemID,
			err)
	}

	// Retrieve the service root
	r.gf = c.Service

	return err
}

func (r *Redfish) Stop() {

}

func (r *Redfish) Gather(acc telegraf.Accumulator) error {
	systems, err := r.gf.Systems()
	if err != nil {
		return err
	}

	// Process only the system defined via ComputerSystemID in the config
	// Collect configured metrics on every chassis
	for _, system := range systems {
		if system.ID != r.ComputerSystemID {
			continue
		}

		chassisList, err := system.Chassis()
		if err != nil {
			return err
		}

		if len(chassisList) == 0 {
			r.Log.Warn("No chassis found, no metric can be produced")
			return nil
		}

		for _, chassis := range chassisList {
			for _, metric := range r.IncludeMetrics {
				var err error
				switch metric {
				case "thermal":
					err = r.gatherThermal(acc, r.host, system, chassis)
				case "power":
					err = r.gatherPower(acc, r.host, system, chassis)
				default:
					return fmt.Errorf("unknown metric requested: %s", metric)
				}
				if err != nil {
					return err
				}
			}
		}
		break
	}
	return nil
}

func setChassisTags(chassis *schemas.Chassis, tags map[string]string) {
	tags["chassis_chassistype"] = string(chassis.ChassisType)
	tags["chassis_manufacturer"] = chassis.Manufacturer
	tags["chassis_model"] = chassis.Model
	tags["chassis_partnumber"] = chassis.PartNumber
	tags["chassis_powerstate"] = string(chassis.PowerState)
	tags["chassis_sku"] = chassis.SKU
	tags["chassis_serialnumber"] = chassis.SerialNumber
	tags["chassis_state"] = string(chassis.Status.State)
	tags["chassis_health"] = string(chassis.Status.Health)
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
