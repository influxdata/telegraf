package jolokia2

// A MetricConfig represents a TOML form of
// a Metric with some optional fields.
type MetricConfig struct {
	Name           string
	Mbean          string
	Paths          []string
	FieldName      *string
	FieldPrefix    *string
	FieldSeparator *string
	TagPrefix      *string
	TagKeys        []string
}

// A Metric represents a specification for a
// Jolokia read request, and the transformations
// to apply to points generated from the responses.
type Metric struct {
	Name           string
	Mbean          string
	Paths          []string
	FieldName      string
	FieldPrefix    string
	FieldSeparator string
	TagPrefix      string
	TagKeys        []string
}

func NewMetric(config MetricConfig, defaultFieldPrefix, defaultFieldSeparator, defaultTagPrefix string) Metric {
	metric := Metric{
		Name:    config.Name,
		Mbean:   config.Mbean,
		Paths:   config.Paths,
		TagKeys: config.TagKeys,
	}

	if config.FieldName != nil {
		metric.FieldName = *config.FieldName
	}

	if config.FieldPrefix == nil {
		metric.FieldPrefix = defaultFieldPrefix
	} else {
		metric.FieldPrefix = *config.FieldPrefix
	}

	if config.FieldSeparator == nil {
		metric.FieldSeparator = defaultFieldSeparator
	} else {
		metric.FieldSeparator = *config.FieldSeparator
	}

	if config.TagPrefix == nil {
		metric.TagPrefix = defaultTagPrefix
	} else {
		metric.TagPrefix = *config.TagPrefix
	}

	return metric
}
