//go:generate ../../../tools/readme_config_includer/generator
package solr

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
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
	s.client = s.createHTTPClient()
	s.collectors = make(map[string]MetricCollector, len(s.Servers))
	return nil
}

// Gather reads the stats from Solr and writes it to the
// Accumulator.
func (s *Solr) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, srv := range s.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()

			// Check the server version from cache or query one
			collector, found := s.collectors[server]
			if !found {
				version, err := s.determineServerAPIVersion(server)
				if err != nil {
					acc.AddError(err)
				}
				s.Log.Debugf("Found API version %d for server %q...", version, server)

				switch version {
				case 0:
					s.Log.Warn("Unable to determine API version! Using API v1...")
					fallthrough
				case 1:
					c, err := newCollectorV1(s.client, s.Username, s.Password, s.Cores)
					if err != nil {
						acc.AddError(fmt.Errorf("creating collector v1 for server %q failed: %w", server, err))
					}
					collector = c
					s.collectors[server] = c
				case 2:
					fallthrough
				default:
					if version > 2 {
						s.Log.Warnf("Unknown API version %q! Using latest known", version)
					}
					c, err := newCollectorV2(s.client, s.Username, s.Password, s.Cores)
					if err != nil {
						acc.AddError(fmt.Errorf("creating collector v2 for server %q failed: %w", server, err))
					}
					collector = c
					s.collectors[server] = c
				}
			}

			collector.Collect(acc, server)
		}(srv)
	}
	wg.Wait()

	return nil
}

func (s *Solr) determineServerAPIVersion(server string) (int, error) {
	url := server + "/solr/admin/info/system?wt=json"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	if s.Username != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}
	req.Header.Set("User-Agent", internal.ProductToken())

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, errors.New(resp.Status)
	}

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return 0, fmt.Errorf("decoding response failed: %w", err)
	}

	lraw, found := info["lucene"]
	if !found {
		return 0, nil
	}
	lucene, ok := lraw.(map[string]interface{})
	if !ok {
		return 0, nil
	}
	vraw, ok := lucene["solr-spec-version"]
	if !ok {
		return 0, nil
	}
	v, ok := vraw.(string)
	if !ok {
		return 0, nil
	}

	// API version 1 is required until v7.x
	version := semver.New(v)
	if version.LessThan(semver.Version{Major: 7}) {
		return 1, nil
	}

	// Starting from 7.0 API version 2 has to be used to get the UPDATE and
	// QUERY metrics.
	return 2, nil
}

func (s *Solr) createHTTPClient() *http.Client {
	tr := &http.Transport{
		ResponseHeaderTimeout: time.Duration(s.HTTPTimeout),
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(s.HTTPTimeout),
	}

	return client
}

func init() {
	inputs.Add("solr", func() telegraf.Input {
		return &Solr{
			HTTPTimeout: config.Duration(time.Second * 5),
		}
	})
}
