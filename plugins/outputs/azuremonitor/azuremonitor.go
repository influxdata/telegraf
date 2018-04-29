package azuremonitor

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// AzureMonitor allows publishing of metrics to the Azure Monitor custom metrics service
type AzureMonitor struct {
	ResourceID          string            `toml:"resource_id"`
	Region              string            `toml:"region"`
	Timeout             internal.Duration `toml:"Timeout"`
	AzureSubscriptionID string            `toml:"azure_subscription"`
	AzureTenantID       string            `toml:"azure_tenant"`
	AzureClientID       string            `toml:"azure_client_id"`
	AzureClientSecret   string            `toml:"azure_client_secret"`
	StringAsDimension   bool              `toml:"string_as_dimension"`

	useMsi           bool `toml:"use_managed_service_identity"`
	metadataService  *AzureInstanceMetadata
	instanceMetadata *VirtualMachineMetadata
	msiToken         *msiToken
	msiResource      string
	bearerToken      string
	expiryWatermark  time.Duration

	oauthConfig *adal.OAuthConfig
	adalToken   adal.OAuthTokenProvider

	client *http.Client

	cache       map[string]*azureMonitorMetric
	period      time.Duration
	delay       time.Duration
	periodStart time.Time
	periodEnd   time.Time

	metrics  chan telegraf.Metric
	shutdown chan struct{}
}

type azureMonitorMetric struct {
	Time time.Time         `json:"time"`
	Data *azureMonitorData `json:"data"`
}

type azureMonitorData struct {
	BaseData *azureMonitorBaseData `json:"baseData"`
}

type azureMonitorBaseData struct {
	Metric         string                `json:"metric"`
	Namespace      string                `json:"namespace"`
	DimensionNames []string              `json:"dimNames"`
	Series         []*azureMonitorSeries `json:"series"`
}

type azureMonitorSeries struct {
	DimensionValues []string `json:"dimValues"`
	Min             float64  `json:"min"`
	Max             float64  `json:"max"`
	Sum             float64  `json:"sum"`
	Count           float64  `json:"count"`
}

var sampleConfig = `
  ## The resource ID against which metric will be logged.  If not
  ## specified, the plugin will attempt to retrieve the resource ID
  ## of the VM via the instance metadata service (optional if running 
  ## on an Azure VM with MSI)
  #resource_id = "/subscriptions/<subscription-id>/resourceGroups/<resource-group>/providers/Microsoft.Compute/virtualMachines/<vm-name>"
  ## Azure region to publish metrics against.  Defaults to eastus.
  ## Leave blank to automatically query the region via MSI.
  #region = "useast"

  ## Write HTTP timeout, formatted as a string.  If not provided, will default
  ## to 5s. 0s means no timeout (not recommended).
  # timeout = "5s"

  ## Whether or not to use managed service identity.
  #use_managed_service_identity = true

  ## Fill in the following values if using Active Directory Service
  ## Principal or User Principal for authentication.
  ## Subscription ID
  #azure_subscription = ""
  ## Tenant ID
  #azure_tenant = ""
  ## Client ID
  #azure_client_id = ""
  ## Client secrete
  #azure_client_secret = ""
`

const (
	defaultRegion = "eastus"

	defaultMSIResource = "https://monitoring.azure.com/"
)

// Connect initializes the plugin and validates connectivity
func (a *AzureMonitor) Connect() error {
	// Set defaults

	// If no direct AD values provided, fall back to MSI
	if a.AzureSubscriptionID == "" && a.AzureTenantID == "" && a.AzureClientID == "" && a.AzureClientSecret == "" {
		a.useMsi = true
	} else if a.AzureSubscriptionID == "" || a.AzureTenantID == "" || a.AzureClientID == "" || a.AzureClientSecret == "" {
		return fmt.Errorf("Must provide values for azureSubscription, azureTenant, azureClient and azureClientSecret, or leave all blank to default to MSI")
	}

	if !a.useMsi {
		// If using direct AD authentication create the AD access client
		oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, a.AzureTenantID)
		if err != nil {
			return fmt.Errorf("Could not initialize AD client: %s", err)
		}
		a.oauthConfig = oauthConfig
	}

	a.metadataService = &AzureInstanceMetadata{}

	// For the metrics API the MSI resource has to be https://ingestion.monitor.azure.com
	a.msiResource = "https://monitoring.azure.com/"

	// Validate the resource identifier
	metadata, err := a.metadataService.GetInstanceMetadata()
	if err != nil {
		return fmt.Errorf("No resource id specified, and Azure Instance metadata service not available.  If not running on an Azure VM, provide a value for resourceId")
	}
	a.ResourceID = metadata.AzureResourceID

	if a.Region == "" {
		a.Region = metadata.Compute.Location
	}

	// Validate credentials
	err = a.validateCredentials()
	if err != nil {
		return err
	}

	a.reset()
	go a.run()

	return nil
}

func (a *AzureMonitor) validateCredentials() error {
	// Use managed service identity
	if a.useMsi {
		// Check expiry on the token
		if a.msiToken != nil {
			expiryDuration := a.msiToken.ExpiresInDuration()
			if expiryDuration > a.expiryWatermark {
				return nil
			}

			// Token is about to expire
			log.Printf("Bearer token expiring in %s; acquiring new token\n", expiryDuration.String())
			a.msiToken = nil
		}

		// No token, acquire an MSI token
		if a.msiToken == nil {
			msiToken, err := a.metadataService.getMsiToken(a.AzureClientID, a.msiResource)
			if err != nil {
				return err
			}
			log.Printf("Bearer token acquired; expiring in %s\n", msiToken.ExpiresInDuration().String())
			a.msiToken = msiToken
			a.bearerToken = msiToken.AccessToken
		}
		// Otherwise directory acquire a token
	} else {
		adToken, err := adal.NewServicePrincipalToken(
			*(a.oauthConfig), a.AzureClientID, a.AzureClientSecret,
			azure.PublicCloud.ActiveDirectoryEndpoint)
		if err != nil {
			return fmt.Errorf("Could not acquire ADAL token: %s", err)
		}
		a.adalToken = adToken
	}

	return nil
}

// Description provides a description of the plugin
func (a *AzureMonitor) Description() string {
	return "Configuration for sending aggregate metrics to Azure Monitor"
}

// SampleConfig provides a sample configuration for the plugin
func (a *AzureMonitor) SampleConfig() string {
	return sampleConfig
}

// Close shuts down an any active connections
func (a *AzureMonitor) Close() error {
	// Close connection to the URL here
	close(a.shutdown)
	return nil
}

// Write writes metrics to the remote endpoint
func (a *AzureMonitor) Write(metrics []telegraf.Metric) error {
	// Assemble basic stats on incoming metrics
	for _, metric := range metrics {
		select {
		case a.metrics <- metric:
		default:
			log.Printf("metrics buffer is full")
		}
	}

	return nil
}

func (a *AzureMonitor) run() {
	// The start of the period is truncated to the nearest minute.
	//
	// Every metric then gets it's timestamp checked and is dropped if it
	// is not within:
	//
	//   start < t < end + truncation + delay
	//
	// So if we start at now = 00:00.2 with a 10s period and 0.3s delay:
	//   now = 00:00.2
	//   start = 00:00
	//   truncation = 00:00.2
	//   end = 00:10
	// 1st interval: 00:00 - 00:10.5
	// 2nd interval: 00:10 - 00:20.5
	// etc.
	//
	now := time.Now()
	a.periodStart = now.Truncate(time.Minute)
	truncation := now.Sub(a.periodStart)
	a.periodEnd = a.periodStart.Add(a.period)
	time.Sleep(a.delay)
	periodT := time.NewTicker(a.period)
	defer periodT.Stop()

	for {
		select {
		case <-a.shutdown:
			if len(a.metrics) > 0 {
				// wait until metrics are flushed before exiting
				continue
			}
			return
		case m := <-a.metrics:
			if m.Time().Before(a.periodStart) ||
				m.Time().After(a.periodEnd.Add(truncation).Add(a.delay)) {
				// the metric is outside the current aggregation period, so
				// skip it.
				continue
			}
			a.add(m)
		case <-periodT.C:
			a.periodStart = a.periodEnd
			a.periodEnd = a.periodStart.Add(a.period)
			a.push()
			a.reset()
		}
	}
}

func (a *AzureMonitor) reset() {
	a.cache = make(map[string]*azureMonitorMetric)
}

func (a *AzureMonitor) add(metric telegraf.Metric) {
	var dimensionNames []string
	var dimensionValues []string
	for i, tag := range metric.TagList() {
		// Azure custom metrics service supports up to 10 dimensions
		if i > 10 {
			continue
		}
		dimensionNames = append(dimensionNames, tag.Key)
		dimensionValues = append(dimensionValues, tag.Value)
	}

	// Azure Monitoe does not support string value types, so convert string
	// fields to dimensions if enabled.
	if a.StringAsDimension {
		for _, f := range metric.FieldList() {
			switch fv := f.Value.(type) {
			case string:
				dimensionNames = append(dimensionNames, f.Key)
				dimensionValues = append(dimensionValues, fv)
				metric.RemoveField(f.Key)
			}
		}
	}

	for _, f := range metric.FieldList() {
		name := metric.Name() + "_" + f.Key
		fv, ok := convert(f.Value)
		if !ok {
			log.Printf("unable to convert field %s (type %T) to float type: %v", f.Key, fv, fv)
			continue
		}

		if azm, ok := a.cache[name]; !ok {
			// hit an uncached metric, create it for first time
			a.cache[name] = &azureMonitorMetric{
				Time: metric.Time(),
				Data: &azureMonitorData{
					BaseData: &azureMonitorBaseData{
						Metric:         name,
						Namespace:      "default",
						DimensionNames: dimensionNames,
						Series: []*azureMonitorSeries{
							newAzureMonitorSeries(dimensionValues, fv),
						},
					},
				},
			}
		} else {
			tmp, i, ok := azm.findSeries(dimensionValues)
			if !ok {
				// add series new series (should be rare)
				n := append(azm.Data.BaseData.Series, newAzureMonitorSeries(dimensionValues, fv))
				a.cache[name].Data.BaseData.Series = n
				continue
			}

			//counter compute
			n := tmp.Count + 1
			tmp.Count = n
			//max/min compute
			if fv < tmp.Min {
				tmp.Min = fv
			} else if fv > tmp.Max {
				tmp.Max = fv
			}
			//sum compute
			tmp.Sum += fv
			//store final data
			a.cache[name].Data.BaseData.Series[i] = tmp
		}
	}
}

func (m *azureMonitorMetric) findSeries(dv []string) (*azureMonitorSeries, int, bool) {
	if len(m.Data.BaseData.DimensionNames) != len(dv) {
		return nil, 0, false
	}
	for i := range m.Data.BaseData.Series {
		if m.Data.BaseData.Series[i].equal(dv) {
			return m.Data.BaseData.Series[i], i, true
		}
	}
	return nil, 0, false
}

func newAzureMonitorSeries(dv []string, fv float64) *azureMonitorSeries {
	return &azureMonitorSeries{
		DimensionValues: append([]string{}, dv...),
		Min:             fv,
		Max:             fv,
		Sum:             fv,
		Count:           1,
	}
}

func (s *azureMonitorSeries) equal(dv []string) bool {
	if len(s.DimensionValues) != len(dv) {
		return false
	}
	for i := range dv {
		if dv[i] != s.DimensionValues[i] {
			return false
		}
	}
	return true
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case int64:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float64:
		return v, true
	case bool:
		if v {
			return 1, true
		}
		return 1, true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func (a *AzureMonitor) push() {
	var body []byte
	for _, metric := range a.cache {
		jsonBytes, err := json.Marshal(&metric)
		if err != nil {
			log.Printf("Error marshalling metrics %s", err)
			return
		}
		body = append(body, jsonBytes...)
		body = append(body, '\n')
	}

	_, err := a.postData(&body)
	if err != nil {
		log.Printf("Error publishing aggregate metrics %s", err)
	}
	return
}

func (a *AzureMonitor) postData(msg *[]byte) (*http.Request, error) {
	if err := a.validateCredentials(); err != nil {
		return nil, fmt.Errorf("Error authenticating: %v", err)
	}

	metricsEndpoint := fmt.Sprintf("https://%s.monitoring.azure.com%s/metrics",
		a.Region, a.ResourceID)

	req, err := http.NewRequest("POST", metricsEndpoint, bytes.NewBuffer(*msg))
	if err != nil {
		log.Printf("Error creating HTTP request")
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.bearerToken)
	req.Header.Set("Content-Type", "application/x-ndjson")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: tr,
		Timeout:   a.Timeout.Duration,
	}
	resp, err := client.Do(req)
	if err != nil {
		return req, err
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		var reply []byte
		reply, err = ioutil.ReadAll(resp.Body)

		if err != nil {
			reply = nil
		}
		return req, fmt.Errorf("Post Error. HTTP response code:%d message:%s reply:\n%s",
			resp.StatusCode, resp.Status, reply)
	}
	return req, nil
}

func init() {
	outputs.Add("azuremonitor", func() telegraf.Output {
		return &AzureMonitor{
			StringAsDimension: true,
			Timeout:           internal.Duration{Duration: time.Second * 5},
			Region:            defaultRegion,
			period:            time.Minute,
			delay:             time.Second * 5,
			metrics:           make(chan telegraf.Metric, 100),
			shutdown:          make(chan struct{}),
		}
	})
}
