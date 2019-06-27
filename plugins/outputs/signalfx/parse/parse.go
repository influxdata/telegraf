package parse

import (
	"fmt"
)

// RemoveSFXDimensions removes dimensions used only to identify special metrics for SignalFx
func RemoveSFXDimensions(metricDims map[string]string) {
	// remove the sf_metric dimension
	delete(metricDims, "sf_metric")
}

// SetPluginDimension sets the plugin dimension to the metric name if it is not already supplied
func SetPluginDimension(metricName string, metricDims map[string]string) {
	// If the plugin doesn't define a plugin name use metric.Name()
	if _, in := metricDims["plugin"]; !in {
		metricDims["plugin"] = metricName
	}
}

// GetMetricName combines telegraf fields and tags into a full metric name
func GetMetricName(metric string, field string, dims map[string]string) (string, bool) {
	// If sf_metric is provided
	if sfmetric, isSFX := dims["sf_metric"]; isSFX {
		return sfmetric, isSFX
	}

	// If it isn't a sf_metric then use metric name
	name := metric

	// Include field in metric name when it adds to the metric name
	if field != "value" {
		name = fmt.Sprintf("%s.%s", name, field)
	}

	return name, false
}

// ExtractProperty of the metric according to the following rules
func ExtractProperty(name string, dims map[string]string) (map[string]interface{}, error) {
	props := make(map[string]interface{}, 1)
	// if the metric is a metadata object
	if name == "objects.host-meta-data" {
		// If property exists remap it
		if _, in := dims["property"]; !in {
			return props, fmt.Errorf("E! objects.host-metadata object doesn't have a property")
		}
		props["property"] = dims["property"]
		delete(dims, "property")
	}
	return props, nil
}
