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
	UseRegex        bool
	StringToNumber  map[string][]map[string]float64
	serializer      serializers.Serializer
}

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

## OPTIONAL:  whether to use "value" for name of simple fields
#simple_fields = false

## OPTIONAL:  character to use between metric and field name.  defaults to . (dot)
#metric_separator = "."

## OPTIONAL:  Convert metric name paths to use metricSeperator character
## When true (default) will convert all _ (underscore) chartacters in final metric name
#convert_paths = true

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

	splunkMetric := SplunkMetric{}

	// -----------------------------------------------------------------------------------------------------------------------
	//  Loop through each measure (plugin)
	// -----------------------------------------------------------------------------------------------------------------------
	for _, m := range measures {
		//fmt.Printf("\n\n Processing %s \n  %+v \n\n",m.Name(), m.Fields())

		// -----------------------------------------------------------------------------------------------------------------------
		//  Loop through tags for each measure    (add ALL Tags (except host) to each "field" element)
		// -----------------------------------------------------------------------------------------------------------------------
		newTags := make(map[string]interface{}) // convert the Tags array from map[string]string to map[string]interface{}
		for tagName, value := range m.Tags() {
			//fmt.Printf(" tag values:  %s:%s \n",tagName,value)
			if tagName == "host" {
				splunkMetric.Host = value
			} else {
				newTags[tagName] = value
			}
		}

		// -----------------------------------------------------------------------------------------------------------------------
		//  Loop through each metric and create a map of all tags + the name and value of the metric  (Splunk Metric format)
		// -----------------------------------------------------------------------------------------------------------------------
		fields := newTags
		for fieldName, value := range m.Fields() {
			fields["metric_name"] = processFieldName(m.Name(), fieldName, s)

			metricValue, buildError := processFieldValue(value, fieldName, s)
			if buildError != nil {
				log.Printf("D! [Splunk] %s\n", buildError.Error())
				continue
			}
			fields["_value"] = metricValue

			// -----------------------------------------------------------------------------------------------------------------
			//  Create a []byte array to send via an HTTP POST
			// -----------------------------------------------------------------------------------------------------------------
			var payload []byte
			var err error

			splunkMetric.Time = m.Time().UnixNano() / 1000000000 // convert metric nanoseconds to unix time
			splunkMetric.Event = "metric"
			splunkMetric.Fields = fields
			payload, err = json.Marshal(splunkMetric)

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
			// Check for existence of s.AuthString to prevent race ('panic: runtime error: invalid memory address or nil pointer dereference')
			if s != nil {
				req.Header.Add("Authorization", "Splunk " + s.AuthString)
			}
			resp, err := s.client.Do(req)
			if err != nil {
				return fmt.Errorf("error POST-ing metrics to Splunk[%s]  Sending Data:%s\n", err, payload)
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode > 209 {
				return fmt.Errorf("received bad status code posting to %s, %d\n\n%s\n", s.SplunkUrl, resp.StatusCode, payload)
			}
		}
	}
	return nil
}

// -----------------------------------------------------------------------------------------------------------------------
//  Check for metric name ending with .value and remove it, sanitize characters and add separators
// -----------------------------------------------------------------------------------------------------------------------
func processFieldName(measureName string, fieldName string, s *Splunk) string {
	var name string
	if !s.SimpleFields && fieldName == "value" {
		name = fmt.Sprintf("%s%s", s.Prefix, measureName)
	} else {
		name = fmt.Sprintf("%s%s%s%s", s.Prefix, measureName, s.MetricSeparator, fieldName)
	}

	if s.UseRegex {
		name = sanitizedRegex.ReplaceAllLiteralString(name, "-")
	} else {
		name = sanitizedChars.Replace(name)
	}

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
		return float64(v.(int64)), nil
	case uint64:
		return float64(v.(uint64)), nil
	case float64:
		return v.(float64), nil
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
			ConvertPaths:    true,
			ConvertBool:     true,
		}
	})
}
