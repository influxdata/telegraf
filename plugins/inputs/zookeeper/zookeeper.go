//go:generate ../../../tools/readme_config_includer/generator
package zookeeper

import (
	"bufio"
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/prometheus"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample.conf
var sampleConfig string

var zookeeperFormatRE = regexp.MustCompile(`^zk_(\w[\w\.\-]*)\s+([\w\.\-]+)`)

// Zookeeper is a zookeeper plugin
type Zookeeper struct {
	Servers         []string        `toml:"servers"`
	Timeout         config.Duration `toml:"timeout"`
	MetricsProvider string          `toml:"metrics_provider"`

	EnableTLS bool `toml:"enable_tls"`
	EnableSSL bool `toml:"enable_ssl" deprecated:"1.7.0;use 'enable_tls' instead"`
	tlsint.ClientConfig

	Log telegraf.Logger `toml:"-"`

	initialized bool
	tlsConfig   *tls.Config

	httpconfig.HTTPClientConfig
	client *http.Client
}

var defaultTimeout = 5 * time.Second

func (z *Zookeeper) dial(ctx context.Context, addr string) (net.Conn, error) {
	var dialer net.Dialer
	if z.EnableTLS || z.EnableSSL {
		deadline, ok := ctx.Deadline()
		if ok {
			dialer.Deadline = deadline
		}
		return tls.DialWithDialer(&dialer, "tcp", addr, z.tlsConfig)
	}
	return dialer.DialContext(ctx, "tcp", addr)
}

func (*Zookeeper) SampleConfig() string {
	return sampleConfig
}

func (z *Zookeeper) Init() error {
	if z.MetricsProvider != "java" && z.MetricsProvider != "prometheus" {
		return fmt.Errorf("unrecognized metrics provider '%s', choose from: \"java\" or \"prometheus\"", z.MetricsProvider)
	}

	return nil
}

// Gather reads stats from all configured servers accumulates stats
func (z *Zookeeper) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	if !z.initialized {
		tlsConfig, err := z.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		z.tlsConfig = tlsConfig
		z.initialized = true
	}

	if z.Timeout < config.Duration(1*time.Second) {
		z.Timeout = config.Duration(defaultTimeout)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(z.Timeout))
	defer cancel()

	if len(z.Servers) == 0 {
		z.Servers = []string{":2181"}
	}

	for _, serverAddress := range z.Servers {
		if z.MetricsProvider == "prometheus" {
			acc.AddError(z.gatherPrometheusMetrics(ctx, serverAddress, acc))
		} else {
			acc.AddError(z.gatherJavaMetrics(ctx, serverAddress, acc))
		}
	}

	return nil
}

func (z *Zookeeper) gatherJavaMetrics(ctx context.Context, address string, acc telegraf.Accumulator) error {
	var zookeeperState string
	_, _, err := net.SplitHostPort(address)
	if err != nil {
		address = address + ":2181"
	}

	c, err := z.dial(ctx, address)
	if err != nil {
		return err
	}
	defer c.Close()

	// Apply deadline to connection
	deadline, ok := ctx.Deadline()
	if ok {
		if err := c.SetDeadline(deadline); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(c, "%s\n", "mntr"); err != nil {
		return err
	}
	rdr := bufio.NewReader(c)
	scanner := bufio.NewScanner(rdr)

	service := strings.Split(address, ":")
	if len(service) != 2 {
		return fmt.Errorf("invalid service address: %s", address)
	}

	fields := make(map[string]interface{})
	for scanner.Scan() {
		line := scanner.Text()
		parts := zookeeperFormatRE.FindStringSubmatch(line)

		if len(parts) != 3 {
			return fmt.Errorf("unexpected line in mntr response: %q", line)
		}

		measurement := strings.TrimPrefix(parts[1], "zk_")
		if measurement == "server_state" {
			zookeeperState = parts[2]
		} else {
			sValue := parts[2]

			iVal, err := strconv.ParseInt(sValue, 10, 64)
			if err == nil {
				fields[measurement] = iVal
			} else {
				fields[measurement] = sValue
			}
		}
	}

	srv := "localhost"
	if service[0] != "" {
		srv = service[0]
	}

	tags := map[string]string{
		"server": srv,
		"port":   service[1],
		"state":  zookeeperState,
	}
	acc.AddFields("zookeeper", fields, tags)

	return nil
}

func (z *Zookeeper) gatherPrometheusMetrics(ctx context.Context, address string, acc telegraf.Accumulator) error {
	// ensure the correct port is used
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		address = address + ":7000"
	}
	if host == "" {
		address = "localhost" + address
	}

	// add protocol and metrics URL
	address = fmt.Sprintf("http://%s/metrics", address)
	source, err := url.ParseRequestURI(address)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return err
	}

	if z.client == nil {
		err := z.startClient()
		if err != nil {
			return err
		}
	}

	resp, err := z.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"received status code %d (%s)",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
		)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading body failed: %v", err)
	}

	// Instantiate a new parser for the new data to avoid trouble with stateful parsers
	parser := prometheus.Parser{
		IgnoreTimestamp: true,
	}
	metrics, err := parser.Parse(b)
	if err != nil {
		return fmt.Errorf("parsing metrics failed: %v", err)
	}

	for _, metric := range metrics {
		if !metric.HasTag("source") {
			metric.AddTag("source", source.Host)
		}
		acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
	}

	return nil
}

func (z *Zookeeper) startClient() error {
	ctx := context.Background()
	client, err := z.HTTPClientConfig.CreateClient(ctx, z.Log)
	if err != nil {
		return err
	}

	z.client = client

	return nil
}

func init() {
	inputs.Add("zookeeper", func() telegraf.Input {
		return &Zookeeper{
			MetricsProvider: "java",
		}
	})
}
