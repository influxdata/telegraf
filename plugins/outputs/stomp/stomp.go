package stomp

import (
	ts "crypto/tls"
	"net"

	"github.com/go-stomp/stomp"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

//STOMP ...
type STOMP struct {
	Host      string `toml:"host"`
	Username  string `toml:"username,omitempty"`
	Password  string `toml:"password,omitempty"`
	QueueName string `toml:"queueName"`
	SSL       bool   `toml:"ssl"`
	tls.ClientConfig
	Conn      *ts.Conn
	NetConn   net.Conn
	Stomp     *stomp.Conn
	serialize serializers.Serializer
	Log       telegraf.Logger `toml:"-"`
}

//Connect ...
func (q *STOMP) Connect() error {
	if q.SSL == true {
		tlsConfig, err := q.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		q.Conn, err = ts.Dial("tcp", q.Host, tlsConfig)
		if err != nil {
			return err
		}
		q.Stomp, err = stomp.Connect(q.Conn, stomp.ConnOpt.HeartBeat(0, 0), stomp.ConnOpt.Login(q.Username, q.Password))
		if err != nil {
			return err
		}
	} else {
		var err error
		q.NetConn, err = net.Dial("tcp", q.Host)
		if err != nil {
			return err
		}
		q.Stomp, err = stomp.Connect(q.NetConn, stomp.ConnOpt.HeartBeat(0, 0), stomp.ConnOpt.Login(q.Username, q.Password))
		if err != nil {
			return err
		}

	}

	q.Log.Info("STOMP Connected...")
	return nil
}

//SetSerializer ...
func (q *STOMP) SetSerializer(serializer serializers.Serializer) {
	q.serialize = serializer
}

//Write ...
func (q *STOMP) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		values, err := q.serialize.Serialize(metric)
		if err != nil {
			return err
		}
		err = q.Stomp.Send(q.QueueName, "text/plain",
			[]byte(values), nil)
		if err != nil {
			return err
		}
	}
	return nil
}

//Close ...
func (q *STOMP) Close() error {
	var err error
	err = q.Stomp.Disconnect()
	if err != nil {
		return err
	}
	err = q.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

//SampleConfig ...
func (q *STOMP) SampleConfig() string {
	return `
[[outputs.STOMP]]
	## Host of Active Mq broker
    	host = "localhost:61613"

	## Queue name for producer messages
    	queueName = "telegraf"


	## Optional username and password if Required to connect Active MQ server.
    	username = ""
    	password = ""


  	## Default No TLS Connecton 
    	# SSL = false

  	## Optional TLS Config
    	# SSL = true
    	# tls_ca = "/etc/telegraf/ca.pem"
    	# tls_cert = "/etc/telegraf/cert.pem"
    	# tls_key = "/etc/telegraf/key.pem"
    ## Use TLS but skip chain & host verification
    	# insecure_skip_verify = false



	## Data format to output.
    	# data_format = "json"
`
}

//Description ...
func (q *STOMP) Description() string {
	return "Telegraf Output Plugin For Stomp"
}
func init() {
	outputs.Add("stomp", func() telegraf.Output { return &STOMP{} })
}
