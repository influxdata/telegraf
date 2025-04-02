package inlong

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/apache/inlong/inlong-sdk/dataproxy-sdk-twins/dataproxy-sdk-golang/dataproxy"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/csv"
)

type mockProducer struct {
	groupID      string
	managerURL   string
	ReceivedData []byte
}

func (m *mockProducer) Send(_ context.Context, msg dataproxy.Message) error {
	m.ReceivedData = msg.Payload
	return nil
}

func (*mockProducer) SendAsync(context.Context, dataproxy.Message, dataproxy.Callback) {
}

func (*mockProducer) Close() {
}

func (*mockProducer) SendMessage(context.Context, dataproxy.Message) error {
	return nil
}

func newMockProducer(groupID, managerURL string) (dataproxy.Client, error) {
	return &mockProducer{
		groupID:    groupID,
		managerURL: managerURL,
	}, nil
}

func TestInlong_Connect(t *testing.T) {
	managerURL := "http://inlong-manager:8080"
	groupID := "test-group"
	i := &Inlong{
		ManagerURL: managerURL,
		GroupID:    groupID,
		producerFunc: func(gid, url string) (dataproxy.Client, error) {
			require.Equal(t, groupID, gid)
			require.Equal(t, managerURL+"/inlong/manager/openapi/dataproxy/getIpList", url)
			return newMockProducer(gid, url)
		},
	}
	defer i.Close()
	require.NoError(t, i.Connect())
}

func TestInlong_Write(t *testing.T) {
	s := &csv.Serializer{Header: true}
	require.NoError(t, s.Init())
	producer := &mockProducer{}
	i := &Inlong{
		producer:   producer,
		serializer: s,
	}
	m := metric.New(
		"cpu",
		map[string]string{
			"topic": "test-topic",
		},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	var metrics []telegraf.Metric
	metrics = append(metrics, m)
	require.NoError(t, i.Write(metrics))
	data := []string{"timestamp,measurement,topic,value", "0,cpu,test-topic,42", ""}
	require.Equal(t, strings.Join(data, getSeparator()), string(producer.ReceivedData))
}

func getSeparator() string {
	switch runtime.GOOS {
	case "windows":
		return "\r\n"
	default:
		return "\n"
	}
}
