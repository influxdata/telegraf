//go:generate ../../../tools/readme_config_includer/generator
package radius

import (
	"context"
	_ "embed"
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
	RadiusClient    radius.Client
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

	r.RadiusClient = radius.Client{
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
	tags := map[string]string{"source": host, "port": port}
	fields := make(map[string]interface{})

	secret, err := r.Secret.Get()
	if err != nil {
		return fmt.Errorf("getting secret failed: %v", err)
	}
	defer config.ReleaseSecret(secret)

	username, err := r.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %v", err)
	}
	defer config.ReleaseSecret(username)

	password, err := r.Password.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %v", err)
	}
	defer config.ReleaseSecret(password)

	// Create the radius packet with PAP authentication
	packet := radius.New(radius.CodeAccessRequest, secret)
	rfc2865.UserName_Set(packet, username)
	rfc2865.UserPassword_Set(packet, password)

	// Do the radius request
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.ResponseTimeout))
	defer cancel()
	startTime := time.Now()
	response, err := r.RadiusClient.Exchange(ctx, packet, server)
	duration := time.Since(startTime)

	if err != nil {
		r.Log.Warnf("error on new request to %s : %s", server, err)
		fields["responsetime"] = time.Duration(r.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(r.ResponseTimeout).Milliseconds()
	} else if response.Code != radius.CodeAccessAccept {
		r.Log.Warnf("Got radius return code: %d", response.Code)
		fields["responsetime"] = time.Duration(r.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(r.ResponseTimeout).Milliseconds()
	} else {
		fields["responsetime"] = duration.Seconds()
		fields["responsetime_ms"] = duration.Milliseconds()
	}

	acc.AddFields("radius", fields, tags)
	return nil
}

func init() {
	inputs.Add("radius", func() telegraf.Input {
		return &Radius{ResponseTimeout: config.Duration(time.Second * 5)}
	})
}
