//go:generate ../../../tools/config_includer/generator
//go:generate ../../../tools/readme_config_includer/generator
package logql

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type LogQL struct {
	URL            string          `toml:"url"`
	Username       config.Secret   `toml:"username"`
	Password       config.Secret   `toml:"password"`
	Token          config.Secret   `toml:"token"`
	Organizations  []string        `toml:"organizations"`
	Timeout        config.Duration `toml:"timeout"`
	InstantQueries []InstantQuery  `toml:"instant"`
	RangeQueries   []RangeQuery    `toml:"range"`
	Log            telegraf.Logger `toml:"-"`
	common_http.TransportConfig

	client *client
}

func (*LogQL) SampleConfig() string {
	return sampleConfig
}

func (l *LogQL) Init() error {
	// Check settings
	if l.URL == "" {
		l.URL = "http://localhost:3100"
	}

	if l.Username.Empty() && !l.Password.Empty() {
		return errors.New("expecting username for basic authentication")
	}

	if !l.Username.Empty() && !l.Token.Empty() {
		return errors.New("cannot use both basic and bearer authentication")
	}

	if len(l.InstantQueries)+len(l.RangeQueries) == 0 {
		return errors.New("no queries configured")
	}

	// Setup the API client
	l.client = &client{
		url:      l.URL,
		username: l.Username,
		password: l.Password,
		token:    l.Token,
		org:      strings.Join(l.Organizations, "|"), // see https://grafana.com/docs/loki/latest/operations/multi-tenancy
		cfg:      l.TransportConfig,
		timeout:  time.Duration(l.Timeout),
	}

	// Setup queries
	for i := range l.InstantQueries {
		if err := l.InstantQueries[i].init(l.client, l.Log); err != nil {
			return err
		}
	}
	for i := range l.RangeQueries {
		if err := l.RangeQueries[i].init(l.client, l.Log); err != nil {
			return err
		}
	}

	return nil
}

func (l *LogQL) Start(telegraf.Accumulator) error {
	// Initialize the API client
	if err := l.client.init(); err != nil {
		return fmt.Errorf("initializing API client failed: %w", err)
	}

	return nil
}

func (l *LogQL) Stop() {
	if l.client != nil {
		l.client.close()
	}
}

func (l *LogQL) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()
	if l.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(l.Timeout))
		defer cancel()
	}

	// Check if the server is ready
	if ready, msg, err := l.client.ready(ctx); err != nil {
		return fmt.Errorf("checking readiness failed: %w", err)
	} else if !ready {
		return fmt.Errorf("server at %q is not ready: %s", l.URL, msg)
	}

	t := time.Now()

	// Do the queries
	for _, q := range l.InstantQueries {
		acc.AddError(q.execute(ctx, acc, t))
	}
	for _, q := range l.RangeQueries {
		acc.AddError(q.execute(ctx, acc, t))
	}

	return nil
}

func init() {
	inputs.Add("logql", func() telegraf.Input {
		return &LogQL{
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
