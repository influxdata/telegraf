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
	"github.com/davecgh/go-spew/spew"
	"github.com/influxdata/telegraf"
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

// Description provides a description of the plugin
func (s *AzureMonitor) Description() string {
	return "Configuration for Azure Monitor to send metrics to"
}

// SampleConfig provides a sample configuration for the plugin
func (s *AzureMonitor) SampleConfig() string {
	return sampleConfig
}

// Connect initializes the plugin and validates connectivity
func (s *AzureMonitor) Connect() error {
	// Set defaults

	// If no direct AD values provided, fall back to MSI
	if s.AzureSubscriptionID == "" && s.AzureTenantID == "" && s.AzureClientID == "" && s.AzureClientSecret == "" {
		s.useMsi = true
	} else if s.AzureSubscriptionID == "" || s.AzureTenantID == "" || s.AzureClientID == "" || s.AzureClientSecret == "" {
		return fmt.Errorf("Must provide values for azureSubscription, azureTenant, azureClient and azureClientSecret, or leave all blank to default to MSI")
	}

	if s.useMsi == false {
		// If using direct AD authentication create the AD access client
		oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, s.AzureTenantID)
		if err != nil {
			return fmt.Errorf("Could not initialize AD client: %s", err)
		}
		s.oauthConfig = oauthConfig

	}

	if s.HTTPPostTimeout == 0 {
		s.HTTPPostTimeout = 10
	}

	s.metadataService = &AzureInstanceMetadata{}

	// For the metrics API the MSI resource has to be https://ingestion.monitor.azure.com
	s.msiResource = "https://ingestion.monitor.azure.com/"

	// Validate the resource identifier
	if s.ResourceID == "" {
		metadata, err := s.metadataService.GetInstanceMetadata()
		if err != nil {
			return fmt.Errorf("No resource id specified, and Azure Instance metadata service not available.  If not running on an Azure VM, provide a value for resourceId")
		}
		s.ResourceID = metadata.AzureResourceID

		if s.Region == "" {
			s.Region = metadata.Compute.Location
		}
	}

	if s.Region == "" {
		s.Region = azureMonitorDefaultRegion
	}

	// Validate credentials
	err := s.validateCredentials()
	if err != nil {
		return err
	}

	return nil
}

// Close shuts down an any active connections
func (s *AzureMonitor) Close() error {
	// Close connection to the URL here
	return nil
}

// Write writes metrics to the remote endpoint
func (s *AzureMonitor) Write(metrics []telegraf.Metric) error {
	// Flatten metrics into an Azure Monitor common schema compatible format
	metricsList, err := s.flattenMetrics(metrics)
	if err != nil {
		log.Printf("Error translating metrics %s", err)
		return err
	}

	for _, v := range metricsList {
		jsonBytes, err := json.Marshal(&v)
		_, err = s.postData(&jsonBytes)
		if err != nil {
			log.Printf("Error publishing metrics %s", err)
			return err
		}
	}

	return nil
}

func (s *AzureMonitor) validateCredentials() error {
	// Use managed service identity
	if s.useMsi {
		// Check expiry on the token
		if s.msiToken != nil {
			expiryDuration := s.msiToken.ExpiresInDuration()
			if expiryDuration > s.expiryWatermark {
				return nil
			}

			// Token is about to expire
			log.Printf("Bearer token expiring in %s; acquiring new token\n", expiryDuration.String())
			s.msiToken = nil
		}

		// No token, acquire an MSI token
		if s.msiToken == nil {
			msiToken, err := s.metadataService.GetMsiToken(s.AzureClientID, s.msiResource)
			if err != nil {
				return err
			}
			log.Printf("Bearer token acquired; expiring in %s\n", msiToken.ExpiresInDuration().String())
			s.msiToken = msiToken
			s.bearerToken = msiToken.AccessToken
		}
		// Otherwise directory acquire a token
	} else {
		adToken, err := adal.NewServicePrincipalToken(
			*(s.oauthConfig), s.AzureClientID, s.AzureClientSecret,
			azure.PublicCloud.ActiveDirectoryEndpoint)
		if err != nil {
			return fmt.Errorf("Could not acquire ADAL token: %s", err)
		}
		s.adalToken = adToken
	}

	return nil
}

type azureMonitorMetric struct {
	Time time.Time        `json:"time"`
	Data azureMonitorData `json:"data"`
}

type azureMonitorData struct {
	BaseData azureMonitorBaseData `json:"baseData"`
}

type azureMonitorBaseData struct {
	Metric         string               `json:"metric"`
	Namespace      string               `json:"namespace"`
	DimensionNames []string             `json:"dimNames"`
	Series         []azureMonitorSeries `json:"series"`
}

type azureMonitorSeries struct {
	DimensionValues []string `json:"dimValues"`
	Min             int64    `json:"min"`
	Max             int64    `json:"max"`
	Sum             int64    `json:"sum"`
	Count           int64    `json:"count"`
}

func (s *AzureMonitor) flattenMetrics(metrics []telegraf.Metric) ([]azureMonitorMetric, error) {
	var azureMetrics []azureMonitorMetric
	for _, metric := range metrics {

		// Get the list of custom dimensions (elevated tags and fields)
		var dimensionNames []string
		var dimensionValues []string
		for name, value := range metric.Fields() {
			dimensionNames = append(dimensionNames, name)
			dimensionValues = append(dimensionValues, s.formatField(value))
		}

		series := azureMonitorSeries{
			DimensionValues: dimensionValues,
		}

		if v, ok := metric.Fields()["min"]; ok {
			series.Min = s.formatInt(v)
		}

		if v, ok := metric.Fields()["max"]; ok {
			series.Max = s.formatInt(v)
		}

		if v, ok := metric.Fields()["sum"]; ok {
			series.Sum = s.formatInt(v)
		}

		if v, ok := metric.Fields()["count"]; ok {
			series.Count = s.formatInt(v)
		} else {
			// Azure Monitor requires count >= 1
			series.Count = 1
		}

		azureMetric := azureMonitorMetric{
			Time: metric.Time(),
			Data: azureMonitorData{
				BaseData: azureMonitorBaseData{
					Metric:         metric.Name(),
					Namespace:      "default",
					DimensionNames: dimensionNames,
					Series:         []azureMonitorSeries{series},
				},
			},
		}

		azureMetrics = append(azureMetrics, azureMetric)
	}
	return azureMetrics, nil
}

func (s *AzureMonitor) formatInt(value interface{}) int64 {
	return 0
}

func (s *AzureMonitor) formatField(value interface{}) string {
	var ret string

	switch v := value.(type) {
	case int:
		ret = strconv.FormatInt(int64(value.(int)), 10)
	case int8:
		ret = strconv.FormatInt(int64(value.(int8)), 10)
	case int16:
		ret = strconv.FormatInt(int64(value.(int16)), 10)
	case int32:
		ret = strconv.FormatInt(int64(value.(int32)), 10)
	case int64:
		ret = strconv.FormatInt(value.(int64), 10)
	case float32:
		ret = strconv.FormatFloat(float64(value.(float32)), 'f', -1, 64)
	case float64:
		ret = strconv.FormatFloat(value.(float64), 'f', -1, 64)
	default:
		spew.Printf("field is of unsupported value type %v\n", v)
	}
	return ret
}

func (s *AzureMonitor) postData(msg *[]byte) (*http.Request, error) {
	metricsEndpoint := fmt.Sprintf("https://%s.monitoring.azure.com%s/metrics",
		s.Region, s.ResourceID)

	req, err := http.NewRequest("POST", metricsEndpoint, bytes.NewBuffer(*msg))
	if err != nil {
		log.Printf("Error creating HTTP request")
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.bearerToken)
	req.Header.Set("Content-Type", "application/json")

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
