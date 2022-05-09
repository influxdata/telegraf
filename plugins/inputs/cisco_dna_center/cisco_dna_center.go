package cisco_dna_center


import (
	"net/url"

	dnac_sdk "github.com/cisco-en-programmability/dnacenter-go-sdk/v3/sdk"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Dnac struct {
	Username    string
	Password    string
	SSLVerify   string
	Report      []string
	Log         telegraf.Logger
	Client      *dnac_sdk.Client
	DnacBaseURL string
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
  ## SSL Verify true/false
  sslverify = "false"
  ## Health types to report
  report = ["client","network"]
`
}

// Init is for setup, and validating config.
func (d *Dnac) Init() error {
	err := d.InitClient()
	if err != nil {
		return err
	}
	err = d.Client.AuthClient()
	if err != nil {
		return err
	}
	return nil
}

func (d *Dnac) InitClient() error {
	var debug string
	if telegraf.Debug {
		debug = "true"
	} else {
		debug = "false"
	}

	var client, err = dnac_sdk.NewClientWithOptionsNoAuth(d.DnacBaseURL, d.Username, d.Password, debug, d.SSLVerify)
	if err != nil {
		d.Log.Errorf("Connection or login to DNAC failed")
		return err
	}
	d.Client = client
	return nil
}

func NewDnac() *Dnac {
	return &Dnac{}
}

func (d *Dnac) Gather(acc telegraf.Accumulator) error {
	var err error

	var dnacURL, urlErr = url.Parse(d.DnacBaseURL)
	if urlErr != nil {
		d.Log.Errorf("Invalid DNAC Base URL Provided")
		return urlErr
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

	for healthType := range d.Report {
		if d.Report[healthType] == "client" {
			getOverallClientHealthQueryParams := &dnac_sdk.GetOverallClientHealthQueryParams{}
			clientHealth, _, err := d.Client.Clients.GetOverallClientHealth(getOverallClientHealthQueryParams)
			if err != nil {
				d.Log.Errorf("Client health request failed")
				return err
			}
			for _, response := range *clientHealth.Response {
				tags := map[string]string{
					"host":    dnacURL.Host,
					"site_id": response.SiteID,
				}
				fields := make(map[string]interface{})
				for _, clientType := range *response.ScoreDetail {
					l1Prefix := clientType.ScoreCategory.ScoreCategory + "_" + clientType.ScoreCategory.Value
					fields[internal.SnakeCase(l1Prefix+"_client_count")] = clientType.ClientCount
					fields[internal.SnakeCase(l1Prefix+"_score_value")] = clientType.ScoreValue
					if clientType.ScoreList == nil {
						continue
					}
					for _, scoreType := range *clientType.ScoreList {
						l2Prefix := scoreType.ScoreCategory.ScoreCategory + "_" + scoreType.ScoreCategory.Value
						fields[internal.SnakeCase(l1Prefix+"_"+l2Prefix+"_client_count")] = scoreType.ClientCount
						if scoreType.ScoreList == nil {
							continue
						}
						for _, rootCause := range *scoreType.ScoreList {
							if rootCause.ScoreCategory.ScoreCategory == "rootCause" {
								l3Prefix := rootCause.ScoreCategory.ScoreCategory + "_" + rootCause.ScoreCategory.Value
								fields[internal.SnakeCase(l1Prefix+"_"+l2Prefix+"_"+l3Prefix+"_client_count")] = rootCause.ClientCount
							}
						}
					}
				}
				acc.AddFields("dnac_client_health", fields, tags)
			}
		} else if d.Report[healthType] == "network" {
			getOverallNetworkHeathQueryParams := &dnac_sdk.GetOverallNetworkHealthQueryParams{}

			networkHealth, _, err := d.Client.Topology.GetOverallNetworkHealth(getOverallNetworkHeathQueryParams)

			if err != nil {
				d.Log.Errorf("Network health request failed")
				return err
			}
			networkHealthTags := map[string]string{
				"host":    dnacURL.Host,
				"site_id": networkHealth.MeasuredBy,
			}
			networkHealthFields := make(map[string]interface{})

			for _, response := range *networkHealth.Response {
				networkHealthFields[internal.SnakeCase("overall_health_score")] = response.HealthScore
				networkHealthFields[internal.SnakeCase("overall_total_count")] = response.TotalCount
				networkHealthFields[internal.SnakeCase("overall_no_health_count")] = response.UnmonCount
				networkHealthFields[internal.SnakeCase("overall_good_count")] = response.GoodCount
				networkHealthFields[internal.SnakeCase("overall_fair_count")] = response.FairCount
				networkHealthFields[internal.SnakeCase("overall_bad_count")] = response.BadCount
			}

			for _, healthDist := range *networkHealth.HealthDistirubution {
				networkHealthFields[internal.SnakeCase(healthDist.Category+"_health_score")] = healthDist.HealthScore
				networkHealthFields[internal.SnakeCase(healthDist.Category+"_total_count")] = healthDist.TotalCount
				networkHealthFields[internal.SnakeCase(healthDist.Category+"_bad_count")] = healthDist.BadCount
				networkHealthFields[internal.SnakeCase(healthDist.Category+"_bad_percentage")] = healthDist.BadPercentage
				networkHealthFields[internal.SnakeCase(healthDist.Category+"_fair_count")] = healthDist.FairCount
				networkHealthFields[internal.SnakeCase(healthDist.Category+"_fair_percentage")] = healthDist.FairPercentage
				networkHealthFields[internal.SnakeCase(healthDist.Category+"_good_count")] = healthDist.GoodCount
				networkHealthFields[internal.SnakeCase(healthDist.Category+"_good_percentage")] = healthDist.GoodPercentage
				networkHealthFields[internal.SnakeCase(healthDist.Category+"_no_health_count")] = healthDist.UnmonCount
				networkHealthFields[internal.SnakeCase(healthDist.Category+"_no_health_percentage")] = healthDist.UnmonPercentage
			}
			acc.AddFields("dnac_network_health", networkHealthFields, networkHealthTags)
		}
	}
	return nil
}

func init() {
	inputs.Add("dnac", func() telegraf.Input { return &Dnac{} })
}
