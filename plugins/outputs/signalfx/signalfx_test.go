package signalfx

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/datapoint/dpsink"
	"github.com/signalfx/golib/event"
)

type sink struct {
	dps []*datapoint.Datapoint
	evs []*event.Event
}

func (s *sink) AddDatapoints(ctx context.Context, points []*datapoint.Datapoint) error {
	s.dps = append(s.dps, points...)
	return nil
}
func (s *sink) AddEvents(ctx context.Context, events []*event.Event) error {
	s.evs = append(s.evs, events...)
	return nil
}

type errorsink struct {
	dps []*datapoint.Datapoint
	evs []*event.Event
}

func (e *errorsink) AddDatapoints(ctx context.Context, points []*datapoint.Datapoint) error {
	return errors.New("not sending datapoints")
}
func (e *errorsink) AddEvents(ctx context.Context, events []*event.Event) error {
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
		Exclude []string
		Include []string
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
				datapoints: []*datapoint.Datapoint{
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "counter",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Counter,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "gauge",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "summary",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "histogram",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "untyped",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "unrecognized",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
				events: []*event.Event{},
			},
		},
		{
			name: "add events of all types",
			fields: fields{
				Include: []string{"event.mymeasurement"},
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
				datapoints: []*datapoint.Datapoint{},
				events: []*event.Event{
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "counter",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "gauge",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "summary",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "histogram",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "untyped",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "unrecognized",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "exclude datapoints and events",
			fields: fields{
				Exclude: []string{"datapoint"},
			},
			measurements: []*measurement{
				{
					name:   "datapoint",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"value": float64(3.14)},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Gauge,
				},
				{
					name:   "event",
					tags:   map[string]string{"host": "192.168.0.1"},
					fields: map[string]interface{}{"value": "hello world"},
					time:   time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
					tp:     telegraf.Gauge,
				},
			},
			want: want{
				datapoints: []*datapoint.Datapoint{},
				events:     []*event.Event{},
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
			want: want{
				datapoints: []*datapoint.Datapoint{
					datapoint.New(
						"datapoint",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "gauge",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
				events: []*event.Event{},
			},
		},
		{
			name: "add event",
			fields: fields{
				Include: []string{"event.mymeasurement"},
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
			want: want{
				datapoints: []*datapoint.Datapoint{},
				events: []*event.Event{
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "untyped",
							"host":          "192.168.0.1",
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
			want: want{
				datapoints: []*datapoint.Datapoint{},
				events:     []*event.Event{},
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
			want: want{
				datapoints: []*datapoint.Datapoint{},
				events:     []*event.Event{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := outputs.Outputs["signalfx"]().(*SignalFx)
			s.Exclude = tt.fields.Exclude
			s.Include = tt.fields.Include

			s.Connect()

			s.client = &sink{
				dps: []*datapoint.Datapoint{},
				evs: []*event.Event{},
			}

			measurements := []telegraf.Metric{}

			for _, measurement := range tt.measurements {
				m, err := metric.New(
					measurement.name, measurement.tags, measurement.fields, measurement.time, measurement.tp,
				)
				if err != nil {
					t.Errorf("Error creating measurement %v", measurement)
				}
				measurements = append(measurements, m)
			}
			s.Write(measurements)
			for !(len(s.client.(*sink).dps) == len(tt.want.datapoints) && len(s.client.(*sink).evs) == len(tt.want.events)) {
				time.Sleep(1 * time.Second)
			}
			if !reflect.DeepEqual(s.client.(*sink).dps, tt.want.datapoints) {
				t.Errorf("Collected datapoints do not match desired.  Collected: %v Desired: %v", s.client.(*sink).dps, tt.want.datapoints)
			}
			if !reflect.DeepEqual(s.client.(*sink).evs, tt.want.events) {
				t.Errorf("Collected events do not match desired.  Collected: %v Desired: %v", s.client.(*sink).evs, tt.want.events)
			}

			err := s.Close()
			if err != nil {
				t.Errorf("Failed to close the plugin %v", err)
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
		Exclude []string
		Include []string
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
				datapoints: []*datapoint.Datapoint{},
				events:     []*event.Event{},
			},
		},
		{
			name: "add events of all types",
			fields: fields{
				Include: []string{"event.mymeasurement"},
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
				datapoints: []*datapoint.Datapoint{},
				events:     []*event.Event{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := outputs.Outputs["signalfx"]().(*SignalFx)
			// constrain the buffer to cover code that emits when batch size is met
			s.BatchSize = 2
			s.Exclude = tt.fields.Exclude
			s.Include = tt.fields.Include

			s.Connect()

			s.client = &errorsink{
				dps: []*datapoint.Datapoint{},
				evs: []*event.Event{},
			}

			for _, measurement := range tt.measurements {
				m, err := metric.New(
					measurement.name, measurement.tags, measurement.fields, measurement.time, measurement.tp,
				)
				if err != nil {
					t.Errorf("Error creating measurement %v", measurement)
				}
				s.Write([]telegraf.Metric{m})
			}
			for !(len(s.client.(*errorsink).dps) == len(tt.want.datapoints) && len(s.client.(*errorsink).evs) == len(tt.want.events)) {
				time.Sleep(1 * time.Second)
			}
			if !reflect.DeepEqual(s.client.(*errorsink).dps, tt.want.datapoints) {
				t.Errorf("Collected datapoints do not match desired.  Collected: %v Desired: %v", s.client.(*errorsink).dps, tt.want.datapoints)
			}
			if !reflect.DeepEqual(s.client.(*errorsink).evs, tt.want.events) {
				t.Errorf("Collected events do not match desired.  Collected: %v Desired: %v", s.client.(*errorsink).evs, tt.want.events)
			}

			err := s.Close()
			if err != nil {
				t.Errorf("Failed to close the plugin %v", err)
			}
		})
	}
}

func TestSignalFx_fillAndSendDatapoints(t *testing.T) {
	type fields struct {
		APIToken    string
		BatchSize   int
		ChannelSize int
		client      dpsink.Sink
		dps         chan *datapoint.Datapoint
		evts        chan *event.Event
	}
	type args struct {
		in  []*datapoint.Datapoint
		buf []*datapoint.Datapoint
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*datapoint.Datapoint
	}{
		{
			name: "test buffer fills until batch size is met",
			fields: fields{
				BatchSize: 3,
				dps:       make(chan *datapoint.Datapoint, 10),
				client: &sink{
					dps: []*datapoint.Datapoint{},
				},
			},
			args: args{
				buf: []*datapoint.Datapoint{},
				in: []*datapoint.Datapoint{
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "counter",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Counter,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "gauge",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "summary",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
			want: []*datapoint.Datapoint{
				datapoint.New(
					"datapoint.mymeasurement",
					map[string]string{
						"plugin":        "datapoint",
						"agent":         "telegraf",
						"telegraf_type": "counter",
						"host":          "192.168.0.1",
					},
					datapoint.NewFloatValue(float64(3.14)),
					datapoint.Counter,
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				datapoint.New(
					"datapoint.mymeasurement",
					map[string]string{
						"plugin":        "datapoint",
						"agent":         "telegraf",
						"telegraf_type": "gauge",
						"host":          "192.168.0.1",
					},
					datapoint.NewFloatValue(float64(3.14)),
					datapoint.Gauge,
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				datapoint.New(
					"datapoint.mymeasurement",
					map[string]string{
						"plugin":        "datapoint",
						"agent":         "telegraf",
						"telegraf_type": "summary",
						"host":          "192.168.0.1",
					},
					datapoint.NewFloatValue(float64(3.14)),
					datapoint.Gauge,
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "test buffer fills until batch size is and has 1 remaining when it breaks",
			fields: fields{
				BatchSize: 2,
				dps:       make(chan *datapoint.Datapoint, 10),
				client: &sink{
					dps: []*datapoint.Datapoint{},
				},
			},
			args: args{
				buf: []*datapoint.Datapoint{},
				in: []*datapoint.Datapoint{
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "counter",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Counter,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "gauge",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "summary",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
			want: []*datapoint.Datapoint{
				datapoint.New(
					"datapoint.mymeasurement",
					map[string]string{
						"plugin":        "datapoint",
						"agent":         "telegraf",
						"telegraf_type": "counter",
						"host":          "192.168.0.1",
					},
					datapoint.NewFloatValue(float64(3.14)),
					datapoint.Counter,
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				datapoint.New(
					"datapoint.mymeasurement",
					map[string]string{
						"plugin":        "datapoint",
						"agent":         "telegraf",
						"telegraf_type": "gauge",
						"host":          "192.168.0.1",
					},
					datapoint.NewFloatValue(float64(3.14)),
					datapoint.Gauge,
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				datapoint.New(
					"datapoint.mymeasurement",
					map[string]string{
						"plugin":        "datapoint",
						"agent":         "telegraf",
						"telegraf_type": "summary",
						"host":          "192.168.0.1",
					},
					datapoint.NewFloatValue(float64(3.14)),
					datapoint.Gauge,
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignalFx{
				APIToken:    tt.fields.APIToken,
				BatchSize:   tt.fields.BatchSize,
				ChannelSize: tt.fields.ChannelSize,
				client:      tt.fields.client,
				dps:         tt.fields.dps,
				evts:        tt.fields.evts,
			}
			for _, e := range tt.args.in {
				s.dps <- e
			}
			s.fillAndSendDatapoints(tt.args.buf)
			if !reflect.DeepEqual(s.client.(*sink).dps, tt.want) {
				t.Errorf("fillAndSendDatapoints() datapoints do not match desired.  Collected: %v Desired: %v", s.client.(*sink).dps, tt.want)
			}
		})
	}
}

func TestSignalFx_fillAndSendDatapointsWithError(t *testing.T) {
	type fields struct {
		APIToken    string
		BatchSize   int
		ChannelSize int
		client      dpsink.Sink
		dps         chan *datapoint.Datapoint
		evts        chan *event.Event
	}
	type args struct {
		in  []*datapoint.Datapoint
		buf []*datapoint.Datapoint
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*datapoint.Datapoint
	}{
		{
			name: "test buffer fills until batch size is met",
			fields: fields{
				BatchSize: 3,
				dps:       make(chan *datapoint.Datapoint, 10),
				client: &errorsink{
					dps: []*datapoint.Datapoint{},
				},
			},
			args: args{
				buf: []*datapoint.Datapoint{},
				in: []*datapoint.Datapoint{
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "counter",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Counter,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "gauge",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					datapoint.New(
						"datapoint.mymeasurement",
						map[string]string{
							"plugin":        "datapoint",
							"agent":         "telegraf",
							"telegraf_type": "summary",
							"host":          "192.168.0.1",
						},
						datapoint.NewFloatValue(float64(3.14)),
						datapoint.Gauge,
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
			want: []*datapoint.Datapoint{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignalFx{
				APIToken:    tt.fields.APIToken,
				BatchSize:   tt.fields.BatchSize,
				ChannelSize: tt.fields.ChannelSize,
				client:      tt.fields.client,
				dps:         tt.fields.dps,
				evts:        tt.fields.evts,
			}
			for _, e := range tt.args.in {
				s.dps <- e
			}
			s.fillAndSendDatapoints(tt.args.buf)
			if !reflect.DeepEqual(s.client.(*errorsink).dps, tt.want) {
				t.Errorf("fillAndSendDatapoints() datapoints do not match desired.  Collected: %v Desired: %v", s.client.(*errorsink).dps, tt.want)
			}
		})
	}
}

func TestSignalFx_fillAndSendEvents(t *testing.T) {
	type fields struct {
		APIToken    string
		BatchSize   int
		ChannelSize int
		client      dpsink.Sink
		dps         chan *datapoint.Datapoint
		evts        chan *event.Event
	}
	type args struct {
		in  []*event.Event
		buf []*event.Event
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*event.Event
	}{
		{
			name: "test buffer fills until batch size is met",
			fields: fields{
				BatchSize: 3,
				evts:      make(chan *event.Event, 10),
				client: &sink{
					evs: []*event.Event{},
				},
			},
			args: args{
				buf: []*event.Event{},
				in: []*event.Event{
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "counter",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "gauge",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "summary",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
			want: []*event.Event{
				event.NewWithProperties(
					"event.mymeasurement",
					event.AGENT,
					map[string]string{
						"plugin":        "event",
						"agent":         "telegraf",
						"telegraf_type": "counter",
						"host":          "192.168.0.1",
					},
					map[string]interface{}{
						"message": "hello world",
					},
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				event.NewWithProperties(
					"event.mymeasurement",
					event.AGENT,
					map[string]string{
						"plugin":        "event",
						"agent":         "telegraf",
						"telegraf_type": "gauge",
						"host":          "192.168.0.1",
					},
					map[string]interface{}{
						"message": "hello world",
					},
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				event.NewWithProperties(
					"event.mymeasurement",
					event.AGENT,
					map[string]string{
						"plugin":        "event",
						"agent":         "telegraf",
						"telegraf_type": "summary",
						"host":          "192.168.0.1",
					},
					map[string]interface{}{
						"message": "hello world",
					},
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "test buffer fills until batch size is and has 1 remaining when it breaks",
			fields: fields{
				BatchSize: 2,
				evts:      make(chan *event.Event, 10),
				client: &sink{
					evs: []*event.Event{},
				},
			},
			args: args{
				buf: []*event.Event{},
				in: []*event.Event{
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "counter",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "gauge",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "summary",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
			want: []*event.Event{
				event.NewWithProperties(
					"event.mymeasurement",
					event.AGENT,
					map[string]string{
						"plugin":        "event",
						"agent":         "telegraf",
						"telegraf_type": "counter",
						"host":          "192.168.0.1",
					},
					map[string]interface{}{
						"message": "hello world",
					},
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				event.NewWithProperties(
					"event.mymeasurement",
					event.AGENT,
					map[string]string{
						"plugin":        "event",
						"agent":         "telegraf",
						"telegraf_type": "gauge",
						"host":          "192.168.0.1",
					},
					map[string]interface{}{
						"message": "hello world",
					},
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				event.NewWithProperties(
					"event.mymeasurement",
					event.AGENT,
					map[string]string{
						"plugin":        "event",
						"agent":         "telegraf",
						"telegraf_type": "summary",
						"host":          "192.168.0.1",
					},
					map[string]interface{}{
						"message": "hello world",
					},
					time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignalFx{
				APIToken:    tt.fields.APIToken,
				BatchSize:   tt.fields.BatchSize,
				ChannelSize: tt.fields.ChannelSize,
				client:      tt.fields.client,
				dps:         tt.fields.dps,
				evts:        tt.fields.evts,
			}
			for _, e := range tt.args.in {
				s.evts <- e
			}
			s.fillAndSendEvents(tt.args.buf)
			if !reflect.DeepEqual(s.client.(*sink).evs, tt.want) {
				t.Errorf("fillAndSendEvents() datapoints do not match desired.  Collected: %v Desired: %v", s.client.(*sink).evs, tt.want)
			}
		})
	}
}

func TestSignalFx_fillAndSendEventsWithErrors(t *testing.T) {
	type fields struct {
		APIToken    string
		BatchSize   int
		ChannelSize int
		client      dpsink.Sink
		dps         chan *datapoint.Datapoint
		evts        chan *event.Event
	}
	type args struct {
		in  []*event.Event
		buf []*event.Event
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*event.Event
	}{
		{
			name: "test buffer fills until batch size is met, but add events returns error",
			fields: fields{
				BatchSize: 3,
				evts:      make(chan *event.Event, 10),
				client: &errorsink{
					evs: []*event.Event{},
				},
			},
			args: args{
				buf: []*event.Event{},
				in: []*event.Event{
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "counter",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "gauge",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
					event.NewWithProperties(
						"event.mymeasurement",
						event.AGENT,
						map[string]string{
							"plugin":        "event",
							"agent":         "telegraf",
							"telegraf_type": "summary",
							"host":          "192.168.0.1",
						},
						map[string]interface{}{
							"message": "hello world",
						},
						time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)),
				},
			},
			want: []*event.Event{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignalFx{
				APIToken:    tt.fields.APIToken,
				BatchSize:   tt.fields.BatchSize,
				ChannelSize: tt.fields.ChannelSize,
				client:      tt.fields.client,
				dps:         tt.fields.dps,
				evts:        tt.fields.evts,
			}
			for _, e := range tt.args.in {
				s.evts <- e
			}
			s.fillAndSendEvents(tt.args.buf)
			if !reflect.DeepEqual(s.client.(*errorsink).evs, tt.want) {
				t.Errorf("fillAndSendEvents() datapoints do not match desired.  Collected: %v Desired: %v", s.client.(*errorsink).evs, tt.want)
			}
		})
	}
}

// this is really just for complete code coverage
func TestSignalFx_Description(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "verify description is correct",
			want: "Send metrics to SignalFx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignalFx{}
			if got := s.Description(); got != tt.want {
				t.Errorf("SignalFx.Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

// this is also just for complete code coverage
func TestSignalFx_SampleConfig(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "verify sample config is returned",
			want: sampleConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignalFx{}
			if got := s.SampleConfig(); got != tt.want {
				t.Errorf("SignalFx.SampleConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
