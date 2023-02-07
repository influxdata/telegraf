package tacacs

import (
	"context"
	"sync"
	"time"

	"github.com/nwaples/tacplus"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Tacacs struct {
	Servers         []string
	Username        string
	Password        string
	RemAddr         string
	Secret          string
	ResponseTimeout config.Duration
	Log             telegraf.Logger
}

var sampleConfig = `
  ## An array of Server IPs to gather from, default localhost
  servers = ["127.0.0.1"]

  ## Request source server IP, normally the server running telegraf
  remaddr = "127.0.0.1"

  ## Credentials for tacacs authentication.
  # username = "myuser"
  # password = "mypassword"
  # secret = "mysecret"

  ## Maximum time to receive response.
  # response_timeout = "5s"
`

func (n *Tacacs) SampleConfig() string {
	return sampleConfig
}

func (n *Tacacs) Description() string {
	return "Test Tacacs authentication"
}

func (n *Tacacs) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if len(n.Servers) == 0 {
		n.Servers = []string{"127.0.0.1"}
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

func (n *Tacacs) pollServer(server string, acc telegraf.Accumulator) error {
	// Create the fields for this metric
	tags := map[string]string{"server": server}
	fields := make(map[string]interface{})

	// Create the tacacs client
	client := &tacplus.Client{
		Addr: server,
		ConnConfig: tacplus.ConnConfig{
			Secret: []byte(n.Secret),
		},
	}

	// Initialize the AuthStart data
	testAuthStart := &tacplus.AuthenStart{
		Action:        tacplus.AuthenActionLogin,
		AuthenType:    tacplus.AuthenTypeASCII,
		AuthenService: tacplus.AuthenServiceLogin,
		PrivLvl:       1,
		Port:          "heartbeat",
		RemAddr:       n.RemAddr,
	}

	// send the start request, the reply should be AuthenStatusGetUser
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(n.ResponseTimeout))
	defer cancel()
	startTime := time.Now()
	reply, session, err := client.SendAuthenStart(ctx, testAuthStart)
	if err != nil {
		n.Log.Warnf("error on new tacacs authentication start request to %s : %s", server, err)
		fields["responsetime"] = time.Duration(n.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(n.ResponseTimeout).Milliseconds()
		acc.AddFields("tacacs", fields, tags)
		return nil
	}
	defer session.Close()
	// Check the returned status
	if reply.Status != tacplus.AuthenStatusGetUser {
		n.Log.Warnf("error on new tacacs authentication start request to %s : Unexpected response code %d", server, reply.Status)
		fields["responsetime"] = time.Duration(n.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(n.ResponseTimeout).Milliseconds()
		acc.AddFields("tacacs", fields, tags)
		return nil
	}

	// Send the first Continue request with the username, the reply should be AuthenStatusGetPass
	reply, err = session.Continue(ctx, n.Username)
	if err != nil {
		n.Log.Warnf("error on new tacacs authentication continue username request to %s : %s", server, err)
		fields["responsetime"] = time.Duration(n.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(n.ResponseTimeout).Milliseconds()
		acc.AddFields("tacacs", fields, tags)
		return nil
	}
	if reply.Status != tacplus.AuthenStatusGetPass {
		n.Log.Warnf("error on first tacacs continue username request to %s : Unexpected response code %d", server, reply.Status)
		fields["responsetime"] = time.Duration(n.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(n.ResponseTimeout).Milliseconds()
		acc.AddFields("tacacs", fields, tags)
		return nil
	}

	reply, err = session.Continue(ctx, n.Password)
	if err != nil {
		n.Log.Warnf("error on new tacacs authentication continue password request to %s : %s", server, err)
		fields["responsetime"] = time.Duration(n.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(n.ResponseTimeout).Milliseconds()
		acc.AddFields("tacacs", fields, tags)
		return nil
	}
	duration := time.Since(startTime)

	if reply.Status != tacplus.AuthenStatusPass {
		n.Log.Warnf("Got tacacs status code: %d", reply.Status)
		fields["responsetime"] = time.Duration(n.ResponseTimeout).Seconds()
		fields["responsetime_ms"] = time.Duration(n.ResponseTimeout).Milliseconds()
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
