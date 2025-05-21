package msgpack

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Serializer struct{}

func (*Serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return marshalMetric(nil, metric)
}
func (*Serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {
	buf := make([]byte, 0)
	for _, m := range metrics {
		var err error
		buf, err = marshalMetric(buf, m)

		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func marshalMetric(buf []byte, metric telegraf.Metric) ([]byte, error) {
	return (&Metric{
		Name:   metric.Name(),
		Time:   MessagePackTime{time: metric.Time()},
		Tags:   metric.Tags(),
		Fields: metric.Fields(),
	}).MarshalMsg(buf)
}

func init() {
	serializers.Add("msgpack",
		func() telegraf.Serializer {
			return &Serializer{}
		},
	)
}
