package qsc_qsys

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"math/rand"
	"net"
	"time"
)

// QSC_QSYS holds configuration for the plugin
type QSC_QSYS struct {
	// Q-SYS core to target
	Server        string
	NamedControls []string `toml:"named_controls"`
	Username      string
	PIN           string

	client net.Conn
}

// Description will appear directly above the plugin definition in the config file
func (m *QSC_QSYS) Description() string {
	return `Retrieve Named Controls and status from a QSC Q-SYS core`
}

var exampleConfig = `
  ## Specify the core address and port
  server = "localhost:1710"

  ## If the core is set up with user accounts set the username and PIN to use
  # username = ""
  # pin = ""

  ## If desired, an array of named controls can be collected
  # named_controls = []
`

// SampleConfig will populate the sample configuration portion of the plugin's configuration
func (m *QSC_QSYS) SampleConfig() string {
	return exampleConfig
}

func (m *QSC_QSYS) init(acc telegraf.Accumulator) error {
	client, connerr := net.Dial("tcp", m.Server)
	if connerr != nil {
		return connerr
	}
	m.client = client
	if m.Username != "" {
		logonerr := m.logon()
		if logonerr != nil {
			return logonerr
		}
	}
	return nil
}

// Gather defines what data the plugin will gather.
func (m *QSC_QSYS) Gather(acc telegraf.Accumulator) error {
	if m.client == nil {
		connerr := m.init(acc)
		if connerr != nil {
			return connerr
		}
	}

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	status, statuserr := m.queryStatus()
	if statuserr != nil {
		acc.AddError(statuserr)
		return statuserr
	}
	tags["server"] = m.Server
	tags["platform"] = status.Platform
	tags["design"] = status.DesignName

	fields["state"] = status.State
	fields["status"] = status.Status.String

	controls, controlerr := m.gatherControls()
	if controlerr != nil {
		acc.AddError(controlerr)
		return controlerr
	}
	for _, control := range controls {
		fields[control.Name] = control.Value
	}

	acc.AddFields("qsys", fields, tags)
	return nil
}

// makeRPCCall sends a JSONRPC message to the core and blocks until it returns or times out
func (m *QSC_QSYS) makeRPCCall(request *JSONRPC) ([]byte, error) {
	// If the upcoming request does not have an ID set, set one to distinguish its reply
	if request.ID == 0 {
		id := rand.Int31()
		request.ID = id
	}

	// Build null-terminated JSON string
	requestString, _ := json.Marshal(request)
	fmt.Fprintf(m.client, "%s\x00", requestString)
	// The core may take some time to reply or send unsolicited other messages,
	// so we have to handle returns async and possibly filter out some data.
	// Make a channel to pass the proper return through, or timeout waiting for it.
	replyWaiter := make(chan []byte, 1)
	defer close(replyWaiter)

	replyReader := bufio.NewReader(m.client)
	go func() {
		for {
			replyString, _ := replyReader.ReadBytes(0)
			// Strip the trailing nullchar that the core sends back
			unterminatedReplyString := replyString[:len(replyString)-1]

			var reply JSONRPC
			jsonerr := json.Unmarshal(unterminatedReplyString, &reply)
			if jsonerr != nil {
				fmt.Printf("Error unmarshaling JSON: %s", jsonerr)
			} else {
				if reply.ID == request.ID {
					replyWaiter <- unterminatedReplyString
					return
				}
			}
		}
	}()

	select {
	case replyString := <-replyWaiter:
		return replyString, nil
	case <-time.After(10 * time.Second):
		return nil, errors.New("did not receive RPC reply to " + request.Method + " in time")
	}
}

func (m *QSC_QSYS) gatherControls() ([]ControlValue, error) {
	request := JSONRPC{
		Version: "2.0",
		Method:  "Control.Get",
		Params:  m.NamedControls,
	}
	replyString, rpcerror := m.makeRPCCall(&request)
	var reply ControlGetReply
	if rpcerror != nil {
		return reply.Result, rpcerror
	}
	jsonerr := json.Unmarshal(replyString, &reply)
	if jsonerr != nil {
		fmt.Printf("Error unmarshaling JSON: %s", jsonerr)
		return reply.Result, jsonerr
	}
	if reply.Error.Code != 0 {
		return reply.Result, errors.New(reply.Error.Message)
	}
	return reply.Result, nil
}

func (m *QSC_QSYS) queryStatus() (EngineStatusReplyData, error) {
	request := JSONRPC{
		Version: "2.0",
		Method:  "StatusGet",
		Params:  0,
	}
	replyString, rpcerror := m.makeRPCCall(&request)
	var reply EngineStatusReply
	if rpcerror != nil {
		return reply.Result, rpcerror
	}
	jsonerr := json.Unmarshal(replyString, &reply)
	if jsonerr != nil {
		fmt.Printf("Error unmarshaling JSON: %s", jsonerr)
		return reply.Result, jsonerr
	}
	return reply.Result, nil
}

func (m *QSC_QSYS) logon() error {
	request := JSONRPC{
		Version: "2.0",
		Method:  "Logon",
		Params: map[string]string{
			"User":     m.Username,
			"Password": m.PIN,
		},
	}
	_, err := m.makeRPCCall(&request)
	return err
}

func init() {
	inputs.Add("qsys", func() telegraf.Input {
		return &QSC_QSYS{}
	})
}
