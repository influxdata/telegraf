package pusher

import (
    "github.com/influxdata/telegraf"
    "github.com/pusher/pusher-http-go"
    "github.com/influxdata/telegraf/plugins/outputs"
    "github.com/influxdata/telegraf/plugins/serializers"
)

type Pusher struct {
    AppId string `toml:"app_id"`
    AppKey string `toml:"app_key"`
    AppSecret string `toml:"app_secret"`

    client *pusher.Client

    serializer serializers.Serializer
}

var sampleConfig = `
  ## Pusher Credentials
  #app_id = ""
  #app_key = ""
  #app_secret = ""

  data_format = "json"
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
    //data := map[string]string{"message": "testing"}
    values, err := p.serializer.Serialize(point)
    p.client.Trigger("test_channel", "test_event", values)

    if err != nil {
        return err
    }

    return nil
}

func (p *Pusher) Connect() error {
    client := pusher.Client{
        AppId: p.AppId,
        Key: p.AppKey,
        Secret: p.AppSecret,
        Secure: true,
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
