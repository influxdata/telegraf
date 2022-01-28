package dnac

// dnac.go

import (
	"net/url"

	dnac_sdk "github.com/cisco-en-programmability/dnacenter-go-sdk/v3/sdk"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Dnac struct {
	DnacBaseURL   string
	Username      string
	Password      string
	SSLVerify     string
	Debug         string
	ClientHealth  bool
	NetworkHealth bool
	Log           telegraf.Logger
	Client        *dnac_sdk.Client
}

func (d *Dnac) Description() string {
	return "a Cisco DNAC plugin"
}

func (d *Dnac) SampleConfig() string {
	return `
  ## Specify DNAC Base URL
  dnacbaseurl = "https://sandboxdnac.cisco.com"
  ## Specify Credentials
  username = "devnetuser"
  password = "Cisco123!"
  ## Debug true/false
  debug = "false"
  ## SSL Verify true/false
  sslverify = "false"
  ## Report Client Health
  clienthealth = true
  ## Report Network Health
  networkhealth = true
`
}

// Init is for setup, and validating config.
func (d *Dnac) Init() error {
	return nil
}

func (d *Dnac) InitClient() error {
	var Client, err = dnac_sdk.NewClientWithOptionsNoAuth(d.DnacBaseURL, d.Username, d.Password, d.Debug, d.SSLVerify)
	if err != nil {
		d.Log.Errorf("Connection or login to DNAC failed")
		return err
	}
	d.Client = Client
	return nil
}

func NewDnac() *Dnac {
	return &Dnac{}
}

func (d *Dnac) Gather(acc telegraf.Accumulator) error {

	var err error

	var dnacUrl, url_err = url.Parse(d.DnacBaseURL)
	if url_err != nil {
		d.Log.Errorf("Invalid DNAC Base URL Provided")
		return url_err
	}

	if d.Client == nil {
		err = d.InitClient()
		if err != nil {
			return err
		}
	}

	err = d.Client.AuthClient()
	if err != nil {
		return err
	}

	if d.ClientHealth {

		getOverallClientHealthQueryParams := &dnac_sdk.GetOverallClientHealthQueryParams{}
		client_health, _, err := d.Client.Clients.GetOverallClientHealth(getOverallClientHealthQueryParams)
		if err != nil {
			d.Log.Errorf("Client health request failed")
			return err
		}

		for _, response := range *client_health.Response {
			tags := map[string]string{
				"host":    dnacUrl.Host,
				"site_id": response.SiteID,
			}
			fields := make(map[string]interface{})

			for _, clientType := range *response.ScoreDetail {
				l1_prefix := clientType.ScoreCategory.ScoreCategory + "_" + clientType.ScoreCategory.Value
				fields[internal.SnakeCase(l1_prefix+"_client_count")] = clientType.ClientCount
				fields[internal.SnakeCase(l1_prefix+"_score_value")] = clientType.ScoreValue
				if clientType.ScoreList != nil {
					for _, scoreType := range *clientType.ScoreList {
						l2_prefix := scoreType.ScoreCategory.ScoreCategory + "_" + scoreType.ScoreCategory.Value
						fields[internal.SnakeCase(l1_prefix+"_"+l2_prefix+"_client_count")] = scoreType.ClientCount
						if scoreType.ScoreList != nil {
							for _, rootCause := range *scoreType.ScoreList {
								if rootCause.ScoreCategory.ScoreCategory == "rootCause" {
									l3_prefix := rootCause.ScoreCategory.ScoreCategory + "_" + rootCause.ScoreCategory.Value
									fields[internal.SnakeCase(l1_prefix+"_"+l2_prefix+"_"+l3_prefix+"_client_count")] = rootCause.ClientCount
								}
							}
						}
					}
				}
			}
			acc.AddFields("dnac_client_health", fields, tags)
		}

	}

	if d.NetworkHealth {
		getOverallNetworkHeathQueryParams := &dnac_sdk.GetOverallNetworkHealthQueryParams{}

		network_health, _, err := d.Client.Topology.GetOverallNetworkHealth(getOverallNetworkHeathQueryParams)

		if err != nil {
			d.Log.Errorf("Network health request failed")
			return err
		}
		network_health_tags := map[string]string{
			"host":    dnacUrl.Host,
			"site_id": network_health.MeasuredBy,
		}
		network_health_fields := make(map[string]interface{})

		for _, response := range *network_health.Response {

			network_health_fields[internal.SnakeCase("overall_health_score")] = response.HealthScore
			network_health_fields[internal.SnakeCase("overall_total_count")] = response.TotalCount
			network_health_fields[internal.SnakeCase("overall_no_health_count")] = response.UnmonCount
			network_health_fields[internal.SnakeCase("overall_good_count")] = response.GoodCount
			network_health_fields[internal.SnakeCase("overall_fair_count")] = response.FairCount
			network_health_fields[internal.SnakeCase("overall_bad_count")] = response.BadCount
		}

		for _, health_dist := range *network_health.HealthDistirubution {
			network_health_fields[internal.SnakeCase(health_dist.Category+"_health_score")] = health_dist.HealthScore
			network_health_fields[internal.SnakeCase(health_dist.Category+"_total_count")] = health_dist.TotalCount
			network_health_fields[internal.SnakeCase(health_dist.Category+"_bad_count")] = health_dist.BadCount
			network_health_fields[internal.SnakeCase(health_dist.Category+"_bad_percentage")] = health_dist.BadPercentage
			network_health_fields[internal.SnakeCase(health_dist.Category+"_fair_count")] = health_dist.FairCount
			network_health_fields[internal.SnakeCase(health_dist.Category+"_fair_percentage")] = health_dist.FairPercentage
			network_health_fields[internal.SnakeCase(health_dist.Category+"_good_count")] = health_dist.GoodCount
			network_health_fields[internal.SnakeCase(health_dist.Category+"_good_percentage")] = health_dist.GoodPercentage
			network_health_fields[internal.SnakeCase(health_dist.Category+"_no_health_count")] = health_dist.UnmonCount
			network_health_fields[internal.SnakeCase(health_dist.Category+"_no_health_percentage")] = health_dist.UnmonPercentage
		}

		acc.AddFields("dnac_network_health", network_health_fields, network_health_tags)
	}

	return nil
}

func init() {
	inputs.Add("dnac", func() telegraf.Input { return &Dnac{} })
}
