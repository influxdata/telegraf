package signalfx

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/testutil"
)

type sink struct {
	datapoints []*datapoint.Datapoint
	events     []*event.Event
}

func (s *sink) AddDatapoints(_ context.Context, points []*datapoint.Datapoint) error {
	s.datapoints = append(s.datapoints, points...)
	return nil
}
func (s *sink) AddEvents(_ context.Context, events []*event.Event) error {
	s.events = append(s.events, events...)
	return nil
}

type errorsink struct {
	datapoints []*datapoint.Datapoint
	events     []*event.Event
}

func (*errorsink) AddDatapoints(context.Context, []*datapoint.Datapoint) error {
	return errors.New("not sending datapoints")
}
func (*errorsink) AddEvents(context.Context, []*event.Event) error {
	return errors.New("not sending events")
}

func TestSignalFx_SignalFx(t *testing.T) {
	type measurement struct {
		name   string
		tags   map[string]string
		fields map[string]interface{}
		time   time.Time
		tp     telegraf.ValueType
	}
	type fields struct {
		IncludedEvents []string
	}
	tests := []struct {
		name         string
		fields       fields
		measurements []*measurement
		want         errorsink
	}{
		{
			name:   "add datapoints of all types",
			fields: fields{},
			measurements: []*measurement{
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Counter,
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Gauge,
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Summary,
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Histogram,
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Untyped,
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"myboolmeasurement": true},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"myboolmeasurement": false},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
				},
			},
			want: errorsink{
				datapoints: []*datapoint.Datapoint{
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"host": "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Counter,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"host": "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"host": "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"host": "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"host": "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"host": "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.myboolmeasurement",
						map[string]string{
							"host": "192.168.0.1",
						},
						datapoint.NewIntValue(int64(1)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.myboolmeasurement",
						map[string]string{
							"host": "192.168.0.1",
						},
						datapoint.NewIntValue(int64(0)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
				events: make([]*event.Event, 0),
			},
		},
		{
			name: "add events of all types",
			fields: fields{
				IncludedEvents: []string{"event.mymeasurement"},
			},
			measurements: []*measurement{
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Counter,
				},
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Gauge,
				},
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Summary,
				},
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Histogram,
				},
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Untyped,
				},
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
				},
			},
			want: errorsink{
				datapoints: make([]*datapoint.Datapoint, 0),
				events: []*event.Event{
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"host": "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"host": "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"host": "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"host": "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"host": "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"host": "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name:   "exclude events by default",
			fields: fields{},
			measurements: []*measurement{
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"value": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Gauge,
				},
			},
			want: errorsink{
				datapoints: make([]*datapoint.Datapoint, 0),
				events:     make([]*event.Event, 0),
			},
		},
		{
			name:   "add datapoint with field named value",
			fields: fields{},
			measurements: []*measurement{
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"value": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Gauge,
				},
			},
			want: errorsink{
				datapoints: []*datapoint.Datapoint{
					datapoint.New(
						"datapoint",
						map[string]string{
							"host": "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
				events: make([]*event.Event, 0),
			},
		},
		{
			name: "add event",
			fields: fields{
				IncludedEvents: []string{"event.mymeasurement"},
			},
			measurements: []*measurement{
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Untyped,
				},
			},
			want: errorsink{
				datapoints: make([]*datapoint.Datapoint, 0),
				events: []*event.Event{
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"host": "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name:   "exclude events that are not explicitly included",
			fields: fields{},
			measurements: []*measurement{
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"value": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Gauge,
				},
			},
			want: errorsink{
				datapoints: make([]*datapoint.Datapoint, 0),
				events:     make([]*event.Event, 0),
			},
		},
		{
			name:   "malformed metadata event",
			fields: fields{},
			measurements: []*measurement{
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1", "sf_metric": "objects.host-meta-data"},
					fields: map[string]interface{}{"value": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Gauge,
				},
			},
			want: errorsink{
				datapoints: make([]*datapoint.Datapoint, 0),
				events:     make([]*event.Event, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := outputs.Outputs["signalfx"]().(*SignalFx)
			s.IncludedEventNames = tt.fields.IncludedEvents
			s.SignalFxRealm = "test"
			s.Log = testutil.Logger{}

			require.NoError(t, s.Connect())

			s.client = &sink{
				datapoints: make([]*datapoint.Datapoint, 0),
				events:     make([]*event.Event, 0),
			}

			measurements := make([]telegraf.Metric, 0, len(tt.measurements))
			for _, measurement := range tt.measurements {
				measurements = append(measurements, metric.New(measurement.name, measurement.tags, measurement.fields, measurement.time, measurement.tp))
			}

			err := s.Write(measurements)
			require.NoError(t, err)
			require.Eventually(t, func() bool { return len(s.client.(*sink).datapoints) == len(tt.want.datapoints) }, 5*time.Second, 10*time.Millisecond)
			require.Eventually(t, func() bool { return len(s.client.(*sink).events) == len(tt.want.events) }, 5*time.Second, 10*time.Millisecond)

			if !reflect.DeepEqual(s.client.(*sink).datapoints, tt.want.datapoints) {
				t.Errorf("Collected datapoints do not match desired.  Collected: %v Desired: %v", s.client.(*sink).datapoints, tt.want.datapoints)
			}
			if !reflect.DeepEqual(s.client.(*sink).events, tt.want.events) {
				t.Errorf("Collected events do not match desired.  Collected: %v Desired: %v", s.client.(*sink).events, tt.want.events)
			}
		})
	}
}

func TestSignalFx_Errors(t *testing.T) {
	type measurement struct {
		name   string
		tags   map[string]string
		fields map[string]interface{}
		time   time.Time
		tp     telegraf.ValueType
	}
	type fields struct {
		IncludedEvents []string
	}
	type want struct {
		datapoints []*datapoint.Datapoint
		events     []*event.Event
	}
	tests := []struct {
		name         string
		fields       fields
		measurements []*measurement
		want         want
	}{
		{
			name:   "add datapoints of all types",
			fields: fields{},
			measurements: []*measurement{
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Counter,
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Gauge,
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Summary,
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Histogram,
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Untyped,
				},
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
				},
			},
			want: want{
				datapoints: make([]*datapoint.Datapoint, 0),
				events:     make([]*event.Event, 0),
			},
		},
		{
			name: "add events of all types",
			fields: fields{
				IncludedEvents: []string{"event.mymeasurement"},
			},
			measurements: []*measurement{
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Counter,
				},
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Gauge,
				},
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Summary,
				},
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Histogram,
				},
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Untyped,
				},
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"mymeasurement": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
				},
			},
			want: want{
				datapoints: make([]*datapoint.Datapoint, 0),
				events:     make([]*event.Event, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := outputs.Outputs["signalfx"]().(*SignalFx)
			// constrain the buffer to cover code that emits when batch size is met
			s.IncludedEventNames = tt.fields.IncludedEvents
			s.SignalFxRealm = "test"
			s.Log = testutil.Logger{}

			require.NoError(t, s.Connect())

			s.client = &errorsink{
				datapoints: make([]*datapoint.Datapoint, 0),
				events:     make([]*event.Event, 0),
			}

			for _, measurement := range tt.measurements {
				m := metric.New(
					measurement.name, measurement.tags, measurement.fields, measurement.time, measurement.tp,
				)

				err := s.Write([]telegraf.Metric{m})
				require.Error(t, err)
			}
			for !(len(s.client.(*errorsink).datapoints) == len(tt.want.datapoints) && len(s.client.(*errorsink).events) == len(tt.want.events)) {
				time.Sleep(1 * time.Second)
			}
			if !reflect.DeepEqual(s.client.(*errorsink).datapoints, tt.want.datapoints) {
				t.Errorf("Collected datapoints do not match desired.  Collected: %v Desired: %v", s.client.(*errorsink).datapoints, tt.want.datapoints)
			}
			if !reflect.DeepEqual(s.client.(*errorsink).events, tt.want.events) {
				t.Errorf("Collected events do not match desired.  Collected: %v Desired: %v", s.client.(*errorsink).events, tt.want.events)
			}
		})
	}
}

func TestGetMetricName(t *testing.T) {
	type args struct {
		metric string
		field  string
		dims   map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantsfx bool
	}{
		{
			name: "fields that equal value should not be append to metricname",
			args: args{
				metric: "datapoint",
				field:  "value",
				dims: map[string]string{
					"testDimKey": "testDimVal",
				},
			},
			want: "datapoint",
		},
		{
			name: "fields other than 'value' with out sf_metric dim should return measurement.fieldname as metric name",
			args: args{
				metric: "datapoint",
				field:  "test",
				dims: map[string]string{
					"testDimKey": "testDimVal",
				},
			},
			want: "datapoint.test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMetricName(tt.args.metric, tt.args.field)
			if got != tt.want {
				t.Errorf("getMetricName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
