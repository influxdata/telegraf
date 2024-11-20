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
	groupId    string
	managerUrl string
}

func (p *MockProducer) Send(ctx context.Context, msg dataproxy.Message) error {

	return nil
}

func (p *MockProducer) SendAsync(ctx context.Context, msg dataproxy.Message, callback dataproxy.Callback) {
	return
}

func (p *MockProducer) Close() {
	return
}

func (p *MockProducer) SendMessage(ctx context.Context, msg dataproxy.Message) error {
	return nil
}

func NewMockProducer(groupId string, managerUrl string) (dataproxy.Client, error) {
	p := &MockProducer{}
	p.groupId = groupId
	p.managerUrl = managerUrl
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
	s.Init()
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
