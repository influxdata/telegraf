//go:generate ../../../tools/readme_config_includer/generator
package consul

import (
	_ "embed"
	"net/http"
	"strings"

	"github.com/hashicorp/consul/api"

	"github.com/influxdata/telegraf"
	telegraf_config "github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Consul struct {
	Address       string `toml:"address"`
	Scheme        string `toml:"scheme"`
	Token         string `toml:"token"`
	Username      string `toml:"username"`
	Password      string `toml:"password"`
	Datacentre    string `toml:"datacentre" deprecated:"1.10.0;1.35.0;use 'datacenter' instead"`
	Datacenter    string `toml:"datacenter"`
	TagDelimiter  string `toml:"tag_delimiter"`
	MetricVersion int    `toml:"metric_version"`
	Log           telegraf.Logger
	tls.ClientConfig

	// client used to connect to Consul agent
	client *api.Client
}

func (*Consul) SampleConfig() string {
	return sampleConfig
}

func (c *Consul) Init() error {
	if c.MetricVersion != 2 {
		telegraf_config.PrintOptionValueDeprecationNotice("inputs.consul", "metric_version", 1,
			telegraf.DeprecationInfo{
				Since:     "1.16.0",
				RemovalIn: "1.40.0",
				Notice:    `please update to 'metric_version = 2'`,
			},
		)
	}

	config := api.DefaultConfig()

	if c.Address != "" {
		config.Address = c.Address
	}

	if c.Scheme != "" {
		config.Scheme = c.Scheme
	}

	if c.Datacentre != "" {
		config.Datacenter = c.Datacentre
	}

	if c.Datacenter != "" {
		config.Datacenter = c.Datacenter
	}

	if c.Token != "" {
		config.Token = c.Token
	}

	if c.Username != "" {
		config.HttpAuth = &api.HttpBasicAuth{
			Username: c.Username,
			Password: c.Password,
		}
	}

	tlsCfg, err := c.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	config.Transport = &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	c.client, err = api.NewClient(config)
	return err
}

func (c *Consul) Gather(acc telegraf.Accumulator) error {
	checks, _, err := c.client.Health().State("any", nil)

	if err != nil {
		return err
	}

	c.gatherHealthCheck(acc, checks)

	return nil
}

func (c *Consul) gatherHealthCheck(acc telegraf.Accumulator, checks []*api.HealthCheck) {
	for _, check := range checks {
		record := make(map[string]interface{})
		tags := make(map[string]string)

		record["passing"] = 0
		record["critical"] = 0
		record["warning"] = 0
		record[check.Status] = 1

		if c.MetricVersion == 2 {
			tags["check_name"] = check.Name
			tags["service_id"] = check.ServiceID
			tags["status"] = check.Status
		} else {
			record["check_name"] = check.Name
			record["service_id"] = check.ServiceID
			record["status"] = check.Status
		}

		tags["node"] = check.Node
		tags["service_name"] = check.ServiceName
		tags["check_id"] = check.CheckID

		for _, checkTag := range check.ServiceTags {
			if c.TagDelimiter != "" {
				splittedTag := strings.SplitN(checkTag, c.TagDelimiter, 2)
				if len(splittedTag) == 1 && checkTag != "" {
					tags[checkTag] = checkTag
				} else if len(splittedTag) == 2 && splittedTag[1] != "" {
					tags[splittedTag[0]] = splittedTag[1]
				}
			} else if checkTag != "" {
				tags[checkTag] = checkTag
			}
		}

		acc.AddFields("consul_health_checks", record, tags)
	}
}

func init() {
	inputs.Add("consul", func() telegraf.Input {
		return &Consul{}
	})
}
