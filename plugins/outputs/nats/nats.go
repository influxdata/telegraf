//go:generate ../../../tools/readme_config_includer/generator
package nats

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

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
	Servers     []string      `toml:"servers"`
	Secure      bool          `toml:"secure"`
	Name        string        `toml:"name"`
	Username    config.Secret `toml:"username"`
	Password    config.Secret `toml:"password"`
	Credentials string        `toml:"credentials"`
	Subject     string        `toml:"subject"`
	Jetstream   *StreamConfig `toml:"jetstream"`
	tls.ClientConfig

	Log telegraf.Logger `toml:"-"`

	conn            *nats.Conn
	jetstreamClient jetstream.JetStream
	serializer      serializers.Serializer
}

// StreamConfig is the configuration for creating stream
// Almost a mirror of https://pkg.go.dev/github.com/nats-io/nats.go/jetstream#StreamConfig but with
// TOML tags.
//
// Some custom types such as RetentionPolicy still point to the source to reuse Stringer interface.
type StreamConfig struct {
	Name                 string                            `toml:"name"`
	Description          string                            `toml:"description,omitempty"`
	Subjects             []string                          `toml:"subjects,omitempty"`
	Retention            jetstream.RetentionPolicy         `toml:"retention"`
	MaxConsumers         int                               `toml:"max_consumers"`
	MaxMsgs              int64                             `toml:"max_msgs"`
	MaxBytes             int64                             `toml:"max_bytes"`
	Discard              jetstream.DiscardPolicy           `toml:"discard"`
	DiscardNewPerSubject bool                              `toml:"discard_new_per_subject,omitempty"`
	MaxAge               time.Duration                     `toml:"max_age"`
	MaxMsgsPerSubject    int64                             `toml:"max_msgs_per_subject"`
	MaxMsgSize           int32                             `toml:"max_msg_size,omitempty"`
	Storage              jetstream.StorageType             `toml:"storage"`
	Replicas             int                               `toml:"num_replicas"`
	NoAck                bool                              `toml:"no_ack,omitempty"`
	Template             string                            `toml:"template_owner,omitempty"`
	Duplicates           time.Duration                     `toml:"duplicate_window,omitempty"`
	Placement            *jetstream.Placement              `toml:"placement,omitempty"`
	Mirror               *jetstream.StreamSource           `toml:"mirror,omitempty"`
	Sources              []*jetstream.StreamSource         `toml:"sources,omitempty"`
	Sealed               bool                              `toml:"sealed,omitempty"`
	DenyDelete           bool                              `toml:"deny_delete,omitempty"`
	DenyPurge            bool                              `toml:"deny_purge,omitempty"`
	AllowRollup          bool                              `toml:"allow_rollup_hdrs,omitempty"`
	Compression          jetstream.StoreCompression        `toml:"compression"`
	FirstSeq             uint64                            `toml:"first_seq,omitempty"`
	SubjectTransform     *jetstream.SubjectTransformConfig `toml:"subject_transform,omitempty"`
	RePublish            *jetstream.RePublish              `toml:"republish,omitempty"`
	AllowDirect          bool                              `toml:"allow_direct"`
	MirrorDirect         bool                              `toml:"mirror_direct"`
	ConsumerLimits       jetstream.StreamConsumerLimits    `toml:"consumer_limits,omitempty"`
	Metadata             map[string]string                 `toml:"metadata,omitempty"`
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
		var streamConfig jetstream.StreamConfig
		n.convertToJetstreamConfig(&streamConfig)
		_, err = n.jetstreamClient.CreateOrUpdateStream(context.Background(), streamConfig)
		if err != nil {
			return fmt.Errorf("failed to create or update stream: %w", err)
		}
		n.Log.Infof("stream (%s) successfully created or updated", n.Jetstream.Name)
	}
	return nil
}

func (n *NATS) convertToJetstreamConfig(streamConfig *jetstream.StreamConfig) {
	telegrafStreamConfig := reflect.ValueOf(n.Jetstream).Elem()
	natsStreamConfig := reflect.ValueOf(streamConfig).Elem()
	for i := 0; i < telegrafStreamConfig.NumField(); i++ {
		destField := natsStreamConfig.FieldByName(telegrafStreamConfig.Type().Field(i).Name)
		if destField.IsValid() && destField.CanSet() {
			destField.Set(telegrafStreamConfig.Field(i))
		}
	}
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
