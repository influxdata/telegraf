package kairosdb

import (
	"errors"
	"io"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	defaultTimeout = 20 * time.Second

	protocolTCP   = "tcp"
	protocolHTTP  = "http"
	protocolHTTPS = "https"
)

// KairosDB is the primary struct for the kairosdb plugin
type KairosDB struct {
	Address  string
	Protocol string
	User     string
	Password string
	Timeout  int64

	innerOutput innerOutput
}

var _ telegraf.Output = (*KairosDB)(nil)

type innerOutput interface {
	io.Closer
	// Connect to the Output
	Connect() error
	// Write takes in group of points to be written to the Output
	Write(metrics []telegraf.Metric) error
}

// Connect implements telegraf.Output
func (k *KairosDB) Connect() error {
	if err := k.initInnerOutput(); err != nil {
		return err
	}
	k.innerOutput.Connect()

	return errors.New("unsupported protocol: " + k.Protocol)
}

func (k *KairosDB) initInnerOutput() error {
	timeout := defaultTimeout
	if k.Timeout > 0 {
		timeout = time.Duration(k.Timeout) * time.Second
	}

	switch k.Protocol {
	case protocolTCP:
		k.innerOutput = &tcpOutput{
			address: k.Address,
			timeout: timeout,
		}
		return nil
	case protocolHTTP, protocolHTTPS:
		k.innerOutput = &httpOutput{
			url:      k.Protocol + "://" + k.Address,
			timeout:  timeout,
			user:     k.User,
			password: k.Password,
		}
		return nil
	}

	return errors.New("unsupported protocol: " + k.Protocol)
}

// Close implements telegraf.Output
func (k *KairosDB) Close() error {
	return k.innerOutput.Close()
}

// Description implements telegraf.Output
func (k *KairosDB) Description() string {
	return "Send telegraf metrics to KairosDB using the Telnet protocol"
}

// SampleConfig implements telegraf.Output
func (k *KairosDB) SampleConfig() string {
	return `
  ## address includes host and port
  address = "kairosdbhost:4242"
  ## protocol can be tcp, http, or https
  protocol = "tcp"
  ## connection/request timeout in seconds
  timeout = "10"
  ## user and password are used for basic auth if the address protocol is http or https
  user = ""
  password = ""
`
}

// Write implements telegraf.Output
func (k *KairosDB) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	return k.innerOutput.Write(metrics)
}

func init() {
	outputs.Add("kairosdb", func() telegraf.Output {
		return &KairosDB{}
	})
}
