package date

import (
	"errors"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/processors"
)

const defaultTimezone = "UTC"

type Date struct {
	TagKey     string          `toml:"tag_key"`
	FieldKey   string          `toml:"field_key"`
	DateFormat string          `toml:"date_format"`
	DateOffset config.Duration `toml:"date_offset"`
	Timezone   string          `toml:"timezone"`

	location *time.Location
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
