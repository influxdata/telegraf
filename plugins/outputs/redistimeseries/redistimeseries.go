package redistimeseries

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/mediocregopher/radix/v3"
	"github.com/mediocregopher/radix/v3/resp/resp2"
)

var sampleConfig = `
  ## The URI of the RedisTimeSeries server
  # URI = "redis://user:password@127.0.0.1:6379/0"

  ## The retention period of metrics in milliseconds, where:
  # -1 means the database's default retention
  # 0 means unlimited
  # retention = 0

  ## Output debug information
  # debug = false
`

// RedisTimeSeries represents a redistimeseries telegraf backend
type RedisTimeSeries struct {
	// TODO ## Optional TLS Config
	// TODO :# tls_ca = "/etc/telegraf/ca.pem"
	// TODO # tls_cert = "/etc/telegraf/cert.pem"
	// TODO # tls_key = "/etc/telegraf/key.pem"
	// TODO ## Use TLS but skip chain & host verification
	// TODO # insecure_skip_verify = true
	// TODO: externalize pipeline window
	// TODO: consider using MADD and TS.CREATE to skip label setting
	// TODO: consider adding introspective metrics (last write, time to parse, to write, errors, ...)
	// TODO: consider using go-redis as it is already in the deps
	URI       string
	Retention int64
	Debug     bool
	pool      *radix.Pool
}

// Connect handles the connection
func (rts *RedisTimeSeries) Connect() (err error) {
	opts := []radix.PoolOpt{
		radix.PoolPipelineWindow(0, 0),
	}
	rts.pool, err = radix.NewPool("tcp", rts.URI, 1, opts...)
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
		timestamp := metric.Time().UnixNano() / int64(time.Microsecond)

		labels := make([]string, len(metric.TagList())*2)
		for i, tag := range metric.TagList() {
			labels[i*2] = tag.Key
			labels[i*2+1] = tag.Value
			name = fmt.Sprintf("%s:%s", name, tag.Value)
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
				err = fmt.Errorf("D! [RedisTimeSeries] Skipping unknown value type %T of field '%s' in metric '%v'",
					field.Value, field.Key, metric)
			}
			if err == nil {
				args := []string{
					key,
					fmt.Sprint(timestamp),
					fmt.Sprint(value),
				}
				if rts.Retention >= 0 {
					args = append(args, "RETENTION", fmt.Sprint(rts.Retention))
				}
				if len(labels) > 0 {
					args = append(args, "LABELS")
					args = append(args, labels...)
				}

				err = rts.pool.Do(radix.Cmd(nil, "TS.ADD", args...))
				var redisErr resp2.Error
				if errors.As(err, &redisErr) {
					log.Printf("E! [RedisTimeSeries] %v", redisErr.E)
					return
				}
			} else {
				// Warnings issued during metric serialization
				if rts.Debug {
					log.Print(err)
				}
				err = nil
			}
		}
	}
	return
}

func init() {
	outputs.Add("RedisTimeSeries", func() telegraf.Output {
		return &RedisTimeSeries{}
	})
}
