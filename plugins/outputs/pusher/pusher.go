package pusher

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/pusher/pusher-http-go"
)

type Pusher struct {
	AppId       string `toml:"app_id"`
	AppKey      string `toml:"app_key"`
	AppSecret   string `toml:"app_secret"`
	ChannelName string `toml:"channel_name"`

	Host string `toml:"host"`

	Secure bool `toml:"secure"`

	client *pusher.Client

	serializer serializers.Serializer
}

var sampleConfig = `
  ## Pusher Credentials
  ## Pusher requires all three of app ID, key and secret for authentication.
  app_id = ""
  app_key = ""
  app_secret = ""
  ## Pusher requires a channel name to be specified
  channel_name = ""
  ## Whether to use https (true) or not (false)
  secure = true
  ## Modify if your Pusher Cluster is not USA (e.g. EU or Asia)
  host = "api.pusherapp.com"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options; read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (p *Pusher) SampleConfig() string {
	return sampleConfig
}

func (p *Pusher) Description() string {
	return "Configuration for Pusher output."
}

func (p *Pusher) SetSerializer(serializer serializers.Serializer) {
	p.serializer = serializer
}

func (p *Pusher) Write(metrics []telegraf.Metric) error {
	for _, m := range metrics {
		err := p.WriteSinglePoint(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pusher) WriteSinglePoint(point telegraf.Metric) error {
	values, err := p.serializer.Serialize(point)

	if err != nil {
		return err
	}

	if _, err = p.client.Trigger(p.ChannelName, point.Name(), values); err != nil {
		return err
	}

	return nil
}

func (p *Pusher) Connect() error {
	client := pusher.Client{
		AppId:  p.AppId,
		Key:    p.AppKey,
		Secret: p.AppSecret,
		Secure: p.Secure,
		Host:   p.Host,
	}
	p.client = &client

	return nil
}

func (p *Pusher) Close() error {
	return nil
}

func init() {
	outputs.Add("pusher", func() telegraf.Output {
		return &Pusher{}
	})
}
