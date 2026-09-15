package jolokia2

import (
	"fmt"
	"slices"
	"strings"
)

type point struct {
	Tags   map[string]string
	Fields map[string]any
}

type pointBuilder struct {
	metric           Metric
	objectAttributes []string
	objectPath       string
	substitutions    []string
}

func NewPointBuilder(metric Metric, attributes []string, path string) *pointBuilder {
	return &pointBuilder{
		metric:           metric,
		objectAttributes: attributes,
		objectPath:       path,
		substitutions:    makeSubstitutionList(metric.Mbean),
	}
}

// Build generates a point for a given mbean name/pattern and value object.
func (pb *pointBuilder) Build(mbean string, value any) ([]point, error) {
	hasPattern := strings.Contains(mbean, "*")
	if !hasPattern || value == nil {
		value = map[string]any{mbean: value}
	}

	valueMap, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("the response of %s's value should be a map", mbean)
	}

	points := make([]point, 0, len(valueMap))
	for mbean, value := range valueMap {
		points = append(points, point{
			Tags:   pb.extractTags(mbean),
			Fields: pb.extractFields(mbean, value),
		})
	}

	return compactPoints(points), nil
}

// extractTags generates the map of tags for a given mbean name/pattern.
func (pb *pointBuilder) extractTags(mbean string) map[string]string {
	propertyMap := makePropertyMap(mbean)
	tagMap := make(map[string]string)

	for key, value := range propertyMap {
		if pb.includeTag(key) {
			tagName := pb.formatTagName(key)
			tagMap[tagName] = value
		}
	}

	return tagMap
}

func (pb *pointBuilder) includeTag(tagName string) bool {
	return slices.Contains(pb.metric.TagKeys, tagName)
}

func (pb *pointBuilder) formatTagName(tagName string) string {
	if tagName == "" {
		return ""
	}

	if tagPrefix := pb.metric.TagPrefix; tagPrefix != "" {
		return tagPrefix + tagName
	}

	return tagName
}

// extractFields generates the map of fields for a given mbean name
// and value object.
func (pb *pointBuilder) extractFields(mbean string, value any) map[string]any {
	fieldMap := make(map[string]any)
	valueMap, ok := value.(map[string]any)

	if ok {
		// complex value
		if len(pb.objectAttributes) == 0 {
			// if there were no attributes requested,
			// then the keys are attributes
			pb.FillFields("", valueMap, fieldMap)
		} else if len(pb.objectAttributes) == 1 {
			// if there was a single attribute requested,
			// then the keys are the attribute's properties
			fieldName := pb.formatFieldName(pb.objectAttributes[0], pb.objectPath)
			pb.FillFields(fieldName, valueMap, fieldMap)
		} else {
			// if there were multiple attributes requested,
			// then the keys are the attribute names
			for _, attribute := range pb.objectAttributes {
				fieldName := pb.formatFieldName(attribute, pb.objectPath)
				pb.FillFields(fieldName, valueMap[attribute], fieldMap)
			}
		}
	} else {
		// scalar value
		var fieldName string
		if len(pb.objectAttributes) == 0 {
			fieldName = pb.formatFieldName(defaultFieldName, pb.objectPath)
		} else {
			fieldName = pb.formatFieldName(pb.objectAttributes[0], pb.objectPath)
		}

		pb.FillFields(fieldName, value, fieldMap)
	}

	if len(pb.substitutions) > 1 {
		pb.applySubstitutions(mbean, fieldMap)
	}

	return fieldMap
}

// formatFieldName generates a field name from the supplied attribute and
// path. The return value has the configured FieldPrefix and FieldSuffix
// instructions applied.
func (pb *pointBuilder) formatFieldName(attribute, path string) string {
	fieldName := attribute
	fieldPrefix := pb.metric.FieldPrefix
	fieldSeparator := pb.metric.FieldSeparator

	if fieldPrefix != "" {
		fieldName = fieldPrefix + fieldName
	}

	if path != "" {
		fieldName = fieldName + fieldSeparator + strings.ReplaceAll(path, "/", fieldSeparator)
	}

	return fieldName
}

// FillFields recurses into the supplied value object, generating a named field
// for every value it discovers.
func (pb *pointBuilder) FillFields(name string, value any, fieldMap map[string]any) {
	if valueMap, ok := value.(map[string]any); ok {
		// keep going until we get to something that is not a map
		for key, innerValue := range valueMap {
			if _, ok := innerValue.([]any); ok {
				continue
			}

			var innerName string
			if name == "" {
				innerName = pb.metric.FieldPrefix + key
			} else {
				innerName = name + pb.metric.FieldSeparator + key
			}

			pb.FillFields(innerName, innerValue, fieldMap)
		}

		return
	}

	if _, ok := value.([]any); ok {
		return
	}

	if pb.metric.FieldName != "" {
		name = pb.metric.FieldName
		if prefix := pb.metric.FieldPrefix; prefix != "" {
			name = prefix + name
		}
	}

	if name == "" {
		name = defaultFieldName
	}

	fieldMap[name] = value
}

// applySubstitutions updates all the keys in the supplied map
// of fields to account for $1-style substitution instructions.
func (pb *pointBuilder) applySubstitutions(mbean string, fieldMap map[string]any) {
	properties := makePropertyMap(mbean)

	for i, subKey := range pb.substitutions[1:] {
		symbol := fmt.Sprintf("$%d", i+1)
		substitution := properties[subKey]

		for fieldName, fieldValue := range fieldMap {
			newFieldName := strings.ReplaceAll(fieldName, symbol, substitution)
			if fieldName != newFieldName {
				fieldMap[newFieldName] = fieldValue
				delete(fieldMap, fieldName)
			}
		}
	}
}

// makePropertyMap returns a the mbean property-key list as
// a dictionary. foo:x=y becomes map[string]string { "x": "y" }
func makePropertyMap(mbean string) map[string]string {
	props := make(map[string]string)
	object := strings.SplitN(mbean, ":", 2)
	domain := object[0]

	if domain != "" && len(object) == 2 {
		list := object[1]

		for keyProperty := range strings.SplitSeq(list, ",") {
			pair := strings.SplitN(keyProperty, "=", 2)

			if len(pair) != 2 {
				continue
			}

			if key := pair[0]; key != "" {
				props[key] = pair[1]
			}
		}
	}

	return props
}

// makeSubstitutionList returns an array of values to
// use as substitutions when renaming fields
// with the $1..$N syntax. The first item in the list
// is always the mbean domain.
func makeSubstitutionList(mbean string) []string {
	subs := make([]string, 0)

	object := strings.SplitN(mbean, ":", 2)
	domain := object[0]

	if domain != "" && len(object) == 2 {
		subs = append(subs, domain)
		list := object[1]

		for keyProperty := range strings.SplitSeq(list, ",") {
			pair := strings.SplitN(keyProperty, "=", 2)

			if len(pair) != 2 {
				continue
			}

			key := pair[0]
			if key == "" {
				continue
			}

			property := pair[1]
			if !strings.Contains(property, "*") {
				continue
			}

			subs = append(subs, key)
		}
	}

	return subs
}
