package suricata

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	// InBufSize is the input buffer size for JSON received via socket.
	// Set to 10MB, as depending on the number of threads the output might be
	// large.
	InBufSize = 10 * 1024 * 1024
)

// Suricata is a Telegraf input plugin for Suricata runtime statistics.
type Suricata struct {
	sync.Mutex

	Source    string `toml:"source"`
	Delimiter string `toml:"delimiter"`

	InputListener *net.UnixListener
	Ctx           context.Context
	Cancel        context.CancelFunc

	wg sync.WaitGroup
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
  source = "/var/run/suricata-stats.sock"
  # Delimiter for flattening field keys, e.g. subitem "alert" of "detect"
  # becomes "detect_alert" when delimiter is "_".
  delimiter = "_"
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
		s.InputListener, err = net.ListenUnix("unix", &net.UnixAddr{
			Name: s.Source,
			Net:  "unix",
		})
		if err != nil {
			return err
		}
		s.Ctx, s.Cancel = context.WithCancel(context.Background())
		s.InputListener.SetUnlinkOnClose(true)
		s.wg.Add(1)
		go s.handleServerConnection(acc)
	}
	return nil
}

// Stop causes the plugin to cease collecting JSON data from the socket provided
// to Suricata.
func (s *Suricata) Stop() {
	s.Lock()
	defer s.Unlock()
	s.InputListener.Close()
	if s.Cancel != nil {
		s.Cancel()
	}
	s.wg.Wait()
	s.InputListener.Close()
	s.InputListener = nil
}

func (s *Suricata) readInput(acc telegraf.Accumulator, conn net.Conn) {
	reader := bufio.NewReaderSize(conn, InBufSize)
	for {
		select {
		case <-s.Ctx.Done():
			return
		default:
			line, rerr := reader.ReadBytes('\n')
			if rerr == nil {
				s.parse(acc, line)
			} else if rerr == io.EOF {
				return
			}
		}
	}
}

func (s *Suricata) handleServerConnection(acc telegraf.Accumulator) {
	var err error
	defer s.wg.Done()
	for {
		select {
		case <-s.Ctx.Done():
			return
		default:
			var conn net.Conn
			conn, err = s.InputListener.Accept()
			if err != nil {
				if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
					acc.AddError(err)
				}
				return
			}
			s.readInput(acc, conn)
		}
	}
}

func flexFlatten(outmap map[string]interface{}, field string, v interface{}, delimiter string) error {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			var err error
			if field == "" {
				err = flexFlatten(outmap, k, v, delimiter)
			} else {
				err = flexFlatten(outmap, fmt.Sprintf("%s%s%s", field, delimiter, k), v, delimiter)
			}
			if err != nil {
				return err
			}
		}
	case float64:
		outmap[field] = v.(float64)
	default:
		return fmt.Errorf("unsupported type %T encountered", t)
	}
	return nil
}

func (s *Suricata) parse(acc telegraf.Accumulator, sjson []byte) {
	// initial parsing
	var result map[string]interface{}
	err := json.Unmarshal([]byte(sjson), &result)
	if err != nil {
		acc.AddError(err)
		return
	}

	// check for presence of relevant stats
	if _, ok := result["stats"]; !ok {
		acc.AddError(fmt.Errorf("input does not contain necessary 'stats' sub-object"))
		return
	}

	if _, ok := result["stats"].(map[string]interface{}); !ok {
		acc.AddError(fmt.Errorf("'stats' sub-object does not have required structure"))
		return
	}

	fields := make(map[string](map[string]interface{}))
	totalmap := make(map[string]interface{})
	for k, v := range result["stats"].(map[string]interface{}) {
		if k == "threads" {
			if v, ok := v.(map[string]interface{}); ok {
				for k, t := range v {
					outmap := make(map[string]interface{})
					if threadStruct, ok := t.(map[string]interface{}); ok {
						err = flexFlatten(outmap, "", threadStruct, s.Delimiter)
						if err != nil {
							acc.AddError(err)
							// we skip this thread as something did not parse correctly
							continue
						}
						fields[k] = outmap
					}
				}
			} else {
				acc.AddError(fmt.Errorf("'threads' sub-object does not have required structure"))
			}
		} else {
			err = flexFlatten(totalmap, k, v, s.Delimiter)
			if err != nil {
				acc.AddError(err)
				// we skip this subitem as something did not parse correctly
			}
		}
	}
	fields["total"] = totalmap

	for k := range fields {
		if k == "Global" {
			acc.AddFields("suricata", fields[k], nil)
		} else {
			acc.AddFields("suricata", fields[k], map[string]string{"thread": k})
		}
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
			Source:    "/var/run/suricata-stats.sock",
			Delimiter: "_",
		}
	})
}
