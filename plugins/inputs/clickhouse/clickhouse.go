package clickhouse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var defaultTimeout = 5 * time.Second

var sampleConfig = `
  ## Username for authorization on ClickHouse server
  ## example: user = "default""
  username = "default"

  ## Password for authorization on ClickHouse server
  ## example: password = "super_secret"

  ## HTTP(s) timeout while getting metrics values
  ## The timeout includes connection time, any redirects, and reading the response body.
  ##   example: timeout = 1s
  # timeout = 5s

  ## List of servers for metrics scraping
  ## metrics scrape via HTTP(s) clickhouse interface
  ## https://clickhouse.tech/docs/en/interfaces/http/
  ##    example: servers = ["http://127.0.0.1:8123","https://custom-server.mdb.yandexcloud.net"]
  servers         = ["http://127.0.0.1:8123"]

  ## If "auto_discovery"" is "true" plugin tries to connect to all servers available in the cluster
  ## with using same "user:password" described in "user" and "password" parameters
  ## and get this server hostname list from "system.clusters" table
  ## see
  ## - https://clickhouse.tech/docs/en/operations/system_tables/#system-clusters
  ## - https://clickhouse.tech/docs/en/operations/server_settings/settings/#server_settings_remote_servers
  ## - https://clickhouse.tech/docs/en/operations/table_engines/distributed/
  ## - https://clickhouse.tech/docs/en/operations/table_engines/replication/#creating-replicated-tables
  ##    example: auto_discovery = false
  # auto_discovery = true

  ## Filter cluster names in "system.clusters" when "auto_discovery" is "true"
  ## when this filter present then "WHERE cluster IN (...)" filter will apply
  ## please use only full cluster names here, regexp and glob filters is not allowed
  ## for "/etc/clickhouse-server/config.d/remote.xml"
  ## <yandex>
  ##  <remote_servers>
  ##    <my-own-cluster>
  ##        <shard>
  ##          <replica><host>clickhouse-ru-1.local</host><port>9000</port></replica>
  ##          <replica><host>clickhouse-ru-2.local</host><port>9000</port></replica>
  ##        </shard>
  ##        <shard>
  ##          <replica><host>clickhouse-eu-1.local</host><port>9000</port></replica>
  ##          <replica><host>clickhouse-eu-2.local</host><port>9000</port></replica>
  ##        </shard>
  ##    </my-onw-cluster>
  ##  </remote_servers>
  ##
  ## </yandex>
  ##
  ## example: cluster_include = ["my-own-cluster"]
  # cluster_include = []

  ## Filter cluster names in "system.clusters" when "auto_discovery" is "true"
  ## when this filter present then "WHERE cluster NOT IN (...)" filter will apply
  ##    example: cluster_exclude = ["my-internal-not-discovered-cluster"]
  # cluster_exclude = []

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

type connect struct {
	Cluster  string `json:"cluster"`
	ShardNum int    `json:"shard_num"`
	Hostname string `json:"host_name"`
	url      *url.URL
}

func init() {
	inputs.Add("clickhouse", func() telegraf.Input {
		return &ClickHouse{
			AutoDiscovery: true,
			ClientConfig: tls.ClientConfig{
				InsecureSkipVerify: false,
			},
			Timeout: internal.Duration{Duration: defaultTimeout},
		}
	})
}

// ClickHouse Telegraf Input Plugin
type ClickHouse struct {
	Username       string            `toml:"username"`
	Password       string            `toml:"password"`
	Servers        []string          `toml:"servers"`
	AutoDiscovery  bool              `toml:"auto_discovery"`
	ClusterInclude []string          `toml:"cluster_include"`
	ClusterExclude []string          `toml:"cluster_exclude"`
	Timeout        internal.Duration `toml:"timeout"`
	client         http.Client
	tls.ClientConfig
}

// SampleConfig returns the sample config
func (*ClickHouse) SampleConfig() string {
	return sampleConfig
}

// Description return plugin description
func (*ClickHouse) Description() string {
	return "Read metrics from one or many ClickHouse servers"
}

// Start ClickHouse input service
func (ch *ClickHouse) Start(telegraf.Accumulator) error {
	timeout := defaultTimeout
	if ch.Timeout.Duration != 0 {
		timeout = ch.Timeout.Duration
	}
	tlsCfg, err := ch.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	ch.client = http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			Proxy:           http.ProxyFromEnvironment,
		},
	}
	return nil
}

// Gather collect data from ClickHouse server
func (ch *ClickHouse) Gather(acc telegraf.Accumulator) (err error) {
	var (
		connects []connect
		exists   = func(host string) bool {
			for _, c := range connects {
				if c.Hostname == host {
					return true
				}
			}
			return false
		}
	)

	for _, server := range ch.Servers {
		u, err := url.Parse(server)
		if err != nil {
			return err
		}
		switch {
		case ch.AutoDiscovery:
			var conns []connect
			if err := ch.execQuery(u, "SELECT cluster, shard_num, host_name FROM system.clusters "+ch.clusterIncludeExcludeFilter(), &conns); err != nil {
				acc.AddError(err)
				continue
			}
			for _, c := range conns {
				if !exists(c.Hostname) {
					c.url = &url.URL{
						Scheme: u.Scheme,
						Host:   net.JoinHostPort(c.Hostname, u.Port()),
					}
					connects = append(connects, c)
				}
			}
		default:
			connects = append(connects, connect{
				url: u,
			})
		}
	}

	for _, conn := range connects {
		if err := ch.tables(acc, &conn); err != nil {
			acc.AddError(err)
		}
		for metric := range commonMetrics {
			if err := ch.commonMetrics(acc, &conn, metric); err != nil {
				acc.AddError(err)
			}
		}
	}
	return nil
}

func (ch *ClickHouse) clusterIncludeExcludeFilter() string {
	if len(ch.ClusterInclude) == 0 && len(ch.ClusterExclude) == 0 {
		return ""
	}
	var (
		escape = func(in string) string {
			return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(in) + "'"
		}
		makeFilter = func(expr string, args []string) string {
			in := make([]string, 0, len(args))
			for _, v := range args {
				in = append(in, escape(v))
			}
			return fmt.Sprintf("cluster %s (%s)", expr, strings.Join(in, ", "))
		}
		includeFilter, excludeFilter string
	)

	if len(ch.ClusterInclude) != 0 {
		includeFilter = makeFilter("IN", ch.ClusterInclude)
	}
	if len(ch.ClusterExclude) != 0 {
		excludeFilter = makeFilter("NOT IN", ch.ClusterExclude)
	}
	if includeFilter != "" && excludeFilter != "" {
		return "WHERE " + includeFilter + " OR " + excludeFilter
	}
	if includeFilter == "" && excludeFilter != "" {
		return "WHERE " + excludeFilter
	}
	if includeFilter != "" && excludeFilter == "" {
		return "WHERE " + includeFilter
	}
	return ""
}

func (ch *ClickHouse) commonMetrics(acc telegraf.Accumulator, conn *connect, metric string) error {
	var result []struct {
		Metric string   `json:"metric"`
		Value  chUInt64 `json:"value"`
	}
	if err := ch.execQuery(conn.url, commonMetrics[metric], &result); err != nil {
		return err
	}

	tags := map[string]string{
		"source": conn.Hostname,
	}
	if len(conn.Cluster) != 0 {
		tags["cluster"] = conn.Cluster
	}
	if conn.ShardNum != 0 {
		tags["shard_num"] = strconv.Itoa(conn.ShardNum)
	}

	fields := make(map[string]interface{})
	for _, r := range result {
		fields[internal.SnakeCase(r.Metric)] = uint64(r.Value)
	}

	acc.AddFields("clickhouse_"+metric, fields, tags)

	return nil
}

func (ch *ClickHouse) tables(acc telegraf.Accumulator, conn *connect) error {
	var parts []struct {
		Database string   `json:"database"`
		Table    string   `json:"table"`
		Bytes    chUInt64 `json:"bytes"`
		Parts    chUInt64 `json:"parts"`
		Rows     chUInt64 `json:"rows"`
	}

	if err := ch.execQuery(conn.url, systemParts, &parts); err != nil {
		return err
	}
	tags := map[string]string{
		"source": conn.Hostname,
	}
	if len(conn.Cluster) != 0 {
		tags["cluster"] = conn.Cluster
	}
	if conn.ShardNum != 0 {
		tags["shard_num"] = strconv.Itoa(conn.ShardNum)
	}
	for _, part := range parts {
		tags["table"] = part.Table
		tags["database"] = part.Database
		acc.AddFields("clickhouse_tables",
			map[string]interface{}{
				"bytes": uint64(part.Bytes),
				"parts": uint64(part.Parts),
				"rows":  uint64(part.Rows),
			},
			tags,
		)
	}
	return nil
}

type clickhouseError struct {
	StatusCode int
	body       []byte
}

func (e *clickhouseError) Error() string {
	return fmt.Sprintf("received error code %d: %s", e.StatusCode, e.body)
}

func (ch *ClickHouse) execQuery(url *url.URL, query string, i interface{}) error {
	q := url.Query()
	q.Set("query", query+" FORMAT JSON")
	url.RawQuery = q.Encode()
	req, _ := http.NewRequest("GET", url.String(), nil)
	if ch.Username != "" {
		req.Header.Add("X-ClickHouse-User", ch.Username)
	}
	if ch.Password != "" {
		req.Header.Add("X-ClickHouse-Key", ch.Password)
	}
	resp, err := ch.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 200))
		return &clickhouseError{
			StatusCode: resp.StatusCode,
			body:       body,
		}
	}
	var response struct {
		Data json.RawMessage
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}
	return json.Unmarshal(response.Data, i)
}

// see https://clickhouse.yandex/docs/en/operations/settings/settings/#session_settings-output_format_json_quote_64bit_integers
type chUInt64 uint64

func (i *chUInt64) UnmarshalJSON(b []byte) error {
	b = bytes.TrimPrefix(b, []byte(`"`))
	b = bytes.TrimSuffix(b, []byte(`"`))
	v, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	*i = chUInt64(v)
	return nil
}

const (
	systemEventsSQL       = "SELECT event AS metric, CAST(value AS UInt64) AS value FROM system.events"
	systemMetricsSQL      = "SELECT          metric, CAST(value AS UInt64) AS value FROM system.metrics"
	systemAsyncMetricsSQL = "SELECT          metric, CAST(value AS UInt64) AS value FROM system.asynchronous_metrics"
	systemParts           = `
		SELECT
			database,
			table,
			SUM(bytes) AS bytes,
			COUNT(*)   AS parts,
			SUM(rows)  AS rows
		FROM system.parts
		WHERE active = 1
		GROUP BY
			database, table
		ORDER BY
			database, table
	`
)

var commonMetrics = map[string]string{
	"events":               systemEventsSQL,
	"metrics":              systemMetricsSQL,
	"asynchronous_metrics": systemAsyncMetricsSQL,
}

var _ telegraf.ServiceInput = &ClickHouse{}
