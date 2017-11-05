package msgpack

import (
	"github.com/influxdata/telegraf"
)

type MsgpackSerializer struct {
}

func (s *MsgpackSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	buf, err := (&Metric{
		Name:   metric.Name(),
		Time:   metric.Time(),
		Tags:   metric.Tags(),
		Fields: metric.Fields(),
	}).MarshalMsg(nil)

	if err != nil {
		return nil, err
	}

	return buf, nil
}
