//go:generate ../../../tools/config_includer/generator
//go:generate ../../../tools/readme_config_includer/generator
package promql

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type PromQL struct {
	URL            string          `toml:"url"`
	Username       config.Secret   `toml:"username"`
	Password       config.Secret   `toml:"password"`
	Token          config.Secret   `toml:"token"`
	Timeout        config.Duration `toml:"timeout"`
	InstantQueries []InstantQuery  `toml:"instant"`
	RangeQueries   []RangeQuery    `toml:"range"`
	Log            telegraf.Logger `toml:"-"`
	common_http.TransportConfig

	client *client
}

func (*PromQL) SampleConfig() string {
	return sampleConfig
}

func (p *PromQL) Init() error {
	// Check settings
	if p.URL == "" {
		return errors.New("'url' cannot be empty")
	}

	if p.Username.Empty() && !p.Password.Empty() {
		return errors.New("expecting username for basic authentication")
	}

	if !p.Username.Empty() && !p.Token.Empty() {
		return errors.New("cannot use both basic and bearer authentication")
	}

	if len(p.InstantQueries)+len(p.RangeQueries) == 0 {
		return errors.New("no queries configured")
	}

	// Setup the API client
	p.client = &client{
		url:      p.URL,
		username: p.Username,
		password: p.Password,
		token:    p.Token,
		cfg:      p.TransportConfig,
	}

	var opts []apiv1.Option
	if p.Timeout > 0 {
		opts = append(opts, apiv1.WithTimeout(time.Duration(p.Timeout)))
	}

	// Setup queries
	for i := range p.InstantQueries {
		if err := p.InstantQueries[i].init(p.client, p.Log, opts...); err != nil {
			return err
		}
	}
	for i := range p.RangeQueries {
		if err := p.RangeQueries[i].init(p.client, p.Log, opts...); err != nil {
			return err
		}
	}

	return nil
}

func (p *PromQL) Start(telegraf.Accumulator) error {
	// Initialize the API client
	c, err := p.client.init()
	if err != nil {
		return fmt.Errorf("initializing API client failed: %w", err)
	}
	p.client = c

	return nil
}

func (p *PromQL) Stop() {
	if p.client != nil {
		p.client.close()
	}
}

func (p *PromQL) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(p.Timeout))
		defer cancel()
	}

	t := time.Now()

	// Do the queries
	for _, q := range p.InstantQueries {
		acc.AddError(q.execute(ctx, acc, t))
	}
	for _, q := range p.RangeQueries {
		acc.AddError(q.execute(ctx, acc, t))
	}

	return nil
}

func init() {
	inputs.Add("promql", func() telegraf.Input {
		return &PromQL{
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
