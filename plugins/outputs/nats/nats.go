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
	"text/template/parse"
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
	Servers              []string      `toml:"servers"`
	Secure               bool          `toml:"secure"`
	Name                 string        `toml:"name"`
	Username             config.Secret `toml:"username"`
	Password             config.Secret `toml:"password"`
	Credentials          string        `toml:"credentials"`
	Subject              string        `toml:"subject"`
	Jetstream            *StreamConfig `toml:"jetstream"`
	ExternalStreamConfig bool          `toml:"external_stream_config"`
	SubjectLayout        []string      `toml:"with_subject_layout"`

	tls.ClientConfig

	Log telegraf.Logger `toml:"-"`

	conn                  *nats.Conn
	jetstreamClient       jetstream.JetStream
	jetstreamStreamConfig *jetstream.StreamConfig
	serializer            telegraf.Serializer
	tplSubject            template.Template
	includeFieldInSubject bool
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

type subMsgPair struct {
	subject string
	metric  telegraf.Metric
}
type metricSubjectTmplCtx struct {
	Name string
	//Field    func() string
	getTag   func(string) string
	getField func() string
}

func (m metricSubjectTmplCtx) GetTag(key string) string {
	return m.getTag(key)
}

func (m metricSubjectTmplCtx) Field() string {
	return m.getField()
}

func createmetricSubjectTmplCtx(metric telegraf.Metric) metricSubjectTmplCtx {
	return metricSubjectTmplCtx{
		Name: metric.Name(),
		getTag: func(key string) string {
			return metric.Tags()[key]
		},
		getField: func() string {
			fields := metric.FieldList()
			if len(fields) == 0 {
				return "emptyFields"
			}
			if len(fields) > 1 {
				return "tooManyFields"
			}

			return fields[0].Key
		},
	}
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
		nats.RetryOnFailedConnect(true),
		nats.ReconnectWait(2 * time.Second),

		// Handlers
		nats.DisconnectHandler(func(nc *nats.Conn) {
			n.Log.Infof("Disconnected from NATS: %v", nc.LastError())
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			n.Log.Infof("Reconnected to NATS at %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			n.Log.Error("Connection permanently closed.")
		}),
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
	n.Log.Debug("Initializing NATS output plugin")
	// If layout is enabled, we will use the subject as the
	// base of the template and add more tokens based on
	// the template.
	if len(n.SubjectLayout) > 0 {
		subParts := []string{n.Subject}
		subParts = append(subParts, n.SubjectLayout...)
		tpl, err := template.New("nats").Parse(strings.Join(subParts, "."))
		if err != nil {
			return fmt.Errorf("failed to parse subject template: %w", err)
		}
		n.tplSubject = *tpl

		if usesFieldField(tpl.Tree.Root) {
			n.Log.Info("Subject template set to include field name")
			n.includeFieldInSubject = true
		}

		tmpSubject := strings.Split(n.Subject, ".")
		if tmpSubject[len(tmpSubject)-1] != ".>" {
			// The base subject does not have a wildcard, so we need to add one
			// to support the dynamic fields.
			n.Subject = n.Subject + ".>"
		}
	}

	if n.Jetstream != nil {
		if strings.TrimSpace(n.Jetstream.Name) == "" {
			return errors.New("stream cannot be empty")
		}

		if n.Jetstream.AsyncAckTimeout == nil {
			to := config.Duration(5 * time.Second)
			n.Jetstream.AsyncAckTimeout = &to
		}

		if len(n.Jetstream.Subjects) == 0 {
			n.Jetstream.Subjects = []string{n.Subject}
		}
		// If the overall-subject is already present anywhere in the Jetstream subject we go from there,
		// otherwise we should append the overall-subject as the last element.
		if !choice.Contains(n.Subject, n.Jetstream.Subjects) {
			n.Jetstream.Subjects = append(n.Jetstream.Subjects, n.Subject)
		}
		var err error
		n.jetstreamStreamConfig, err = n.getJetstreamConfig()
		if err != nil {
			return fmt.Errorf("failed to parse jetstream config: %w", err)
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

	subMsgPairList := make([]subMsgPair, 0)
	var mc metricSubjectTmplCtx
	var bufSubject bytes.Buffer

	if len(n.SubjectLayout) == 0 {
		// Nothing custom to do here
		// All metrics will be sent to the default provided subject
		for _, metric := range metrics {
			subMsgPairList = append(subMsgPairList, subMsgPair{
				subject: n.Subject,
				metric:  metric,
			})
		}
	}

	if len(n.SubjectLayout) > 0 && !n.includeFieldInSubject {
		// The user indicated they want to use a subject layout
		// but they do not want to include the field in the subject name.
		// We can just send the metric as is to the custom subject
		for _, metric := range metrics {
			bufSubject.Reset()
			mc = createmetricSubjectTmplCtx(metric)
			err := n.tplSubject.Execute(&bufSubject, mc)
			if err != nil {
				return fmt.Errorf("failed to execute subject template: %w", err)
			}
			subMsgPairList = append(subMsgPairList, subMsgPair{
				subject: bufSubject.String(),
				metric:  metric,
			})
		}

	}

	if len(n.SubjectLayout) > 0 && n.includeFieldInSubject {
		// The user indicated they want to use a subject layout
		// and they want to include the field name in the subject name.
		// We need to split the metric into separate messages based on the field.
		for _, metric := range metrics {
			for _, field := range metric.FieldList() {
				bufSubject.Reset()
				metricCopy := splitMetricByField(metric, field.Key)
				mc = createmetricSubjectTmplCtx(metricCopy)
				err := n.tplSubject.Execute(&bufSubject, mc)
				if err != nil {
					return fmt.Errorf("failed to execute subject template: %w", err)
				}
				subMsgPairList = append(subMsgPairList, subMsgPair{
					subject: bufSubject.String(),
					metric:  metricCopy,
				})
			}
		}
	}

	var pafs []jetstream.PubAckFuture
	if n.Jetstream != nil && n.Jetstream.AsyncPublish {
		pafs = make([]jetstream.PubAckFuture, len(subMsgPairList))
	}

	for i, pair := range subMsgPairList {
		if strings.Contains(pair.subject, "..") {
			n.Log.Errorf("double dots are not allowed in the subject: %s, most likely caused by a missing value in the template with_subject_layout", pair.subject)
			continue
		}

		buf, err := n.serializer.Serialize(pair.metric)
		if err != nil {
			n.Log.Warnf("Could not serialize metric: %v", err)
			continue
		}

		//n.Log.Debugf("Publishing on Subject: %s, Metrics: %s", pair.subject, string(buf))
		if n.Jetstream != nil {
			if n.Jetstream.AsyncPublish {
				pafs[i], err = n.jetstreamClient.PublishAsync(pair.subject, buf, jetstream.WithExpectStream(n.Jetstream.Name))
			} else {
				_, err = n.jetstreamClient.Publish(context.Background(), pair.subject, buf, jetstream.WithExpectStream(n.Jetstream.Name))
			}
		} else {
			err = n.conn.Publish(pair.subject, buf)
		}
		if err != nil {
			return fmt.Errorf("failed to send NATS message: %w, subject: %s, metric: %s", err, pair.subject, string(buf))
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

func init() {
	outputs.Add("nats", func() telegraf.Output {
		return &NATS{}
	})
}

// Check the template for any references to `.Field`.
// If the template includes a `.Field` reference, we will need to split the metric
// into separate messages based on the field.
func usesFieldField(node parse.Node) bool {
	switch n := node.(type) {
	case *parse.ListNode:
		for _, sub := range n.Nodes {
			if usesFieldField(sub) {
				return true
			}
		}
	case *parse.ActionNode:
		return usesFieldField(n.Pipe)
	case *parse.PipeNode:
		for _, cmd := range n.Cmds {
			if usesFieldField(cmd) {
				return true
			}
		}
	case *parse.CommandNode:
		for _, arg := range n.Args {
			if usesFieldField(arg) {
				return true
			}
		}
	case *parse.FieldNode:
		// .Field will be represented as []string{"Field"}
		return len(n.Ident) == 1 && n.Ident[0] == "Field"
	}
	return false
}

// splitMetricByField will create a new metric that only contains the specified field.
// This is used when the user wants to include the field name in the subject.
func splitMetricByField(metric telegraf.Metric, field string) telegraf.Metric {
	metricCopy := metric.Copy()

	for _, f := range metric.FieldList() {
		if f.Key != field {
			// Remove all fields that are not the specified field
			metricCopy.RemoveField(f.Key)
		}
	}

	return metricCopy
}
