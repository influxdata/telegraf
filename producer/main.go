package main

import (
	"bufio"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

func main() {

	// Open file with zabbix data
	t := time.Now()
	f, err := os.Open("zabbix.txt")
	defer f.Close()
	logErr(err, "error opening file")
	log.Printf("opened file in %v\n", time.Now().Sub(t))
	// Make a new scanner to feed points into Queue
	scan := bufio.NewScanner(f)

	t = time.Now()
	// Create producer to publish to Queue
	p := newProducer()
	defer p.conn.Close()
	defer p.ch.Close()
	log.Printf("established connection in %v\n", time.Now().Sub(t))
	cntr := 0
	cl := NewConcurrencyLimiter(1000)
	t = time.Now()
	for scan.Scan() {
		body := scan.Text()
		cl.Increment()
		go p.publish(body, cl)
		cntr++
		if cntr%10000 == 0 {
			log.Printf("sent %v lines in %v", cntr, time.Now().Sub(t))
		}
	}
	log.Printf("sent %v lines in %v", cntr, time.Now().Sub(t))
}

type producer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

func (p producer) publish(body string, cl *ConcurrencyLimiter) {
	err := p.ch.Publish("", p.q.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         []byte(body),
	})
	if err != nil {
		logErr(err, "Failed to publish a message")
	}
	cl.Decrement()
}

func newProducer() producer {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	logErr(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	logErr(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		"task_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	logErr(err, "Failed to declare a queue")
	return producer{
		conn: conn,
		ch:   ch,
		q:    q,
	}
}

func logErr(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s\n", msg, err)
	}
}

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
