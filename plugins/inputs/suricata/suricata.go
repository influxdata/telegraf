package suricata

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/nytlabs/gojsonexplode"
)

// Suricata is a Telegraf input plugin for Suricata runtime statistics.
type Suricata struct {
	sync.Mutex

	Source        string `toml:"source"`
	InputListener *net.UnixListener
	JSON          []byte
	CloseChan     chan bool
	ClosedChan    chan bool
}

// Description returns the plugin description.
func (s *Suricata) Description() string {
	return "Suricata stats plugin"
}

const sampleConfig = `
  ## Data sink for Suricata stats log
  # This is expected to be a filename of a
  # unix socket to be created for listening.
  # Will be overwritten if a socket or file
  # with that name already exists.
  source = "/tmp/suricata-stats.sock"
`

// SampleConfig returns a sample TOML section to illustrate configuration
// options.
func (s *Suricata) SampleConfig() string {
	return sampleConfig
}

// Start initiates background collection of JSON data from the socket
// provided to Suricata.
func (s *Suricata) Start(acc telegraf.Accumulator) error {
	var err error
	s.Lock()
	defer s.Unlock()
	os.Remove(s.Source)
	s.InputListener, err = net.ListenUnix("unix", &net.UnixAddr{
		Name: s.Source,
		Net:  "unix",
	})
	if err != nil {
		return err
	}
	s.CloseChan = make(chan bool)
	s.InputListener.SetUnlinkOnClose(true)
	go s.handleServerConnection(acc)
	return nil
}

// Stop causes the plugin to cease collection JSON data from the socket provided
// to Suricata.
func (s *Suricata) Stop() {
	s.Lock()
	defer s.Unlock()
	if s.CloseChan != nil {
		s.ClosedChan = make(chan bool)
		close(s.CloseChan)
		<-s.ClosedChan
		s.CloseChan = nil
	}
}

func (s *Suricata) handleServerConnection(acc telegraf.Accumulator) {
	var err error
	for {
		select {
		case <-s.CloseChan:
			close(s.ClosedChan)
			return
		default:
			var conn net.Conn
			conn, err = s.InputListener.Accept()
			if err != nil {
				acc.AddError(err)
				continue
			}
			reader := bufio.NewReaderSize(conn, 10485760)
			for {
				select {
				case <-s.CloseChan:
					conn.Close()
					s.InputListener.Close()
					close(s.ClosedChan)
					return
				default:
					line, isPrefix, rerr := reader.ReadLine()
					if rerr == nil || rerr != io.EOF {
						if isPrefix {
							acc.AddError(errors.New("incomplete line read from input"))
							continue
						} else {
							s.parse(acc, line)
						}
					}
				}
			}
		}
	}
}

func (s *Suricata) parse(acc telegraf.Accumulator, sjson []byte) {
	if len(sjson) == 0 {
		return
	}

	out, err := gojsonexplode.Explodejsonstr(string(sjson), ".")
	if err != nil {
		acc.AddError(err)
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(out), &result)
	if err != nil {
		acc.AddError(err)
		return
	}

	fields := make(map[string](map[string]interface{}))

	for k, v := range result {
		key := strings.Replace(k, "stats.", "", 1)
		if strings.HasPrefix(key, "threads.") {
			key = strings.Replace(key, "threads.", "", 1)
			threadkeys := strings.Split(key, ".")
			if _, ok := fields[threadkeys[0]]; !ok {
				fields[threadkeys[0]] = make(map[string]interface{})
			}
			fields[threadkeys[0]][strings.Join(threadkeys[1:], ".")] = v
		} else {
			if _, ok := fields["total"]; !ok {
				fields["total"] = make(map[string]interface{})
			}
			fields["total"][key] = v
		}
	}
	for k := range fields {
		acc.AddFields("suricata", fields[k], map[string]string{"thread": k})
	}
}

// Gather measures and submits one full set of telemetry to Telegraf.
// Not used here, submission is completely input-driven.
func (s *Suricata) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("suricata", func() telegraf.Input {
		return &Suricata{
			Source: "/tmp/suricata-stats.sock",
		}
	})
}
