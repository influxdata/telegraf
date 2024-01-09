//go:generate ../../../tools/readme_config_includer/generator
package nats

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/influxdata/toml"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

//go:embed sample.conf
var sampleConfig string

type NATS struct {
	Servers     []string                `toml:"servers"`
	Secure      bool                    `toml:"secure"`
	Name        string                  `toml:"name"`
	Username    config.Secret           `toml:"username"`
	Password    config.Secret           `toml:"password"`
	Credentials string                  `toml:"credentials"`
	Subject     string                  `toml:"subject"`
	Jetstream   *JetstreamConfigWrapper `toml:"jetstream"`
	tls.ClientConfig

	Log telegraf.Logger `toml:"-"`

	conn            *nats.Conn
	jetstreamClient jetstream.JetStream
	serializer      serializers.Serializer
}

type JetstreamConfigWrapper struct {
	jetstream.StreamConfig
}

func (jw *JetstreamConfigWrapper) UnmarshalTOML(data []byte) error {
	var tomlMap map[string]interface{}

	if err := toml.Unmarshal(data, &tomlMap); err != nil {
		return err
	}

	// Extract the deeply nested table by specifying the keys(in order)
	keys := []string{"outputs", "nats", "jetstream"}

	nestedTable, err := extractNestedTable(tomlMap, keys...)
	if err != nil {
		return err
	}
	jsonBytes, err := json.Marshal(nestedTable)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, &jw.StreamConfig)
}

// recursive function to extract a nested table
func extractNestedTable(tomlMap map[string]interface{}, keys ...string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return tomlMap, nil
	}

	key := keys[0]
	remainingKeys := keys[1:]

	value, ok := tomlMap[key]
	if !ok {
		return nil, fmt.Errorf("key '%s' not found in TOML data", key)
	}

	innerMap, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("value of key '%s' is not a table", key)
	}

	return extractNestedTable(innerMap, remainingKeys...)
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
	if err != nil {
		return err
	}

	if n.Jetstream != nil {
		n.jetstreamClient, err = jetstream.New(n.conn)
		if err != nil {
			return fmt.Errorf("failed to connect to jetstream: %w", err)
		}

		if len(n.Jetstream.Subjects) == 0 {
			n.Jetstream.Subjects = []string{n.Subject}
		}
		if !choice.Contains(n.Subject, n.Jetstream.Subjects) {
			n.Jetstream.Subjects = append(n.Jetstream.Subjects, n.Subject)
		}
		_, err = n.jetstreamClient.CreateOrUpdateStream(context.Background(), n.Jetstream.StreamConfig)
		if err != nil {
			return fmt.Errorf("failed to create or update stream: %w", err)
		}
		n.Log.Infof("stream (%s) successfully created or updated", n.Jetstream.Name)
	}
	return nil
}

func (n *NATS) Init() error {
	if n.Jetstream != nil {
		if strings.TrimSpace(n.Jetstream.Name) == "" {
		    return errors.New("stream cannot be empty")
	    }
	}

	return nil
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
		// use the same Publish API for nats core and jetstream
		err = n.conn.Publish(n.Subject, buf)
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
