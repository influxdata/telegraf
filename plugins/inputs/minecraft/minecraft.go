package minecraft

// minecraft.go

import (
	"fmt"
	"regexp"
	"strconv"

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

// Minecraft represents a connection to a minecraft server
type Minecraft struct {
	Server   string
	Port     string
	Password string
}

// Description gives a brief description.
func (s *Minecraft) Description() string {
	return "it collects stats from Minecraft servers"
}

// SampleConfig returns our sampleConfig.
func (s *Minecraft) SampleConfig() string {
	return sampleConfig
}

// Gather uses the RCON protocal to collect username and
// scoreboard stats from a minecraft server.
func (s *Minecraft) Gather(acc telegraf.Accumulator) error {
	client := &RCON{
		Server:   s.Server,
		Port:     s.Port,
		Password: s.Password,
	}

	scores, err := client.Gather()
	if err != nil {
		return err
	}

	for _, score := range scores {
		username, err := ParseUsername(score)
		if err != nil {
			return err
		}
		tags := map[string]string{
			"username": username,
			"server":   client.Server,
		}

		stats, err := ParseScoreboard(score)
		if err != nil {
			return err
		}
		var fields map[string]interface{}
		for _, stat := range stats {
			fields[stat.Name] = stat.Value
		}

		acc.AddFields("minecraft", fields, tags)
	}

	return nil
}

// ParseUsername takes an input string from rcon, to parse
// the username.
func ParseUsername(input string) (string, error) {
	var re = regexp.MustCompile(`for\s(.*):-`)

	usernameMatches := re.FindAllStringSubmatch(input, -1)
	if usernameMatches == nil {
		return "", fmt.Errorf("no username was matched")
	}
	return usernameMatches[0][1], nil
}

// Score is an individual tracked scoreboard stat.
type Score struct {
	Name  string
	Value int
}

// ParseScoreboard takes an input string from rcon, to parse
// scoreboard stats.
func ParseScoreboard(input string) ([]Score, error) {
	var re = regexp.MustCompile(`(?U):\s(\d+)\s\((.*)\)`)
	scoreMatches := re.FindAllStringSubmatch(input, -1)
	if scoreMatches == nil {
		return nil, fmt.Errorf("No scores found")
	}

	var scores []Score

	for _, match := range scoreMatches {
		//fmt.Println(match)
		number := match[1]
		name := match[2]
		n, err := strconv.Atoi(number)
		//Not necessary in current state, because regex can only match integers,
		// maybe become necessary if regex is modified to match more types of
		//numbers
		if err != nil {
			return nil, fmt.Errorf("Failed to parse statistic")
		}
		s := Score{
			Name:  name,
			Value: n,
		}
		//	fmt.Println(s)
		scores = append(scores, s)
	}
	//fmt.Println(scores)
	return scores, nil
}

func init() {
	inputs.Add("minecraft", func() telegraf.Input { return &Minecraft{} })
}
