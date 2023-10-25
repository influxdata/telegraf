package amqp_consumer

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"

	_ "net/http/pprof"
)

func TestAutoEncoding(t *testing.T) {
	enc, err := internal.NewGzipEncoder()
	require.NoError(t, err)
	payload, err := enc.Encode([]byte(`measurementName fieldKey="gzip" 1556813561098000000`))
	require.NoError(t, err)

	var a AMQPConsumer
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	a.deliveries = make(map[telegraf.TrackingID]amqp091.Delivery)
	a.parser = parser
	a.decoder, err = internal.NewContentDecoder("auto")
	require.NoError(t, err)

	acc := &testutil.Accumulator{}

	d := amqp091.Delivery{
		ContentEncoding: "gzip",
		Body:            payload,
	}
	err = a.onMessage(acc, d)
	require.NoError(t, err)
	acc.AssertContainsFields(t, "measurementName", map[string]interface{}{"fieldKey": "gzip"})

	encIdentity, err := internal.NewIdentityEncoder()
	require.NoError(t, err)
	payload, err = encIdentity.Encode([]byte(`measurementName2 fieldKey="identity" 1556813561098000000`))
	require.NoError(t, err)

	d = amqp091.Delivery{
		ContentEncoding: "not_gzip",
		Body:            payload,
	}

	err = a.onMessage(acc, d)
	require.NoError(t, err)
	acc.AssertContainsFields(t, "measurementName2", map[string]interface{}{"fieldKey": "identity"})
}

func TestQueueConsume(t *testing.T) {
	// pprof
	go http.ListenAndServe(":80", nil)

	rmq := AMQPConsumer{
		Brokers: []string{
			"amqp://username:password@127.0.0.1:5672",
		},
		Exchange:     "test_exchange",
		ExchangeType: "direct",
		Queue:        "test_queue",
		//QueuePassive:              true,
		QueueConsumeCheck:         true,
		QueueConsumeCheckInterval: time.Second * 10,
		BindingKey:                "#",
		Log:                       log{},
	}

	if err := rmq.Start(new(ac)); err != nil {
		t.Fatal(err)
	}

	t.Log("waiting...")

	select {}
	// delete the "test_queue" queue to test checkQueueConsume.
	// test result:
	// 		Errorf Error inspecting queue: Exception (404) Reason: "NOT_FOUND - no queue 'test_queue' in vhost '/'"
	// 		Errorf Error inspect queue test_queue: no consumers
	// 		panic: QueueConsumeCheckFailCallback test_queue
}

type ac struct {
}

func (a ac) AddFields(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {

}

func (a ac) AddGauge(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {

}

func (a ac) AddCounter(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {

}

func (a ac) AddSummary(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {

}

func (a ac) AddHistogram(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {

}

func (a ac) AddMetric(metric telegraf.Metric) {

}

func (a ac) SetPrecision(precision time.Duration) {

}

func (a ac) AddError(err error) {

}

func (a ac) WithTracking(maxTracked int) telegraf.TrackingAccumulator {
	return tkAc{}
}

type tkAc struct {
}

func (t2 tkAc) AddFields(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {

}

func (t2 tkAc) AddGauge(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {

}

func (t2 tkAc) AddCounter(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {

}

func (t2 tkAc) AddSummary(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {

}

func (t2 tkAc) AddHistogram(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time) {

}

func (t2 tkAc) AddMetric(metric telegraf.Metric) {

}

func (t2 tkAc) SetPrecision(precision time.Duration) {

}

func (t2 tkAc) AddError(err error) {

}

func (t2 tkAc) WithTracking(maxTracked int) telegraf.TrackingAccumulator {
	return nil
}

func (t2 tkAc) AddTrackingMetric(m telegraf.Metric) telegraf.TrackingID {
	return 0
}

func (t2 tkAc) AddTrackingMetricGroup(group []telegraf.Metric) telegraf.TrackingID {
	return 0
}

func (t2 tkAc) Delivered() <-chan telegraf.DeliveryInfo {
	return nil
}

type log struct {
}

func (l log) Errorf(format string, args ...interface{}) {
	fmt.Println("Errorf", fmt.Sprintf(format, args...))
}

func (l log) Error(args ...interface{}) {
	fmt.Println("Error", args)
}

func (l log) Debugf(format string, args ...interface{}) {
	fmt.Println("Debugf", fmt.Sprintf(format, args...))
}

func (l log) Debug(args ...interface{}) {
	fmt.Println("Debug", args)
}

func (l log) Warnf(format string, args ...interface{}) {
	fmt.Println("Warnf", fmt.Sprintf(format, args...))
}

func (l log) Warn(args ...interface{}) {
	fmt.Println("Warn", args)
}

func (l log) Infof(format string, args ...interface{}) {
	fmt.Println("Infof", fmt.Sprintf(format, args...))
}

func (l log) Info(args ...interface{}) {
	fmt.Println("Info", args)
}
