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
	RemAddr         string          `toml:"remaddr"`
	ResponseTimeout config.Duration `toml:"response_timeout"`
	Log             telegraf.Logger `toml:"-"`
}

//go:embed sample.conf
var sampleConfig string

func (t *Tacacs) SampleConfig() string {
	return sampleConfig
}

func (t *Tacacs) Init() error {
	if len(t.Servers) == 0 {
		t.Servers = []string{"127.0.0.1"}
	}
	if t.ResponseTimeout < config.Duration(time.Second) {
		t.ResponseTimeout = config.Duration(time.Second * 5)
	}
	return nil
}

func (t *Tacacs) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	for _, server := range t.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			acc.AddError(t.pollServer(server, acc))
		}(server)
	}

	wg.Wait()
	return nil
}

func (t *Tacacs) pollServer(server string, acc telegraf.Accumulator) error {
	// Create the fields for this metric
	tags := map[string]string{"source": server}
	fields := make(map[string]interface{})

	secret, err := t.Secret.Get()
	if err != nil {
		return fmt.Errorf("getting secret failed: %v", err)
	}
	username, err := t.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %v", err)
	}
	password, err := t.Password.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %v", err)
	}

	// Create the tacacs client
	client := &tacplus.Client{
		Addr: server,
		ConnConfig: tacplus.ConnConfig{
			Secret: secret,
		},
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t.ResponseTimeout))
	defer cancel()
	startTime := time.Now()
	reply, session, err := client.SendAuthenStart(ctx, testAuthStart)
	if err != nil {
		t.Log.Warnf("error on new tacacs authentication start request to %s : %s", server, err)
		fields["responsetime"] = time.Duration(t.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(t.ResponseTimeout).Milliseconds()
		acc.AddFields("tacacs", fields, tags)
		return nil
	}
	defer session.Close()
	// Check the returned status
	if reply.Status != tacplus.AuthenStatusGetUser {
		t.Log.Warnf("error on new tacacs authentication start request to %s : Unexpected response code %d", server, reply.Status)
		fields["responsetime"] = time.Duration(t.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(t.ResponseTimeout).Milliseconds()
		acc.AddFields("tacacs", fields, tags)
		return nil
	}

	// Send the first Continue request with the username, the reply should be AuthenStatusGetPass
	reply, err = session.Continue(ctx, string(username))
	if err != nil {
		t.Log.Warnf("error on new tacacs authentication continue username request to %s : %s", server, err)
		fields["responsetime"] = time.Duration(t.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(t.ResponseTimeout).Milliseconds()
		acc.AddFields("tacacs", fields, tags)
		return nil
	}
	if reply.Status != tacplus.AuthenStatusGetPass {
		t.Log.Warnf("error on first tacacs continue username request to %s : Unexpected response code %d", server, reply.Status)
		fields["responsetime"] = time.Duration(t.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(t.ResponseTimeout).Milliseconds()
		acc.AddFields("tacacs", fields, tags)
		return nil
	}

	reply, err = session.Continue(ctx, string(password))
	if err != nil {
		t.Log.Warnf("error on new tacacs authentication continue password request to %s : %s", server, err)
		fields["responsetime"] = time.Duration(t.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(t.ResponseTimeout).Milliseconds()
		acc.AddFields("tacacs", fields, tags)
		return nil
	}
	duration := time.Since(startTime)

	if reply.Status != tacplus.AuthenStatusPass {
		t.Log.Warnf("Got tacacs status code: %d", reply.Status)
		fields["responsetime"] = time.Duration(t.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(t.ResponseTimeout).Milliseconds()
	} else {
		fields["responsetime"] = duration.Seconds()
		fields["responsetime_ms"] = duration.Milliseconds()
	}

	acc.AddFields("tacacs", fields, tags)
	return nil
}

func init() {
	inputs.Add("tacacs", func() telegraf.Input {
		return &Tacacs{}
	})
}
