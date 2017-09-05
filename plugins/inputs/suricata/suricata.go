package suricata

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/nytlabs/gojsonexplode"
)

var singleDotRegexp = regexp.MustCompilePOSIX(`[^.]\.[^.]`)

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
	if s.InputListener == nil {
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
	}
	return nil
}

// Stop causes the plugin to cease collection JSON data from the socket provided
// to Suricata.
func (s *Suricata) Stop() {
	s.Lock()
	defer s.Unlock()
	if s.CloseChan != nil {
		s.InputListener.Close()
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
			s.InputListener = nil
			close(s.ClosedChan)
			return
		default:
			var conn net.Conn
			conn, err = s.InputListener.Accept()
			if err != nil {
				continue
			}
			reader := bufio.NewReaderSize(conn, 10485760)
		out:
			for {
				select {
				case <-s.CloseChan:
					conn.Close()
					s.InputListener.Close()
					s.InputListener = nil
					close(s.ClosedChan)
					return
				default:
					line, isPrefix, rerr := reader.ReadLine()
					if rerr == nil {
						if isPrefix {
							acc.AddError(errors.New("incomplete line read from input"))
							continue
						} else {
							s.parse(acc, line)
						}
					} else if rerr == io.EOF {
						break out
					}
				}
			}
		}
	}
}

func splitAtSingleDot(in string) []string {
	res := singleDotRegexp.FindAllStringIndex(in, -1)
	if res == nil {
		return []string{in}
	}
	ret := make([]string, 0)
	startpos := 0
	for _, v := range res {
		ret = append(ret, in[startpos:v[0]+1])
		startpos = v[1] - 1
	}
	return append(ret, in[startpos:])
}

func addPercentage(thread string, val1 string, val2 string, valt string,
	fields map[string](map[string]interface{})) {
	if _, ok := fields[thread][val1]; ok {
		if _, ok := fields[thread][val2]; ok {
			f1 := fields[thread][val1].(float64)
			f2 := fields[thread][val2].(float64)
			fields[thread][valt] = f2 / (f1 + f2)
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
			threadkeys := splitAtSingleDot(key)
			threadkey := threadkeys[0]
			if _, ok := fields[threadkey]; !ok {
				fields[threadkey] = make(map[string]interface{})
			}
			fields[threadkey][strings.Join(threadkeys[1:], ".")] = v
		} else {
			if _, ok := fields["total"]; !ok {
				fields["total"] = make(map[string]interface{})
			}
			fields["total"][key] = v
		}
	}

	addPercentage("total", "capture.kernel_packets", "capture.kernel_drops",
		"capture.kernel_drop_percentage", fields)
	addPercentage("total", "capture.kernel_packets_delta", "capture.kernel_drops_delta",
		"capture.kernel_drop_delta_percentage", fields)

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
