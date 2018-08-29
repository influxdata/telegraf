package splunk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Splunk struct {
	Prefix string
	Source string

	SplunkUrl  string
	AuthString string
	Timeout    internal.Duration
	client     *http.Client

	SimpleFields    bool
	MetricSeparator string
	ConvertPaths    bool
	ConvertBool     bool
	ReplaceSpecials bool
	UseRegex        bool
	StringToNumber  map[string][]map[string]float64
	serializer      serializers.Serializer
}

// Descriptions of the flags in Splunk struct
/*
SimpleFields    - boolean to determine whether or not to use xxx.value as the metric name (true) or ommit the .value from the metric name (false)
MetricSeparator - character to use between metric and field name.  defaults to . (dot)
ConvertPaths    - boolean to convert all _ (underscore) chartacters in metric name to MetricSeparator
ConvertBool     - boolean to convert all true/false values to 1/0
ReplaceSpecials - boolean to sanitize special characters in metric names with "-"
UseRegex        - boolean to use Regex to sanitize metric and tag names from invalid characters
StringToNumber  - map used internally to convert string values to numerics
*/

// catch many of the invalid chars that could appear in a metric or tag name
var sanitizedChars = strings.NewReplacer(
	"!", "-", "@", "-", "#", "-", "$", "-", "%", "-", "^", "-", "&", "-",
	"*", "-", "(", "-", ")", "-", "+", "-", "`", "-", "'", "-", "\"", "-",
	"[", "-", "]", "-", "{", "-", "}", "-", ":", "-", ";", "-", "<", "-",
	">", "-", ",", "-", "?", "-", "/", "-", "\\", "-", "|", "-", " ", "-",
	"=", "-",
)

// instead of Replacer which may miss some special characters we can use a regex pattern, but this is significantly slower than Replacer
var sanitizedRegex = regexp.MustCompile("[^a-zA-Z\\d_.-]")

var tagValueReplacer = strings.NewReplacer("\"", "\\\"", "*", "-")

var pathReplacer = strings.NewReplacer("_", "_")

var sampleConfig = `
## REQUIRED
## URL of the Splunk Enterprise HEC endpoint (i.e.: http://localhost:8088/services/collector) 
SplunkUrl = "http://localhost:8088/services/collector"

## REQUIRED
## Splunk Authorization Token for sending data to a Splunk HTTPEventCollector (HEC). 
##   Note:  This Token should map to a 'metrics' index in Splunk. 
AuthString = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"

## OPTIONAL:  prefix for metrics keys
#prefix = "my.specific.prefix."

## OPTIONAL:  whether to use "value" for name of simple fields.    
##  Default is false which will result in using only the measurement name as the metric name, not "value"
#simple_fields = false

## OPTIONAL:  character to use between metric and field name.  defaults to . (dot)
#metric_separator = "."

## OPTIONAL:  Convert metric name paths to use metricSeperator character
## When true will convert all _ (underscore) chartacters in final metric name
#convert_paths = false

## OPTIONAL:  Replace special characters in metric names with "-".  
## This can be useful if metric names contain special characters  
#replace_special_chars = false

## OPTIONAL:  Use Regex to sanitize metric and tag names from invalid characters  
## Regex is more thorough, but significantly slower
#use_regex = false

## OPTIONAL:  whether to convert boolean values to numeric values, with false -> 0.0 and true -> 1.0.  default true
#convert_bool = true

`

type SplunkMetric struct {
	Time   int64                  `json:"time"`
	Event  string                 `json:"event"`
	Source string                 `json:"source,omitempty"`
	Host   string                 `json:"host"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

func (s *Splunk) Connect() error {
	if s.AuthString == "" {
		return fmt.Errorf("An Authrorization String is required to send Telegraf data to Splunk.")
	}

	s.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: s.Timeout.Duration,
	}
	return nil
}

func (s *Splunk) SetSerializer(serializer serializers.Serializer) {
	s.serializer = serializer
}

func (s *Splunk) Write(measures []telegraf.Metric) error {
	//  If there are no metrics, just stop now.
	if len(measures) == 0 {
		return nil
	}

	var allMetrics = []SplunkMetric{}
	splunkMetric := SplunkMetric{}

	// -----------------------------------------------------------------------------------------------------------------------
	//  Loop through each measure (input plugin)
	// -----------------------------------------------------------------------------------------------------------------------
	for _, m := range measures {
		//fmt.Printf("\n\n Processing %s \n  %+v \n\n",m.Name(), m.Fields())

		// -----------------------------------------------------------------------------------------------------------------------
		//  Loop through tags for each measure    (add ALL Tags (except host) to each "field" element)
		// -----------------------------------------------------------------------------------------------------------------------
		fields := make(map[string]interface{}) // convert the Tags array from map[string]string to map[string]interface{}
		for tagName, value := range m.Tags() {
			//fmt.Printf(" tag values:  %s:%s \n",tagName,value)
			// if the tagName == 'host', set the Splunk metric 'header' value == host
			if tagName == "host" {
				splunkMetric.Host = value
			} else {
				fields[tagName] = value
			}
		}

		// -----------------------------------------------------------------------------------------------------------------------
		//  Loop through each metric and create a map of all tags + the name and value of the metric  (Splunk Metric format)
		// -----------------------------------------------------------------------------------------------------------------------
		for fieldName, value := range m.Fields() {
			fields["metric_name"] = processFieldName(m.Name(), fieldName, s)

			metricValue, buildError := processFieldValue(value, fieldName, s)
			if buildError != nil {
				log.Printf("D! [Splunk] %s\n", buildError.Error())
				continue
			}
			fields["_value"] = metricValue

			splunkMetric.Time = m.Time().UnixNano() // convert metric nanoseconds to unix time
			splunkMetric.Event = "metric"
			splunkMetric.Fields = fields

		} //for fieldName, value := range m.Fields() {

		// add this metric to the splunkMetric array
		allMetrics = append(allMetrics, splunkMetric)

	} //for _, m := range measures {

	// -----------------------------------------------------------------------------------------------------------------
	//  Now we can send this set of metrics to Splunk
	//
	//  Create a []byte array to send via an HTTP POST
	// -----------------------------------------------------------------------------------------------------------------
	var payload []byte
	var err error
	payload, err = json.Marshal(allMetrics)

	if err != nil {
		return fmt.Errorf("unable to marshal data, %s\n", err.Error())
	}
	log.Printf("D! Output [Splunk] %s\n", payload)

	// -----------------------------------------------------------------------------------------------------------------
	//  Send the data to Splunk
	// -----------------------------------------------------------------------------------------------------------------
	req, err := http.NewRequest("POST", s.SplunkUrl, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("unable to create http.Request \n    URL:%s\n\n", s.SplunkUrl)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Splunk "+s.AuthString)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error POST-ing metrics to Splunk[%s]  Sending Data:%s\n", err, payload)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("received bad status code posting to %s, %d\n\n%s\n", s.SplunkUrl, resp.StatusCode, payload)
	}

	return nil
}

// -----------------------------------------------------------------------------------------------------------------------
//  Check for metric name ending with .value and remove it, sanitize characters if desired and add separators if needed
// -----------------------------------------------------------------------------------------------------------------------
func processFieldName(measureName string, fieldName string, s *Splunk) string {
	// Add any optional metric prefix to the metric name
	name := s.Prefix + measureName

	// If the fieldName!="value" add it to the metric name
	if s.SimpleFields || fieldName != "value" {
		name += s.MetricSeparator + fieldName
	}

	// Sanitize metric name by removing special characters
	if s.ReplaceSpecials {
		name = sanitizedChars.Replace(name)
	}

	// Sanitize metric name using regex
	if s.UseRegex {
		name = sanitizedRegex.ReplaceAllLiteralString(name, "-")
	}

	// Convert all _ (underscore) chartacters in final metric name to metricSeparator ("." by default)
	if s.ConvertPaths {
		name = pathReplacer.Replace(name)
	}
	return name
}

// -----------------------------------------------------------------------------------------------------------------------
//  Parse the value depending on the data type
// -----------------------------------------------------------------------------------------------------------------------
func processFieldValue(v interface{}, name string, s *Splunk) (float64, error) {
	switch p := v.(type) {
	case bool:
		if s.ConvertBool {
			if p {
				return 1, nil
			} else {
				return 0, nil
			}
		}
	case int64:
		return float64(p), nil
	case uint64:
		return float64(p), nil
	case float64:
		return float64(p), nil
	case string:
		for prefix, mappings := range s.StringToNumber {
			if strings.HasPrefix(name, prefix) {
				for _, mapping := range mappings {
					val, hasVal := mapping[string(p)]
					if hasVal {
						return val, nil
					}
				}
			}
		}
		return 0, fmt.Errorf("unexpected type: %T, with value: %v, for: %s", v, v, name)
	default:
		return 0, fmt.Errorf("unexpected type: %T, with value: %v, for: %s", v, v, name)
	}

	return 0, fmt.Errorf("unexpected type: %T, with value: %v, for: %s", v, v, name)
}

func (s *Splunk) SampleConfig() string {
	return sampleConfig
}

func (s *Splunk) Description() string {
	return "Configuration for Splunk server to send metrics to"
}

func (s *Splunk) Close() error {
	return nil
}

func init() {
	outputs.Add("splunk", func() telegraf.Output {
		return &Splunk{
			MetricSeparator: ".",
			ConvertPaths:    false,
			ConvertBool:     true,
		}
	})
}
