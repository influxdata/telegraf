// Package xxhash provides a Telegraf plugin that computes
// a deterministic xxHash64 from selected tag and field values.
package xxhash

import (
	_ "embed"
	"fmt"

	"strconv"
	"strings"
	"time"

	"encoding"
	"regexp"

	"github.com/cespare/xxhash/v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type XXHash struct {
	Keys      []string `toml:"keys"`
	KeysMode  string   `toml:"keys_mode"`
	TagHash   string   `toml:"tag_hash"`
	FieldHash string   `toml:"field_hash"`

	exactKeys map[string]struct{}
	regexes   []*regexp.Regexp
}

func (*XXHash) SampleConfig() string {
	return sampleConfig
}

func (*XXHash) Description() string {
	return "Compute xxHash64 of selected metric tags/fields and store as tag/field"
}

func (p *XXHash) Init() error {
	p.exactKeys = make(map[string]struct{})

	for _, k := range p.Keys {
		switch p.KeysMode {
		case "exact":
			p.exactKeys[k] = struct{}{}
		case "regex":
			re, err := regexp.Compile(k)
			if err != nil {
				return fmt.Errorf("invalid regex in keys: %q: %w", k, err)
			}
			p.regexes = append(p.regexes, re)
		case "auto":
			if looksLikeRegex(k) {
				re, err := regexp.Compile(k)
				if err != nil {
					return fmt.Errorf("invalid regex in keys: %q: %w", k, err)
				}
				p.regexes = append(p.regexes, re)
			} else {
				p.exactKeys[k] = struct{}{}
			}
		default:
			return fmt.Errorf("invalid keys_mode: %q", p.KeysMode)
		}
	}
	return nil
}

func (p *XXHash) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		var combined uint64 = 0

		// XOR hashes of matching tag key=values, using Sum64String
		for k, v := range metric.Tags() {
			if p.matchKey(k) {
				combined ^= xxhash.Sum64String(k + "=" + v)
			}
		}

		// XOR hashes of matching field key=values
		for k, v := range metric.Fields() {
			if p.matchKey(k) {
				buf := append([]byte(k+"="), toBytes(v)...)
				combined ^= xxhash.Sum64(buf)
			}
		}

		if combined == 0 {
			// optionally skip metrics that have no matching keys
			continue
		}

		if p.FieldHash != "" {
			metric.AddField(p.FieldHash, int64(combined))
		}
		if p.TagHash != "" {
			metric.AddTag(p.TagHash, strconv.FormatUint(combined, 10))
		}
	}
	return metrics
}

func (p *XXHash) matchKey(k string) bool {
	if _, ok := p.exactKeys[k]; ok {
		return true
	}
	for _, re := range p.regexes {
		if re.MatchString(k) {
			return true
		}
	}
	return false
}

func toBytes(val interface{}) []byte {
	switch v := val.(type) {
	case string:
		return []byte(v)
	case []byte:
		return v
	case bool:
		if v {
			return []byte("1")
		}
		return []byte("0")
	case int:
		return []byte(strconv.FormatInt(int64(v), 10))
	case int64:
		return []byte(strconv.FormatInt(v, 10))
	case uint64:
		return []byte(strconv.FormatUint(v, 10))
	case float64:
		return []byte(strconv.FormatFloat(v, 'g', -1, 64))
	case float32:
		return []byte(strconv.FormatFloat(float64(v), 'g', -1, 32))
	case time.Time:
		return []byte(v.UTC().Format(time.RFC3339Nano))
	case encoding.TextMarshaler:
		if b, err := v.MarshalText(); err == nil {
			return b
		}
	}
	return []byte(fmt.Sprintf("%v", val))
}

func looksLikeRegex(s string) bool {
	return strings.ContainsAny(s, `^$.+*?[]{}()|\\`)
}

func init() {
	processors.Add("xxhash", func() telegraf.Processor {
		return &XXHash{
			KeysMode: "auto",
		}
	})
}
