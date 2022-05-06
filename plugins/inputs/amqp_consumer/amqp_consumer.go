package amqp_consumer

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

const (
	defaultMaxUndeliveredMessages = 1000
)

type empty struct{}
type semaphore chan empty

// AMQPConsumer is the top level struct for this plugin
type AMQPConsumer struct {
	URL                    string            `toml:"url" deprecated:"1.7.0;use 'brokers' instead"`
	Brokers                []string          `toml:"brokers"`
	Username               string            `toml:"username"`
	Password               string            `toml:"password"`
	Exchange               string            `toml:"exchange"`
	ExchangeType           string            `toml:"exchange_type"`
	ExchangeDurability     string            `toml:"exchange_durability"`
	ExchangePassive        bool              `toml:"exchange_passive"`
	ExchangeArguments      map[string]string `toml:"exchange_arguments"`
	MaxUndeliveredMessages int               `toml:"max_undelivered_messages"`

	// Queue Name
	Queue           string `toml:"queue"`
	QueueDurability string `toml:"queue_durability"`
	QueuePassive    bool   `toml:"queue_passive"`

	// Binding Key
	BindingKey string `toml:"binding_key"`

	// Controls how many messages the server will try to keep on the network
	// for consumers before receiving delivery acks.
	PrefetchCount int

	// AMQP Auth method
	AuthMethod string
	tls.ClientConfig

	ContentEncoding string `toml:"content_encoding"`
	Log             telegraf.Logger

	deliveries map[telegraf.TrackingID]amqp.Delivery

	parser  parsers.Parser
	conn    *amqp.Connection
	wg      *sync.WaitGroup
	cancel  context.CancelFunc
	decoder internal.ContentDecoder
}

type externalAuth struct{}

func (a *externalAuth) Mechanism() string {
	return "EXTERNAL"
}
func (a *externalAuth) Response() string {
	return "\000"
}

const (
	DefaultAuthMethod = "PLAIN"

	DefaultBroker = "amqp://localhost:5672/influxdb"

	DefaultExchangeType       = "topic"
	DefaultExchangeDurability = "durable"

	DefaultQueueDurability = "durable"

	DefaultPrefetchCount = 50
)

func (a *AMQPConsumer) SetParser(parser parsers.Parser) {
	a.parser = parser
}

// All gathering is done in the Start function
func (a *AMQPConsumer) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (a *AMQPConsumer) createConfig() (*amqp.Config, error) {
	// make new tls config
	tlsCfg, err := a.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	var auth []amqp.Authentication
	if strings.ToUpper(a.AuthMethod) == "EXTERNAL" {
		auth = []amqp.Authentication{&externalAuth{}}
	} else if a.Username != "" || a.Password != "" {
		auth = []amqp.Authentication{
			&amqp.PlainAuth{
				Username: a.Username,
				Password: a.Password,
			},
		}
	}

	config := amqp.Config{
		TLSClientConfig: tlsCfg,
		SASL:            auth, // if nil, it will be PLAIN
	}
	return &config, nil
}

// Start satisfies the telegraf.ServiceInput interface
func (a *AMQPConsumer) Start(acc telegraf.Accumulator) error {
	amqpConf, err := a.createConfig()
	if err != nil {
		return err
	}

	a.decoder, err = internal.NewContentDecoder(a.ContentEncoding)
	if err != nil {
		return err
	}

	msgs, err := a.connect(amqpConf)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	a.wg = &sync.WaitGroup{}
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.process(ctx, msgs, acc)
	}()

	go func() {
		for {
			err := <-a.conn.NotifyClose(make(chan *amqp.Error))
			if err == nil {
				break
			}

			a.Log.Infof("Connection closed: %s; trying to reconnect", err)
			for {
				msgs, err := a.connect(amqpConf)
				if err != nil {
					a.Log.Errorf("AMQP connection failed: %s", err)
					time.Sleep(10 * time.Second)
					continue
				}

				a.wg.Add(1)
				go func() {
					defer a.wg.Done()
					a.process(ctx, msgs, acc)
				}()
				break
			}
		}
	}()

	return nil
}

func (a *AMQPConsumer) connect(amqpConf *amqp.Config) (<-chan amqp.Delivery, error) {
	brokers := a.Brokers
	if len(brokers) == 0 {
		brokers = []string{a.URL}
	}

	p := rand.Perm(len(brokers))
	for _, n := range p {
		broker := brokers[n]
		a.Log.Debugf("Connecting to %q", broker)
		conn, err := amqp.DialConfig(broker, *amqpConf)
		if err == nil {
			a.conn = conn
			a.Log.Debugf("Connected to %q", broker)
			break
		}
		a.Log.Debugf("Error connecting to %q", broker)
	}

	if a.conn == nil {
		return nil, errors.New("could not connect to any broker")
	}

	ch, err := a.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %s", err.Error())
	}

	if a.Exchange != "" {
		exchangeDurable := true
		if a.ExchangeDurability == "transient" {
			exchangeDurable = false
		}

		exchangeArgs := make(amqp.Table, len(a.ExchangeArguments))
		for k, v := range a.ExchangeArguments {
			exchangeArgs[k] = v
		}

		err = a.declareExchange(
			ch,
			exchangeDurable,
			exchangeArgs)
		if err != nil {
			return nil, err
		}
	}

	q, err := a.declareQueue(ch)
	if err != nil {
		return nil, err
	}

	if a.BindingKey != "" {
		err = ch.QueueBind(
			q.Name,       // queue
			a.BindingKey, // binding-key
			a.Exchange,   // exchange
			false,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to bind a queue: %s", err)
		}
	}

	err = ch.Qos(
		a.PrefetchCount,
		0,     // prefetch-size
		false, // global
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %s", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed establishing connection to queue: %s", err)
	}

	return msgs, err
}

func (a *AMQPConsumer) declareExchange(
	channel *amqp.Channel,
	exchangeDurable bool,
	exchangeArguments amqp.Table,
) error {
	var err error
	if a.ExchangePassive {
		err = channel.ExchangeDeclarePassive(
			a.Exchange,
			a.ExchangeType,
			exchangeDurable,
			false, // delete when unused
			false, // internal
			false, // no-wait
			exchangeArguments,
		)
	} else {
		err = channel.ExchangeDeclare(
			a.Exchange,
			a.ExchangeType,
			exchangeDurable,
			false, // delete when unused
			false, // internal
			false, // no-wait
			exchangeArguments,
		)
	}
	if err != nil {
		return fmt.Errorf("error declaring exchange: %v", err)
	}
	return nil
}

func (a *AMQPConsumer) declareQueue(channel *amqp.Channel) (*amqp.Queue, error) {
	var queue amqp.Queue
	var err error

	queueDurable := true
	if a.QueueDurability == "transient" {
		queueDurable = false
	}

	if a.QueuePassive {
		queue, err = channel.QueueDeclarePassive(
			a.Queue,      // queue
			queueDurable, // durable
			false,        // delete when unused
			false,        // exclusive
			false,        // no-wait
			nil,          // arguments
		)
	} else {
		queue, err = channel.QueueDeclare(
			a.Queue,      // queue
			queueDurable, // durable
			false,        // delete when unused
			false,        // exclusive
			false,        // no-wait
			nil,          // arguments
		)
	}
	if err != nil {
		return nil, fmt.Errorf("error declaring queue: %v", err)
	}
	return &queue, nil
}

// Read messages from queue and add them to the Accumulator
func (a *AMQPConsumer) process(ctx context.Context, msgs <-chan amqp.Delivery, ac telegraf.Accumulator) {
	a.deliveries = make(map[telegraf.TrackingID]amqp.Delivery)

	acc := ac.WithTracking(a.MaxUndeliveredMessages)
	sem := make(semaphore, a.MaxUndeliveredMessages)

	for {
		select {
		case <-ctx.Done():
			return
		case track := <-acc.Delivered():
			if a.onDelivery(track) {
				<-sem
			}
		case sem <- empty{}:
			select {
			case <-ctx.Done():
				return
			case track := <-acc.Delivered():
				if a.onDelivery(track) {
					<-sem
					<-sem
				}
			case d, ok := <-msgs:
				if !ok {
					return
				}
				err := a.onMessage(acc, d)
				if err != nil {
					acc.AddError(err)
					<-sem
				}
			}
		}
	}
}

func (a *AMQPConsumer) onMessage(acc telegraf.TrackingAccumulator, d amqp.Delivery) error {
	onError := func() {
		// Discard the message from the queue; will never be able to process
		// this message.
		rejErr := d.Ack(false)
		if rejErr != nil {
			a.Log.Errorf("Unable to reject message: %d: %v", d.DeliveryTag, rejErr)
			a.conn.Close()
		}
	}

	body, err := a.decoder.Decode(d.Body)
	if err != nil {
		onError()
		return err
	}

	metrics, err := a.parser.Parse(body)
	if err != nil {
		onError()
		return err
	}

	id := acc.AddTrackingMetricGroup(metrics)
	a.deliveries[id] = d
	return nil
}

func (a *AMQPConsumer) onDelivery(track telegraf.DeliveryInfo) bool {
	delivery, ok := a.deliveries[track.ID()]
	if !ok {
		// Added by a previous connection
		return false
	}

	if track.Delivered() {
		err := delivery.Ack(false)
		if err != nil {
			a.Log.Errorf("Unable to ack written delivery: %d: %v", delivery.DeliveryTag, err)
			a.conn.Close()
		}
	} else {
		err := delivery.Reject(false)
		if err != nil {
			a.Log.Errorf("Unable to reject failed delivery: %d: %v", delivery.DeliveryTag, err)
			a.conn.Close()
		}
	}

	delete(a.deliveries, track.ID())
	return true
}

func (a *AMQPConsumer) Stop() {
	a.cancel()
	a.wg.Wait()
	err := a.conn.Close()
	if err != nil && err != amqp.ErrClosed {
		a.Log.Errorf("Error closing AMQP connection: %s", err)
		return
	}
}

func init() {
	inputs.Add("amqp_consumer", func() telegraf.Input {
		return &AMQPConsumer{
			URL:                    DefaultBroker,
			AuthMethod:             DefaultAuthMethod,
			ExchangeType:           DefaultExchangeType,
			ExchangeDurability:     DefaultExchangeDurability,
			QueueDurability:        DefaultQueueDurability,
			PrefetchCount:          DefaultPrefetchCount,
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		}
	})
}
