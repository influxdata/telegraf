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
	Field               string `toml:"field"`
	SourceFormat        string `toml:"source_timestamp_format"`
	SourceTimezone      string `toml:"source_timestamp_timezone"`
	DestinationFormat   string `toml:"destination_timestamp_format"`
	DestinationTimezone string `toml:"destination_timestamp_timezone"`

	sourceLocation      *time.Location
	destinationLocation *time.Location
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

	if t.SourceTimezone == "" {
		t.SourceTimezone = "UTC"
	}

	// LoadLocation returns UTC if timezone is the empty string.
	var err error
	t.sourceLocation, err = time.LoadLocation(t.SourceTimezone)
	if err != nil {
		return fmt.Errorf("invalid source_timestamp_timezone %q: %w", t.SourceTimezone, err)
	}

	if t.DestinationTimezone == "" {
		t.DestinationTimezone = "UTC"
	}
	t.destinationLocation, err = time.LoadLocation(t.DestinationTimezone)
	if err != nil {
		return fmt.Errorf("invalid source_timestamp_timezone %q: %w", t.DestinationTimezone, err)
	}

	return nil
}

func (t *Timestamp) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		if field, ok := point.GetField(t.Field); ok {
			timestamp, err := internal.ParseTimestamp(t.SourceFormat, field, t.sourceLocation)
			if err != nil {
				continue
			}

			switch t.DestinationFormat {
			case "unix":
				point.AddField(t.Field, timestamp.Unix())
			case "unix_ms":
				point.AddField(t.Field, timestamp.UnixNano()/1000000)
			case "unix_us":
				point.AddField(t.Field, timestamp.UnixNano()/1000)
			case "unix_ns":
				point.AddField(t.Field, timestamp.UnixNano())
			default:
				inLocation := timestamp.In(t.destinationLocation)
				point.AddField(t.Field, inLocation.Format(t.DestinationFormat))
			}
		}
	}

	return in
}

func init() {
	processors.Add("timestamp", func() telegraf.Processor {
		return &Timestamp{}
	})
}
