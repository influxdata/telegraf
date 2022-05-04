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
	Source    string `toml:"source"`
	Delimiter string `toml:"delimiter"`
	Alerts    bool   `toml:"alerts"`

	inputListener *net.UnixListener
	cancel        context.CancelFunc

	Log telegraf.Logger `toml:"-"`

	wg sync.WaitGroup
}

// Start initiates background collection of JSON data from the socket
// provided to Suricata.
func (s *Suricata) Start(acc telegraf.Accumulator) error {
	var err error
	s.inputListener, err = net.ListenUnix("unix", &net.UnixAddr{
		Name: s.Source,
		Net:  "unix",
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.inputListener.SetUnlinkOnClose(true)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		go s.handleServerConnection(ctx, acc)
	}()
	return nil
}

// Stop causes the plugin to cease collecting JSON data from the socket provided
// to Suricata.
func (s *Suricata) Stop() {
	// Ignore the returned error as we cannot do anything about it anyway
	//nolint:errcheck,revive
	s.inputListener.Close()
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *Suricata) readInput(ctx context.Context, acc telegraf.Accumulator, conn net.Conn) error {
	reader := bufio.NewReaderSize(conn, InBufSize)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			line, rerr := reader.ReadBytes('\n')
			if rerr != nil {
				return rerr
			}
			if len(line) > 0 {
				err := s.parse(acc, line)
				if err != nil {
					acc.AddError(err)
				}
			}
		}
	}
}

func (s *Suricata) handleServerConnection(ctx context.Context, acc telegraf.Accumulator) {
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var conn net.Conn
			conn, err = s.inputListener.Accept()
			if err != nil {
				if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
					acc.AddError(err)
				}
				continue
			}
			err = s.readInput(ctx, acc, conn)
			// we want to handle EOF as an opportunity to wait for a new
			// connection -- this could, for example, happen when Suricata is
			// restarted while Telegraf is running.
			if err != io.EOF {
				acc.AddError(err)
				return
			}
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
	case []interface{}:
		for _, v := range t {
			err := flexFlatten(outmap, field, v, delimiter)
			if err != nil {
				return err
			}
		}
	case string:
		outmap[field] = v
	case float64:
		outmap[field] = t
	default:
		return fmt.Errorf("unsupported type %T encountered", t)
	}
	return nil
}

func (s *Suricata) parseAlert(acc telegraf.Accumulator, result map[string]interface{}) {
	if _, ok := result["alert"].(map[string]interface{}); !ok {
		s.Log.Debug("'alert' sub-object does not have required structure")
		return
	}

	totalmap := make(map[string]interface{})
	for k, v := range result["alert"].(map[string]interface{}) {
		//source and target fields are maps
		err := flexFlatten(totalmap, k, v, s.Delimiter)
		if err != nil {
			s.Log.Debugf("Flattening alert failed: %v", err)
			// we skip this subitem as something did not parse correctly
			continue
		}
	}

	//threads field do not exist in alert output, always global
	acc.AddFields("suricata_alert", totalmap, nil)
}

func (s *Suricata) parseStats(acc telegraf.Accumulator, result map[string]interface{}) {
	if _, ok := result["stats"].(map[string]interface{}); !ok {
		s.Log.Debug("The 'stats' sub-object does not have required structure")
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
						err := flexFlatten(outmap, "", threadStruct, s.Delimiter)
						if err != nil {
							s.Log.Debugf("Flattening alert failed: %v", err)
							// we skip this thread as something did not parse correctly
							continue
						}
						fields[k] = outmap
					}
				}
			} else {
				s.Log.Debug("The 'threads' sub-object does not have required structure")
			}
		} else {
			err := flexFlatten(totalmap, k, v, s.Delimiter)
			if err != nil {
				s.Log.Debugf("Flattening alert failed: %v", err)
				// we skip this subitem as something did not parse correctly
				continue
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

func (s *Suricata) parse(acc telegraf.Accumulator, sjson []byte) error {
	// initial parsing
	var result map[string]interface{}
	err := json.Unmarshal(sjson, &result)
	if err != nil {
		return err
	}
	// check for presence of relevant stats or alert
	_, ok := result["stats"]
	_, ok2 := result["alert"]
	if !ok && !ok2 {
		s.Log.Debugf("Invalid input without 'stats' or 'alert' object: %v", result)
		return fmt.Errorf("input does not contain 'stats' or 'alert' object")
	}
	if ok {
		s.parseStats(acc, result)
	} else if ok2 && s.Alerts {
		s.parseAlert(acc, result)
	}
	return nil
}

// Gather measures and submits one full set of telemetry to Telegraf.
// Not used here, submission is completely input-driven.
func (s *Suricata) Gather(_ telegraf.Accumulator) error {
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
