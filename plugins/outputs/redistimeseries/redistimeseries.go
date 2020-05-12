package redistimeseries

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/mediocregopher/radix/v3"
	"github.com/mediocregopher/radix/v3/resp/resp2"
)

var (
	pluginName      = "RedisTimeSeries"
	defaultURL      = "redis://localhost:6379"
	defaultPoolSize = 1
	sampleConfig    = `
  ## The URLs of the RedisTimeSeries servers
  # urls = ["redis://127.0.0.1:6379"]

  ## When set to TRUE indicates that Redis is clustered
  # cluster = false

  ## Number of connections in the pool
  # pool_size = 1

  ## Connection pool pipeline window duration
  # pipeline_window_duration = "0ms"

  ## Connection pool pipeline window limit
  # pipeline_window_limit = 0

  ## When set to TRUE uses the database's default retention
  # default_retention = false

  ## When 'default_retention' is FALSE this is the retention period for
  ## RedisTimeSeries values where 0 means unlimited
  # retention = "0ms"

  ## Output debug information
  # debug = false
`
)

// TODO: cluster support
// TODO ## Optional TLS Config
// TODO :# tls_ca = "/etc/telegraf/ca.pem"
// TODO # tls_cert = "/etc/telegraf/cert.pem"
// TODO # tls_key = "/etc/telegraf/key.pem"
// TODO ## Use TLS but skip chain & host verification
// TODO # insecure_skip_verify = true
// TODO: consider using MADD and TS.CREATE to skip label setting
// TODO: consider adding introspective metrics (last write, time to parse, to write, errors, ...)
// TODO: consider using go-redis as it is already in the deps

// RedisTimeSeries represents a redistimeseries telegraf backend
type RedisTimeSeries struct {
	URLs                   []string          `toml:"urls"`
	Cluster                bool              `toml:"cluster"`
	PoolSize               int               `toml:"pool_size"`
	PipelineWindowDuration internal.Duration `toml:"pipeline_window_duration"`
	PipelineWindowLimit    int               `toml:"pipeline_window_limit"`
	DefaultRetention       bool              `toml:"default_retention"`
	Retention              internal.Duration `toml:"retention"`
	Debug                  bool              `toml:"debug"`
	rententionClause       []string
	pool                   *radix.Pool
}

// Connect handles the connection
func (rts *RedisTimeSeries) Connect() (err error) {
	if rts.Cluster {
		// TOOD: :)
		return fmt.Errorf("cluster not supported yet")
	}

	if !rts.DefaultRetention {
		rts.rententionClause = []string{"RETENTION", fmt.Sprint(rts.Retention.Duration.Milliseconds())}
	}

	poolSize := rts.PoolSize
	if poolSize == 0 {
		poolSize = defaultPoolSize
		if rts.Debug {
			log.Printf("[%s] set connection pool size to %d by default\n", pluginName, poolSize)
		}
	}

	urls := make([]string, 0, len(rts.URLs))
	urls = append(urls, rts.URLs...)

	if len(urls) == 0 {
		urls = append(urls, defaultURL)
	}

	if len(urls) > 1 && !rts.Cluster {
		// TOOD: :)
		return fmt.Errorf("urls list should be len 0 or 1 when not a cluster")
	}

	for _, u := range urls {
		parts, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("error parsing url [%q]: %v", u, err)
		}
		switch parts.Scheme {
		case "redis":
			continue
		default:
			return fmt.Errorf("unsupported scheme [%q]: %q", u, parts.Scheme)
		}
	}

	opts := []radix.PoolOpt{
		radix.PoolPipelineWindow(rts.PipelineWindowDuration.Duration, rts.PipelineWindowLimit),
	}
	rts.pool, err = radix.NewPool("tcp", urls[0], poolSize, opts...)
	return
}

// Close handles closing the connection
func (rts *RedisTimeSeries) Close() (err error) {
	err = rts.pool.Close()
	return
}

// Description returns a description
func (rts *RedisTimeSeries) Description() string {
	return "Configuration for sending metrics to RedisTimeSeries"
}

// SampleConfig returns a sample config
func (rts *RedisTimeSeries) SampleConfig() string {
	return sampleConfig
}

type byTimestamp []telegraf.Metric

func (b byTimestamp) Len() int           { return len(b) }
func (b byTimestamp) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byTimestamp) Less(i, j int) bool { return b[i].Time().UnixNano() < b[j].Time().UnixNano() }

// Write writes metrics to RedisTimeSeries
func (rts *RedisTimeSeries) Write(metrics []telegraf.Metric) (err error) {
	if len(metrics) == 0 {
		return nil
	}

	// Ensure metrics are in ascending timestamp order as RTS is picky
	sort.Sort(byTimestamp(metrics))

	for _, metric := range metrics {
		name := metric.Name()
		timestamp := metric.Time().UnixNano() / int64(time.Millisecond)

		labels := make([]string, len(metric.TagList())*2)
		for i, tag := range metric.TagList() {
			if tag.Key == "host" {
				name = fmt.Sprintf("%s:%s", tag.Value, name)
			} else {
				name = fmt.Sprintf("%s:%s", name, tag.Value)
			}
			labels[i*2] = tag.Key
			labels[i*2+1] = tag.Value
		}

		for _, field := range metric.FieldList() {
			key := fmt.Sprintf("%s:%s", name, field.Key)
			var value float64
			switch v := field.Value.(type) {
			case float64:
				value = v
			case float32:
			case uint64:
			case int64:
				value = float64(v)
			default:
				if rts.Debug {
					log.Printf("[%s] skipping over unknown value type %T of field '%s' in metric '%v'",
						pluginName, field.Value, field.Key, metric)
				}
				continue
			}

			args := []string{
				key,
				fmt.Sprint(timestamp),
				fmt.Sprint(value),
			}

			if !rts.DefaultRetention {
				args = append(args, rts.rententionClause...)
			}

			if len(labels) > 0 {
				args = append(args, "LABELS")
				args = append(args, labels...)
			}

			err = rts.pool.Do(radix.Cmd(nil, "TS.ADD", args...))
			if err != nil {
				var redisErr resp2.Error
				if errors.As(err, &redisErr) {
					// Handle Redis errors
				} else {
					// Handle generic errors
				}
				if rts.Debug {
					log.Printf("[%s] error while storing value '%v' of field '%s' in metric '%v'\n",
						pluginName, field.Value, field.Key, metric)
				}
				return
			}
		}
	}
	return
}

func init() {
	outputs.Add(pluginName, func() telegraf.Output {
		return &RedisTimeSeries{}
	})
}
