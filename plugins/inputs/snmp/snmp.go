//go:generate ../../../tools/readme_config_includer/generator
package snmp

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/snmp"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Snmp struct {
	// The SNMP agent to query. Format is [SCHEME://]ADDR[:PORT] (e.g.
	// udp://1.2.3.4:161).  If the scheme is not specified then "udp" is used.
	Agents []string `toml:"agents"`

	// The tag used to name the agent host
	AgentHostTag string `toml:"agent_host_tag"`

	// Stop collection when receiving errors from an agent
	StopOnError bool `toml:"stop_on_error"`

	snmp.ClientConfig

	Tables []snmp.Table `toml:"table"`

	// Name & Fields are the elements of a Table.
	// Telegraf chokes if we try to embed a Table. So instead we have to embed the
	// fields of a Table, and construct a Table during runtime.
	Name   string       `toml:"name"`
	Fields []snmp.Field `toml:"field"`

	Log telegraf.Logger `toml:"-"`

	connectionCache []snmp.Connection

	// Protects connectionCache + lastReset
	cacheMu   sync.Mutex
	lastReset []time.Time

	// Prevent thrash: minimum duration between resets per agent
	resetCooldown time.Duration

	translator snmp.Translator
}

func (*Snmp) SampleConfig() string {
	return sampleConfig
}

func (s *Snmp) SetTranslator(name string) {
	s.Translator = name
}

func (s *Snmp) Init() error {
	var err error
	switch s.Translator {
	case "gosmi":
		s.translator, err = snmp.NewGosmiTranslator(s.Path, s.Log)
		if err != nil {
			return err
		}
	case "netsnmp":
		s.translator = snmp.NewNetsnmpTranslator(s.Log)
	default:
		return errors.New("invalid translator value")
	}

	s.connectionCache = make([]snmp.Connection, len(s.Agents))
	s.lastReset = make([]time.Time, len(s.Agents))

	// Reasonable default cooldown; prevents constant reconnect loops on noisy links.
	// You can tune this; 30s–2m is usually fine.
	s.resetCooldown = 30 * time.Second

	for i := range s.Tables {
		if err := s.Tables[i].Init(s.translator); err != nil {
			return fmt.Errorf("initializing table %s: %w", s.Tables[i].Name, err)
		}
	}

	for i := range s.Fields {
		if err := s.Fields[i].Init(s.translator); err != nil {
			return fmt.Errorf("initializing field %s: %w", s.Fields[i].Name, err)
		}
	}

	if len(s.AgentHostTag) == 0 {
		s.AgentHostTag = "agent_host"
	}
	if s.AgentHostTag != "source" {
		config.PrintOptionValueDeprecationNotice("inputs.snmp", "agent_host_tag", s.AgentHostTag, telegraf.DeprecationInfo{
			Since:  "1.29.0",
			Notice: `set to "source" for consistent usage across plugins or safely ignore this message and continue to use the current value`,
		})
	}

	s.GosnmpDebugLogger = s.Log

	return nil
}

func (s *Snmp) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for i, agent := range s.Agents {
		wg.Add(1)
		go func(i int, agent string) {
			defer wg.Done()
			gs, err := s.getConnection(i)
			if err != nil {
				acc.AddError(fmt.Errorf("agent %s: %w", agent, err))
				return
			}

			// First is the top-level fields. We treat the fields as table prefixes with an empty index.
			t := snmp.Table{
				Name:   s.Name,
				Fields: s.Fields,
			}
			topTags := make(map[string]string)

			if err := s.gatherTable(acc, gs, tTop, topTags, false); err != nil {
				// If it's a v3 session/auth mismatch, reset and stop early to avoid extra walks.
				if s.isSnmpV3SessionInvalid(err) {
					s.resetConnection(i, agent, gs, err)
					return
				}
				acc.AddError(fmt.Errorf("agent %s: %w", agent, err))
				if s.StopOnError {
					return
				}
			}

			// Now is the real tables.
			for _, t := range s.Tables {
				if err := s.gatherTable(acc, gs, t, topTags, true); err != nil {
					if s.isSnmpV3SessionInvalid(err) {
						s.resetConnection(i, agent, gs, err)
						return
					}
					acc.AddError(fmt.Errorf("agent %s: gathering table %s: %w", agent, t.Name, err))
					if s.StopOnError {
						return
					}
				}
			}
		}(i, agent)
	}
	wg.Wait()

	return nil
}

func (s *Snmp) gatherTable(acc telegraf.Accumulator, gs snmp.Connection, t snmp.Table, topTags map[string]string, walk bool) error {
	rt, err := t.Build(gs, walk)
	if err != nil {
		return err
	}

	for _, tr := range rt.Rows {
		if !walk {
			// top-level table. Add tags to topTags.
			for k, v := range tr.Tags {
				topTags[k] = v
			}
		} else {
			// real table. Inherit any specified tags.
			for _, k := range t.InheritTags {
				if v, ok := topTags[k]; ok {
					tr.Tags[k] = v
				}
			}
		}
		if _, ok := tr.Tags[s.AgentHostTag]; !ok {
			tr.Tags[s.AgentHostTag] = gs.Host()
		}
		acc.AddFields(rt.Name, tr.Fields, tr.Tags, rt.Time)
	}

	return nil
}

// getConnection creates a snmpConnection (*gosnmp.GoSNMP) object and caches the
// result using `agentIndex` as the cache key.  This is done to allow multiple
// connections to a single address.  It is an error to use a connection in
// more than one goroutine.
func (s *Snmp) getConnection(idx int) (snmp.Connection, error) {
	// Read cached connection under lock
	s.cacheMu.Lock()
	gs := s.connectionCache[idx]
	s.cacheMu.Unlock()

	if gs != nil {
		if err := gs.Reconnect(); err != nil {
			return gs, fmt.Errorf("reconnecting: %w", err)
		}
		return gs, nil
	}

	agent := s.Agents[idx]

	newConn, err := snmp.NewWrapper(s.ClientConfig)
	if err != nil {
		return nil, err
	}
	if err := newConn.SetAgent(agent); err != nil {
		return nil, err
	}
	if err := newConn.Connect(); err != nil {
		return nil, fmt.Errorf("setting up connection: %w", err)
	}

	// Store in cache under lock
	s.cacheMu.Lock()
	s.connectionCache[idx] = newConn
	s.cacheMu.Unlock()

	return newConn, nil
}

// resetConnection drops the cached connection so the next gather will do a fresh Connect().
// Includes a cooldown to avoid thrashing on noisy links.
func (s *Snmp) resetConnection(idx int, agent string, gs snmp.Connection, cause error) {
	now := time.Now()

	s.cacheMu.Lock()
	last := s.lastReset[idx]
	if !last.IsZero() && now.Sub(last) < s.resetCooldown {
		// Within cooldown; don't thrash
		s.cacheMu.Unlock()
		s.Log.Warnf("SNMPv3 session error on agent %s but reset is in cooldown (%s). Cause: %v",
			agent, s.resetCooldown, cause)
		return
	}

	s.lastReset[idx] = now
	// Drop cached connection
	s.connectionCache[idx] = nil
	s.cacheMu.Unlock()

	// Best-effort close if supported (do NOT assume internal fields)
	type closer interface{ Close() error }
	if c, ok := gs.(closer); ok {
		_ = c.Close()
	}

	s.Log.Warnf("SNMPv3 session/auth mismatch detected on agent %s; cleared cached connection for next cycle. Cause: %v",
		agent, cause)
}

// isSnmpV3SessionInvalid tries to detect errors consistent with SNMPv3 engine/time state mismatch
// after device reboot/snmpd restart (common symptoms).
func (s *Snmp) isSnmpV3SessionInvalid(err error) bool {
	// Unwrap to the root message chain
	msg := strings.ToLower(s.unwrapErrorString(err))

	// Common gosnmp/USM symptoms seen when engineBoots/engineTime changes or auth fails:
	needles := []string{
		"incoming packet is not authentic", // auth failure / wrong keys or engine context mismatch symptoms
		"not in time window",               // engineTime/boots mismatch
		"unknown engine id",                // engineID changed
		"unknown engineid",
		"usm",                              // USM-related failures often include this
	}

	// If the plugin isn't using v3, avoid resetting for these messages.
	// (Version is in ClientConfig; keeping it conservative.)
	if s.Version != 3 {
		return false
	}

	for _, n := range needles {
		if strings.Contains(msg, n) {
			return true
		}
	}
	return false
}

func (s *Snmp) unwrapErrorString(err error) string {
	// Build a compact string including wrapped errors without blowing up logs
	var parts []string
	seen := 0
	for err != nil && seen < 6 {
		parts = append(parts, err.Error())
		err = errors.Unwrap(err)
		seen++
	}
	return strings.Join(parts, " | ")
}


func init() {
	inputs.Add("snmp", func() telegraf.Input {
		return &Snmp{
			Name: "snmp",
			ClientConfig: snmp.ClientConfig{
				Retries:        3,
				MaxRepetitions: 10,
				Timeout:        config.Duration(5 * time.Second),
				Version:        2,
				Path:           []string{"/usr/share/snmp/mibs"},
				Community:      "public",
			},
		}
	})
}
