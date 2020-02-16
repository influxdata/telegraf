package clickhouse

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var defaultTimeout = 5 * time.Second

var sampleConfig = `
  ## HTTP(s) timeout while getting metrics values in seconds
  ## The timeout includes connection time, any redirects, and reading the response body.
  ##   example: timeout = 1
  # timeout = 5 # seconds

  ## List of servers for metrics scraping
  ## metrics scrape via HTTP(s) clickhouse interface
  ## https://clickhouse.tech/docs/en/interfaces/http/
  ##    example: servers = ["http://default:@127.0.0.1:8123","https://user:password@custom-server.mdb.yandexclud.net"]
  servers         = ["http://username:password@127.0.0.1:8123"]

  ## If "auto_discovery"" is "true" plugin tries to connect to all servers available in the cluster
  ## with using same "user:password" described in "servers"" parameter
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
  ## when this filter present then "WHERE cluster IN (...)" filter will apply
  ## if "cluster_include" not empty, this parameter will ignored
  ##    example: cluster_exclude = ["my-internal-not-discovered-cluster"]
  # cluster_exclude = []

  ## Parameter which controls whether a client verifies the server's certificate chain and host name.
  ## If http_tls_insecure_skip_verify is true, TLS accepts any certificate
  ## presented by the server and any host name in that certificate.
  # http_tls_insecure_skip_verify = true
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
			AutoDiscovery:      true,
			InsecureSkipVerify: true,
			Timeout:            5,
		}
	})
}

// ClickHouse Telegraf Input Plugin
type ClickHouse struct {
	Servers            []string `toml:"servers"`
	AutoDiscovery      bool     `toml:"auto_discovery"`
	ClusterInclude     []string `toml:"cluster_include"`
	ClusterExclude     []string `toml:"cluster_exclude"`
	Timeout            int      `toml:"timeout"`
	InsecureSkipVerify bool     `toml:"http_tls_insecure_skip_verify"`
	client             http.Client
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
	if ch.Timeout != 0 {
		timeout = time.Duration(ch.Timeout) * time.Second
	}
	ch.client = http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: ch.InsecureSkipVerify,
			},
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
						User:   u.User,
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
			return "cluster " + expr + " (" + strings.Join(in, ", ") + ")"
		}
	)

	if len(ch.ClusterInclude) != 0 {
		return "WHERE " + makeFilter("IN", ch.ClusterInclude)
	}

	return "WHERE " + makeFilter("NOT IN", ch.ClusterExclude)
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
	return fmt.Sprintf("ClickHouse server returned an error code: %d\n%s", e.StatusCode, e.body)
}

func (ch *ClickHouse) execQuery(url *url.URL, query string, i interface{}) error {
	q := url.Query()
	q.Set("query", query+" FORMAT JSON")
	url.RawQuery = q.Encode()
	resp, err := ch.client.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
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
