package minecraft

// minecraft.go

import (
	"fmt"
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  # server address for minecraft
  server = "localhost"
  # port for RCON
  port = "25575"
  # password RCON for mincraft server
  password = "replace_me"
`

type Minecraft struct {
	Server   string
	Port     string
	Password string
}

func (s *Minecraft) Description() string {
	return "it collects stats from Minecraft servers"
}

func (s *Minecraft) SampleConfig() string {
	return sampleConfig
}

func (s *Minecraft) Gather(acc telegraf.Accumulator) error {
	if s.Port == " " {
		acc.AddFields("state", map[string]interface{}{"value": "pretty good"}, nil)
	} else {
		acc.AddFields("state", map[string]interface{}{"value": "not great"}, nil)
	}

	return nil
}

func ParseUsername(input string) (string, error) {
	var re = regexp.MustCompile(`for\s(.*):-`)

	usernameMatches := re.FindAllStringSubmatch(input, -1)
	if usernameMatches == nil {
		return "", fmt.Errorf("no username was matched")
	}
	return usernameMatches[0][1], nil
}

func init() {
	inputs.Add("minecraft", func() telegraf.Input { return &Minecraft{} })
}
