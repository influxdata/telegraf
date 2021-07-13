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
	"github.com/influxdata/telegraf/filter"
)

const DefaultTemplate = "host.tags.measurement.field"

var (
	strictAllowedChars          = regexp.MustCompile(`[^a-zA-Z0-9-:._=\p{L}]`)
	compatibleAllowedCharsName  = regexp.MustCompile(`[^ "-:\<>-\]_a-~\p{L}]`)
	compatibleAllowedCharsValue = regexp.MustCompile(`[^ -:<-~\p{L}]`)
	compatibleLeadingTildeDrop  = regexp.MustCompile(`^[~]*(.*)`)
	hyphenChars                 = strings.NewReplacer(
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

type GraphiteTemplate struct {
	Filter filter.Filter
	Value  string
}

type GraphiteSerializer struct {
	Prefix          string
	Template        string
	TagSupport      bool
	TagSanitizeMode string
	Separator       string
	Templates       []*GraphiteTemplate
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
			bucket := SerializeBucketNameWithTags(metric.Name(), metric.Tags(), s.Prefix, s.Separator, fieldName, s.TagSanitizeMode)
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
		template := s.Template
		for _, graphiteTemplate := range s.Templates {
			if graphiteTemplate.Filter.Match(metric.Name()) {
				template = graphiteTemplate.Value
				break
			}
		}

		bucket := SerializeBucketName(metric.Name(), metric.Tags(), template, s.Prefix)
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
				strictSanitize(InsertField(bucket, fieldName)),
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
		}
		return "0"
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
// to generate this, or DefaultTemplate.
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
		template = DefaultTemplate
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

func InitGraphiteTemplates(templates []string) ([]*GraphiteTemplate, string, error) {
	var graphiteTemplates []*GraphiteTemplate
	defaultTemplate := ""

	for i, t := range templates {
		parts := strings.Fields(t)

		if len(parts) == 0 {
			return nil, "", fmt.Errorf("missing template at position: %d", i)
		}
		if len(parts) == 1 {
			if parts[0] == "" {
				return nil, "", fmt.Errorf("missing template at position: %d", i)
			}

			// Override default template
			defaultTemplate = t
			continue
		}

		if len(parts) > 2 {
			return nil, "", fmt.Errorf("invalid template format: '%s'", t)
		}

		tFilter, err := filter.Compile([]string{parts[0]})

		if err != nil {
			return nil, "", err
		}

		graphiteTemplates = append(graphiteTemplates, &GraphiteTemplate{
			Filter: tFilter,
			Value:  parts[1],
		})
	}

	return graphiteTemplates, defaultTemplate, nil
}

// SerializeBucketNameWithTags will take the given measurement name and tags and
// produce a graphite bucket. It will use the Graphite11Serializer.
// http://graphite.readthedocs.io/en/latest/tags.html
func SerializeBucketNameWithTags(
	measurement string,
	tags map[string]string,
	prefix string,
	separator string,
	field string,
	tagSanitizeMode string,
) string {
	var out string
	var tagsCopy []string
	for k, v := range tags {
		if k == "name" {
			k = "_name"
		}
		if tagSanitizeMode == "compatible" {
			tagsCopy = append(tagsCopy, compatibleSanitize(k, v))
		} else {
			tagsCopy = append(tagsCopy, strictSanitize(k+"="+v))
		}
	}
	sort.Strings(tagsCopy)

	if prefix != "" {
		out = prefix + separator
	}

	out += measurement

	if field != "value" {
		out += separator + field
	}

	out = strictSanitize(out)

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

	var tagStr string
	for i, k := range keys {
		tagValue := strings.Replace(tags[k], ".", "_", -1)
		if i == 0 {
			tagStr += tagValue
		} else {
			tagStr += "." + tagValue
		}
	}
	return tagStr
}

func strictSanitize(value string) string {
	// Apply special hyphenation rules to preserve backwards compatibility
	value = hyphenChars.Replace(value)
	// Apply rule to drop some chars to preserve backwards compatibility
	value = dropChars.Replace(value)
	// Replace any remaining illegal chars
	return strictAllowedChars.ReplaceAllLiteralString(value, "_")
}

func compatibleSanitize(name string, value string) string {
	name = compatibleAllowedCharsName.ReplaceAllLiteralString(name, "_")
	value = compatibleAllowedCharsValue.ReplaceAllLiteralString(value, "_")
	value = compatibleLeadingTildeDrop.FindStringSubmatch(value)[1]
	return name + "=" + value
}
