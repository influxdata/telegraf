//go:generate ../../../tools/readme_config_includer/generator
package nats

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

//go:embed sample.conf
var sampleConfig string

type NATS struct {
	Servers     []string         `toml:"servers"`
	Secure      bool             `toml:"secure"`
	Name        string           `toml:"name"`
	Username    config.Secret    `toml:"username"`
	Password    config.Secret    `toml:"password"`
	Credentials string           `toml:"credentials"`
	Subject     string           `toml:"subject"`
	Jetstream   *JetstreamConfig `toml:"jetstream"`
	tls.ClientConfig

	Log telegraf.Logger `toml:"-"`

	conn            *nats.Conn
	jetstreamClient jetstream.JetStream
	serializer      serializers.Serializer
}

type JetstreamConfig struct {
	AutoCreateStream bool   `toml:"auto_create_stream"`
	Stream           string `toml:"stream"`
	StreamJSON       string `toml:"stream_config_json"`
	// Other jetsream options

	// storing local copy of stream config
	streamConfig jetstream.StreamConfig
}

func (*NATS) SampleConfig() string {
	return sampleConfig
}

func (n *NATS) SetSerializer(serializer serializers.Serializer) {
	n.serializer = serializer
}

func (n *NATS) Connect() error {
	var err error

	opts := []nats.Option{
		nats.MaxReconnects(-1),
	}

	// override authentication, if any was specified
	if !n.Username.Empty() && !n.Password.Empty() {
		username, err := n.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		password, err := n.Password.Get()
		if err != nil {
			username.Destroy()
			return fmt.Errorf("getting password failed: %w", err)
		}
		opts = append(opts, nats.UserInfo(username.String(), password.String()))
		username.Destroy()
		password.Destroy()
	}

	if n.Credentials != "" {
		opts = append(opts, nats.UserCredentials(n.Credentials))
	}

	if n.Name != "" {
		opts = append(opts, nats.Name(n.Name))
	}

	if n.Secure {
		tlsConfig, err := n.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}

		opts = append(opts, nats.Secure(tlsConfig))
	}

	// try and connect
	n.conn, err = nats.Connect(strings.Join(n.Servers, ","), opts...)
	if n.Jetstream != nil {
		// connect to jetstream
		n.jetstreamClient, err = jetstream.New(n.conn)
		if err != nil {
			return fmt.Errorf("failed to connect to jetstream: %w", err)
		}
		if n.Jetstream.Stream == "" {
			return errors.New("stream cannot be empty")
		}
		streamExists := n.streamExists(n.Jetstream.Stream)
		if !streamExists {
			if !n.Jetstream.AutoCreateStream {
				return fmt.Errorf("stream %s does not exist", n.Jetstream.Stream)
			}
			streamConfigJSON := strings.TrimSpace(n.Jetstream.StreamJSON)
			var streamCfg jetstream.StreamConfig
			if len(streamConfigJSON) > 0 {
				err = json.Unmarshal([]byte(streamConfigJSON), &streamCfg)
				if err != nil {
					return fmt.Errorf("invalid jetstream config %w", err)
				}
			}
			streamCfg.Name = n.Jetstream.Stream
			streamCfg.Subjects = []string{n.Subject}
			n.Jetstream.streamConfig = streamCfg
			err = n.createStream(streamCfg)
			if err != nil {
				return fmt.Errorf("failed to create stream: %w", err)
			}
		}
	}
	return err
}

func (n *NATS) streamExists(stream string) bool {
	_, err := n.jetstreamClient.Stream(context.Background(), stream)
	return err == nil
}

func (n *NATS) createStream(streamCfg jetstream.StreamConfig) error {
	_, err := n.jetstreamClient.CreateStream(context.Background(), streamCfg)
	return err
}

func (n *NATS) Close() error {
	n.conn.Close()
	return nil
}

func (n *NATS) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	for _, metric := range metrics {
		buf, err := n.serializer.Serialize(metric)
		if err != nil {
			n.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}
		if n.Jetstream != nil {
			_, err = n.jetstreamClient.Publish(context.Background(), n.Subject, buf)
		} else {
			err = n.conn.Publish(n.Subject, buf)
		}
		if err != nil {
			return fmt.Errorf("FAILED to send NATS message: %w", err)
		}
	}
	return nil
}

func init() {
	outputs.Add("nats", func() telegraf.Output {
		return &NATS{}
	})
}
