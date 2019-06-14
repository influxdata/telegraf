package date

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
##Specify the date tags to add
tagKey = "month"
dateFormat = "%m"

`

type Date struct {
	TagKey     string `toml:"tagKey"`
	DateFormat string `toml:"dateFormat"`
}

func (d *Date) SampleConfig() string {
	return sampleConfig
}

func (d *Date) Description() string {
	return "Dates measurements, tags, and fields that pass through this filter."
}

func (d *Date) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		point.AddTag(d.TagKey, point.Time().Format(d.DateFormat))

	}

	return in

}

func init() {
	processors.Add("date", func() telegraf.Processor {
		return &Date{}
	})
}

/**
 *

[processors.date]
  jdfj

  ##Set Months to True or False
  tagKey = "month"
  dateFormat = "%m" // January

[processors.date]
  jdfj

  ##Set Months to True or False
  tagKey = "day_of_week"
  dateFormat = "%d" // Wednesday


  # [[processors.regex.fields]]
  #   key = "request"
  #   pattern = ".*category=(\\w+).*"
  #   replacement = "${1}"
  #   result_key = "search_category"


*/
