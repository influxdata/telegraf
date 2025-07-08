//go:generate ../../../tools/readme_config_includer/generator
package nats

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
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

	conn                  *nats.Conn
	jetstreamClient       jetstream.JetStream
	jetstreamStreamConfig *jetstream.StreamConfig
	serializer            telegraf.Serializer
	tplSubject            template.Template
	includeFieldInSubject bool
	subjectIsDynamic      bool
}

// StreamConfig is the configuration for creating stream
// Almost a mirror of https://pkg.go.dev/github.com/nats-io/nats.go/jetstream#StreamConfig but with TOML tags
type StreamConfig struct {
	Name                  string                            `toml:"name"`
	Description           string                            `toml:"description"`
	Subjects              []string                          `toml:"subjects"`
	Retention             string                            `toml:"retention"`
	MaxConsumers          int                               `toml:"max_consumers"`
	MaxMsgs               int64                             `toml:"max_msgs"`
	MaxBytes              int64                             `toml:"max_bytes"`
	Discard               string                            `toml:"discard"`
	DiscardNewPerSubject  bool                              `toml:"discard_new_per_subject"`
	MaxAge                config.Duration                   `toml:"max_age"`
	MaxMsgsPerSubject     int64                             `toml:"max_msgs_per_subject"`
	MaxMsgSize            int32                             `toml:"max_msg_size"`
	Storage               string                            `toml:"storage"`
	Replicas              int                               `toml:"num_replicas"`
	NoAck                 bool                              `toml:"no_ack"`
	Template              string                            `toml:"template_owner"`
	Duplicates            config.Duration                   `toml:"duplicate_window"`
	Placement             *jetstream.Placement              `toml:"placement"`
	Mirror                *jetstream.StreamSource           `toml:"mirror"`
	Sources               []*jetstream.StreamSource         `toml:"sources"`
	Sealed                bool                              `toml:"sealed"`
	DenyDelete            bool                              `toml:"deny_delete"`
	DenyPurge             bool                              `toml:"deny_purge"`
	AllowRollup           bool                              `toml:"allow_rollup_hdrs"`
	Compression           string                            `toml:"compression"`
	FirstSeq              uint64                            `toml:"first_seq"`
	SubjectTransform      *jetstream.SubjectTransformConfig `toml:"subject_transform"`
	RePublish             *jetstream.RePublish              `toml:"republish"`
	AllowDirect           bool                              `toml:"allow_direct"`
	MirrorDirect          bool                              `toml:"mirror_direct"`
	ConsumerLimits        jetstream.StreamConsumerLimits    `toml:"consumer_limits"`
	Metadata              map[string]string                 `toml:"metadata"`
	AsyncPublish          bool                              `toml:"async_publish"`
	AsyncAckTimeout       *config.Duration                  `toml:"async_ack_timeout"`
	DisableStreamCreation bool                              `toml:"disable_stream_creation"`
}

func (*NATS) SampleConfig() string {
	return sampleConfig
}

func (n *NATS) SetSerializer(serializer telegraf.Serializer) {
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

		if n.Jetstream.DisableStreamCreation {
			stream, err := n.jetstreamClient.Stream(context.Background(), n.Jetstream.Name)
			if err != nil {
				if errors.Is(err, nats.ErrStreamNotFound) {
					return fmt.Errorf("stream %q does not exist and disable_stream_creation is true", n.Jetstream.Name)
				}
				return fmt.Errorf("failed to get stream info, name: %s, err: %w", n.Jetstream.Name, err)
			}
			subjects := stream.CachedInfo().Config.Subjects
			n.Log.Infof("Connected to existing stream %q with subjects: %v", n.Jetstream.Name, subjects)
			return nil
		}
		_, err = n.jetstreamClient.CreateOrUpdateStream(context.Background(), *n.jetstreamStreamConfig)
		if err != nil {
			return fmt.Errorf("failed to create or update stream: %w", err)
		}
		n.Log.Infof("Stream %q successfully created or updated", n.Jetstream.Name)
	}
	return nil
}

func (n *NATS) getJetstreamConfig() (*jetstream.StreamConfig, error) {
	var retention jetstream.RetentionPolicy
	switch n.Jetstream.Retention {
	case "", "limits":
		retention = jetstream.LimitsPolicy
	case "interest":
		retention = jetstream.InterestPolicy
	case "workqueue":
		retention = jetstream.WorkQueuePolicy
	default:
		return nil, fmt.Errorf("invalid 'retention' setting %q", n.Jetstream.Retention)
	}

	var discard jetstream.DiscardPolicy
	switch n.Jetstream.Discard {
	case "", "old":
		discard = jetstream.DiscardOld
	case "new":
		discard = jetstream.DiscardNew
	default:
		return nil, fmt.Errorf("invalid 'discard' setting %q", n.Jetstream.Discard)
	}

	var storage jetstream.StorageType
	switch n.Jetstream.Storage {
	case "memory":
		storage = jetstream.MemoryStorage
	case "", "file":
		storage = jetstream.FileStorage
	default:
		return nil, fmt.Errorf("invalid 'storage' setting %q", n.Jetstream.Storage)
	}

	var compression jetstream.StoreCompression
	switch n.Jetstream.Compression {
	case "s2":
		compression = jetstream.S2Compression
	case "", "none":
		compression = jetstream.NoCompression
	default:
		return nil, fmt.Errorf("invalid 'compression' setting %q", n.Jetstream.Compression)
	}

	streamConfig := &jetstream.StreamConfig{
		Name:                 n.Jetstream.Name,
		Description:          n.Jetstream.Description,
		Subjects:             n.Jetstream.Subjects,
		Retention:            retention,
		MaxConsumers:         n.Jetstream.MaxConsumers,
		MaxMsgs:              n.Jetstream.MaxMsgs,
		MaxBytes:             n.Jetstream.MaxBytes,
		Discard:              discard,
		DiscardNewPerSubject: n.Jetstream.DiscardNewPerSubject,
		MaxAge:               time.Duration(n.Jetstream.MaxAge),
		MaxMsgsPerSubject:    n.Jetstream.MaxMsgsPerSubject,
		MaxMsgSize:           n.Jetstream.MaxMsgSize,
		Storage:              storage,
		Replicas:             n.Jetstream.Replicas,
		NoAck:                n.Jetstream.NoAck,
		Template:             n.Jetstream.Template,
		Duplicates:           time.Duration(n.Jetstream.Duplicates),
		Placement:            n.Jetstream.Placement,
		Mirror:               n.Jetstream.Mirror,
		Sources:              n.Jetstream.Sources,
		Sealed:               n.Jetstream.Sealed,
		DenyDelete:           n.Jetstream.DenyDelete,
		DenyPurge:            n.Jetstream.DenyPurge,
		AllowRollup:          n.Jetstream.AllowRollup,
		Compression:          compression,
		FirstSeq:             n.Jetstream.FirstSeq,
		SubjectTransform:     n.Jetstream.SubjectTransform,
		RePublish:            n.Jetstream.RePublish,
		AllowDirect:          n.Jetstream.AllowDirect,
		MirrorDirect:         n.Jetstream.MirrorDirect,
		ConsumerLimits:       n.Jetstream.ConsumerLimits,
		Metadata:             n.Jetstream.Metadata,
	}
	return streamConfig, nil
}

func (n *NATS) Init() error {
	tpl, err := template.New("nats").Parse(n.Subject)
	if err != nil {
		return fmt.Errorf("failed to parse subject template: %w", err)
	}
	n.tplSubject = *tpl

	n.subjectIsDynamic = isSubjectDynamic(n.tplSubject, n.Subject)

	if strings.Contains(n.Subject, `.Tag "FieldName"`) {
		n.includeFieldInSubject = true
	}

	n.Log.Info("subject is dynamic: ", n.subjectIsDynamic)
	n.Log.Info("subject includes fieldname: ", n.includeFieldInSubject)

	if n.Jetstream == nil {
		return nil
	}

	// JETSTREAM-ONLY code beyond this line
	// Validate stream name
	if strings.TrimSpace(n.Jetstream.Name) == "" {
		return errors.New("stream cannot be empty")
	}

	if n.Jetstream.AsyncAckTimeout == nil {
		to := config.Duration(5 * time.Second)
		n.Jetstream.AsyncAckTimeout = &to
	}
	// Handle dynamic subject case
	if n.subjectIsDynamic {
		if len(n.Jetstream.Subjects) > 0 {
			n.Log.Info("skip adding subject to Jetstream subjects because it is dynamic")
			var err error
			n.jetstreamStreamConfig, err = n.getJetstreamConfig()
			return err
		}
		return errors.New("jetstream subjects must be set when using a dynamic subject")
	}

	// JETSTREAM-ONLY and STATIC SUBJECT code beyond this line
	// Set default subject if none provided
	if len(n.Jetstream.Subjects) == 0 {
		n.Jetstream.Subjects = []string{n.Subject}
	}

	// Append subject if not already included
	if !choice.Contains(n.Subject, n.Jetstream.Subjects) {
		n.Jetstream.Subjects = append(n.Jetstream.Subjects, n.Subject)
	}

	// Generate Jetstream config
	cfg, err := n.getJetstreamConfig()
	if err != nil {
		return fmt.Errorf("failed to parse jetstream config: %w", err)
	}
	n.jetstreamStreamConfig = cfg
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

	// If the FieldName is included in the subject, we need to split the metric into multiple metrics
	// each with a single field.
	if n.includeFieldInSubject {
		var newMetrics []telegraf.Metric
		for _, m := range metrics {
			newMetrics = append(newMetrics, splitMetricByField(m)...)
		}
		metrics = newMetrics
	}

	var bufSubject bytes.Buffer
	var err error
	var ack jetstream.PubAckFuture

	subjectMetricMap := make(map[string][]telegraf.Metric)
	for _, m := range metrics {
		bufSubject.Reset()
		err = n.tplSubject.Execute(&bufSubject, m.(telegraf.TemplateMetric))
		if err != nil {
			return fmt.Errorf("failed to execute subject template: %w", err)
		}
		subjectMetricMap[bufSubject.String()] = append(subjectMetricMap[bufSubject.String()], m)
	}

	var pafs []jetstream.PubAckFuture
	if n.Jetstream != nil && n.Jetstream.AsyncPublish {
		pafs = make([]jetstream.PubAckFuture, 0, len(metrics))
	}

	for sub, metrics := range subjectMetricMap {
		if strings.Contains(sub, "..") {
			n.Log.Errorf("invalid subject: %s, incorrect template", sub)
			continue
		}

		if strings.HasSuffix(sub, ".") {
			n.Log.Errorf("invalid subject: %s, incorrect template", sub)
			continue
		}
		for _, m := range metrics {
			buf, err := n.serializer.Serialize(m)
			if err != nil {
				n.Log.Warnf("Could not serialize metric: %v", err)
				continue
			}

			n.Log.Debugf("Publishing on Subject: %s, Metrics: %s", sub, string(buf))
			if n.Jetstream != nil {
				if n.Jetstream.AsyncPublish {
					ack, err = n.jetstreamClient.PublishAsync(sub, buf, jetstream.WithExpectStream(n.Jetstream.Name))
					pafs = append(pafs, ack)
				} else {
					_, err = n.jetstreamClient.Publish(context.Background(), sub, buf, jetstream.WithExpectStream(n.Jetstream.Name))
				}
			} else {
				err = n.conn.Publish(sub, buf)
			}
			if err != nil {
				return fmt.Errorf("failed to send NATS message: %w, subject: %s, metric: %s", err, sub, string(buf))
			}
		}
	}

	if pafs != nil {
		// Check Ack from async publish
		select {
		case <-n.jetstreamClient.PublishAsyncComplete():
			for i := range pafs {
				select {
				case <-pafs[i].Ok():
					continue
				case err := <-pafs[i].Err():
					return fmt.Errorf("publish acknowledgement is an error: %w (retrying)", err)
				}
			}
		case <-time.After(time.Duration(*n.Jetstream.AsyncAckTimeout)):
			return fmt.Errorf("waiting for acknowledgement timed out, %d messages pending", n.jetstreamClient.PublishAsyncPending())
		}
	}
	return nil
}

func isSubjectDynamic(tpl template.Template, subject string) bool {
	var buf bytes.Buffer
	err := tpl.Execute(&buf, nil)
	if err != nil || buf.String() != subject {
		return true
	}
	return false
}

func splitMetricByField(m telegraf.Metric) []telegraf.Metric {
	metrics := make([]telegraf.Metric, 0, len(m.FieldList()))
	for _, field := range m.FieldList() {
		metric := m.Copy()
		for _, f := range m.FieldList() {
			if f.Key != field.Key {
				metric.RemoveField(f.Key)
			}
		}
		metric.AddTag("FieldName", field.Key)
		metrics = append(metrics, metric)
	}
	return metrics
}

func init() {
	outputs.Add("nats", func() telegraf.Output {
		return &NATS{}
	})
}
