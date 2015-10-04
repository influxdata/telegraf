package mqtt

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	paho "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/influxdb/influxdb/client"
	t "github.com/influxdb/telegraf"
	"github.com/influxdb/telegraf/outputs"
)

const MaxClientIdLen = 8
const MaxRetryCount = 3
const ClientIdPrefix = "telegraf"

type MQTT struct {
	Servers     []string `toml:"servers"`
	Username    string
	Password    string
	Database    string
	Timeout     t.Duration
	TopicPrefix string
	Hostname    string

	Client *paho.Client
	Opts   *paho.ClientOptions
	sync.Mutex
}

var sampleConfig = `
 	servers = ["localhost:1883"] # required.

        # MQTT outputs send metrics
        #    "<topic_prefix>/host/<hostname>/mem_used_percent"
        # topic_prefix = "prefix"

        # Set hostname used in the sending topic. if empty use os.Hostname().
        # This is not inherit from agent config
        # hostname = "mytelegraf"

        # username and password to connect MQTT server.
 	# username = "telegraf"
 	# password = "metricsmetricsmetricsmetrics"
`

func (m *MQTT) Connect() error {
	m.setHostname()

	var err error
	m.Lock()
	defer m.Unlock()

	m.Opts, err = m.CreateOpts()
	if err != nil {
		return err
	}

	m.Client = paho.NewClient(m.Opts)
	if token := m.Client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (m *MQTT) Close() error {
	if m.Client.IsConnected() {
		m.Client.Disconnect(20)
	}
	return nil
}

func (m *MQTT) SampleConfig() string {
	return sampleConfig
}

func (m *MQTT) Description() string {
	return "Configuration for MQTT server to send metrics to"
}

func (m *MQTT) Write(bp client.BatchPoints) error {
	m.Lock()
	defer m.Unlock()
	if len(bp.Points) == 0 {
		return nil
	}

	for _, p := range bp.Points {
		var t []string
		if m.TopicPrefix != "" {
			t = append(t, m.TopicPrefix)
		}
		t = append(t, "host", m.Hostname, p.Measurement)
		topic := strings.Join(t, "/")

		var value string
		if p.Raw != "" {
			value = p.Raw
		} else {
			value = getValue(p.Fields["value"])
		}
		err := m.publish(topic, value)
		if err != nil {
			return fmt.Errorf("Could not write to MQTT server, %s", err)
		}
	}

	return nil
}

func getValue(v interface{}) string {
	var ret string
	switch v.(type) {
	default:
		ret = fmt.Sprintf("%v", v)
	case bool:
		ret = fmt.Sprintf("%t", v)
	case float32, float64:
		ret = fmt.Sprintf("%f", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		ret = fmt.Sprintf("%d", v)
	case string, []byte:
		ret = fmt.Sprintf("%s", v)
	}
	return ret
}

func (m *MQTT) publish(topic, body string) error {
	token := m.Client.Publish(topic, 0, false, body)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (m *MQTT) CreateOpts() (*paho.ClientOptions, error) {
	opts := paho.NewClientOptions()

	clientId := getRandomClientId()
	opts.SetClientID(clientId)

	TLSConfig := &tls.Config{InsecureSkipVerify: false}
	ca := "" // TODO
	scheme := "tcp"
	if ca != "" {
		scheme = "ssl"
		certPool, err := getCertPool(ca)
		if err != nil {
			return nil, err
		}
		TLSConfig.RootCAs = certPool
	}
	TLSConfig.InsecureSkipVerify = true // TODO
	opts.SetTLSConfig(TLSConfig)

	user := m.Username
	if user == "" {
		opts.SetUsername(user)
	}
	password := m.Password
	if password != "" {
		opts.SetPassword(password)
	}

	if len(m.Servers) == 0 {
		return opts, fmt.Errorf("could not get host infomations")
	}
	for _, host := range m.Servers {
		server := fmt.Sprintf("%s://%s", scheme, host)

		opts.AddBroker(server)
	}
	opts.SetAutoReconnect(true)
	return opts, nil
}

func getRandomClientId() string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, MaxClientIdLen)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return ClientIdPrefix + "-" + string(bytes)
}

func getCertPool(pemPath string) (*x509.CertPool, error) {
	certs := x509.NewCertPool()

	pemData, err := ioutil.ReadFile(pemPath)
	if err != nil {
		return nil, err
	}
	certs.AppendCertsFromPEM(pemData)
	return certs, nil
}

// setHostname overwrites default hostname.
// TODO: should use agent.Hostname
func (m *MQTT) setHostname() {
	if m.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = ""
		}
		m.Hostname = hostname
	}
}

func init() {
	outputs.Add("mqtt", func() outputs.Output {
		return &MQTT{}
	})
}
