package rabbit_mq_parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/streadway/amqp"
)

// init registers the input with telegraf
func init() {
	inputs.Add("rabbit_mq_parser", func() telegraf.Input {
		return &RabbitMQParser{}
	})
}

// ##################
// # RabbitMQParser #
// ##################

// RabbitMQParser is the top level struct for this plugin
type RabbitMQParser struct {
	RabbitmqAddress string
	QueueName       string
	Prefetch        int
	DroppedLog      string

	conn  *amqp.Connection
	ch    *amqp.Channel
	q     amqp.Queue
	drops int
	acks  int
	log   *os.File
	cl    *ConcurrencyLimiter

	sync.Mutex
}

// Description satisfies the telegraf.ServiceInput interface
func (rmq *RabbitMQParser) Description() string {
	return "RabbitMQ client with specialized parser"
}

// SampleConfig satisfies the telegraf.ServiceInput interface
func (rmq *RabbitMQParser) SampleConfig() string {
	return `
  ## Address and port for the rabbitmq server to pull from 
  rabbitmq_address = "amqp://guest:guest@localhost:5672/"
  queue_name = "task_queue"
  prefetch = 1000
  dropped_log = "/Users/johnzampolin/.rabbitmq/drops.log"
`
}

// Gather satisfies the telegraf.ServiceInput interface
// All gathering is done in the Start function
func (rmq *RabbitMQParser) Gather(_ telegraf.Accumulator) error {
	numMessages := rmq.drops + rmq.acks
	percentDrops := (float64(rmq.drops) / float64(numMessages)) * 100.0
	log.Printf("Dropped %.2f%% of %d metrics in the last interval", percentDrops, numMessages)
	rmq.drops = 0
	rmq.acks = 0
	return nil
}

// Start satisfies the telegraf.ServiceInput interface
// Yanked from "https://www.rabbitmq.com/tutorials/tutorial-two-go.html"
func (rmq *RabbitMQParser) Start(acc telegraf.Accumulator) error {
	// Create drops file
	f, err := os.Create(rmq.DroppedLog)
	if err != nil {
		panic(err)
	}
	rmq.log = f

	// Limit number of workers to the number of CPU on system
	rmq.cl = NewConcurrencyLimiter(runtime.NumCPU() * 2)

	// Create queue connection and assign it to RabbitMQParser
	conn, err := amqp.Dial(rmq.RabbitmqAddress)
	if err != nil {
		return fmt.Errorf("%v: Failed to connect to RabbitMQ", err)
	}
	rmq.conn = conn

	// Create channel and assign it to RabbitMQParser
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("%v: Failed to open a channel", err)
	}
	rmq.ch = ch

	// Declare a queue and assign it to RabbitMQParser
	q, err := ch.QueueDeclare(rmq.QueueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("%v: Failed to declare a queue", err)
	}
	rmq.q = q

	// Declare QoS on queue
	err = ch.Qos(rmq.Prefetch, 0, false)
	if err != nil {
		return fmt.Errorf("%v: failed to set Qos", err)
	}

	// Register the RabbitMQ parser as a consumer of the queue
	// And start the lister passing in the Accumulator
	msgs := rmq.registerConsumer()
	go rmq.listen(msgs, acc)

	// Log that service has started
	log.Println("Starting RabbitMQ service...")
	return nil
}

// Stop satisfies the telegraf.ServiceInput interface
func (rmq *RabbitMQParser) Stop() {
	rmq.Lock()
	defer rmq.Unlock()
	rmq.conn.Close()
	rmq.log.Close()
	rmq.ch.Close()
}

// Yanked from "https://www.rabbitmq.com/tutorials/tutorial-two-go.html"
func (rmq *RabbitMQParser) registerConsumer() <-chan amqp.Delivery {
	messages, err := rmq.ch.Consume(rmq.QueueName, "", false, false, false, false, nil)
	if err != nil {
		panic(fmt.Errorf("%v: failed establishing connection to queue", err))
	}
	return messages
}

// Iterate over messages as they are coming in
// and launch new goroutine to handle load
func (rmq *RabbitMQParser) listen(msgs <-chan amqp.Delivery, acc telegraf.Accumulator) {
	for d := range msgs {
		rmq.cl.Increment()
		rmq.handleMessage(d, acc)
	}
}

// handleMessage parses the incoming messages into *client.Point
// and then adds them to the Accumulator
func (rmq *RabbitMQParser) handleMessage(d amqp.Delivery, acc telegraf.Accumulator) {
	msg := rmq.SanitizeMsg(d)
	// If point is not valid then we will drop message
	if msg == nil {
		err := rmq.logDropped(string(d.Body))
		if err != nil {
			panic(err)
		}
		rmq.drops++
		d.Ack(false)
		rmq.cl.Decrement()
		return
	}
	d.Ack(false)
	acc.AddFields(msg.Name(), msg.Fields(), msg.Tags(), msg.Time())
	rmq.acks++
	rmq.cl.Decrement()
}

// SanitizeMsg breaks message cleanly into the different parts
// turns them into an IR and returns a point
func (rmq *RabbitMQParser) SanitizeMsg(msg amqp.Delivery) *client.Point {
	ir := &IRMessage{}
	data, err := parseBody(msg.Body)
	if err != nil {
		err := rmq.logDropped(string(msg.Body))
		if err != nil {
			panic(err)
		}
		rmq.drops++
		return nil
	}
	for key, val := range data {
		value := fmt.Sprintf("%v", val)
		switch {
		case key == "host":
			ir.Host = value
		case key == "value":
			i, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil
			}
			ir.Value = i
		case key == "key":
			ir.Key = value
		case key == "server":
			ir.Server = value
		case key == "clock":
			ir.Clock = time.Unix(int64(val.(float64)), 0)
		}
	}
	m := ir.point()
	if m == nil {
		err := rmq.logDropped(string(msg.Body))
		if err != nil {
			panic(err)
		}
		rmq.drops++
		return nil
	}
	return m
}

// logDropped writes dropped points to a file
func (rmq *RabbitMQParser) logDropped(drop string) error {
	rmq.Lock()
	defer rmq.Unlock()
	// write some text to file
	_, err := rmq.log.WriteString(fmt.Sprintf("%v\n", drop))
	if err != nil {
		return err
	}
	// save changes
	err = rmq.log.Sync()
	if err != nil {
		return err
	}
	return nil
}

// #############
// # IRMessage #
// #############

// IRMessage is an intermediate representation of the
// point as it moves through the parser
type IRMessage struct {
	Host   string
	Clock  time.Time
	Value  float64
	Key    string
	Server string
}

func (ir *IRMessage) point() *client.Point {
	meas, tags, fields := structureKey(ir.Key, ir.Value)
	tags["host"] = ir.Host
	tags["server"] = ir.Server
	pt, err := client.NewPoint(meas, tags, fields, ir.Clock)
	if err != nil {
		panic(fmt.Errorf("%v: creating float point", err))
	}
	return pt
}

func parseBody(msg []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	// Try to parse, if not replace single with double quotes
	// then return err
	if err := json.Unmarshal(msg, &data); err != nil {
		rp := bytes.Replace(msg, []byte("'"), []byte(`"`), -1)
		if err := json.Unmarshal(rp, &data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

// ################
// # Parsing Tree #
// ################

// This is awful decision tree parsing, but it works...
// Layout:
//   I've split all of the keys on the "[" and then [0] of that split on "."
//   Then I walk down all the different combinations there
//   The first switch statement is on length of the bracket split
//   within each case of the bracket switch statement there is a switch
//   on the length of the period split.
func structureKey(key string, value interface{}) (string, map[string]string, map[string]interface{}) {
	// Beginning of Influx point
	meas := ""
	tags := make(map[string]string, 0)
	fields := make(map[string]interface{}, 0)

	// BracketSplit splits the metics on the "["
	bs := strings.Split(key, "[")
	// PeriodSplit splits the first part of the metric on "."s
	ps := strings.Split(bs[0], ".")

	// Switch on the results of the bracket split
	switch len(bs) {

	// No brackets so len(split) == 1
	case 1:
		meas = ps[0]
		// Switch on the results of the period split
		switch len(ps) {
		// meas.field
		case 2:
			fields[ps[1]] = value

		// meas.field*
		case 3:
			switch {
			case ps[1] == "lbv":
				meas = jwp(ps[0], ps[1])
				fields[ps[2]] = value
			default:
				fields[fmt.Sprintf("%v.%v", ps[1], ps[2])] = value
			}
		// meas.field.field.context
		case 4:
			switch {
			case strings.Contains(ps[3], "-"):
				meas = jwp(ps[0], ps[1])
				fields[ps[2]] = value
				tags["context"] = ps[3]
			case ps[1] == "lbv" && ps[2] == "cs":
				meas = jwp(ps[0], ps[1])
				fields[jwp(ps[2], ps[3])] = value
			default:
				fields[jw2p(ps[1], ps[2], ps[3])] = value

			}
		// netscaler.lbv.(rps|srv).(rack)
		case 6:
			meas = jwp(ps[0], ps[1])
			tags["rack"] = jw2p(ps[3], ps[4], ps[5])
			fields[ps[2]] = value
		// Default - Deal with "CPU-", "Memory-", "Incoming-", "Outgoing-"
		default:
			switch {
			// "CPU-"
			case strings.Contains(key, "CPU-"):
				s := strings.Split(key, "CPU-")
				meas = "CPU"
				tags["host"] = s[1]
				fields["value"] = value
			// "Memory-"
			case strings.Contains(key, "Memory-"):
				s := strings.Split(key, "Memory-")
				meas = "Memory"
				tags["host"] = s[1]
				// "Incoming-"
				fields["value"] = value
			case strings.Contains(key, "Incoming-"):
				s := strings.Split(key, "Incoming-")
				meas = "Incoming"
				tags["host"] = s[1]
				fields["value"] = value
			// "Outgoing-"
			case strings.Contains(key, "Outgoing-"):
				s := strings.Split(key, "Outgoing-")
				meas = "Outgoing"
				tags["host"] = s[1]
				fields["value"] = value
			// Default!
			default:
				meas = key
				fields["value"] = value
			}
		}
	// Brackets so len(split) == 2
	// longest case
	case 2:

		// Switch on the results of the period split
		switch len(ps) {

		// period split only contains measurement
		case 1:
			meas = ps[0]
			bracket := trim(bs[1])
			// Arcane parsing rules
			slash := strings.Contains(bs[1], "/")
			comma := strings.Contains(bs[1], ",")
			dash := strings.Contains(bs[1], "-")
			vlan := strings.Contains(bs[1], "Vlan")
			inter := strings.Contains(meas, "if")
			switch {

			// Bracket contains something like 1/40 -> ignore
			case slash:
				fields["value"] = value
			// bracket is field name with some changes
			case comma:
				// switch "," and " " to "."
				bracket = rp(rp(bracket, ",", "."), " ", ".")
				fields[bracket] = value
			// bracket contains a port number
			case dash:
				ds := strings.Split(bracket, "-")
				tags[ds[0]] = ds[1]
				fields["value"] = value
			// Bracket contains a Vlan number
			case vlan:
				s := strings.Split(bracket, "Vlan")
				tags["Vlan"] = s[1]
				fields["value"] = value
			// Bracket contains an interface name
			case inter:
				tags["interface"] = bracket
				fields["value"] = value
			// Default
			default:
				meas = key
				fields["value"] = value
			}

		// period split contains more information as well as brackets
		case 2:
			meas = ps[0]
			bracket := trim(bs[1])
			// Switch on length of bracket
			switch {

			// short brakets
			case len(bracket) < 10:
				bracket = rp(bracket, ",", "")
				if bracket != "" {
					tags["process"] = bracket
				}
				fields[ps[1]] = value

			// medium brakets
			case len(bracket) < 25:
				// remove all {,}," from bracket
				bracket = rp(rp(rp(bracket, "\"", ""), "{", ""), "}", "")
				fields[bracket] = value

			// long brackets are system.run[curl ....]
			case len(bracket) > 25 && len(bracket) < 150:
				fields[ps[1]] = bracket
				tags["status_code"] = fmt.Sprint(value)

			// Default
			default:
				meas = ps[0]
				f := strings.Split(bracket, "FailureStatus")
				tags["ps"] = f[0]
				fields["powersupply.failurestatus"] = value
			}

		// len(period_split) == 3 and contains more information
		case 3:
			meas = ps[0]
			bracket := trim(bs[1])
			// Switch on bracket content
			switch {
			// netscaler.lbv.*
			case ps[0] == "netscaler" && ps[1] == "lbv":
				meas = jwp(ps[0], ps[1])
				tags["context"] = bracket
				fields[ps[2]] = value

			// bracket contains context
			case strings.Contains(bracket, "-"):
				fields[jwp(ps[1], ps[2])] = value
				tags["context"] = bracket

			// bracket contains file system info
			case strings.Contains(bracket, "/"):
				t := strings.Split(bracket, ",")
				tags["path"] = t[0]
				if len(t) > 1 {
					fields[jw2p(ps[1], ps[2], t[1])] = value
				} else {
					fields[jwp(ps[1], ps[2])] = value
				}

			// TODO: find a non default case that fits all "net","system","vm" mess down here
			default:
				bracketCommaSplit := strings.Split(bracket, ",")

				// Switch on bracket contents then measurement (set on line 119)
				switch {

				// system cpu and swap meas
				case bracketCommaSplit[0] == "":
					fields[jwp(ps[1], bracketCommaSplit[1])] = value

				// net meas
				case meas == "net":
					tags["interface"] = bracketCommaSplit[0]
					if len(bracketCommaSplit) > 1 {
						fields[jw2p(ps[1], ps[2], bracketCommaSplit[1])] = value
					} else {
						fields[jwp(ps[1], ps[2])] = value
					}

				// vm measurement
				case meas == "vm":
					fields[jw2p(ps[1], ps[2], bracketCommaSplit[0])] = value

				// system measurment
				case meas == "system":
					// for per-cpu metrics we need to pull out cpu as tag
					if ps[1] == "cpu" {
						fields[jw2p(ps[1], ps[2], bracketCommaSplit[0])] = value
						tags["cpu"] = bracketCommaSplit[1]
					} else {
						// For system health checks we need to store system checked (mem, disk, cpu, etc...) with diff tags
						fields[jwp(ps[1], ps[2])] = value
						tags["system"] = bracketCommaSplit[0]
					}

				// web measurement
				case meas == "web":
					if ps[2] == "time" {
						fields["value"] = value
					} else {
						fields[jwp(ps[1], ps[2])] = value
					}
					tags["system"] = "ZabbixGUI"
				// app measurement
				case meas == "app":
					fields[jwp(ps[1], ps[2])] = value
					tags["provider"] = bracket
				// Default
				default:
					meas = key
					fields["value"] = value
				}
			}

		// len(period_split) == 5 and contains most of the metadata
		case 5:
			meas = ps[0]
			bracket := trim(bs[1])
			// Switch on measurement name
			switch {

			// custom measurement -> custom.vfs.dev
			case meas == "custom":
				meas = jw2p(ps[0], ps[1], ps[2])
				tags["drive"] = bracket
				fields[jwp(ps[3], ps[4])] = value

			// app measurement
			case meas == "app":
				tags["name"] = jwp(ps[1], ps[2])
				fields[jwp(ps[3], ps[4])] = value

			// default
			default:
				meas = key
				fields["value"] = value
			}

		// Default case for len(period_split) == 5
		default:
			meas = key
			fields["value"] = value
		}

	// Multiple brackets -> grpavg["app-searchautocomplete","system.cpu.util[,user]",last,0]
	default:
		sp := strings.Split(strings.Split(key, "grpavg[")[1], ",")
		tags["app"] = trimS(sp[0])
		s := strings.Split(trimS(sp[1]), ".")
		meas = s[0]
		field := fmt.Sprintf("%v.%v.%v.%v", s[1], s[2], sp[3], sp[4])
		fields[field] = value
	}
	// Return the start of a point
	return meas, tags, fields
}

// #####################
// # Parsing Utilities #
// #####################

// join with period
func jwp(s1, s2 string) string {
	return fmt.Sprintf("%v.%v", s1, s2)
}

// join with 2 period
func jw2p(s1, s2, s3 string) string {
	return fmt.Sprintf("%v.%v.%v", s1, s2, s3)
}

// replace
func rp(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

// trims last char from string
func trim(s string) string {
	return s[0 : len(s)-1]
}

// trimS the first and last char from string
func trimS(s string) string {
	return s[1 : len(s)-1]
}

// ######################
// # ConcurrencyLimiter #
// ######################

// ConcurrencyLimiter is a go routine safe struct that can be used to
// ensure that no more than a specifid max number of goroutines are
// executing.
type ConcurrencyLimiter struct {
	inc   chan chan struct{}
	dec   chan struct{}
	max   int
	count int
}

// NewConcurrencyLimiter returns a configured limiter that will
// ensure that calls to Increment will block if the max is hit.
func NewConcurrencyLimiter(max int) *ConcurrencyLimiter {
	c := &ConcurrencyLimiter{
		inc: make(chan chan struct{}),
		dec: make(chan struct{}, max),
		max: max,
	}
	go c.handleLimits()
	return c
}

// Increment will increase the count of running goroutines by 1.
// if the number is currently at the max, the call to Increment
// will block until another goroutine decrements.
func (c *ConcurrencyLimiter) Increment() {
	r := make(chan struct{})
	c.inc <- r
	<-r
}

// Decrement will reduce the count of running goroutines by 1
func (c *ConcurrencyLimiter) Decrement() {
	c.dec <- struct{}{}
}

// handleLimits runs in a goroutine to manage the count of
// running goroutines.
func (c *ConcurrencyLimiter) handleLimits() {
	for {
		r := <-c.inc
		if c.count >= c.max {
			<-c.dec
			c.count--
		}
		c.count++
		r <- struct{}{}
	}
}
