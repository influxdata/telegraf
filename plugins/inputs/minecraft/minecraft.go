package minecraft

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## server address for minecraft
  # server = "localhost"
  ## port for RCON
  # port = "25575"
  ## password RCON for mincraft server
  # password = ""
`

var (
	playerNameRegex = regexp.MustCompile(`for\s([^:]+):-`)
	scoreboardRegex = regexp.MustCompile(`(?U):\s(\d+)\s\((.*)\)`)
)

// Client is an interface for a client which gathers data from a minecraft server
type Client interface {
	Gather(producer RCONClientProducer) ([]string, error)
}

// Minecraft represents a connection to a minecraft server
type Minecraft struct {
	Server    string
	Port      string
	Password  string
	client    Client
	clientSet bool
}

// Description gives a brief description.
func (s *Minecraft) Description() string {
	return "Collects scores from a minecraft server's scoreboard using the RCON protocol"
}

// SampleConfig returns our sampleConfig.
func (s *Minecraft) SampleConfig() string {
	return sampleConfig
}

// Gather uses the RCON protocol to collect player and
// scoreboard stats from a minecraft server.
//var hasClient bool = false
func (s *Minecraft) Gather(acc telegraf.Accumulator) error {
	// can't simply compare s.client to nil, because comparing an interface
	// to nil often does not produce the desired result
	if !s.clientSet {
		var err error
		s.client, err = NewRCON(s.Server, s.Port, s.Password)
		if err != nil {
			return err
		}
		s.clientSet = true
	}

	// (*RCON).Gather() takes an RCONClientProducer for testing purposes
	d := defaultClientProducer{
		Server: s.Server,
		Port:   s.Port,
	}

	scores, err := s.client.Gather(d)
	if err != nil {
		return err
	}

	for _, score := range scores {
		player, err := ParsePlayerName(score)
		if err != nil {
			return err
		}
		tags := map[string]string{
			"player": player,
			"server": s.Server + ":" + s.Port,
		}

		stats, err := ParseScoreboard(score)
		if err != nil {
			return err
		}
		var fields = make(map[string]interface{}, len(stats))
		for _, stat := range stats {
			fields[stat.Name] = stat.Value
		}

		acc.AddFields("minecraft", fields, tags)
	}

	return nil
}

// ParsePlayerName takes an input string from rcon, to parse
// the player.
func ParsePlayerName(input string) (string, error) {
	playerMatches := playerNameRegex.FindAllStringSubmatch(input, -1)
	if playerMatches == nil {
		return "", fmt.Errorf("no player was matched")
	}
	return playerMatches[0][1], nil
}

// Score is an individual tracked scoreboard stat.
type Score struct {
	Name  string
	Value int
}

// ParseScoreboard takes an input string from rcon, to parse
// scoreboard stats.
func ParseScoreboard(input string) ([]Score, error) {
	scoreMatches := scoreboardRegex.FindAllStringSubmatch(input, -1)
	if scoreMatches == nil {
		return nil, fmt.Errorf("No scores found")
	}

	var scores []Score

	for _, match := range scoreMatches {
		number := match[1]
		name := match[2]
		n, err := strconv.Atoi(number)
		// Not necessary in current state, because regex can only match integers,
		// maybe become necessary if regex is modified to match more types of
		// numbers
		if err != nil {
			return nil, fmt.Errorf("Failed to parse score")
		}
		s := Score{
			Name:  name,
			Value: n,
		}
		scores = append(scores, s)
	}
	return scores, nil
}

func init() {
	inputs.Add("minecraft", func() telegraf.Input {
		return &Minecraft{
			Server: "localhost",
			Port:   "25575",
		}
	})
}
