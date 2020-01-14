package date

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
  ## New tag to create
  tag_key = "month"

  ## Date format string, must be a representation of the Go "reference time"
  ## which is "Mon Jan 2 15:04:05 -0700 MST 2006".
  date_format = "Jan"

  ## Offset duration added to the date string when writing the new tag.
  # date_offset = "0s"

  ## Timezone to use when creating the tag.  This can be set to one of
  ## "UTC", "Local", or to a location name in the IANA Time Zone database.
  ##   example: timezone = "America/Los_Angeles"
  # timezone = "UTC"
`

const defaultTimezone = "UTC"

type Date struct {
	TagKey     string            `toml:"tag_key"`
	DateFormat string            `toml:"date_format"`
	DateOffset internal.Duration `toml:"date_offset"`
	Timezone   string            `toml:"timezone"`

	location *time.Location
}

func (d *Date) SampleConfig() string {
	return sampleConfig
}

func (d *Date) Description() string {
	return "Dates measurements, tags, and fields that pass through this filter."
}

func (d *Date) Init() error {
	var err error
	// LoadLocation returns UTC if timezone is the empty string.
	d.location, err = time.LoadLocation(d.Timezone)
	return err
}

func (d *Date) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		tm := point.Time().In(d.location).Add(d.DateOffset.Duration)
		point.AddTag(d.TagKey, tm.Format(d.DateFormat))
	}

	return in
}

func init() {
	processors.Add("date", func() telegraf.Processor {
		return &Date{
			Timezone: defaultTimezone,
		}
	})
}
