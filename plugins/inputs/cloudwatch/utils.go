package cloudwatch

import (
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/influxdata/telegraf/internal"
)

func metricMatch(cm *cloudwatchMetric, m types.Metric) bool {
	if !slices.Contains(cm.MetricNames, *m.MetricName) {
		return false
	}
	// Dimensions need to match completely so exit early if the length mismatches
	if len(cm.Dimensions) != len(m.Dimensions) {
		return false
	}
	// Sort the dimensions for efficient comparison
	slices.SortStableFunc(m.Dimensions, func(a, b types.Dimension) int {
		return strings.Compare(*a.Name, *b.Name)
	})
	return slices.EqualFunc(cm.Dimensions, m.Dimensions, func(rd *dimension, vd types.Dimension) bool {
		return rd.Name == *vd.Name && (rd.valueMatcher == nil || rd.valueMatcher.Match(*vd.Value))
	})
}

func sanitizeMeasurement(namespace string) string {
	namespace = strings.ReplaceAll(namespace, "/", "_")
	namespace = snakeCase(namespace)
	return "cloudwatch_" + namespace
}

func snakeCase(s string) string {
	s = internal.SnakeCase(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "__", "_")
	return s
}

func ctod(cDimensions []types.Dimension) *map[string]string {
	dimensions := make(map[string]string, len(cDimensions))
	for i := range cDimensions {
		dimensions[snakeCase(*cDimensions[i].Name)] = *cDimensions[i].Value
	}
	return &dimensions
}
