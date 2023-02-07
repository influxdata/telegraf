package radius

import (
	"context"
	_ "embed"
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
	Servers         []string
	Username        string
	Password        string
	Secret          string
	ResponseTimeout config.Duration
	Log             telegraf.Logger
}

//go:embed sample.conf
var sampleConfig string

func (n *Radius) SampleConfig() string {
	return sampleConfig
}

func (n *Radius) Description() string {
	return "Test Radius authentication"
}

func (n *Radius) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if len(n.Servers) == 0 {
		n.Servers = []string{"127.0.0.1:1812"}
	}
	if n.ResponseTimeout < config.Duration(time.Second) {
		n.ResponseTimeout = config.Duration(time.Second * 5)
	}

	for _, server := range n.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			acc.AddError(n.pollServer(server, acc))
		}(server)
	}

	wg.Wait()
	return nil
}

func (n *Radius) pollServer(server string, acc telegraf.Accumulator) error {
	// Create the fields for this metric
	host, port, err := net.SplitHostPort(server)
	tags := map[string]string{"server": host, "port": port}
	fields := make(map[string]interface{})

	// Create the radius Client
	var client = &radius.Client{
		Retry: 0,
	}

	// Create the radius packet with PAP authentication
	packet := radius.New(radius.CodeAccessRequest, []byte(n.Secret))
	rfc2865.UserName_SetString(packet, n.Username)
	rfc2865.UserPassword_SetString(packet, n.Password)

	// Do the radius request
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(n.ResponseTimeout))
	defer cancel()
	startTime := time.Now()
	response, err := client.Exchange(ctx, packet, server)
	duration := time.Since(startTime)

	if err != nil {
		n.Log.Warnf("error on new request to %s : %s", server, err)
		fields["responsetime"] = time.Duration(n.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(n.ResponseTimeout).Milliseconds()
	} else if response.Code != radius.CodeAccessAccept {
		n.Log.Warnf("Got radius return code: %d", response.Code)
		fields["responsetime"] = time.Duration(n.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(n.ResponseTimeout).Milliseconds()
	} else {
		fields["responsetime"] = duration.Seconds()
		fields["responsetime_ms"] = duration.Milliseconds()
	}

	acc.AddFields("radius", fields, tags)
	return nil
}

func init() {
	inputs.Add("radius", func() telegraf.Input {
		return &Radius{}
	})
}
