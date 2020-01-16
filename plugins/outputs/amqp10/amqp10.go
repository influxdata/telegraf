package amqp10

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/Azure/go-amqp"
)

type AMQP10 struct {
	Brokers  []string          `toml:"brokers"`
	Username string            `toml:"username"`
	Password string            `toml:"password"`
	Timeout  internal.Duration `toml:"timeout"`
	Topic    string            `toml:"topic"`

	serializer   serializers.Serializer
	sentMessages int

	client      *amqp.Client
	session     *amqp.Session
	connOptions []amqp.ConnOption
}

var sampleConfig = `
  ## Brokers to publish to.  If multiple brokers are specified a random broker
  ## will be selected anytime a connection is established.  This can be
  ## helpful for load balancing when not using a dedicated load balancer.
  ## The SASPolicyKey has to be URL encoded!
  brokers = ["amqps://[SASPolicyName]:[SASPolicyKey]@[namespace].servicebus.windows.net"]

  ## Target address to send the message to.
  topic = "/target"

  ## Authentication credentials.
  # username = ""
  # password = ""

  ## Connection timeout.  If not provided, will default to 5s.  0s means no
  ## timeout (not recommended).
  # timeout = "5s"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
`

func (q *AMQP10) SampleConfig() string {
	return sampleConfig
}

func (q *AMQP10) Description() string {
	return "Publishes metrics to an AMQP 1.0 broker"
}

func (q *AMQP10) SetSerializer(serializer serializers.Serializer) {
	q.serializer = serializer
}

func (q *AMQP10) Connect() error {
	p := rand.Perm(len(q.Brokers))
	for _, n := range p {
		broker := q.Brokers[n]
		log.Printf("D! Output [amqp10] connecting to %q", broker)

		if len(q.Username) > 0 && len(q.Password) > 0 {
			q.connOptions = append(q.connOptions, amqp.ConnSASLPlain(q.Username, q.Password))
		}

		client, err := amqp.Dial(broker, q.connOptions...)

		if err == nil {
			q.client = client
			log.Printf("D! Output [amqp10] connected to %q", broker)
			break
		}
		log.Printf("D! Output [amqp10] error connecting to %q: %v", broker, err)
	}

	if q.client == nil {
		return errors.New("E! Output [amqp10] could not connect to any broker")
	}

	session, err := q.client.NewSession()
	if err != nil {
		return fmt.Errorf("E! Output [amqp10] error creating session: %v", err)
	}
	q.session = session

	return nil
}

func (q *AMQP10) Close() error {
	if q.client != nil {
		return q.client.Close()
	}
	return nil
}

func (q *AMQP10) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	for _, metric := range metrics {

		body, err := q.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		ctx := context.Background()

		// Send the message
		{
			sender, err := q.session.NewSender(
				amqp.LinkTargetAddress(q.Topic),
			)
			if err != nil {
				return fmt.Errorf("E! Output [amqp10] error creating sender: %v", err)
			}

			// TODO: use timeout var
			ctx, cancel := context.WithTimeout(ctx, q.Timeout.Duration)

			// Send message
			err = sender.Send(ctx, amqp.NewMessage(body))
			if err != nil {
				log.Printf("E! Output [amqp10] error sending message: %v", err)
			}

			sender.Close(ctx)
			cancel()
		}
	}

	return nil
}

func (q *AMQP10) serialize(metrics []telegraf.Metric) ([]byte, error) {

	var buf bytes.Buffer
	for _, metric := range metrics {
		octets, err := q.serializer.Serialize(metric)
		if err != nil {
			return nil, err
		}
		_, err = buf.Write(octets)
		if err != nil {
			return nil, err
		}
	}
	body := buf.Bytes()
	return body, nil

}

func init() {
	outputs.Add("amqp10", func() telegraf.Output {
		return &AMQP10{
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}