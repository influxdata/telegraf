package poller

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	"github.com/influxdata/telegraf/internal/config"
	"github.com/influxdata/telegraf/internal/models"

	influxconfig "github.com/influxdata/config"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"github.com/streadway/amqp"
)

// Poller runs telegraf and collects data based on the given config
type Poller struct {
	Config      *config.Config
	AMQPconn    *amqp.Connection
	AMQPchannel *amqp.Channel
	rawTasks    chan []byte
}

// NewPoller returns an Poller struct based off the given Config
func NewPoller(config *config.Config) (*Poller, error) {
	p := &Poller{
		Config: config,
	}

	if p.Config.Poller.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		p.Config.Poller.Hostname = hostname
	}

	config.Tags["host"] = p.Config.Poller.Hostname

	return p, nil
}

func (p *Poller) getTask(conn *amqp.Connection, queueName string, consumerTag string, toto chan []byte) error {
	//    defer conn.Close()
	tasks, err := p.AMQPchannel.Consume(queueName, consumerTag+"_"+queueName, false, false, false, false, nil)
	if err != nil {
		// TODO BETER HANDLING
		return fmt.Errorf("basic.consume: %v", err)
	}
	for task := range tasks {
		//log.Printf("%s \n", task)
		toto <- task.Body
		err := task.Nack(false, false)
		if err != nil {
			//TODO ????
		}
	}
	return nil
}

// Conenctio AMQP server
func (p *Poller) AMQPConnect() error {
	p.rawTasks = make(chan []byte)
	var err error
	// Prepare config
	// TODO Handle vhost
	conf := amqp.Config{
		//        Vhost: "/telegraf",
		Heartbeat: time.Duration(0) * time.Second,
	}
	// Dial server
	p.AMQPconn, err = amqp.DialConfig(p.Config.Poller.AMQPUrl, conf)
	if err != nil {
		return err
	}
	return nil
}

// Create Channel
func (p *Poller) AMQPCreateChannel() error {
	var err error
	// Create Channel
	p.AMQPchannel, err = p.AMQPconn.Channel()
	if err != nil {
		return err
	}

	for _, AMQPlabel := range p.Config.Poller.AMQPlabels {
		// Subscribing to queue
		go p.getTask(p.AMQPconn, AMQPlabel, p.Config.Poller.Hostname, p.rawTasks)
	}
	return nil
}

// Connect connects to all configured outputs
func (p *Poller) Connect() error {
	for _, o := range p.Config.Outputs {
		o.Quiet = p.Config.Poller.Quiet

		switch ot := o.Output.(type) {
		case telegraf.ServiceOutput:
			if err := ot.Start(); err != nil {
				log.Printf("Service for output %s failed to start, exiting\n%s\n",
					o.Name, err.Error())
				return err
			}
		}

		if p.Config.Poller.Debug {
			log.Printf("Attempting connection to output: %s\n", o.Name)
		}
		err := o.Output.Connect()
		if err != nil {
			log.Printf("Failed to connect to output %s, retrying in 15s, "+
				"error was '%s' \n", o.Name, err)
			time.Sleep(15 * time.Second)
			err = o.Output.Connect()
			if err != nil {
				return err
			}
		}
		if p.Config.Poller.Debug {
			log.Printf("Successfully connected to output: %s\n", o.Name)
		}
	}
	return nil
}

// Close closes the connection to all configured outputs
func (p *Poller) Close() error {
	var err error
	for _, o := range p.Config.Outputs {
		err = o.Output.Close()
		switch ot := o.Output.(type) {
		case telegraf.ServiceOutput:
			ot.Stop()
		}
	}
	// TODO close AMQP connection
	return err
}

func panicRecover(input *internal_models.RunningInput) {
	if err := recover(); err != nil {
		trace := make([]byte, 2048)
		runtime.Stack(trace, true)
		log.Printf("FATAL: Input [%s] panicked: %s, Stack:\n%s\n",
			input.Name, err, trace)
		log.Println("PLEASE REPORT THIS PANIC ON GITHUB with " +
			"stack trace, configuration, and OS information: " +
			"https://github.com/influxdata/telegraf/issues/new")
	}
}

func (p *Poller) gather(input *internal_models.RunningInput, metricC chan telegraf.Metric) error {
	defer panicRecover(input)

	var outerr error
	start := time.Now()
	acc := agent.NewAccumulator(input.Config, metricC)
	acc.SetDebug(p.Config.Poller.Debug)
	acc.SetDefaultTags(p.Config.Tags)

	if err := input.Input.Gather(acc); err != nil {
		log.Printf("Error in input [%s]: %s", input.Name, err)
	}

	elapsed := time.Since(start)
	if !p.Config.Poller.Quiet {
		log.Printf("Gathered metric, from polling, from %s in %s\n",
			input.Name, elapsed)
	}

	return outerr
}

func (p *Poller) getInput(rawInput []byte) (*internal_models.RunningInput, error) {
	// Transform rawInput from Message body to input plugin object
	table, err := toml.Parse(rawInput)
	if err != nil {
		return nil, errors.New("invalid configuration")
	}

	for name, val := range table.Fields {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return nil, errors.New("invalid configuration")
		}

		switch name {
		case "inputs", "plugins":
			for pluginName, pluginVal := range subTable.Fields {

				name := pluginName

				var table *ast.Table
				switch pluginSubTable := pluginVal.(type) {
				case *ast.Table:
					table = pluginSubTable
				case []*ast.Table:
					table = pluginSubTable[0]
					// TODO handle this case
					/*
					   for _, t := range pluginSubTable {
					       if err = c.addInput(pluginName, t); err != nil {
					           return err
					       }
					   }*/
				default:
					return nil, fmt.Errorf("Unsupported config format: %s",
						pluginName)
				}

				// TODO factorize copy/paste from config/addInput
				// Legacy support renaming io input to diskio
				if name == "io" {
					name = "diskio"
				}

				creator, ok := inputs.Inputs[name]
				if !ok {
					return nil, fmt.Errorf("Undefined but requested input: %s", name)
				}
				input := creator()

				// If the input has a SetParser function, then this means it can accept
				// arbitrary types of input, so build the parser and set it.
				switch t := input.(type) {
				case parsers.ParserInput:
					parser, err := config.BuildParser(name, table)
					if err != nil {
						return nil, err
					}
					t.SetParser(parser)
				}

				pluginConfig, err := config.BuildInput(name, table)
				if err != nil {
					return nil, err
				}

				if err := influxconfig.UnmarshalTable(table, input); err != nil {
					return nil, err
				}

				rp := &internal_models.RunningInput{
					Name:   name,
					Input:  input,
					Config: pluginConfig,
				}

				return rp, nil
			}
		default:
			// TODO log bad conf
			continue
		}
	}
	return nil, nil
}

// Test verifies that we can 'Gather' from all inputs with their configured
// Config struct
func (p *Poller) Test() error {
	//TODO remove it ?????
	shutdown := make(chan struct{})
	defer close(shutdown)
	metricC := make(chan telegraf.Metric)

	// dummy receiver for the point channel
	go func() {
		for {
			select {
			case <-metricC:
				// do nothing
			case <-shutdown:
				return
			}
		}
	}()

	for _, input := range p.Config.Inputs {
		acc := agent.NewAccumulator(input.Config, metricC)
		acc.SetDebug(true)

		fmt.Printf("* Plugin: %s, Collection 1\n", input.Name)
		if input.Config.Interval != 0 {
			fmt.Printf("* Internal: %s\n", input.Config.Interval)
		}

		if err := input.Input.Gather(acc); err != nil {
			return err
		}

		// Special instructions for some inputs. cpu, for example, needs to be
		// run twice in order to return cpu usage percentages.
		switch input.Name {
		case "cpu", "mongodb", "procstat":
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("* Plugin: %s, Collection 2\n", input.Name)
			if err := input.Input.Gather(acc); err != nil {
				return err
			}
		}

	}
	return nil
}

// flush writes a list of metrics to all configured outputs
func (p *Poller) flush() {
	var wg sync.WaitGroup

	wg.Add(len(p.Config.Outputs))
	for _, o := range p.Config.Outputs {
		go func(output *internal_models.RunningOutput) {
			defer wg.Done()
			err := output.Write()
			if err != nil {
				log.Printf("Error writing to output [%s]: %s\n",
					output.Name, err.Error())
			}
		}(o)
	}

	wg.Wait()
}

// flusher monitors the metrics input channel and flushes on the minimum interval
func (p *Poller) flusher(shutdown chan struct{}, metricC chan telegraf.Metric) error {
	// Inelegant, but this sleep is to allow the Gather threads to run, so that
	// the flusher will flush after metrics are collected.
	time.Sleep(time.Millisecond * 200)

	ticker := time.NewTicker(p.Config.Poller.FlushInterval.Duration)

	for {
		select {
		case <-shutdown:
			log.Println("Hang on, flushing any cached metrics before shutdown")
			p.flush()
			return nil
		case <-ticker.C:
			p.flush()
		case m := <-metricC:
			for _, o := range p.Config.Outputs {
				o.AddMetric(m)
			}
		}
	}
}

// Run runs the agent daemon, gathering every Interval
func (p *Poller) Run(shutdown chan struct{}) error {
	var wg sync.WaitGroup

	p.Config.Agent.FlushInterval.Duration = agent.JitterInterval(
		p.Config.Agent.FlushInterval.Duration,
		p.Config.Agent.FlushJitter.Duration)

	log.Printf("Agent Config: Debug:%#v, Quiet:%#v, Hostname:%#v, "+
		"Flush Interval:%s \n",
		p.Config.Poller.Debug, p.Config.Poller.Quiet,
		p.Config.Poller.Hostname, p.Config.Poller.FlushInterval.Duration)

	log.Print("Message queue connection\n")
	err := p.AMQPConnect()
	if err == nil {
		log.Print("Channel creation\n")
		err = p.AMQPCreateChannel()
	}
	if err == nil {
		log.Print("Message queue connected\n")
	} else {
		log.Printf("Message queue connection error: %s\n", err)
	}

	// channel shared between all input threads for accumulating metrics
	metricC := make(chan telegraf.Metric, 10000)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := p.flusher(shutdown, metricC); err != nil {
			log.Printf("Flusher routine failed, exiting: %s\n", err.Error())
			close(shutdown)
		}
	}()

	defer wg.Wait()

	c := make(chan *amqp.Error)
reconnection:
	for {
		// We need to be sure that channel is open
		// TODO handle channelS!!!
		if p.AMQPchannel != nil && p.AMQPconn != nil {
			select {
			case <-shutdown:
				return nil

			case rawTask := <-p.rawTasks:
				go func(rawTask []byte) {
					// Get input obj from message
					input, err := p.getInput(rawTask)
					if err != nil {
						log.Printf(err.Error())
					} else {
						// Do the check
						if err := p.gather(input, metricC); err != nil {
							log.Printf(err.Error())
						}
					}
				}(rawTask)
			case err := <-p.AMQPconn.NotifyClose(c):
				// handle connection errors
				// and reconnections
				log.Printf("Connection error: %s\n", err)
				break reconnection
			case err := <-p.AMQPchannel.NotifyClose(c):
				// handle channel errors
				// and reconnections
				log.Printf("Channel error: %s\n", err)
				break reconnection
			}
		} else {
			break
		}
	}

	// Handle restart
	log.Print("Message queue reconnection in 3 seconds\n")
	ticker := time.NewTicker(time.Duration(3) * time.Second)
	select {
	case <-shutdown:
		return nil
	case <-ticker.C:
	}
	// Send shutdown signal to restart routines
	log.Print("Shutdown signal send to routines\n")
	shutdown <- struct{}{}

	return p.Run(shutdown)
}
