package qsys_qrc

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

// Qsys_QRC holds configuration for the plugin
type Qsys_QRC struct {
	// Q-SYS core to target
	Core string

	client net.Conn
}

// Description will appear directly above the plugin definition in the config file
func (m *Qsys_QRC) Description() string {
	return `Retrieve Named Controls from a QSC Q-SYS core`
}

var exampleConfig = `
  ## Specify the core address and port
  core = "localhost:1710"
`

// SampleConfig will populate the sample configuration portion of the plugin's configuration
func (m *Qsys_QRC) SampleConfig() string {
	return exampleConfig
}

func (m *Qsys_QRC) init(acc telegraf.Accumulator) error {

	client, connerr := net.Dial("tcp", m.Core)
	if connerr != nil {
		return connerr
	}
	m.client = client
	return nil
}

// Gather defines what data the plugin will gather.
func (m *Qsys_QRC) Gather(acc telegraf.Accumulator) error {
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
	} else {
		tags["server"] = m.Core
		fields["state"] = status.State
		fields["design"] = status.DesignName
		fields["status"] = status.Status.String
		acc.AddFields("qsys", fields, tags)
	}
	return nil
}

// makeRPCCall sends a JSONRPC message to the core and blocks until it returns or times out
func (m *Qsys_QRC) makeRPCCall(request *JSONRPC) ([]byte, error) {
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
		return nil, errors.New("did not receive RPC reply in time")
	}
}

func (m *Qsys_QRC) queryStatus() (EngineStatusReplyData, error) {
	id := rand.Int31()
	request := JSONRPC{
		Version: "2.0",
		Method:  "StatusGet",
		Params:  0,
		ID:      id,
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

func init() {
	inputs.Add("qsys", func() telegraf.Input {
		return &Qsys_QRC{}
	})
}
