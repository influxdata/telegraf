//go:generate ../../../tools/readme_config_includer/generator
package radius

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Radius struct {
	Servers         []string        `toml:"servers"`
	Username        config.Secret   `toml:"username"`
	Password        config.Secret   `toml:"password"`
	Secret          config.Secret   `toml:"secret"`
	ResponseTimeout config.Duration `toml:"response_timeout"`
	Log             telegraf.Logger `toml:"-"`
	client          radius.Client
}

//go:embed sample.conf
var sampleConfig string

func (r *Radius) SampleConfig() string {
	return sampleConfig
}

func (r *Radius) Init() error {
	if len(r.Servers) == 0 {
		r.Servers = []string{"127.0.0.1:1812"}
	}

	r.client = radius.Client{
		Retry: 0,
	}

	return nil
}

func (r *Radius) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	for _, server := range r.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			acc.AddError(r.pollServer(acc, server))
		}(server)
	}

	wg.Wait()
	return nil
}

func (r *Radius) pollServer(acc telegraf.Accumulator, server string) error {
	// Create the fields for this metric
	host, port, err := net.SplitHostPort(server)
	if err != nil {
		return fmt.Errorf("splitting host and port failed: %w", err)
	}
	tags := map[string]string{"source": host, "source_port": port}
	fields := make(map[string]interface{})

	secret, err := r.Secret.Get()
	if err != nil {
		return fmt.Errorf("getting secret failed: %w", err)
	}
	defer config.ReleaseSecret(secret)

	username, err := r.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	defer config.ReleaseSecret(username)

	password, err := r.Password.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %w", err)
	}
	defer config.ReleaseSecret(password)

	// Create the radius packet with PAP authentication
	packet := radius.New(radius.CodeAccessRequest, secret)
	err = rfc2865.UserName_Set(packet, username)
	if err != nil {
		return fmt.Errorf("setting username for radius auth failed: %w", err)
	}
	err = rfc2865.UserPassword_Set(packet, password)
	if err != nil {
		return fmt.Errorf("setting password for radius auth failed: %w", err)
	}

	// Do the radius request
	ctx := context.Background()
	if r.ResponseTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(r.ResponseTimeout))
		defer cancel()
	}

	startTime := time.Now()
	response, err := r.client.Exchange(ctx, packet, server)
	duration := time.Since(startTime)

	if err != nil {
		if !errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		fields["responsetime_ms"] = time.Duration(r.ResponseTimeout).Milliseconds()
		tags["response_code"] = "timeout"
	} else if response.Code != radius.CodeAccessAccept {
		fields["responsetime_ms"] = time.Duration(r.ResponseTimeout).Milliseconds()
		tags["response_code"] = response.Code.String()
	} else {
		fields["responsetime_ms"] = duration.Milliseconds()
		tags["response_code"] = response.Code.String()
	}

	acc.AddFields("radius", fields, tags)
	return nil
}

func init() {
	inputs.Add("radius", func() telegraf.Input {
		return &Radius{ResponseTimeout: config.Duration(time.Second * 5)}
	})
}
