package senders

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

// Gets a metric line in the Wavefront metrics data format:
// <metricName> <metricValue> [<timestamp>] source=<source> [pointTags]
// Example: "new-york.power.usage 42422.0 1533531013 source=localhost datacenter=dc1"
func MetricLine(name string, value float64, ts int64, source string, tags map[string]string, defaultSource string) (string, error) {

	if name == "" {
		return "", fmt.Errorf("empty metric name")
	}

	if source == "" {
		source = defaultSource
	}

	sb := bytes.NewBufferString("")
	sb.WriteString(strconv.Quote(name))
	sb.WriteString(" ")
	sb.WriteString(strconv.FormatFloat(value, 'f', -1, 64))

	if ts != 0 {
		sb.WriteString(" ")
		sb.WriteString(strconv.FormatInt(ts, 10))
	}

	sb.WriteString(" source=")
	sb.WriteString(strconv.Quote(source))

	for k, v := range tags {
		if v == "" {
			return "", fmt.Errorf("metric point tag value cannot be blank")
		}
		sb.WriteString(" ")
		sb.WriteString(strconv.Quote(k))
		sb.WriteString("=")
		sb.WriteString(strconv.Quote(v))
	}
	sb.WriteString("\n")
	return sb.String(), nil
}

// Gets a histogram line in the Wavefront histogram data format:
// {!M | !H | !D} [<timestamp>] #<count> <mean> [centroids] <histogramName> source=<source> [pointTags]
// Example: "!M 1533531013 #20 30.0 #10 5.1 request.latency source=appServer1 region=us-west"
func HistoLine(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string, defaultSource string) (string, error) {

	if name == "" {
		return "", fmt.Errorf("empty distribution name")
	}

	if len(centroids) == 0 {
		return "", fmt.Errorf("distribution should have at least one centroid")
	}

	if len(hgs) == 0 {
		return "", fmt.Errorf("histogram granularities cannot be empty")
	}

	if source == "" {
		source = defaultSource
	}

	sb := bytes.NewBufferString("")
	for hg := range hgs {
		sb.WriteString(hg.String())
		if ts != 0 {
			sb.WriteString(" ")
			sb.WriteString(strconv.FormatInt(ts, 10))
		}
		for _, centroid := range centroids {
			sb.WriteString(" #")
			sb.WriteString(strconv.Itoa(centroid.Count))
			sb.WriteString(" ")
			sb.WriteString(strconv.FormatFloat(centroid.Value, 'f', -1, 64))
		}
		sb.WriteString(" ")
		sb.WriteString(strconv.Quote(name))
		sb.WriteString(" source=")
		sb.WriteString(strconv.Quote(source))

		for k, v := range tags {
			if v == "" {
				return "", fmt.Errorf("histogram tag value cannot be blank")
			}
			sb.WriteString(" ")
			sb.WriteString(strconv.Quote(k))
			sb.WriteString("=")
			sb.WriteString(strconv.Quote(v))
		}
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

// Gets a span line in the Wavefront span data format:
// <tracingSpanName> source=<source> [pointTags] <start_millis> <duration_milli_seconds>
// Example:
// "getAllUsers source=localhost traceId=7b3bf470-9456-11e8-9eb6-529269fb1459 spanId=0313bafe-9457-11e8-9eb6-529269fb1459
//    parent=2f64e538-9457-11e8-9eb6-529269fb1459 application=Wavefront http.method=GET 1533531013 343500"
func SpanLine(name string, startMillis, durationMillis int64, source, traceId, spanId string, parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog, defaultSource string) (string, error) {

	if name == "" {
		return "", fmt.Errorf("empty span name")
	}

	if source == "" {
		source = defaultSource
	}

	//TODO: verify if strings are uuid?

	sb := bytes.NewBufferString("")
	sb.WriteString(strconv.Quote(name))
	sb.WriteString(" source=")
	sb.WriteString(strconv.Quote(source))
	sb.WriteString(" traceId=")
	sb.WriteString(traceId)
	sb.WriteString(" spanId=")
	sb.WriteString(spanId)

	for _, parent := range parents {
		sb.WriteString(" parent=")
		sb.WriteString(parent)
	}

	for _, item := range followsFrom {
		sb.WriteString(" followsFrom=")
		sb.WriteString(item)
	}

	for _, tag := range tags {
		if tag.Key == "" || tag.Value == "" {
			return "", fmt.Errorf("span tag key/value cannot be blank")
		}
		sb.WriteString(" ")
		sb.WriteString(strconv.Quote(tag.Key))
		sb.WriteString("=")
		sb.WriteString(strconv.Quote(tag.Value))
	}
	sb.WriteString(" ")
	sb.WriteString(strconv.FormatInt(startMillis, 10))
	sb.WriteString(" ")
	sb.WriteString(strconv.FormatInt(durationMillis, 10))
	sb.WriteString("\n")

	return sb.String(), nil
}
