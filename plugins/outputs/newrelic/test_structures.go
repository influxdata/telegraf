package newrelic

import (
  "time"
  "github.com/influxdata/influxdb/client/v2"
)

func DemoTagList() map[string]string {
	return map[string]string{
     "Fluff": "Naa",
     "Hoof": "Bar",
		 "Zoo": "Goo",
		 "host": "Hulu",
	}
}

type DemoMetric struct {
  MyName string
  TagList map[string]string
}

func (dm DemoMetric) Name() string { return dm.MyName }
func (dm DemoMetric) Tags() map[string]string { return dm.TagList }
func (dm DemoMetric) Time() time.Time { return time.Now() }
func (dm DemoMetric) UnixNano() int64 { return 0 }
func (dm DemoMetric) Fields() map[string]interface{} { return nil }
func (dm DemoMetric) String() string { return "StringRepresenation" }
func (dm DemoMetric) PrecisionString(precison string) string { return "PrecisionString" }
func (dm DemoMetric) Point() *client.Point { return nil }
