package appdynamics_ma

// appdynamics_ma.go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

/*************************************************************************
* Define AppD structs
*************************************************************************/
type AppDynamicsMA struct {
	Host       string `toml:"host"`
	Port       string `toml:"port"`
	MetricPath string `toml:"metricPath"`
}

// JSON structure for the AppD metric
type AppDynamicsJson struct {
	MetricName     string `json:"metricName"`
	AggregatorType string `json:"aggregatorType"`
	Value          int    `json:"value"`
}

// Slice of type AppDynamicsJson to be able to collect multiple records
type AppDynamicsSlice struct {
	Appdynamics_MA []AppDynamicsJson
}

/*************************************************************************
* Define AppD constants for default values
*************************************************************************/
// Defaults assume locally running AppDynamics Machine Agent using default
// port.  Also assumes full Server Visibility license for metric path.
const (
	defaultHost       = "http://127.0.0.1"
	defaultPort       = "8293"
	defaultMetricPath = "Custom Metrics|Telegraf|"
)

/*************************************************************************
* Define AppD functions
*************************************************************************/
/**************************************
* Description
**************************************/
func (a *AppDynamicsMA) Description() string {
	return "A plugin that sends metrics to an AppDynamics Machine Agent"
}

/**************************************
* SampleConfig
**************************************/
func (a *AppDynamicsMA) SampleConfig() string {
	return `
  ## AppDynamics Machine Agent Host (Required)
  host = "http://127.0.0.1"

  ## AppDynamics Machine Agent HTTP Listener Port (Required)
  port = "8293"

  ## AppDynamics Metric Path (Required)
  metricPath = "Custom Metrics|Telegraf|"
`
}

/**************************************
* Connect
**************************************/
// There is no formal connection required
func (a *AppDynamicsMA) Connect() error {
	if a.Host == "" {
		return fmt.Errorf("host is a required field for AppDynamics output.")
	}
	if a.Port == "" {
		return fmt.Errorf("port is a required field for AppDynamics output.")
	}
	if a.MetricPath == "" {
		return fmt.Errorf("metricPath is a required field for AppDynamics output.")
	}
	return nil
}

/**************************************
* Close
**************************************/
// There is no formal close required
func (a *AppDynamicsMA) Close() error {
	return nil
}

/**************************************
* BuildMetrics
**************************************/
func BuildMetrics(metrics []telegraf.Metric, metricPath string) *AppDynamicsSlice {
	var appdSlice AppDynamicsSlice
	var metricPathBase string
	var buildMetric bool
	index := 0

	// Loop through each Telegraf metric and build AppD-friendly metrics
	for _, m := range metrics {
		// Initialize vars
		metricPathBase = metricPath

		metricPathBase += m.Name() + "|"
		// For each Tag, add Tag Value to metric path
		tagList := m.TagList()
		index = 0
		for _, tag := range tagList {
			// Extract the Tag Value for each Tag, and add it to our metric path
			tagValueString := string(tag.Value)
			metricPathBase += tagValueString + "|"
			index += 1
		}
		// For each Field, build an AppD metric
		fieldList := m.FieldList()
		index = 0
		for _, field := range fieldList {
			buildMetric = true
			// Populate local string variable with Field Key
			fieldKey := string(field.Key)
			// Populate local int variable with field value
			var fieldValueInt int
			switch v := field.Value.(type) {
			case int64:
				fieldValueInt = int(v)
			case uint64:
				fieldValueInt = int(v)
			case float64:
				fieldValueInt = int(v)
			case bool:
				fieldValueInt = int(0)
				if v {
					fieldValueInt = int(1)
				}
			default:
				// We don't want to err out, just drop this metric,
				//log that we dropped it, and move on.
				buildMetric = false
				fmt.Printf("Dropping unsupported Field, Key: %s, Type: %T\n", fieldKey, field.Value)
			}
			// Build local AppDynamicsJson JSON struct for each field
			if buildMetric {
				appd := AppDynamicsJson{
					MetricName:     metricPathBase + fieldKey,
					AggregatorType: "AVERAGE",
					Value:          fieldValueInt}
				// Append to slice
				appdSlice.Appdynamics_MA = append(appdSlice.Appdynamics_MA, appd)
			}
		}
	}
	return &appdSlice
}

/**************************************
* Write
**************************************/
func (a *AppDynamicsMA) Write(metrics []telegraf.Metric) error {

	// Generate AppD metrics from the Telegraf metrics and store in slice
	appdSlice := BuildMetrics(metrics, a.MetricPath)

	// Marshal into JSON encoding, the result is a byte array that we can trim
	// to our needs.
	appdJson, _ := json.Marshal(appdSlice)
	// Per AppD MA HTTP Listener format requirements, trim the
	// Appdynamics_MA json object out of the byte array
	appdJsonFinal := appdJson[18 : len(appdJson)-1]

	/*************************************************************************
	* POST the JSON data to the AppDynamics MA
	*************************************************************************/
	url := a.Host + ":" + a.Port + "/api/v1/metrics"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(appdJsonFinal))
	if err != nil {
		return fmt.Errorf("unable to create http.Request, %s\n", err.Error())
	}
	req.Header.Add("Content-Type", "application/json")

	var httpClient http.Client
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error POSTing metrics, %s\n", err.Error())
	}
	defer resp.Body.Close()

	// Successful Response Code is always 204
	if resp.StatusCode != 204 {
		return fmt.Errorf("error POSTing metrics, %s\n", err.Error())
	}
	return nil
}

/**************************************
* Init
**************************************/
func init() {
	outputs.Add("appdynamics_ma", func() telegraf.Output {
		return &AppDynamicsMA{
			Host:       defaultHost,
			Port:       defaultPort,
			MetricPath: defaultMetricPath}
	})
}
