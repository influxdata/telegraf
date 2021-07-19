package squid

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Squid struct {
	Url             string
	ResponseTimeout internal.Duration
	tls.ClientConfig

	client *http.Client
}

const sampleConfig string = `
  ## url of the squid proxy manager counters page
  url = "http://localhost:3128"

  ## Maximum time to receive response.
  response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (s *Squid) SampleConfig() string {
	return sampleConfig
}

func (o *Squid) Description() string {
	return "Squid web proxy cache plugin"
}

// return an initialized Squid
func NewSquid() *Squid {
	return &Squid{
		Url:             "http://localhost:3128",
		ResponseTimeout: internal.Duration{Duration: time.Second * 5},
	}
}

// Gather metrics
func (s *Squid) Gather(acc telegraf.Accumulator) error {
	if s.client == nil {
		tlsCfg, err := s.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		s.client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsCfg,
			},
			Timeout: s.ResponseTimeout.Duration,
		}
	}

	acc.AddError(s.gatherCounters(s.Url+"/squid-internal-mgr/counters", acc))

	return nil
}

// gather counters
func (s *Squid) gatherCounters(url string, acc telegraf.Accumulator) error {
	resp, err := s.client.Get(url)
	if err != nil {
		return fmt.Errorf("unable to GET \"%s\": %s", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK status code returned from \"%s\": %d", url, resp.StatusCode)
	}

	fields := parseBody(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to parse body from \"%s\": %s", url, err)
	}

	tags := map[string]string{
		"source": s.Url,
	}

	acc.AddFields("squid", fields, tags)

	return nil
}

// parseBody accepts a response body as an io.Reader and uses bufio.NewScanner
// to walk the body. It returns the metric fields expected format is "this.key
// = 0.000\n"
func parseBody(body io.Reader) map[string]interface{} {
	fields := map[string]interface{}{}
	sc := bufio.NewScanner(body)
	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			// skip if this line isn't long enough
			if len(parts) != 2 {
				continue
			}

			// skip sample_time
			if parts[0] == "sample_time" {
				continue
			}

			key := strings.TrimSpace(parts[0])
			key = strings.Replace(key, ".", "_", -1)
			valueStr := strings.TrimSpace(parts[1])

			// src/mgr/CountersAction.h defines these all as double,
			// so turn them into 64-bit floats
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				continue
			}

			// store this field
			fields[key] = value
		}
	}
	return fields
}

func init() {
	inputs.Add("squid", func() telegraf.Input { return NewSquid() })
}
