package clarify

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/jsonrpc"
	"github.com/clarify/clarify-go/views"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

var errTimeout = errors.New("timeout: operation timed out")

const validResponse = `{
	"signalsByInput" : {
		"test1.value" : {
			"id": "c8bvu9fqfsjctpv7b6fg",
			"created" : true
		}
	}
}`

type MockHandler struct {
	jsonResult string
	sleep      time.Duration
}

func (m *MockHandler) Do(ctx context.Context, _ jsonrpc.Request, result any) error {
	err := json.Unmarshal([]byte(m.jsonResult), result)
	if m.sleep > 0 {
		timer := time.NewTimer(m.sleep)
		select {
		case <-ctx.Done():
			timer.Stop()
			return errTimeout
		case <-timer.C:
			timer.Stop()
			return nil
		}
	}
	return err
}

func TestGenerateID(t *testing.T) {
	clfy := &Clarify{
		Log:          testutil.Logger{},
		IDTags:       []string{"tag1", "tag2"},
		ClarifyIDTag: "clarify_input_id",
	}
	var idTests = []struct {
		inMetric telegraf.Metric
		outID    []string
		err      error
	}{
		{
			testutil.MustMetric(
				"cpu+='''..2!@#$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890",
				map[string]string{
					"tag1": "78sx",
				},
				map[string]interface{}{
					"time_idle": math.NaN(),
				},
				time.Now()),
			[]string{"cpu.time_idle.78sx"},
			errIDTooLong,
		},
		{
			testutil.MustMetric(
				"cpu@@",
				map[string]string{
					"tag1": "78sx",
					"tag2": "33t2",
				},
				map[string]interface{}{
					"time_idle": math.NaN(),
				},
				time.Now()),
			[]string{"cpu__.time_idle.78sx.33t2"},
			nil,
		},
		{
			testutil.MustMetric(
				"temperature",
				map[string]string{},
				map[string]interface{}{
					"cpu1": 12,
					"cpu2": 13,
				},
				time.Now()),
			[]string{"temperature.cpu1", "temperature.cpu2"},
			nil,
		},
		{
			testutil.MustMetric(
				"legacy_measurement",
				map[string]string{
					"clarify_input_id": "e5e82f63-3700-4997-835d-eb366b7294a2",
					"xid":              "78sx",
				},
				map[string]interface{}{
					"value": 1337,
				},
				time.Now()),
			[]string{"e5e82f63-3700-4997-835d-eb366b7294a2"},
			nil,
		},
	}
	for _, tt := range idTests {
		for n, f := range tt.inMetric.FieldList() {
			id, err := clfy.generateID(tt.inMetric, f)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
				require.True(t, slices.Contains(tt.outID, id), "\nexpected %+v\ngot %+v\n", tt.outID[n], id)
			}
		}
	}
}

func TestProcessMetrics(t *testing.T) {
	clfy := &Clarify{
		Log:          testutil.Logger{},
		IDTags:       []string{"tag1", "tag2", "node_id"},
		ClarifyIDTag: "clarify_input_id",
	}
	var idTests = []struct {
		inMetric   telegraf.Metric
		outFrame   views.DataFrame
		outSignals map[string]views.SignalSave
	}{
		{
			testutil.MustMetric(
				"cpu1",
				map[string]string{
					"tag1": "78sx",
				},
				map[string]interface{}{
					"time_idle": 1337.3,
				},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
			views.DataFrame{
				"cpu1.time_idle.78sx": views.DataSeries{
					1257894000000000: 1337.3,
				},
			},
			map[string]views.SignalSave{
				"cpu1.time_idle.78sx": {
					SignalSaveAttributes: views.SignalSaveAttributes{
						Name: "cpu1.time_idle",
						Labels: map[string][]string{
							"tag1": {"78sx"},
						},
					},
				},
			},
		},
		{
			testutil.MustMetric(
				"cpu2",
				map[string]string{
					"tag1": "78sx",
					"tag2": "33t2",
				},
				map[string]interface{}{
					"time_idle": 200,
				},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
			views.DataFrame{
				"cpu2.time_idle.78sx.33t2": views.DataSeries{
					1257894000000000: 200,
				},
			},
			map[string]views.SignalSave{
				"cpu2.time_idle.78sx.33t2": {
					SignalSaveAttributes: views.SignalSaveAttributes{
						Name: "cpu2.time_idle",
						Labels: map[string][]string{
							"tag1": {"78sx"},
							"tag2": {"33t2"},
						},
					},
				},
			},
		},
		{
			testutil.MustMetric(
				"temperature",
				map[string]string{},
				map[string]interface{}{
					"cpu1": 12,
					"cpu2": 13,
				},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
			views.DataFrame{
				"temperature.cpu1": views.DataSeries{
					1257894000000000: 12,
				},
				"temperature.cpu2": views.DataSeries{
					1257894000000000: 13,
				},
			},
			map[string]views.SignalSave{
				"temperature.cpu1": {
					SignalSaveAttributes: views.SignalSaveAttributes{
						Name: "temperature.cpu1",
					},
				},
				"temperature.cpu2": {
					SignalSaveAttributes: views.SignalSaveAttributes{
						Name: "temperature.cpu2",
					},
				},
			},
		},
		{
			testutil.MustMetric(
				"legacy_measurement",
				map[string]string{
					"clarify_input_id": "e5e82f63-3700-4997-835d-eb366b7294a2",
					"xid":              "78sx",
				},
				map[string]interface{}{
					"value": 123.333,
				},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
			views.DataFrame{
				"e5e82f63-3700-4997-835d-eb366b7294a2": views.DataSeries{
					1257894000000000: 123.333,
				},
			},
			map[string]views.SignalSave{
				"e5e82f63-3700-4997-835d-eb366b7294a2": {
					SignalSaveAttributes: views.SignalSaveAttributes{
						Name: "legacy_measurement.value",
						Labels: map[string][]string{
							"clarify-input-id": {"e5e82f63-3700-4997-835d-eb366b7294a2"},
							"xid":              {"78sx"},
						},
					},
				},
			},
		},
		{
			testutil.MustMetric(
				"opc_metric",
				map[string]string{
					"node_id": "ns=1;s=Omron PLC.Objects.new_Controller_0.GlobalVars.counter1",
				},
				map[string]interface{}{
					"value":   12345.6789,
					"quality": "GOOD",
				},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
			views.DataFrame{
				"opc_metric.value.ns_1_s_Omron_PLC.Objects.new_Controller_0.GlobalVars.counter1": views.DataSeries{
					1257894000000000: 12345.6789,
				},
			},
			map[string]views.SignalSave{
				"opc_metric.value.ns_1_s_Omron_PLC.Objects.new_Controller_0.GlobalVars.counter1": {
					SignalSaveAttributes: views.SignalSaveAttributes{
						Name: "opc_metric.value",
						Labels: map[string][]string{
							"node-id": {"ns=1;s=Omron PLC.Objects.new_Controller_0.GlobalVars.counter1"},
						},
					},
				},
			},
		},
	}
	for _, tt := range idTests {
		of, os := clfy.processMetrics([]telegraf.Metric{tt.inMetric})
		require.EqualValues(t, tt.outFrame, of)
		require.EqualValues(t, tt.outSignals, os)
	}
}

func TestTimeout(t *testing.T) {
	clfy := &Clarify{
		Log:     testutil.Logger{},
		Timeout: config.Duration(1 * time.Millisecond),
		client: clarify.NewClient("c8bvu9fqfsjctpv7b6fg", &MockHandler{
			sleep:      6 * time.Millisecond,
			jsonResult: validResponse,
		}),
	}

	metrics := []telegraf.Metric{}
	err := clfy.Write(metrics)
	require.ErrorIs(t, err, errTimeout)
}

func TestInit(t *testing.T) {
	username := config.NewSecret([]byte("user"))

	clfy := &Clarify{
		Log:     testutil.Logger{},
		Timeout: config.Duration(1 * time.Millisecond),
		client: clarify.NewClient("c8bvu9fqfsjctpv7b6fg", &MockHandler{
			sleep:      6 * time.Millisecond,
			jsonResult: validResponse,
		}),
		Username:        username,
		CredentialsFile: "file",
	}
	require.ErrorIs(t, clfy.Init(), errCredentials)
}
