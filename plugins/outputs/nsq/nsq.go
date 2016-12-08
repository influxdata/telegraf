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
	// OutputPrecision string (parsed as a duration,
	// only used for JSON output)
	OutputPrecision string

	producer *nsq.Producer

	serializer serializers.Serializer
}

var sampleConfig = `
  ## Location of nsqd instance listening on TCP
  server = "localhost:4150"
  ## NSQ topic for producer messages
  topic = "telegraf"
  ## The output_precision parameter can be used to specify the units that should
  ## be used when creating timestamps and is only used when the data_format is
  ## set to "json"; in that case valid values are "1ns", "1us" (or "1µs"), "1ms",
  ## or "1s"; for the other supported data_format types, the precision will depend
  ## on the data_format (seconds for "graphite" data, nanoseconds for "influx"
  ## data); if unspecified, then the timestamps output with "json" data
  ## will be in seconds
  output_precision = ""

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
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

	for _, metric := range metrics {
		buf, err := n.serializer.Serialize(metric, n.OutputPrecision)
		if err != nil {
			return err
		}

		err = n.producer.Publish(n.Topic, buf)
		if err != nil {
			return fmt.Errorf("FAILED to send NSQD message: %s", err)
		}
	}
	return nil
}

func init() {
	outputs.Add("nsq", func() telegraf.Output {
		return &NSQ{}
	})
}
