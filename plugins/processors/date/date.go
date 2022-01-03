package date

import (
	"errors"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
	## New tag to create
	tag_key = "month"

	## New field to create (cannot set both field_key and tag_key)
	# field_key = "month"

	## Date format string, must be a representation of the Go "reference time"
	## which is "Mon Jan 2 15:04:05 -0700 MST 2006".
	date_format = "Jan"

	## If destination is a field, date format can also be one of
	## "unix", "unix_ms", "unix_us", or "unix_ns", which will insert an integer field.
	# date_format = "unix"

	## Offset duration added to the date string when writing the new tag.
	# date_offset = "0s"

	## Timezone to use when creating the tag or field using a reference time
	## string.  This can be set to one of "UTC", "Local", or to a location name
	## in the IANA Time Zone database.
	##   example: timezone = "America/Los_Angeles"
	# timezone = "UTC"
`

const defaultTimezone = "UTC"

type Date struct {
	TagKey     string          `toml:"tag_key"`
	FieldKey   string          `toml:"field_key"`
	DateFormat string          `toml:"date_format"`
	DateOffset config.Duration `toml:"date_offset"`
	Timezone   string          `toml:"timezone"`

	location *time.Location
}

func (d *Date) SampleConfig() string {
	return sampleConfig
}

func (d *Date) Description() string {
	return "Dates measurements, tags, and fields that pass through this filter."
}

func (d *Date) Init() error {
	// Check either TagKey or FieldKey specified
	if len(d.FieldKey) > 0 && len(d.TagKey) > 0 {
		return errors.New("Only one of field_key or tag_key can be specified")
	} else if len(d.FieldKey) == 0 && len(d.TagKey) == 0 {
		return errors.New("One of field_key or tag_key must be specified")
	}

	var err error
	// LoadLocation returns UTC if timezone is the empty string.
	d.location, err = time.LoadLocation(d.Timezone)
	return err
}

func (d *Date) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		tm := point.Time().In(d.location).Add(time.Duration(d.DateOffset))
		if len(d.TagKey) > 0 {
			point.AddTag(d.TagKey, tm.Format(d.DateFormat))
		} else if len(d.FieldKey) > 0 {
			switch d.DateFormat {
			case "unix":
				point.AddField(d.FieldKey, tm.Unix())
			case "unix_ms":
				point.AddField(d.FieldKey, tm.UnixNano()/1000000)
			case "unix_us":
				point.AddField(d.FieldKey, tm.UnixNano()/1000)
			case "unix_ns":
				point.AddField(d.FieldKey, tm.UnixNano())
			default:
				point.AddField(d.FieldKey, tm.Format(d.DateFormat))
			}
		}
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
