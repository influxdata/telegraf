//go:generate ../../../tools/readme_config_includer/generator
package solr

import (
	_ "embed"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type MetricCollector interface {
	Collect(acc telegraf.Accumulator, server string)
}

// Solr is a plugin to read stats from one or many Solr servers
type Solr struct {
	Servers     []string        `toml:"servers"`
	Username    string          `toml:"username"`
	Password    string          `toml:"password"`
	HTTPTimeout config.Duration `toml:"timeout"`
	Cores       []string        `toml:"cores"`
	Log         telegraf.Logger `toml:"-"`

	client     *http.Client
	collectors map[string]MetricCollector
}

func (*Solr) SampleConfig() string {
	return sampleConfig
}

func (s *Solr) Init() error {
	s.client = &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: time.Duration(s.HTTPTimeout),
		},
		Timeout: time.Duration(s.HTTPTimeout),
	}

	s.collectors = make(map[string]MetricCollector, len(s.Servers))
	return nil
}

func (s *Solr) Start(acc telegraf.Accumulator) error {
	for _, server := range s.Servers {
		acc.AddError(s.updateCollector(server))
	}
	return nil
}

func (s *Solr) Stop() {}

func (s *Solr) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, srv := range s.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()

			// Check the server version from cache or query one
			if err := s.updateCollector(server); err != nil {
				acc.AddError(err)
				return
			}
			collector := s.collectors[server]
			collector.Collect(acc, server)
		}(srv)
	}
	wg.Wait()

	return nil
}

func (s *Solr) updateCollector(server string) error {
	if _, found := s.collectors[server]; found {
		return nil
	}
	version, err := s.determineServerAPIVersion(server)
	if err != nil {
		s.Log.Errorf("Getting version for %q failed: %v", server, err)
	}
	s.Log.Debugf("Found API version %d for server %q...", version, server)

	switch version {
	case 0:
		s.Log.Warn("Unable to determine API version! Using API v1...")
		fallthrough
	case 1:
		c, err := newCollectorV1(s, s.Cores)
		if err != nil {
			return fmt.Errorf("creating collector v1 for server %q failed: %w", server, err)
		}
		s.collectors[server] = c
	case 2:
		fallthrough
	default:
		if version > 2 {
			s.Log.Warnf("Unknown API version %q! Using latest known", version)
		}
		c, err := newCollectorV2(s, s.Cores)
		if err != nil {
			return fmt.Errorf("creating collector v2 for server %q failed: %w", server, err)
		}
		s.collectors[server] = c
	}

	return nil
}

func init() {
	inputs.Add("solr", func() telegraf.Input {
		return &Solr{
			HTTPTimeout: config.Duration(time.Second * 5),
		}
	})
}
