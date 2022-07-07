//go:generate ../../../tools/readme_config_includer/generator
package iotdb

// iotdb.go

import (
	_ "embed"
	"errors"
	"fmt"

	// Register IoTDB go client
	"github.com/apache/iotdb-client-go/client"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type IoTDB struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Timeout  int    `toml:"timeout"`
	session  *client.Session

	Log telegraf.Logger `toml:"-"`
}

func (*IoTDB) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (s *IoTDB) Init() error {
	return nil
}

func (s *IoTDB) Connect() error {
	// Make any connection required here
	// Check the configuration
	if s.Timeout < 0 {
		var errorMsg string
		errorMsg = fmt.Sprintf("Configuration Error: The value of 'timeout' should be greater than or equal to 0, but it's value is:%d.", s.Timeout)
		s.Log.Errorf(errorMsg)
		return errors.New(errorMsg)
	}

	config := &client.Config{
		Host:     s.Host,
		Port:     s.Port,
		UserName: s.User,
		Password: s.Password,
	}
	s.session = client.NewSession(config)
	if err := s.session.Open(false, s.Timeout); err != nil {
		s.Log.Errorf("Connect Error: Fail to connect host:'%s', port:'%s', err:%v", s.Host, s.Port, err)
		return err
	}

	return nil
}

func (s *IoTDB) Close() error {
	// Close any connections here.
	// Write will not be called once Close is called, so there is no need to synchronize.
	return nil
}

// Write should write immediately to the output, and not buffer writes
// (Telegraf manages the buffer for you). Returning an error will fail this
// batch of writes and the entire batch will be retried automatically.
func (s *IoTDB) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		// write `metric` to the output sink here
	}
	return nil
}

func init() {
	outputs.Add("iotdb", func() telegraf.Output { return newIoTDB() })
}

func newIoTDB() *IoTDB {
	return &IoTDB{}
}
