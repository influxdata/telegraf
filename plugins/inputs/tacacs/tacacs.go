//go:generate ../../../tools/readme_config_includer/generator
package tacacs

import (
	"context"
	_ "embed"
	"fmt"
	"sync"
	"time"

	"github.com/nwaples/tacplus"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Tacacs struct {
	Servers         []string        `toml:"servers"`
	Username        config.Secret   `toml:"username"`
	Password        config.Secret   `toml:"password"`
	Secret          config.Secret   `toml:"secret"`
	RemAddr         string          `toml:"request_ip"`
	ResponseTimeout config.Duration `toml:"response_timeout"`
	Log             telegraf.Logger `toml:"-"`
	client          tacplus.Client
}

//go:embed sample.conf
var sampleConfig string

func (t *Tacacs) SampleConfig() string {
	return sampleConfig
}

func (t *Tacacs) Init() error {
	if len(t.Servers) == 0 {
		t.Servers = []string{"127.0.0.1:49"}
	}

	t.client = tacplus.Client{}

	return nil
}

func (t *Tacacs) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	for _, server := range t.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			acc.AddError(t.pollServer(acc, server))
		}(server)
	}

	wg.Wait()
	return nil
}

func (t *Tacacs) pollServer(acc telegraf.Accumulator, server string) error {
	// Create the fields for this metric
	tags := map[string]string{"source": server}
	fields := make(map[string]interface{})

	secret, err := t.Secret.Get()
	if err != nil {
		return fmt.Errorf("getting secret failed: %w", err)
	}
	defer config.ReleaseSecret(secret)

	username, err := t.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	defer config.ReleaseSecret(username)

	password, err := t.Password.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %w", err)
	}
	defer config.ReleaseSecret(password)

	t.client.Addr = server
	t.client.ConnConfig = tacplus.ConnConfig{
		Secret: secret,
	}

	// Initialize the AuthStart data
	testAuthStart := &tacplus.AuthenStart{
		Action:        tacplus.AuthenActionLogin,
		AuthenType:    tacplus.AuthenTypeASCII,
		AuthenService: tacplus.AuthenServiceLogin,
		PrivLvl:       1,
		Port:          "heartbeat",
		RemAddr:       t.RemAddr,
	}

	// send the start request, the reply should be AuthenStatusGetUser
	ctx := context.Background()
	if t.ResponseTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(t.ResponseTimeout))
		defer cancel()
	}

	startTime := time.Now()
	reply, session, err := t.client.SendAuthenStart(ctx, testAuthStart)
	if err != nil {
		return fmt.Errorf("error on new tacacs authentication start request to %s : %w", server, err)
	}
	defer session.Close()
	// Check the returned status
	if reply.Status != tacplus.AuthenStatusGetUser {
		acc.AddFields("tacacs", fields, tags)
		return fmt.Errorf("error on new tacacs authentication start request to %s : Unexpected response code %d", server, reply.Status)
	}

	// Send the first Continue request with the username, the reply should be AuthenStatusGetPass
	reply, err = session.Continue(ctx, string(username))
	if err != nil {
		return fmt.Errorf("error on tacacs authentication continue username request to %s : %w", server, err)
	}
	if reply.Status != tacplus.AuthenStatusGetPass {
		return fmt.Errorf("error on first tacacs authentication continue username request to %s : Unexpected response code %d", server, reply.Status)
	}

	reply, err = session.Continue(ctx, string(password))
	if err != nil {
		return fmt.Errorf("error on second tacacs authentication continue password request to %s : %w", server, err)
	}
	duration := time.Since(startTime)
	if reply.Status != tacplus.AuthenStatusPass {
		acc.AddFields("tacacs", fields, tags)
		return fmt.Errorf("error on second tacacs authentication continue password request to %s : Unexpected response code %d", server, reply.Status)
	}

	fields["responsetime_ms"] = duration.Milliseconds()
	acc.AddFields("tacacs", fields, tags)
	return nil
}

func init() {
	inputs.Add("tacacs", func() telegraf.Input {
		return &Tacacs{RemAddr: "127.0.0.1", ResponseTimeout: config.Duration(time.Second * 5)}
	})
}
