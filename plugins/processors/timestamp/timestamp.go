//go:generate ../../../tools/readme_config_includer/generator
package timestamp

import (
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Timestamp struct {
	SourceField       string `toml:"source_timestamp_field"`
	SourceFormat      string `toml:"source_timestamp_format"`
	SourceTimezone    string `toml:"source_timestamp_timezone"`
	DestinationFormat string `toml:"destination_timestamp_format"`

	location *time.Location
}

func (*Timestamp) SampleConfig() string {
	return sampleConfig
}

func (t *Timestamp) Init() error {
	switch t.SourceFormat {
	case "":
		return errors.New("source_timestamp_format is required")
	case "unix", "unix_ms", "unix_us", "unix_ns":
	default:
		if time.Now().Format(t.SourceFormat) == t.SourceFormat {
			return fmt.Errorf("invalid timestamp format %q", t.SourceFormat)
		}
	}

	switch t.DestinationFormat {
	case "":
		return errors.New("source_timestamp_format is required")
	case "unix", "unix_ms", "unix_us", "unix_ns":
	default:
		if time.Now().Format(t.DestinationFormat) == t.DestinationFormat {
			return fmt.Errorf("invalid timestamp format %q", t.DestinationFormat)
		}
	}

	// LoadLocation returns UTC if timezone is the empty string.
	var err error
	t.location, err = time.LoadLocation(t.SourceTimezone)
	return err
}

func (t *Timestamp) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		if field, ok := point.GetField(t.SourceField); ok {
			var timestamp time.Time
			switch t.SourceFormat {
			case "unix":
				ts, err := internal.ToInt64(field)
				if err != nil {
					continue
				}
				timestamp = time.Unix(ts, 0)
			case "unix_ms":
				ts, err := internal.ToInt64(field)
				if err != nil {
					continue
				}
				timestamp = time.UnixMilli(ts)
			case "unix_us":
				ts, err := internal.ToInt64(field)
				if err != nil {
					continue
				}
				timestamp = time.UnixMilli(ts)
			case "unix_ns":
				ts, err := internal.ToInt64(field)
				if err != nil {
					continue
				}
				timestamp = time.Unix(0, ts)
			default:
				stringField, err := internal.ToString(field)
				if err != nil {
					continue
				}
				timestamp, err = time.ParseInLocation(t.SourceFormat, stringField, t.location)
				if err != nil {
					continue
				}
			}

			switch t.DestinationFormat {
			case "unix":
				point.AddField(t.SourceField, timestamp.Unix())
			case "unix_ms":
				point.AddField(t.SourceField, timestamp.UnixNano()/1000000)
			case "unix_us":
				point.AddField(t.SourceField, timestamp.UnixNano()/1000)
			case "unix_ns":
				point.AddField(t.SourceField, timestamp.UnixNano())
			default:
				point.AddField(t.SourceField, timestamp.Format(t.DestinationFormat))
			}
		}
	}

	return in
}

func init() {
	processors.Add("timestamp", func() telegraf.Processor {
		return &Timestamp{
			SourceTimezone: "UTC",
		}
	})
}
