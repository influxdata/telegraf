package syslog

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/influxdata/go-syslog/rfc5425"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Syslog is a syslog plugin
type Syslog struct {
	Address string `toml:"server"`

	Cacert             string `toml:"tls_cacert"`
	Cert               string `toml:"tls_cert"`
	Key                string `toml:"tls_key"`
	InsecureSkipVerify bool

	mu sync.Mutex
	wg sync.WaitGroup

	listener net.Listener
}

var sampleConfig = `
    ## Specify an ip or hostname with port - eg., localhost:6514, 10.0.0.1:6514

    ## Address and port to host the syslog receiver.
    ## If no server is specified, then localhost is used as the host.
    ## If no port is specified, 6514 is used (RFC5425#section-4.1).
    server = [":6514"]

    ## TLS Config
    # tls_cacert = "/etc/telegraf/ca.pem"
    # tls_cert = "/etc/telegraf/cert.pem"
    # tls_key = "/etc/telegraf/key.pem"
    ## If false, skip chain & host verification
    # insecure_skip_verify = true
`

// SampleConfig returns sample configuration message
func (s *Syslog) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description
func (s *Syslog) Description() string {
	return "Influx syslog receiver as per RFC5425"
}

// Gather ...
func (s *Syslog) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Start starts the service.
func (s *Syslog) Start(acc telegraf.Accumulator) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// tags := map[string]string{
	// 	"address": s.Address,
	// }

	var err error
	var listener net.Listener
	if tlsConfig, err := internal.GetTLSConfig(s.Cert, s.Key, s.Cacert, s.InsecureSkipVerify); err != nil && tlsConfig != nil {
		listener, err = tls.Listen("tcp", s.Address, tlsConfig)
	} else {
		listener, err = net.Listen("tcp", s.Address)
	}
	if err != nil {
		return err
	}
	defer listener.Close()
	s.listener = listener

	for {
		log.Println("accepting ...")
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		p := rfc5425.NewParser(conn, rfc5425.WithBestEffort())
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()

			log.Println("ini handling ...")
			//handleConnection(conn)
			p.ParseExecuting(func(r *rfc5425.Result) {
				outputOne(*r)
			})

			log.Println("end handling ...")
		}()
	}
}

func outputOne(r rfc5425.Result) {
	if r.Error != nil {
		spew.Dump(r.Error)
	}
	if r.Message != nil {
		spew.Dump(r.Message)
	}
	if r.MessageError != nil {
		spew.Dump(r.MessageError)
	}
	fmt.Println()
}

// Stop cleans up all resources
func (s *Syslog) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.listener.Close()
	s.wg.Wait()

	log.Println("I! Stopped syslog receiver listening on ", s.Address)
}

func init() {
	inputs.Add("syslog", func() telegraf.Input {
		return &Syslog{
			Address: ":6514",
		}
	})
}
