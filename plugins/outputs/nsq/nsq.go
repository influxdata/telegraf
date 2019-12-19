package nsq

import (
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/nsqio/go-nsq"
)

type NSQ struct {
	Server       string
	Topic        string
	producer     *nsq.Producer
	BatchMessage bool `toml:"batch"`
	ConnectRate  internal.Duration

	connectable  bool
	last_connect time.Time
	serializer   serializers.Serializer
}

var sampleConfig = `
  ## Location of nsqd instance listening on TCP
  server = "localhost:4150"
  ## NSQ topic for producer messages
  topic = "telegraf"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

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

func (n *NSQ) SampleConfig() string {
	return sampleConfig
}

func (n *NSQ) Description() string {
	return "Send telegraf measurements to NSQD"
}

func (n *NSQ) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	metricsmap := make(map[string][]telegraf.Metric)

	if !n.connectable && time.Since(n.last_connect) < n.ConnectRate.Duration {
		// We are not connected yet, and it's not time to try again. Throw away everything
		return nil
	}

	for _, metric := range metrics {
		if n.BatchMessage {
			metricsmap[n.Topic] = append(metricsmap[n.Topic], metric)
		} else {
			buf, err := n.serializer.Serialize(metric)
			if err != nil {
				return err
			}

			err = n.producer.Publish(n.Topic, buf)
			if err != nil {
				if strings.Contains(err.Error(), "timeout") {
					n.last_connect = time.Now()
					n.connectable = false
					return nil
				}
				return fmt.Errorf("FAILED to send NSQD message: %s", err)
			}
		}
	}

	for key := range metricsmap {
		buf, err := n.serializer.SerializeBatch(metricsmap[key])

		if err != nil {
			log.Printf("D! [outputs.nsq] Could not serialize metric: %v", err)
			continue
		}
		err = n.producer.Publish(n.Topic, buf)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				n.last_connect = time.Now()
				n.connectable = false
				return nil
			}
			return fmt.Errorf("Could not write to NSQ server, %s", err)
		}
	}

	n.connectable = true
	return nil
}

func init() {
	outputs.Add("nsq", func() telegraf.Output {
		return &NSQ{connectable: false, last_connect: time.Now()}
	})
}
