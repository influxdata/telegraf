package newrelic

import(
  "fmt"
  "encoding/json"
  "os"
  "strings"
  "bytes"
  "github.com/influxdata/telegraf"
)

type NewRelicComponent struct {
  Duration int
  TMetric telegraf.Metric
  GuidBase string
	tags *NewRelicTags
}

func (nrc NewRelicComponent) Tags() *NewRelicTags {
	if nrc.tags == nil {
    nrc.tags = &NewRelicTags{};
		nrc.tags.Fill(nrc.TMetric.Tags())
	}
	return nrc.tags
}

func (nrc NewRelicComponent) Name() string {
  return nrc.TMetric.Name()
}

func metricValue(value interface{}) int {
	result := 0
	switch value.(type) {
	case int32:
		result = int(value.(int32))
	case int64:
		result = int(value.(int64))
	case float32:
		result = int(value.(float32))
	case float64:
		result = int(value.(float64))
	default:
		result = 0
  }
	return result
}

func (nrc* NewRelicComponent) MetricName(originalName string) string {
  var nameBuffer bytes.Buffer
  nameBuffer.WriteString("Component/")
  nameBuffer.WriteString(strings.Title(nrc.TMetric.Name()))
  nameBuffer.WriteString("/")
  nameBuffer.WriteString(strings.Title(originalName))
  tags := nrc.Tags()
  for _, key := range tags.SortedKeys {
    nameBuffer.WriteString(fmt.Sprintf("/%s-%s", key, tags.GetTag(key)))
  }
  nameBuffer.WriteString("[Units]")
  return nameBuffer.String()
}

func (nrc *NewRelicComponent) Metrics() map[string]int {
	result := make(map[string]int)
	for k,v := range(nrc.TMetric.Fields()) {
		result[nrc.MetricName(k)] = metricValue(v)
	}
  return result
}

func (nrc NewRelicComponent) Hostname() string {
  result := nrc.Tags().Hostname
  if result == "" {
    osname, err := os.Hostname()
    if err == nil { result = "unknown" } else { result = osname }
  }
  return result
}

func (nrc *NewRelicComponent) Guid() string {
	return fmt.Sprintf("%s-%s", nrc.GuidBase, strings.ToLower(nrc.TMetric.Name()))
}

func (nrc NewRelicComponent) MarshalJSON() ([]byte, error) {
	myData := map[string]interface{} {
		"name": nrc.Hostname(),
		"guid": nrc.Guid(),
		"duration": nrc.Duration,
		"metrics": nrc.Metrics(),
	}
	return json.Marshal(myData)
}
