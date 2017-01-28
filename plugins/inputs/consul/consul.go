package consul

import (
	"net/http"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/consul/structs"
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

	CollectServiceHealth bool `toml:"service_health"`

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
  ## Most of these values defaults to the one configured on a Consul's agent level.
  ## Optional Consul server address (default: "localhost")
  # address = "localhost"
  ## Optional URI scheme for the Consul server (default: "http")
  # scheme = "http"
  ## Optional ACL token used in every request (default: "")
  # token = ""
  ## Optional username used for request HTTP Basic Authentication (default: "")
  # username = ""
  ## Optional password used for HTTP Basic Authentication (default: "")
  # password = ""
  ## Optional data centre to query the health checks from (default: "")
  # datacentre = ""
  ## Optional should we gather service health checks (default: false)
  # service_health = false
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

		record["check_id"] = check.CheckID
		record["check_name"] = check.Name
		record["service_id"] = check.ServiceID
		record["status"] = check.Status

		tags["node"] = check.Node
		tags["service_name"] = check.ServiceName

		acc.AddFields("consul_health_checks", record, tags)
	}
}

func (c *Consul) GatherServiceHealth(services map[string][]string, acc telegraf.Accumulator) error {
	for s := range services {
		service, _, err := c.client.Health().Service(s, "", false, &api.QueryOptions{})

		if err != nil {
			return err
		}

		for _, elem := range service {
			if len(elem.Checks) == 0 {
				continue
			}

			record := make(map[string]interface{})
			tags := make(map[string]string)

			record["healthy"] = 1.0
			tags["service_name"] = s

			for _, check := range elem.Checks {
				if len(tags["node"]) == 0 {
					tags["node"] = check.Node
				}
				if check.Status != structs.HealthPassing {
					record["healthy"] = 0.0
					break
				}
			}

			acc.AddFields("consul_service_health", record, tags)
		}
	}

	return nil
}

func (c *Consul) GatherServerStats(acc telegraf.Accumulator) error {
	peers, err := c.client.Status().Peers()

	if err != nil {
		return err
	}

	record := make(map[string]interface{})

	record["peers"] = float64(len(peers))

	leader, err := c.client.Status().Leader()

	if err != nil {
		return err
	}

	if len(leader) == 0 {
		record["leader"] = 1.0
	} else {
		record["leader"] = 0.0
	}

	nodes, _, err := c.client.Catalog().Nodes(&api.QueryOptions{})

	if err != nil {
		return err
	}

	record["nodes"] = float64(len(nodes))

	services, _, err := c.client.Catalog().Services(&api.QueryOptions{})

	if err != nil {
		return err
	}

	record["services"] = float64(len(services))

	acc.AddFields("consul_server_stats", record, map[string]string{})

	if c.CollectServiceHealth {
		err = c.GatherServiceHealth(services, acc)

		if err != nil {
			return err
		}
	}

	return nil
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

	err = c.GatherServerStats(acc)

	if err != nil {
		return err
	}

	return nil
}

func init() {
	inputs.Add("consul", func() telegraf.Input {
		return &Consul{}
	})
}
