package carbon2

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const sanitizedChars = "!@#$%^&*()+`'\"[]{};<>,?/\\|="

type Serializer struct {
	Format              string `toml:"carbon2_format"`
	SanitizeReplaceChar string `toml:"carbon2_sanitize_replace_char"`

	sanitizeReplacer *strings.Replacer
}

func (s *Serializer) Init() error {
	if len(s.SanitizeReplaceChar) > 1 {
		return errors.New("sanitize replace char has to be a singular character")
	}

	if s.SanitizeReplaceChar == "" {
		s.SanitizeReplaceChar = ":"
	}
	s.sanitizeReplacer = createSanitizeReplacer(sanitizedChars, rune(s.SanitizeReplaceChar[0]))

	switch s.Format {
	case "":
		// default value
		s.Format = "field_separate"
	case "field_separate", "metric_includes_field":
		// valid choices, do nothing
	default:
		return fmt.Errorf("unknown carbon2 format: %s", s.Format)
	}

	return nil
}

func (s *Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return s.createObject(metric), nil
}

func (s *Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var batch bytes.Buffer
	for _, metric := range metrics {
		batch.Write(s.createObject(metric))
	}
	return batch.Bytes(), nil
}

func (s *Serializer) createObject(metric telegraf.Metric) []byte {
	var m bytes.Buffer

	for fieldName, fieldValue := range metric.Fields() {
		if isString(fieldValue) {
			continue
		}

		name := s.sanitizeReplacer.Replace(metric.Name())

		switch s.Format {
		case "field_separate":
			m.WriteString(serializeMetricFieldSeparate(name, fieldName))
		case "metric_includes_field":
			m.WriteString(serializeMetricIncludeField(name, fieldName))
		}

		for _, tag := range metric.TagList() {
			m.WriteString(strings.ReplaceAll(tag.Key, " ", "_"))
			m.WriteString("=")
			value := tag.Value
			if len(value) == 0 {
				value = "null"
			}
			m.WriteString(strings.ReplaceAll(value, " ", "_"))
			m.WriteString(" ")
		}
		m.WriteString(" ")
		m.WriteString(formatValue(fieldValue))
		m.WriteString(" ")
		m.WriteString(strconv.FormatInt(metric.Time().Unix(), 10))
		m.WriteString("\n")
	}
	return m.Bytes()
}

func serializeMetricFieldSeparate(name, fieldName string) string {
	return fmt.Sprintf("metric=%s field=%s ",
		strings.ReplaceAll(name, " ", "_"),
		strings.ReplaceAll(fieldName, " ", "_"),
	)
}

func serializeMetricIncludeField(name, fieldName string) string {
	return fmt.Sprintf("metric=%s_%s ",
		strings.ReplaceAll(name, " ", "_"),
		strings.ReplaceAll(fieldName, " ", "_"),
	)
}

func formatValue(fieldValue interface{}) string {
	switch v := fieldValue.(type) {
	case bool:
		// Print bools as 0s and 1s
		return fmt.Sprintf("%d", bool2int(v))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func isString(v interface{}) bool {
	switch v.(type) {
	case string:
		return true
	default:
		return false
	}
}

func bool2int(b bool) int {
	// Slightly more optimized than a usual if ... return ... else return ... .
	// See: https://0x0f.me/blog/golang-compiler-optimization/
	var i int
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}

// createSanitizeReplacer creates string replacer replacing all provided
// characters with the replaceChar.
func createSanitizeReplacer(sanitizedChars string, replaceChar rune) *strings.Replacer {
	sanitizeCharPairs := make([]string, 0, 2*len(sanitizedChars))
	for _, c := range sanitizedChars {
		sanitizeCharPairs = append(sanitizeCharPairs, string(c), string(replaceChar))
	}
	return strings.NewReplacer(sanitizeCharPairs...)
}

func init() {
	serializers.Add("carbon2",
		func() serializers.Serializer {
			return &Serializer{}
		},
	)
}

// InitFromConfig is a compatibility function to construct the parser the old way
func (s *Serializer) InitFromConfig(cfg *serializers.Config) error {
	s.Format = cfg.Carbon2Format
	s.SanitizeReplaceChar = cfg.Carbon2SanitizeReplaceChar

	return nil
}
