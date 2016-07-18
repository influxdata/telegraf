package graphite

import (
	"fmt"
	"sort"
	"strings"

	"github.com/influxdata/telegraf"
)

const DEFAULT_TEMPLATE = "host.tags.measurement.field"

var (
	fieldDeleter   = strings.NewReplacer(".FIELDNAME", "", "FIELDNAME.", "")
	sanitizedChars = strings.NewReplacer("/", "-", "@", "-", "*", "-", " ", "_", "..", ".")
)

type GraphiteSerializer struct {
	Prefix   string
	Template string
}

func (s *GraphiteSerializer) Serialize(metric telegraf.Metric) ([]string, error) {
	out := []string{}

	// Convert UnixNano to Unix timestamps
	timestamp := metric.UnixNano() / 1000000000

	bucket := SerializeBucketName(metric.Name(), metric.Tags(), s.Template, s.Prefix)
	if bucket == "" {
		return out, nil
	}

	for fieldName, value := range metric.Fields() {
		// Convert value to string
		valueS := fmt.Sprintf("%#v", value)
		point := fmt.Sprintf("%s %s %d",
			// insert "field" section of template
			InsertField(bucket, fieldName),
			valueS,
			timestamp)
		out = append(out, point)
	}
	return out, nil
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
		return sanitizedChars.Replace(strings.Join(out, "."))
	}
	return sanitizedChars.Replace(prefix + "." + strings.Join(out, "."))
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
