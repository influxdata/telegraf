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
	"github.com/influxdata/telegraf/plugins/outputs"
)

// AzureMonitor allows publishing of metrics to the Azure Monitor custom metrics service
type AzureMonitor struct {
	ResourceID          string `toml:"resourceId"`
	Region              string `toml:"region"`
	HTTPPostTimeout     int    `toml:"httpPostTimeout"`
	AzureSubscriptionID string `toml:"azureSubscription"`
	AzureTenantID       string `toml:"azureTenant"`
	AzureClientID       string `toml:"azureClientId"`
	AzureClientSecret   string `toml:"azureClientSecret"`

	useMsi           bool
	metadataService  *AzureInstanceMetadata
	instanceMetadata *VirtualMachineMetadata
	msiToken         *MsiToken
	msiResource      string
	bearerToken      string
	expiryWatermark  time.Duration

	oauthConfig *adal.OAuthConfig
	adalToken   adal.OAuthTokenProvider

	client *http.Client

	cache       map[uint64]azureMonitorMetric
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
resourceId = "/subscriptions/3e9c2afc-52b3-4137-9bba-02b6eb204331/resourceGroups/someresourcegroup-rg/providers/Microsoft.Compute/virtualMachines/somevmname"
## Azure region to publish metrics against.  Defaults to eastus
region = "useast"
## Maximum duration to wait for HTTP post (in seconds).  Defaults to 15
httpPostTimeout = 15
## Whether or not to use managed service identity (defaults to true).
useManagedServiceIdentity = true

## Leave this section blank to use Managed Service Identity.
## TODO
azureSubscription = "TODO"
## TODO 
azureTenant = "TODO"
## TODO
azureClientId = "TODO"
## TODO
azureClientSecret = "TODO"
`

const (
	azureMonitorDefaultRegion = "eastus"
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

	if a.useMsi == false {
		// If using direct AD authentication create the AD access client
		oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, a.AzureTenantID)
		if err != nil {
			return fmt.Errorf("Could not initialize AD client: %s", err)
		}
		a.oauthConfig = oauthConfig

	}

	if a.HTTPPostTimeout == 0 {
		a.HTTPPostTimeout = 10
	}

	a.metadataService = &AzureInstanceMetadata{}

	// For the metrics API the MSI resource has to be https://ingestion.monitor.azure.com
	a.msiResource = "https://monitoring.azure.com/"

	// Validate the resource identifier
	if a.ResourceID == "" {
		metadata, err := a.metadataService.GetInstanceMetadata()
		if err != nil {
			return fmt.Errorf("No resource id specified, and Azure Instance metadata service not available.  If not running on an Azure VM, provide a value for resourceId")
		}
		a.ResourceID = metadata.AzureResourceID

		if a.Region == "" {
			a.Region = metadata.Compute.Location
		}
	}

	if a.Region == "" {
		a.Region = azureMonitorDefaultRegion
	}

	// Validate credentials
	err := a.validateCredentials()
	if err != nil {
		return err
	}

	a.reset()
	go a.run()

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
	log.Printf("metrics collected: %+v", metrics)

	// Assemble stats on incoming metrics
	for _, metric := range metrics {
		select {
		case a.metrics <- metric:
		default:
			log.Printf("metrics buffer is full")
		}
	}

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
			msiToken, err := a.metadataService.GetMsiToken(a.AzureClientID, a.msiResource)
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

func (a *AzureMonitor) add(metric telegraf.Metric) {
	id := metric.HashID()
	if azm, ok := a.cache[id]; !ok {
		// hit an uncached metric, create caches for first time:
		var dimensionNames []string
		var dimensionValues []string
		for i, tag := range metric.TagList() {
			// Azure custom metrics service supports up to 10 dimensions
			if i > 9 {
				continue
			}
			dimensionNames = append(dimensionNames, tag.Key)
			dimensionValues = append(dimensionValues, tag.Value)
		}
		// Field keys are stored as the last dimension
		dimensionNames = append(dimensionNames, "field")

		var seriesList []*azureMonitorSeries
		// Store each field as a separate series with field key as a new dimension
		for _, field := range metric.FieldList() {
			azmseries := newAzureMonitorSeries(field, dimensionValues)
			seriesList = append(seriesList, azmseries)
		}

		if len(seriesList) < 1 {
			log.Printf("no valid fields for metric: %s", metric)
			return
		}

		a.cache[id] = azureMonitorMetric{
			Time: metric.Time(),
			Data: &azureMonitorData{
				BaseData: &azureMonitorBaseData{
					Metric:         metric.Name(),
					Namespace:      "default",
					DimensionNames: dimensionNames,
					Series:         seriesList,
				},
			},
		}
	} else {
		for _, f := range metric.FieldList() {
			fv, ok := convert(f.Value)
			if !ok {
				continue
			}

			tmp, ok := azm.findSeriesWithField(f.Key)
			if !ok {
				// hit an uncached field of a cached metric
				var dimensionValues []string
				for i, tag := range metric.TagList() {
					// Azure custom metrics service supports up to 10 dimensions
					if i > 9 {
						continue
					}
					dimensionValues = append(dimensionValues, tag.Value)
				}
				azm.Data.BaseData.Series = append(azm.Data.BaseData.Series, newAzureMonitorSeries(f, dimensionValues))
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
			a.cache[id].Data.BaseData.Series = append(a.cache[id].Data.BaseData.Series, tmp)
		}
	}
}

func (b *azureMonitorMetric) findSeriesWithField(f string) (*azureMonitorSeries, bool) {
	if len(b.Data.BaseData.Series) > 0 {
		for _, s := range b.Data.BaseData.Series {
			if f == s.DimensionValues[len(s.DimensionValues)-1] {
				return s, true
			}
		}
	}
	return nil, false
}

func newAzureMonitorSeries(f *telegraf.Field, dv []string) *azureMonitorSeries {
	fv, ok := convert(f.Value)
	if !ok {
		log.Printf("unable to convert field %s (type %T) to float type: %v", f.Key, fv, fv)
		return nil
	}
	return &azureMonitorSeries{
		DimensionValues: append(append([]string{}, dv...), f.Key),
		Min:             fv,
		Max:             fv,
		Sum:             fv,
		Count:           1,
	}
}

func (a *AzureMonitor) reset() {
	a.cache = make(map[uint64]azureMonitorMetric)
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Printf("converted string: %s to %v", v, f)
			return 0, false
		}
		return f, true
	default:
		log.Printf("did not convert %T: %s", v, v)
		return 0, false
	}
}

func (a *AzureMonitor) push() {
	var body []byte
	for _, metric := range a.cache {
		jsonBytes, err := json.Marshal(&metric)
		log.Printf("marshalled point %s", jsonBytes)
		if err != nil {
			log.Printf("Error marshalling metrics %s", err)
			return
		}
		body = append(body, jsonBytes...)
		body = append(body, '\n')
	}

	log.Printf("Publishing metrics %s", body)
	_, err := a.postData(&body)
	if err != nil {
		log.Printf("Error publishing metrics %s", err)
		return
	}

	return
}

func (a *AzureMonitor) postData(msg *[]byte) (*http.Request, error) {
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
		// TODO - fix this
		//Timeout: time.Duration(s.HTTPPostTimeout * time.Second),
		Timeout: time.Duration(10 * time.Second),
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

func init() {
	outputs.Add("azuremonitor", func() telegraf.Output {
		return &AzureMonitor{
			period:   time.Minute,
			delay:    time.Second * 5,
			metrics:  make(chan telegraf.Metric, 100),
			shutdown: make(chan struct{}),
		}
	})
}
