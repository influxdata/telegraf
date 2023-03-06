package amqp_consumer

import (
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestAutoEncoding(t *testing.T) {
	enc := internal.NewGzipEncoder()
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

	encIdentity := internal.NewIdentityEncoder()
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
