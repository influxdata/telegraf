// +build !windows

package opensmtpd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func SmtpCTL(output string, Timeout int, useSudo bool) func(string, int, bool) (*bytes.Buffer, error) {
	return func(string, int, bool) (*bytes.Buffer, error) {
		return bytes.NewBuffer([]byte(output)), nil
	}
}

func TestGather(t *testing.T) {

	acc := &testutil.Accumulator{}
	v := &Opensmtpd{
		run:   SmtpCTL(smOutput, 1000, false),
		Stats: []string{"*"},
	}
	v.Gather(acc)

	assert.True(t, acc.HasMeasurement("opensmtpd"))

	//         Manual validation
	//         assert.Equal(t, acc.NFields(), len(parsedSmOutput))
	//         for field, value := range parsedSmOutput {
	//           t.Logf("field: %s - value: %f", field, value)
	//           assert.True(t, acc.HasFloatField("opensmtpd", field) )
	//           acc_value, _ := acc.FloatField("opensmtpd", field)
	//           assert.Equal(t,acc_value, value)
	//         }

	acc.AssertContainsFields(t, "opensmtpd", parsedSmOutput)
}

func TestParseFullOutput(t *testing.T) {
	acc := &testutil.Accumulator{}
	v := &Opensmtpd{
		run:   SmtpCTL(fullOutput, 1000, true),
		Stats: []string{"*"},
	}
	err := v.Gather(acc)

	assert.NoError(t, err)

	assert.True(t, acc.HasMeasurement("opensmtpd"))

	assert.Len(t, acc.Metrics, 1)
	assert.Equal(t, acc.NFields(), 36)
}

func TestFilterSomeStats(t *testing.T) {
	acc := &testutil.Accumulator{}
	v := &Opensmtpd{
		run:   SmtpCTL(fullOutput, 1000, false),
		Stats: []string{"mda.*","scheduler.*","uptime", "smtp.*"},
	}
	err := v.Gather(acc)

	assert.NoError(t, err)
	assert.True(t, acc.HasMeasurement("opensmtpd"))
	assert.Equal(t, acc.NMetrics(), uint64(1))

	assert.Equal(t, acc.NFields(), 18)
	acc.AssertContainsFields(t, "opensmtpd", parsedFSSOutput)
}

func TestFieldConfig(t *testing.T) {
	expect := map[string]int{
		"*":            36,
		"":             0,
		"uptime":         1,
		"smtp.*":       3,
		"uptime.human": 0,
	}

	for fieldCfg, expected := range expect {
		acc := &testutil.Accumulator{}
		v := &Opensmtpd{
			run:   SmtpCTL(fullOutput, 1000, true),
			Stats: strings.Split(fieldCfg, ","),
		}
		err := v.Gather(acc)

		assert.NoError(t, err)

		// If nothing to collect measurement doesn't exists
		if fieldCfg == "" || fieldCfg == "uptime.human" {
			assert.False(t, acc.HasMeasurement("opensmtpd"))
		} else {
			assert.True(t, acc.HasMeasurement("opensmtpd"))
		}

		flat := flatten(acc.Metrics)
		assert.Equal(t, expected, len(flat))
	}
}

func flatten(metrics []*testutil.Metric) map[string]interface{} {
	flat := map[string]interface{}{}
	for _, m := range metrics {
		buf := &bytes.Buffer{}
		for k, v := range m.Tags {
			buf.WriteString(fmt.Sprintf("%s=%s", k, v))
		}
		for k, v := range m.Fields {
			flat[fmt.Sprintf("%s %s", buf.String(), k)] = v
		}
	}
	return flat
}

var smOutput = `control.session=2
mda.envelope=0
mda.pending=0
mda.running=0
mda.user=0
queue.evpcache.load.hit=2
queue.evpcache.size=1
scheduler.delivery.ok=1
scheduler.envelope=0
scheduler.envelope.incoming=1
scheduler.envelope.inflight=0
scheduler.ramqueue.envelope=1
scheduler.ramqueue.message=1
scheduler.ramqueue.update=1
smtp.session=1
smtp.session.local=2
uptime=21
uptime.human=21s`

var parsedSmOutput = map[string]interface{}{
        "control_session": float64(2),
        "mda_envelope": float64(0),
        "mda_pending": float64(0),
        "mda_running": float64(0),
        "mda_user": float64(0),
        "queue_evpcache_load_hit": float64(2),
        "queue_evpcache_size": float64(1),
        "scheduler_delivery_ok": float64(1),
        "scheduler_envelope": float64(0),
        "scheduler_envelope_incoming": float64(1),
        "scheduler_envelope_inflight": float64(0),
        "scheduler_ramqueue_envelope": float64(1),
        "scheduler_ramqueue_message": float64(1),
        "scheduler_ramqueue_update": float64(1),
        "smtp_session": float64(1),
        "smtp_session_local": float64(2),
        "uptime": float64(21),
}

var parsedFSSOutput = map[string]interface{}{
        "mda_envelope": float64(0),
        "mda_pending": float64(0),
        "mda_running": float64(0),
        "mda_user": float64(0),
        "scheduler_delivery_ok": float64(1922951),
        "scheduler_delivery_permfail": float64(45967),
        "scheduler_delivery_tempfail": float64(493),
        "scheduler_envelope": float64(0),
        "scheduler_envelope_expired": float64(17),
        "scheduler_envelope_incoming": float64(0),
        "scheduler_envelope_inflight": float64(0),
        "scheduler_ramqueue_envelope": float64(0),
        "scheduler_ramqueue_message": float64(0),
        "scheduler_ramqueue_update": float64(0),
        "smtp_session": float64(0),
        "smtp_session_inet4": float64(1903412),
        "smtp_session_local": float64(10827),
        "uptime": float64(9253995),
}

var fullOutput = `bounce.envelope=0
bounce.message=0
bounce.session=0
control.session=1
mda.envelope=0
mda.pending=0
mda.running=0
mda.user=0
mta.connector=1
mta.domain=1
mta.envelope=0
mta.host=6
mta.relay=1
mta.route=1
mta.session=1
mta.source=1
mta.task=0
mta.task.running=5
queue.bounce=11495
queue.evpcache.load.hit=3927539
queue.evpcache.size=0
queue.evpcache.update.hit=508
scheduler.delivery.ok=1922951
scheduler.delivery.permfail=45967
scheduler.delivery.tempfail=493
scheduler.envelope=0
scheduler.envelope.expired=17
scheduler.envelope.incoming=0
scheduler.envelope.inflight=0
scheduler.ramqueue.envelope=0
scheduler.ramqueue.message=0
scheduler.ramqueue.update=0
smtp.session=0
smtp.session.inet4=1903412
smtp.session.local=10827
uptime=9253995
uptime.human=107d2h33m15s`
