//go:generate ../../../tools/config_includer/generator
//go:generate ../../../tools/readme_config_includer/generator
package slurm

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Slurm struct {
	URL              string          `toml:"url"`
	Username         string          `toml:"username"`
	Token            string          `toml:"token"`
	APIVersion       string          `toml:"api_version"`
	EnabledEndpoints []string        `toml:"enabled_endpoints"`
	ResponseTimeout  config.Duration `toml:"response_timeout"`
	Log              telegraf.Logger `toml:"-"`
	tls.ClientConfig

	api         slurmAPI
	baseURL     *url.URL
	endpointMap map[string]bool
}

func (*Slurm) SampleConfig() string {
	return sampleConfig
}

func (s *Slurm) Init() error {
	if len(s.EnabledEndpoints) == 0 {
		s.EnabledEndpoints = []string{"diag", "jobs", "nodes", "partitions", "reservations"}
	}

	s.endpointMap = make(map[string]bool, len(s.EnabledEndpoints))
	for _, endpoint := range s.EnabledEndpoints {
		switch e := strings.ToLower(endpoint); e {
		case "diag", "jobs", "nodes", "partitions", "reservations":
			s.endpointMap[e] = true
		default:
			return fmt.Errorf("unknown endpoint %q", endpoint)
		}
	}

	if s.URL == "" {
		return errors.New("empty URL provided")
	}

	u, err := url.Parse(s.URL)
	if err != nil {
		return err
	}

	if u.Hostname() == "" {
		return fmt.Errorf("empty hostname for url %q", s.URL)
	}

	s.baseURL = u

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid scheme %q", u.Scheme)
	}

	tlsCfg, err := s.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	if u.Scheme == "http" && tlsCfg != nil {
		s.Log.Warn("non-empty TLS configuration for a URL with an http scheme. Ignoring it...")
		tlsCfg = nil
	}

	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
		Timeout:   time.Duration(s.ResponseTimeout),
	}
	userAgent := internal.ProductToken()

	switch s.APIVersion {
	case "", "0038":
		s.api = newV0038Client(u.Host, u.Scheme, userAgent, httpClient, s.Username, s.Token)
	case "0041":
		s.api = newV0041Client(u.Host, u.Scheme, userAgent, httpClient, s.Username, s.Token)
	default:
		return fmt.Errorf("unsupported api_version %q, valid values are \"0038\" and \"0041\"", s.APIVersion)
	}

	return nil
}

func (s *Slurm) Gather(acc telegraf.Accumulator) error {
	source := s.baseURL.Hostname()

	if s.endpointMap["diag"] {
		if err := s.api.gatherDiag(acc, source); err != nil {
			return err
		}
	}
	if s.endpointMap["jobs"] {
		if err := s.api.gatherJobs(acc, source); err != nil {
			return err
		}
	}
	if s.endpointMap["nodes"] {
		if err := s.api.gatherNodes(acc, source); err != nil {
			return err
		}
	}
	if s.endpointMap["partitions"] {
		if err := s.api.gatherPartitions(acc, source); err != nil {
			return err
		}
	}
	if s.endpointMap["reservations"] {
		if err := s.api.gatherReservations(acc, source); err != nil {
			return err
		}
	}

	return nil
}

func parseTres(tres string) map[string]interface{} {
	tresKVs := strings.Split(tres, ",")
	parsedValues := make(map[string]interface{}, len(tresKVs))

	for _, tresVal := range tresKVs {
		parsedTresVal := strings.Split(tresVal, "=")
		if len(parsedTresVal) != 2 {
			continue
		}

		tag := parsedTresVal[0]
		val := parsedTresVal[1]
		var factor float64 = 1

		if tag == "mem" {
			var ok bool
			factor, ok = map[string]float64{
				"K": 1.0 / 1024.0,
				"M": 1,
				"G": 1024,
				"T": 1024 * 1024,
				"P": 1024 * 1024 * 1024,
			}[strings.ToUpper(val[len(val)-1:])]
			if !ok {
				continue
			}
			val = val[:len(val)-1]
		}

		parsedFloat, err := strconv.ParseFloat(val, 64)
		if err == nil {
			parsedValues[tag] = parsedFloat * factor
			continue
		}
		parsedValues[tag] = val
	}

	return parsedValues
}

func init() {
	inputs.Add("slurm", func() telegraf.Input {
		return &Slurm{
			ResponseTimeout: config.Duration(5 * time.Second),
		}
	})
}
