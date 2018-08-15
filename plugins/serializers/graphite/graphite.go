package graphite

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

const DEFAULT_TEMPLATE = "host.tags.measurement.field"

var (
	allowedChars = regexp.MustCompile(`[^a-zA-Z0-9-:._=\p{L}]`)
	hypenChars   = strings.NewReplacer(
		"/", "-",
		"@", "-",
		"*", "-",
	)
	dropChars = strings.NewReplacer(
		`\`, "",
		"..", ".",
	)

	fieldDeleter = strings.NewReplacer(".FIELDNAME", "", "FIELDNAME.", "")
)

type GraphiteSerializer struct {
	Prefix     string
	Template   string
	TagSupport bool
}

func (s *GraphiteSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	out := []byte{}

	// Convert UnixNano to Unix timestamps
	timestamp := metric.Time().UnixNano() / 1000000000

	switch s.TagSupport {
	case true:
		for fieldName, value := range metric.Fields() {
			fieldValue := formatValue(value)
			if fieldValue == "" {
				continue
			}
			bucket := SerializeBucketNameWithTags(metric.Name(), metric.Tags(), s.Prefix, fieldName)
			metricString := fmt.Sprintf("%s %s %d\n",
				// insert "field" section of template
				bucket,
				//bucket,
				fieldValue,
				timestamp)
			point := []byte(metricString)
			out = append(out, point...)
		}
	default:
		bucket := SerializeBucketName(metric.Name(), metric.Tags(), s.Template, s.Prefix)
		if bucket == "" {
			return out, nil
		}

		for fieldName, value := range metric.Fields() {
			fieldValue := formatValue(value)
			if fieldValue == "" {
				continue
			}
			metricString := fmt.Sprintf("%s %s %d\n",
				// insert "field" section of template
				sanitize(InsertField(bucket, fieldName)),
				fieldValue,
				timestamp)
			point := []byte(metricString)
			out = append(out, point...)
		}
	}
	return out, nil
}

func (s *GraphiteSerializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	var batch bytes.Buffer
	for _, m := range metrics {
		buf, err := s.Serialize(m)
		if err != nil {
			return nil, err
		}
		_, err = batch.Write(buf)
		if err != nil {
			return nil, err
		}
	}
	return batch.Bytes(), nil
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return ""
	case bool:
		if v {
			return "1"
		} else {
			return "0"
		}
	case uint64:
		return strconv.FormatUint(v, 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		if math.IsNaN(v) {
			return ""
		}

		if math.IsInf(v, 0) {
			return ""
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	}

	return ""
}

// SerializeBucketName will take the given measurement name and tags and
// produce a graphite bucket. It will use the GraphiteSerializer.Template
// to generate this, or DEFAULT_TEMPLATE.
//
// NOTE: SerializeBucketName replaces the "field" portion of the template with
// FIELDNAME. It is up to the user to replace this. This is so that
// SerializeBucketName can be called just once per measurement, rather than
// once per field. See GraphiteSerializer.InsertField() function.
func SerializeBucketName(
	measurement string,
	tags map[string]string,
	template string,
	prefix string,
) string {
	if template == "" {
		template = DEFAULT_TEMPLATE
	}
	tagsCopy := make(map[string]string)
	for k, v := range tags {
		tagsCopy[k] = v
	}

	var out []string
	templateParts := strings.Split(template, ".")
	for _, templatePart := range templateParts {
		switch templatePart {
		case "measurement":
			out = append(out, measurement)
		case "tags":
			// we will replace this later
			out = append(out, "TAGS")
		case "field":
			// user of SerializeBucketName needs to replace this
			out = append(out, "FIELDNAME")
		default:
			// This is a tag being applied
			if tagvalue, ok := tagsCopy[templatePart]; ok {
				out = append(out, strings.Replace(tagvalue, ".", "_", -1))
				delete(tagsCopy, templatePart)
			}
		}
	}

	// insert remaining tags into output name
	for i, templatePart := range out {
		if templatePart == "TAGS" {
			out[i] = buildTags(tagsCopy)
			break
		}
	}

	if len(out) == 0 {
		return ""
	}

	if prefix == "" {
		return strings.Join(out, ".")
	}
	return prefix + "." + strings.Join(out, ".")
}

// SerializeBucketNameWithTags will take the given measurement name and tags and
// produce a graphite bucket. It will use the Graphite11Serializer.
// http://graphite.readthedocs.io/en/latest/tags.html
func SerializeBucketNameWithTags(
	measurement string,
	tags map[string]string,
	prefix string,
	field string,
) string {
	var out string
	var tagsCopy []string
	for k, v := range tags {
		if k == "name" {
			k = "_name"
		}
		tagsCopy = append(tagsCopy, sanitize(k+"="+v))
	}
	sort.Strings(tagsCopy)

	if prefix != "" {
		out = prefix + "."
	}

	out += measurement

	if field != "value" {
		out += "." + field
	}

	out = sanitize(out)

	if len(tagsCopy) > 0 {
		out += ";" + strings.Join(tagsCopy, ";")
	}

	return out
}

// InsertField takes the bucket string from SerializeBucketName and replaces the
// FIELDNAME portion. If fieldName == "value", it will simply delete the
// FIELDNAME portion.
func InsertField(bucket, fieldName string) string {
	// if the field name is "value", then dont use it
	if fieldName == "value" {
		return fieldDeleter.Replace(bucket)
	}
	return strings.Replace(bucket, "FIELDNAME", fieldName, 1)
}

func buildTags(tags map[string]string) string {
	var keys []string
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var tag_str string
	for i, k := range keys {
		tag_value := strings.Replace(tags[k], ".", "_", -1)
		if i == 0 {
			tag_str += tag_value
		} else {
			tag_str += "." + tag_value
		}
	}
	return tag_str
}

func sanitize(value string) string {
	// Apply special hypenation rules to preserve backwards compatibility
	value = hypenChars.Replace(value)
	// Apply rule to drop some chars to preserve backwards compatibility
	value = dropChars.Replace(value)
	// Replace any remaining illegal chars
	return allowedChars.ReplaceAllLiteralString(value, "_")
}
