package openmetrics

import (
	"bytes"
	"errors"
	"fmt"
	"hash/maphash"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/textparse"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TextToMetricFamilies(data []byte) ([]*MetricFamily, error) {
	var metrics []*MetricFamily

	parser := textparse.NewOpenMetricsParser(data, nil)

	seed := maphash.MakeSeed()
	mf := &MetricFamily{}
	mfMetric := &Metric{}
	mfMetricKey := uint64(0)
	mfMetricPoint := &MetricPoint{}
	for {
		entry, err := parser.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if mf.Name != "" {
					if mfMetricPoint.Value != nil {
						mfMetric.MetricPoints = append(mfMetric.MetricPoints, mfMetricPoint)
					}
					if len(mfMetric.MetricPoints) > 0 {
						mf.Metrics = append(mf.Metrics, mfMetric)
					}
					metrics = append(metrics, mf)
				}
				break
			}
			return nil, fmt.Errorf("parsing failed: %w", err)
		}

		switch entry {
		case textparse.EntryInvalid:
			continue
		case textparse.EntryType:
			name, mtype := parser.Type()
			if len(name) == 0 {
				return nil, errors.New("empty metric-family name")
			}

			if mf.Name == "" {
				mf.Name = string(name)
			} else if mf.Name != string(name) {
				if mfMetricPoint.Value != nil {
					mfMetric.MetricPoints = append(mfMetric.MetricPoints, mfMetricPoint)
				}
				if len(mfMetric.MetricPoints) > 0 {
					mf.Metrics = append(mf.Metrics, mfMetric)
				}
				metrics = append(metrics, mf)
				mf = &MetricFamily{Name: string(name)}
				mfMetric = &Metric{}
				mfMetricKey = 0
				mfMetricPoint = &MetricPoint{}
			}

			switch mtype {
			case model.MetricTypeCounter:
				mf.Type = MetricType_COUNTER
			case model.MetricTypeGauge:
				mf.Type = MetricType_GAUGE
			case model.MetricTypeHistogram:
				mf.Type = MetricType_HISTOGRAM
			case model.MetricTypeGaugeHistogram:
				mf.Type = MetricType_GAUGE_HISTOGRAM
			case model.MetricTypeSummary:
				mf.Type = MetricType_SUMMARY
			case model.MetricTypeInfo:
				mf.Type = MetricType_INFO
			case model.MetricTypeStateset:
				mf.Type = MetricType_STATE_SET
			case model.MetricTypeUnknown:
				mf.Type = MetricType_UNKNOWN
			}
		case textparse.EntryHelp:
			name, mhelp := parser.Help()
			if len(name) == 0 {
				return nil, errors.New("empty metric-family name")
			}

			if mf.Name == "" {
				mf.Name = string(name)
			} else if mf.Name != string(name) {
				if mfMetricPoint.Value != nil {
					mfMetric.MetricPoints = append(mfMetric.MetricPoints, mfMetricPoint)
				}
				if len(mfMetric.MetricPoints) > 0 {
					mf.Metrics = append(mf.Metrics, mfMetric)
				}
				metrics = append(metrics, mf)
				mf = &MetricFamily{Name: string(name)}
				mfMetric = &Metric{}
				mfMetricKey = 0
				mfMetricPoint = &MetricPoint{}
			}
			mf.Help = string(mhelp)
		case textparse.EntrySeries:
			series, ts, value := parser.Series()

			// Extract the metric name and labels
			dn, _, _ := bytes.Cut(series, []byte("{"))
			if len(dn) == 0 {
				return nil, errors.New("empty metric name")
			}
			sampleName := string(dn)

			var metricLabels labels.Labels
			parser.Labels(&metricLabels)

			// There might be metrics without meta-information, however in this
			// case the metric is of type UNKNOWN according to the spec and do
			// only contain a single metric. Therefore, we can use the metric
			// name as metric-family name
			if mf.Name == "" {
				mf.Name = sampleName
			}

			// The name contained in the sample is constructed using the metric
			// name and an optional sample-type suffix used for more complex
			// types (e.g. histograms).
			sampleType, seriesLabels := extractSampleType(sampleName, mf.Name, mf.Type, &metricLabels)

			// Check if we are still in the same metric, if not, add the
			// previous one to the metric family and create a new one...
			key := getSeriesKey(seriesLabels, seed)
			if mfMetricKey != key {
				if mfMetricPoint.Value != nil {
					mfMetric.MetricPoints = append(mfMetric.MetricPoints, mfMetricPoint)
				}
				if len(mfMetric.MetricPoints) > 0 {
					mf.Metrics = append(mf.Metrics, mfMetric)
				}
				mfMetric = &Metric{}
				mfMetricKey = key
				mfMetricPoint = &MetricPoint{}
				mfMetric.Labels = make([]*Label, 0, seriesLabels.Len())
				seriesLabels.Range(func(l labels.Label) {
					mfMetric.Labels = append(mfMetric.Labels, &Label{
						Name:  l.Name,
						Value: l.Value,
					})
				})
			}

			// Check if we are still in the same metric-point
			var mpTimestamp int64
			if mfMetricPoint.Timestamp != nil {
				mpTimestamp = mfMetricPoint.Timestamp.Seconds * int64(time.Second)
				mpTimestamp += int64(mfMetricPoint.Timestamp.Nanos)
			}
			var timestamp int64
			if ts != nil {
				timestamp = *ts * int64(time.Millisecond)
			}
			if mpTimestamp != timestamp {
				if mfMetricPoint.Value != nil {
					mfMetric.MetricPoints = append(mfMetric.MetricPoints, mfMetricPoint)
				}
				mfMetricPoint = &MetricPoint{}
				if ts != nil {
					mfMetricPoint.Timestamp = timestamppb.New(time.Unix(0, timestamp))
				}
			}

			// Fill in the metric-point
			mfMetricPoint.set(mf.Name, mf.Type, sampleType, value, &metricLabels)
		case textparse.EntryComment:
			// ignore comments
		case textparse.EntryUnit:
			name, munit := parser.Unit()
			if len(name) == 0 {
				return nil, errors.New("empty metric-family name")
			}

			if mf.Name == "" {
				mf.Name = string(name)
			} else if mf.Name != string(name) {
				if mfMetricPoint.Value != nil {
					mfMetric.MetricPoints = append(mfMetric.MetricPoints, mfMetricPoint)
				}
				if len(mfMetric.MetricPoints) > 0 {
					mf.Metrics = append(mf.Metrics, mfMetric)
				}
				metrics = append(metrics, mf)
				mf = &MetricFamily{Name: string(name)}
				mfMetric = &Metric{}
				mfMetricKey = 0
				mfMetricPoint = &MetricPoint{}
			}
			mf.Unit = string(munit)
		case textparse.EntryHistogram:
			// not supported yet
		default:
			return nil, fmt.Errorf("unknown entry type %v", entry)
		}
	}

	return metrics, nil
}

func getSeriesKey(seriesLabels *labels.Labels, seed maphash.Seed) uint64 {
	sorted := make([]string, 0, seriesLabels.Len())
	seriesLabels.Range(func(l labels.Label) {
		sorted = append(sorted, l.Name+"="+l.Value)
	})
	slices.Sort(sorted)

	var h maphash.Hash
	h.SetSeed(seed)
	for _, p := range sorted {
		h.WriteString(p)
		h.WriteByte(0)
	}
	return h.Sum64()
}

func extractSampleType(raw, name string, mtype MetricType, metricLabels *labels.Labels) (string, *labels.Labels) {
	suffix := strings.TrimLeft(strings.TrimPrefix(raw, name), "_")
	var seriesLabelArray []labels.Label
	metricLabels.Range(func(l labels.Label) {
		// filter out special labels
		switch {
		case l.Name == "__name__":
		case mtype == MetricType_STATE_SET && l.Name == name:
		case mtype == MetricType_HISTOGRAM && l.Name == "le":
		case mtype == MetricType_GAUGE_HISTOGRAM && l.Name == "le":
		case mtype == MetricType_SUMMARY && l.Name == "quantile":
		default:
			seriesLabelArray = append(seriesLabelArray, labels.Label{Name: l.Name, Value: l.Value})
		}
	})
	seriesLabels := labels.New(seriesLabelArray...)
	return suffix, &seriesLabels
}

func (mp *MetricPoint) set(mname string, mtype MetricType, stype string, value float64, mlabels *labels.Labels) {
	switch mtype {
	case MetricType_UNKNOWN:
		mp.Value = &MetricPoint_UnknownValue{
			UnknownValue: &UnknownValue{
				Value: &UnknownValue_DoubleValue{DoubleValue: value},
			},
		}
	case MetricType_GAUGE:
		mp.Value = &MetricPoint_GaugeValue{
			GaugeValue: &GaugeValue{
				Value: &GaugeValue_DoubleValue{DoubleValue: value},
			},
		}
	case MetricType_COUNTER:
		var v *MetricPoint_CounterValue
		if mp.Value != nil {
			v = mp.Value.(*MetricPoint_CounterValue)
		} else {
			v = &MetricPoint_CounterValue{
				CounterValue: &CounterValue{},
			}
		}
		switch stype {
		case "total":
			v.CounterValue.Total = &CounterValue_DoubleValue{DoubleValue: value}
		case "created":
			t := time.Unix(0, int64(value*float64(time.Second)))
			v.CounterValue.Created = timestamppb.New(t)
		}
		mp.Value = v
	case MetricType_STATE_SET:
		var v *MetricPoint_StateSetValue
		if mp.Value != nil {
			v = mp.Value.(*MetricPoint_StateSetValue)
		} else {
			v = &MetricPoint_StateSetValue{
				StateSetValue: &StateSetValue{},
			}
		}

		var name string
		mlabels.Range(func(l labels.Label) {
			if l.Name == mname && name == "" {
				name = l.Value
			}
		})
		v.StateSetValue.States = append(v.StateSetValue.States, &StateSetValue_State{
			Enabled: value > 0,
			Name:    name,
		})
		mp.Value = v
	case MetricType_INFO:
		mp.Value = &MetricPoint_InfoValue{
			InfoValue: &InfoValue{},
		}
	case MetricType_HISTOGRAM, MetricType_GAUGE_HISTOGRAM:
		var v *MetricPoint_HistogramValue
		if mp.Value != nil {
			v = mp.Value.(*MetricPoint_HistogramValue)
		} else {
			v = &MetricPoint_HistogramValue{
				HistogramValue: &HistogramValue{},
			}
		}

		switch stype {
		case "sum", "gsum":
			v.HistogramValue.Sum = &HistogramValue_DoubleValue{DoubleValue: value}
		case "count", "gcount":
			v.HistogramValue.Count = uint64(value)
		case "created":
			t := time.Unix(0, int64(value*float64(time.Second)))
			v.HistogramValue.Created = timestamppb.New(t)
		case "bucket":
			var boundLabel string
			mlabels.Range(func(l labels.Label) {
				if l.Name == "le" && boundLabel == "" {
					boundLabel = l.Value
				}
			})
			var bound float64
			if boundLabel == "+Inf" {
				bound = math.Inf(1)
			} else {
				var err error
				if bound, err = strconv.ParseFloat(boundLabel, 64); err != nil {
					bound = math.NaN()
				}
			}

			v.HistogramValue.Buckets = append(v.HistogramValue.Buckets, &HistogramValue_Bucket{
				Count:      uint64(value),
				UpperBound: bound,
			})
		}
		mp.Value = v
	case MetricType_SUMMARY:
		var v *MetricPoint_SummaryValue
		if mp.Value != nil {
			v = mp.Value.(*MetricPoint_SummaryValue)
		} else {
			v = &MetricPoint_SummaryValue{
				SummaryValue: &SummaryValue{},
			}
		}

		switch stype {
		case "sum":
			v.SummaryValue.Sum = &SummaryValue_DoubleValue{DoubleValue: value}
		case "count":
			v.SummaryValue.Count = uint64(value)
		case "created":
			t := time.Unix(0, int64(value*float64(time.Second)))
			v.SummaryValue.Created = timestamppb.New(t)
		default:
			var quantileLabel string
			mlabels.Range(func(l labels.Label) {
				if l.Name == "quantile" && quantileLabel == "" {
					quantileLabel = l.Value
				}
			})
			var quantile float64
			if quantileLabel == "+Inf" {
				quantile = math.MaxFloat64
			} else {
				var err error
				if quantile, err = strconv.ParseFloat(quantileLabel, 64); err != nil {
					quantile = math.NaN()
				}
			}

			v.SummaryValue.Quantile = append(v.SummaryValue.Quantile, &SummaryValue_Quantile{
				Quantile: quantile,
				Value:    value,
			})
		}
		mp.Value = v
	}
}
