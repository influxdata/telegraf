package inlong

import (
	"context"
	"github.com/apache/inlong/inlong-sdk/dataproxy-sdk-twins/dataproxy-sdk-golang/dataproxy"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/csv"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type MockProducer struct {
	groupID    string
	managerURL string
}

func (p *MockProducer) Send(context.Context, dataproxy.Message) error {
	return nil
}

func (p *MockProducer) SendAsync(context.Context, dataproxy.Message, dataproxy.Callback) {
}

func (p *MockProducer) Close() {
}

func (p *MockProducer) SendMessage(context.Context, dataproxy.Message) error {
	return nil
}

func NewMockProducer(groupID string, managerURL string) (dataproxy.Client, error) {
	p := &MockProducer{}
	p.groupID = groupID
	p.managerURL = managerURL
	return p, nil
}

func TestInlong_Connect(t *testing.T) {
	t.Run("", func(t *testing.T) {
		i := &Inlong{
			producerFunc: NewMockProducer,
		}
		require.NoError(t, i.Connect())
	})
}

func TestInlong_Write(t *testing.T) {
	s := &csv.Serializer{Header: true}
	require.NoError(t, s.Init())
	t.Run("", func(t *testing.T) {
		producer := &MockProducer{}
		i := &Inlong{
			producer:   producer,
			serializer: s,
		}
		m := metric.New(
			"cpu",
			map[string]string{
				"topic": "xyzzy",
			},
			map[string]interface{}{
				"value": 42.0,
			},
			time.Unix(0, 0),
		)
		var metrics []telegraf.Metric
		metrics = append(metrics, m)
		require.NoError(t, i.Write(metrics))
	})
}
