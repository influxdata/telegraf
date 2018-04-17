package consul

import (
	"net/http"

	"github.com/hashicorp/consul/api"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Consul struct {
	Address    string
	Scheme     string
	Token      string
	Username   string
	Password   string
	Datacentre string

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	// client used to connect to Consul agnet
	client *api.Client
}

var sampleConfig = `
  ## Consul server address
  # address = "localhost"

  ## URI scheme for the Consul server, one of "http", "https"
  # scheme = "http"

  ## ACL token used in every request
  # token = ""

  ## HTTP Basic Authentication username and password.
  # username = ""
  # password = ""

  ## Data centre to query the health checks from
  # datacentre = ""

  ## SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## If false, skip chain & host verification
  # insecure_skip_verify = true
`

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

	if c.Token != "" {
		config.Token = c.Token
	}

	if c.Username != "" {
		config.HttpAuth = &api.HttpBasicAuth{
			Username: c.Username,
			Password: c.Password,
		}
	}

	tlsCfg, err := internal.GetTLSConfig(
		c.SSLCert, c.SSLKey, c.SSLCA, c.InsecureSkipVerify)

	if err != nil {
		return nil, err
	}

	config.HttpClient.Transport = &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	return api.NewClient(config)
}

func (c *Consul) GatherHealthCheck(acc telegraf.Accumulator, checks []*api.HealthCheck) {
	for _, check := range checks {
		record := make(map[string]interface{})
		tags := make(map[string]string)

		record["check_name"] = check.Name
		record["service_id"] = check.ServiceID

		record["status"] = check.Status
		record["passing"] = 0
		record["critical"] = 0
		record["warning"] = 0
		record[check.Status] = 1

		tags["node"] = check.Node
		tags["service_name"] = check.ServiceName
		tags["check_id"] = check.CheckID

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
