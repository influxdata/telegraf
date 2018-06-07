package signalfx

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/event"
	"github.com/signalfx/golib/sfxclient"
)

func TestSignalFx_GetObjects(t *testing.T) {
	dp, _ := metric.New("datapoint",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"mymeasurement": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC))
	ev, _ := metric.New("event",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"mymeasurement": "hello world"},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC))
	type fields struct {
		APIToken           string
		BatchSize          int
		ChannelSize        int
		DatapointIngestURL string
		EventIngestURL     string
		Exclude            []string
		Include            []string
		exclude            map[string]bool
		include            map[string]bool
		ctx                context.Context
		client             *sfxclient.HTTPSink
		dps                chan *datapoint.Datapoint
		evts               chan *event.Event
		done               chan struct{}
	}
	type args struct {
		datapoints []telegraf.Metric
		events     []telegraf.Metric
		dps        chan *datapoint.Datapoint
		evts       chan *event.Event
	}
	type want struct {
		datapoints []*datapoint.Datapoint
		events     []*event.Event
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name:   "add datapoints",
			fields: fields{},
			args: args{
				datapoints: []telegraf.Metric{
					dp,
				},
				dps:  make(chan *datapoint.Datapoint, 10),
				evts: make(chan *event.Event, 10),
			},
			want: want{
				datapoints: []*datapoint.Datapoint{
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
				},
				events: []*event.Event{},
			},
		},
		{
			name: "add events",
			fields: fields{
				Include: []string{"event.mymeasurement"},
			},
			args: args{
				events: []telegraf.Metric{
					ev,
				},
				dps:  make(chan *datapoint.Datapoint, 10),
				evts: make(chan *event.Event, 10),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SignalFx{
				APIToken:           tt.fields.APIToken,
				BatchSize:          tt.fields.BatchSize,
				ChannelSize:        tt.fields.ChannelSize,
				DatapointIngestURL: tt.fields.DatapointIngestURL,
				EventIngestURL:     tt.fields.EventIngestURL,
				Exclude:            tt.fields.Exclude,
				Include:            tt.fields.Include,
				exclude:            tt.fields.exclude,
				include:            tt.fields.include,
				ctx:                tt.fields.ctx,
				client:             tt.fields.client,
				dps:                tt.fields.dps,
				evts:               tt.fields.evts,
				done:               tt.fields.done,
			}
			s.GetObjects(tt.args.datapoints, tt.args.dps, tt.args.evts)
			var collectedDatapoints = []*datapoint.Datapoint{}
			for i := 0; i < len(tt.args.datapoints); i++ {
				dpt := <-tt.args.dps
				collectedDatapoints = append(collectedDatapoints, dpt)
			}
			if !reflect.DeepEqual(collectedDatapoints, tt.want.datapoints) {
				t.Errorf("Collected datapoints do not match desired.  Collected: %v Desired: %v", collectedDatapoints, tt.want.datapoints)
			}
			s.GetObjects(tt.args.events, tt.args.dps, tt.args.evts)
			var collectedEvents = []*event.Event{}
			for i := 0; i < len(tt.args.events); i++ {
				evt := <-tt.args.evts
				collectedEvents = append(collectedEvents, evt)
			}
			if !reflect.DeepEqual(collectedEvents, tt.want.events) {
				t.Errorf("Collected events do not match desired.  Collected: %v Desired: %v", collectedEvents, tt.want.events)
			}
		})
	}
}
