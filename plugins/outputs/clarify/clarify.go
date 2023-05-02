package clarify

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/views"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Clarify struct {
	Username        string          `toml:"username"`
	Password        string          `toml:"password"`
	CredentialsFile string          `toml:"credentials_file"`
	IDTags          []string        `toml:"id_tags"`
	Log             telegraf.Logger `toml:"-"`

	client *clarify.Client
}

//go:embed sample.conf
var sampleConfig string

func (c *Clarify) Connect() error {
	ctx := context.Background()
	if c.CredentialsFile != "" {
		creds, err := clarify.CredentialsFromFile(c.CredentialsFile)
		if err != nil {
			return err
		}
		c.client = creds.Client(ctx)
		return nil
	}
	if c.Username != "" && c.Password != "" {
		creds := clarify.BasicAuthCredentials(c.Username, c.Password)
		c.client = creds.Client(ctx)
		return nil
	}
	return fmt.Errorf("no Clarify credentials provided")
}

func verifyValue(v interface{}) (float64, error) {
	var value float64
	switch v := v.(type) {
	case bool:
		value = float64(0)
		if v {
			value = float64(1)
		}
	case uint8:
		value = float64(v)
	case uint16:
		value = float64(v)
	case uint32:
		value = float64(v)
	case uint64:
		value = float64(v)
	case int8:
		value = float64(v)
	case int16:
		value = float64(v)
	case int32:
		value = float64(v)
	case int64:
		value = float64(v)
	case float32:
		value = float64(v)
	case float64:
		value = v
	default:
		return value, fmt.Errorf("unsupported field type: %T", v)
	}
	return value, nil
}

func (c *Clarify) Write(metrics []telegraf.Metric) error {
	signals := make(map[string]views.SignalSave)
	frame := views.DataFrame{}

	for _, m := range metrics {
		for _, f := range m.FieldList() {
			if value, err := verifyValue(f.Value); err == nil {
				id := c.generateID(m, f)
				ts := fields.AsTimestamp(m.Time())

				if _, ok := frame[id]; ok {
					frame[id][ts] = value
				} else {
					frame[id] = views.DataSeries{ts: value}
				}

				s := views.SignalSave{}
				s.Name = fmt.Sprintf("%s.%s", m.Name(), f.Key)

				for _, t := range m.TagList() {
					labelName := strings.ReplaceAll(t.Key, " ", "-")
					labelName = strings.ReplaceAll(labelName, "_", "-")
					labelName = strings.ToLower(labelName)
					s.Labels.Add(labelName, t.Value)
				}

				signals[id] = s
			} else {
				c.Log.Infof("Unable to add field `%s` for metric `%s` due to error '%v', skipping", f.Key, m.Name(), err)
			}
		}
	}

	if _, err := c.client.Insert(frame).Do(context.Background()); err != nil {
		return err
	}

	if _, err := c.client.SaveSignals(signals).Do(context.Background()); err != nil {
		return err
	}

	return nil
}

func (c *Clarify) generateID(m telegraf.Metric, f *telegraf.Field) string {
	var id string
	cid, exist := m.GetTag("clarify_input_id")
	if exist && len(m.FieldList()) == 1 {
		id = cid
	} else {
		id = fmt.Sprintf("%s.%s", m.Name(), f.Key)
		for _, idTag := range c.IDTags {
			if m.HasTag(idTag) {
				id = fmt.Sprintf("%s.%s", id, m.Tags()[idTag])
			}
		}
	}
	return strings.ToLower(id)
}

func (c *Clarify) SampleConfig() string {
	return sampleConfig
}

func (c *Clarify) Close() error {
	c.client = nil
	return nil
}

func init() {
	outputs.Add("clarify", func() telegraf.Output {
		return &Clarify{}
	})
}
