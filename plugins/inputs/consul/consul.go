package consul

import (
	"net/http"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Consul struct {
	Address    string
	Scheme     string
	Token      string
	Username   string
	Password   string
	Datacentre string // deprecated in 1.10; use Datacenter
	Datacenter string
	tls.ClientConfig
	TagDelimiter  string
	MetricVersion int
	Log           telegraf.Logger

	// client used to connect to Consul agnet
	client *api.Client
}

var sampleConfig = `
  ## Consul server address
  # address = "localhost:8500"

  ## URI scheme for the Consul server, one of "http", "https"
  # scheme = "http"

  ## Metric version controls the mapping from Consul metrics into
  ## Telegraf metrics.
  ##
  ##   example: metric_version = 1; deprecated in 1.15
  ##            metric_version = 2; recommended version
  # metric_version = 1

  ## ACL token used in every request
  # token = ""

  ## HTTP Basic Authentication username and password.
  # username = ""
  # password = ""

  ## Data center to query the health checks from
  # datacenter = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = true

  ## Consul checks' tag splitting
  # When tags are formatted like "key:value" with ":" as a delimiter then
  # they will be splitted and reported as proper key:value in Telegraf
  # tag_delimiter = ":"
`

func (c *Consul) Init() error {
	if c.MetricVersion != 2 {
		c.Log.Warnf("Use of deprecated configuration: 'metric_version = 1'; please update to 'metric_version = 2'")
	}

	return nil
}

func (c *Consul) Description() string {
	return "Gather health check statuses from services registered in Consul"
}

func (c *Consul) SampleConfig() string {
	return sampleConfig
}

func (c *Consul) createAPIClient() (*api.Client, error) {
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
		return nil, err
	}

	config.Transport = &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	return api.NewClient(config)
}

func (c *Consul) GatherHealthCheck(acc telegraf.Accumulator, checks []*api.HealthCheck) {
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

func (c *Consul) Gather(acc telegraf.Accumulator) error {
	if c.client == nil {
		newClient, err := c.createAPIClient()

		if err != nil {
			return err
		}

		c.client = newClient
	}

	checks, _, err := c.client.Health().State("any", nil)

	if err != nil {
		return err
	}

	c.GatherHealthCheck(acc, checks)

	return nil
}

func init() {
	inputs.Add("consul", func() telegraf.Input {
		return &Consul{}
	})
}
