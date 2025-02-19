//go:generate ../../../tools/readme_config_includer/generator

package clarify

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/views"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Clarify struct {
	Username        config.Secret   `toml:"username"`
	Password        config.Secret   `toml:"password"`
	CredentialsFile string          `toml:"credentials_file"`
	Timeout         config.Duration `toml:"timeout"`
	IDTags          []string        `toml:"id_tags"`
	ClarifyIDTag    string          `toml:"clarify_id_tag"`
	Log             telegraf.Logger `toml:"-"`

	client *clarify.Client
}

var errIDTooLong = errors.New("id too long (>128)")
var errCredentials = errors.New("only credentials_file OR username/password can be specified")

const defaultTimeout = config.Duration(20 * time.Second)
const allowedIDRunes = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_:.#+/`

//go:embed sample.conf
var sampleConfig string

func (c *Clarify) Init() error {
	if c.Timeout <= 0 {
		c.Timeout = defaultTimeout
	}
	// Not blocking as it doesn't do any http requests, just sets up the necessary Oauth2 client.
	ctx := context.Background()
	switch {
	case c.CredentialsFile != "":
		if !c.Username.Empty() || !c.Password.Empty() {
			return errCredentials
		}
		creds, err := clarify.CredentialsFromFile(c.CredentialsFile)
		if err != nil {
			return err
		}
		c.client = creds.Client(ctx)
		return nil
	case !c.Username.Empty() && !c.Password.Empty():
		username, err := c.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		password, err := c.Password.Get()
		if err != nil {
			username.Destroy()
			return fmt.Errorf("getting password failed: %w", err)
		}
		creds := clarify.BasicAuthCredentials(username.String(), password.String())
		username.Destroy()
		password.Destroy()
		c.client = creds.Client(ctx)
		return nil
	}
	return errors.New("no credentials provided")
}

func (*Clarify) Connect() error {
	return nil
}

func (c *Clarify) Write(metrics []telegraf.Metric) error {
	frame, signals := c.processMetrics(metrics)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.Timeout))
	defer cancel()

	if _, err := c.client.Insert(frame).Do(ctx); err != nil {
		return fmt.Errorf("inserting failed: %w", err)
	}

	if _, err := c.client.SaveSignals(signals).Do(ctx); err != nil {
		return fmt.Errorf("saving signals failed: %w", err)
	}

	return nil
}

func (c *Clarify) processMetrics(metrics []telegraf.Metric) (views.DataFrame, map[string]views.SignalSave) {
	signals := make(map[string]views.SignalSave)
	frame := views.DataFrame{}

	for _, m := range metrics {
		for _, f := range m.FieldList() {
			value, err := internal.ToFloat64(f.Value)
			if err != nil {
				c.Log.Warnf("Skipping field %q of metric %q: %s", f.Key, m.Name(), err.Error())
				continue
			}
			id, err := c.generateID(m, f)
			if err != nil {
				c.Log.Warnf("Skipping field %q of metric %q: %s", f.Key, m.Name(), err.Error())
				continue
			}
			ts := fields.AsTimestamp(m.Time())

			if _, ok := frame[id]; ok {
				frame[id][ts] = value
			} else {
				frame[id] = views.DataSeries{ts: value}
			}

			s := views.SignalSave{}
			s.Name = m.Name() + "." + f.Key

			for _, t := range m.TagList() {
				labelName := strings.ReplaceAll(t.Key, " ", "-")
				labelName = strings.ReplaceAll(labelName, "_", "-")
				labelName = strings.ToLower(labelName)
				s.Labels.Add(labelName, t.Value)
			}

			signals[id] = s
		}
	}
	return frame, signals
}

func normalizeID(id string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(allowedIDRunes, r) {
			return r
		}
		return '_'
	}, id)
}

func (c *Clarify) generateID(m telegraf.Metric, f *telegraf.Field) (string, error) {
	var id string
	if c.ClarifyIDTag != "" {
		if cid, exist := m.GetTag(c.ClarifyIDTag); exist && len(m.FieldList()) == 1 {
			id = cid
		}
	}
	if id == "" {
		parts := make([]string, 0, len(c.IDTags)+2)
		parts = append(parts, m.Name(), f.Key)

		for _, idTag := range c.IDTags {
			if k, found := m.GetTag(idTag); found {
				parts = append(parts, k)
			}
		}
		id = strings.Join(parts, ".")
	}
	id = normalizeID(id)
	if len(id) > 128 {
		return id, errIDTooLong
	}
	return id, nil
}

func (*Clarify) SampleConfig() string {
	return sampleConfig
}

func (c *Clarify) Close() error {
	c.client = nil
	return nil
}

func init() {
	outputs.Add("clarify", func() telegraf.Output {
		return &Clarify{
			Timeout: defaultTimeout,
		}
	})
}
