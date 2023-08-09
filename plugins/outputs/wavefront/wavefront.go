//go:generate ../../../tools/readme_config_includer/generator
package wavefront

import (
	_ "embed"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	serializer "github.com/influxdata/telegraf/plugins/serializers/wavefront"
)

//go:embed sample.conf
var sampleConfig string

const maxTagLength = 254

type Wavefront struct {
	URL                  string                          `toml:"url"`
	Token                config.Secret                   `toml:"token"`
	Host                 string                          `toml:"host" deprecated:"2.4.0;use url instead"`
	Port                 int                             `toml:"port" deprecated:"2.4.0;use url instead"`
	Prefix               string                          `toml:"prefix"`
	SimpleFields         bool                            `toml:"simple_fields"`
	MetricSeparator      string                          `toml:"metric_separator"`
	ConvertPaths         bool                            `toml:"convert_paths"`
	ConvertBool          bool                            `toml:"convert_bool"`
	HTTPMaximumBatchSize int                             `toml:"http_maximum_batch_size"`
	Timeout              config.Duration                 `toml:"timeout"`
	UseRegex             bool                            `toml:"use_regex"`
	UseStrict            bool                            `toml:"use_strict"`
	TruncateTags         bool                            `toml:"truncate_tags"`
	ImmediateFlush       bool                            `toml:"immediate_flush"`
	SourceOverride       []string                        `toml:"source_override"`
	StringToNumber       map[string][]map[string]float64 `toml:"string_to_number" deprecated:"1.9.0;use the enum processor instead"`
	tls.ClientConfig
	sender wavefront.Sender
	Log    telegraf.Logger `toml:"-"`
}

// instead of Sanitize which may miss some special characters we can use a regex pattern, but this is significantly slower than Sanitize
var sanitizedRegex = regexp.MustCompile(`[^a-zA-Z\d_.-]`)

var tagValueReplacer = strings.NewReplacer("*", "-")

var pathReplacer = strings.NewReplacer("_", "_")

func (*Wavefront) SampleConfig() string {
	return sampleConfig
}

func (w *Wavefront) senderURLFromURLAndToken() (string, error) {
	newURL, err := url.Parse(w.URL)
	if err != nil {
		return "", fmt.Errorf("could not parse the provided URL: %s", w.URL)
	}

	token := "DUMMY_TOKEN"
	if !w.Token.Empty() {
		b, err := w.Token.Get()
		if err != nil {
			return "", fmt.Errorf("getting token failed: %w", err)
		}
		token = string(b)
		config.ReleaseSecret(b)
	}
	newURL.User = url.User(token)

	return newURL.String(), nil
}

func senderURLFromHostAndPort(host string, port int) string {
	return fmt.Sprintf("http://%s:%d", host, port)
}

func (w *Wavefront) Connect() error {
	flushSeconds := 5
	if w.ImmediateFlush {
		flushSeconds = 86400 // Set a very long flush interval if we're flushing directly
	}
	var connectionURL string
	if w.URL != "" {
		w.Log.Debugf("Connecting over http/https using url: %s", w.URL)
		connectionURLWithToken, err := w.senderURLFromURLAndToken()
		if err != nil {
			return err
		}
		connectionURL = connectionURLWithToken
	} else {
		w.Log.Warn("Configuration with host/port is deprecated. Please use url.")
		w.Log.Debugf("Connecting over http using Host: %q and Port: %d", w.Host, w.Port)
		connectionURL = senderURLFromHostAndPort(w.Host, w.Port)
	}

	tlsConfig, err := w.TLSConfig()
	if err != nil {
		return err
	}

	sender, err := wavefront.NewSender(connectionURL,
		wavefront.BatchSize(w.HTTPMaximumBatchSize),
		wavefront.FlushIntervalSeconds(flushSeconds),
		wavefront.TLSConfigOptions(tlsConfig),
		wavefront.Timeout(time.Duration(w.Timeout)),
	)

	if err != nil {
		return fmt.Errorf("could not create Wavefront Sender for the provided url")
	}

	w.sender = sender

	if w.ConvertPaths && w.MetricSeparator == "_" {
		w.ConvertPaths = false
	}
	if w.ConvertPaths {
		pathReplacer = strings.NewReplacer("_", w.MetricSeparator)
	}
	return nil
}

func (w *Wavefront) Write(metrics []telegraf.Metric) error {
	for _, m := range metrics {
		for _, point := range w.buildMetrics(m) {
			err := w.sender.SendMetric(point.Metric, point.Value, point.Timestamp, point.Source, point.Tags)
			if err != nil {
				if isRetryable(err) {
					// The internal buffer in the Wavefront SDK is full. To prevent data loss,
					// we flush the buffer (which is a blocking operation) and try again.
					w.Log.Debug("SDK buffer overrun, forcibly flushing the buffer")
					if err = w.sender.Flush(); err != nil {
						return fmt.Errorf("wavefront flushing error: %w", err)
					}
					// Try again.
					err = w.sender.SendMetric(point.Metric, point.Value, point.Timestamp, point.Source, point.Tags)
					if err != nil {
						if isRetryable(err) {
							return fmt.Errorf("wavefront sending error: %w", err)
						}
					}
				}
				w.Log.Errorf("Non-retryable error during Wavefront.Write: %v", err)
				w.Log.Debugf("Non-retryable metric data: %+v", point)
			}
		}
	}
	if w.ImmediateFlush {
		w.Log.Debugf("Flushing batch of %d points", len(metrics))
		return w.sender.Flush()
	}
	return nil
}

func (w *Wavefront) buildMetrics(m telegraf.Metric) []*serializer.MetricPoint {
	ret := make([]*serializer.MetricPoint, 0)

	for fieldName, value := range m.Fields() {
		var name string
		if !w.SimpleFields && fieldName == "value" {
			name = fmt.Sprintf("%s%s", w.Prefix, m.Name())
		} else {
			name = fmt.Sprintf("%s%s%s%s", w.Prefix, m.Name(), w.MetricSeparator, fieldName)
		}

		if w.UseRegex {
			name = sanitizedRegex.ReplaceAllLiteralString(name, "-")
		} else {
			name = serializer.Sanitize(w.UseStrict, name)
		}

		if w.ConvertPaths {
			name = pathReplacer.Replace(name)
		}

		metric := &serializer.MetricPoint{
			Metric:    name,
			Timestamp: m.Time().Unix(),
		}

		metricValue, buildError := buildValue(value, metric.Metric, w)
		if buildError != nil {
			w.Log.Debugf("Error building tags: %s\n", buildError.Error())
			continue
		}
		metric.Value = metricValue

		source, tags := w.buildTags(m.Tags())
		metric.Source = source
		metric.Tags = tags

		ret = append(ret, metric)
	}
	return ret
}

func (w *Wavefront) buildTags(mTags map[string]string) (string, map[string]string) {
	// Remove all empty tags.
	for k, v := range mTags {
		if v == "" {
			delete(mTags, k)
		}
	}

	// find source, use source_override property if needed
	var source string
	if s, ok := mTags["source"]; ok {
		source = s
		delete(mTags, "source")
	} else {
		sourceTagFound := false
		for _, s := range w.SourceOverride {
			for k, v := range mTags {
				if k == s {
					source = v
					if mTags["host"] != "" {
						mTags["telegraf_host"] = mTags["host"]
					}

					sourceTagFound = true
					delete(mTags, k)
					break
				}
			}
			if sourceTagFound {
				break
			}
		}

		if !sourceTagFound {
			source = mTags["host"]
		}
	}
	source = tagValueReplacer.Replace(source)

	// remove default host tag
	delete(mTags, "host")

	// sanitize tag keys and values
	tags := make(map[string]string)
	for k, v := range mTags {
		var key string
		if w.UseRegex {
			key = sanitizedRegex.ReplaceAllLiteralString(k, "-")
		} else {
			key = serializer.Sanitize(w.UseStrict, k)
		}
		val := tagValueReplacer.Replace(v)
		if w.TruncateTags {
			if len(key) > maxTagLength {
				w.Log.Warnf("Tag key length > 254. Skipping tag: %s", key)
				continue
			}
			if len(key)+len(val) > maxTagLength {
				w.Log.Debugf("Key+value length > 254: %s", key)
				val = val[:maxTagLength-len(key)]
			}
		}
		tags[key] = val
	}

	return source, tags
}

func buildValue(v interface{}, name string, w *Wavefront) (float64, error) {
	switch p := v.(type) {
	case bool:
		if w.ConvertBool {
			if p {
				return 1, nil
			}
			return 0, nil
		}
	case int64:
		return float64(v.(int64)), nil
	case uint64:
		return float64(v.(uint64)), nil
	case float64:
		return v.(float64), nil
	case string:
		for prefix, mappings := range w.StringToNumber {
			if strings.HasPrefix(name, prefix) {
				for _, mapping := range mappings {
					val, hasVal := mapping[p]
					if hasVal {
						return val, nil
					}
				}
			}
		}
		return 0, fmt.Errorf("unexpected type: %T, with value: %v, for: %s", v, v, name)
	default:
		return 0, fmt.Errorf("unexpected type: %T, with value: %v, for: %s", v, v, name)
	}
	return 0, fmt.Errorf("unexpected type: %T, with value: %v, for: %s", v, v, name)
}

func (w *Wavefront) Close() error {
	w.sender.Close()
	return nil
}

func init() {
	outputs.Add("wavefront", func() telegraf.Output {
		return &Wavefront{
			MetricSeparator:      ".",
			ConvertPaths:         true,
			ConvertBool:          true,
			TruncateTags:         false,
			ImmediateFlush:       true,
			HTTPMaximumBatchSize: 10000,
			Timeout:              config.Duration(10 * time.Second),
		}
	})
}

// TODO: Currently there's no canonical way to exhaust all
// retryable/non-retryable errors from wavefront, so this implementation just
// handles known non-retryable errors in a case-by-case basis and assumes all
// other errors are retryable.
// A support ticket has been filed against wavefront to provide a canonical way
// to distinguish between retryable and non-retryable errors (link is not
// public).
func isRetryable(err error) bool {
	if err != nil {
		// "empty metric name" errors are non-retryable as retry will just keep
		// getting the same error again and again.
		if strings.Contains(err.Error(), "empty metric name") {
			return false
		}
	}
	return true
}
