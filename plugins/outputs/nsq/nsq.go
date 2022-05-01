package nsq

import (
	"fmt"

	"github.com/nsqio/go-nsq"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type NSQ struct {
	Server string
	Topic  string
	Log    telegraf.Logger `toml:"-"`

	producer   *nsq.Producer
	serializer serializers.Serializer
}

func (n *NSQ) SetSerializer(serializer serializers.Serializer) {
	n.serializer = serializer
}

func (n *NSQ) Connect() error {
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(n.Server, config)

	if err != nil {
		return err
	}

	n.producer = producer
	return nil
}

func (n *NSQ) Close() error {
	n.producer.Stop()
	return nil
}

func (n *NSQ) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		buf, err := n.serializer.Serialize(metric)
		if err != nil {
			n.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		err = n.producer.Publish(n.Topic, buf)
		if err != nil {
			return fmt.Errorf("failed to send NSQD message: %s", err)
		}
	}
	return nil
}

func init() {
	outputs.Add("nsq", func() telegraf.Output {
		return &NSQ{}
	})
}
