//go:generate ../../../tools/readme_config_includer/generator
package inlong

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/url"

	"github.com/apache/inlong/inlong-sdk/dataproxy-sdk-twins/dataproxy-sdk-golang/dataproxy"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type Inlong struct {
	ManagerURL string          `toml:"url"`
	GroupID    string          `toml:"group_id"`
	StreamID   string          `toml:"stream_id"`
	Log        telegraf.Logger `toml:"-"`

	producer   dataproxy.Client
	serializer telegraf.Serializer
}

func (*Inlong) SampleConfig() string {
	return sampleConfig
}

func (i *Inlong) Init() error {
	if i.ManagerURL == "" {
		return errors.New("'url' must not be empty")
	}
	if i.GroupID == "" {
		return errors.New("''group_id' must not be empty")
	}
	if i.StreamID == "" {
		return errors.New("'stream_id' must not be empty")
	}
	parsedURL, err := url.Parse(i.ManagerURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %s", i.ManagerURL)
	}

	switch parsedURL.Scheme {
	case "http", "https":
		if parsedURL.Host == "" {
			return fmt.Errorf("no host in URL: %s", i.ManagerURL)
		}
	default:
		return fmt.Errorf("invalid URL scheme: %s", parsedURL.Scheme)
	}
	elements := []string{"inlong", "manager", "openapi", "dataproxy", "getIpList"}
	i.ManagerURL = parsedURL.JoinPath(elements...).String()
	return nil
}

func (i *Inlong) SetSerializer(serializer telegraf.Serializer) {
	i.serializer = serializer
}

func (i *Inlong) Connect() error {
	producer, err := dataproxy.NewClient(
		dataproxy.WithGroupID(i.GroupID),
		dataproxy.WithURL(i.ManagerURL),
	)
	if err != nil {
		return &internal.StartupError{
			Err:   fmt.Errorf("connecting to manager %q with group-id %q failed: %w", i.ManagerURL, i.GroupID, err),
			Retry: true,
		}
	}
	i.producer = producer
	return nil
}

func (i *Inlong) Close() error {
	i.producer.Close()
	return nil
}

func (i *Inlong) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		b, err := i.serializer.Serialize(metric)
		if err != nil {
			return fmt.Errorf("could not serialize metric: %w", err)
		}
		err = i.producer.Send(context.Background(), dataproxy.Message{
			GroupID:  i.GroupID,
			StreamID: i.StreamID,
			Payload:  b,
		})
		if err != nil {
			return fmt.Errorf("could not send metric to GroupID: %s StreamID: %s: %w", i.GroupID, i.StreamID, err)
		}
	}
	return nil
}

func init() {
	outputs.Add("inlong", func() telegraf.Output {
		return &Inlong{}
	})
}
