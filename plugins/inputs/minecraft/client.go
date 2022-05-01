package minecraft

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf/plugins/inputs/minecraft/internal/rcon"
)

var (
	scoreboardRegexLegacy = regexp.MustCompile(`(?U):\s(?P<value>\d+)\s\((?P<name>.*)\)`)
	scoreboardRegex       = regexp.MustCompile(`\[(?P<name>[^\]]+)\]: (?P<value>\d+)`)
)

// Connection is an established connection to the Minecraft server.
type Connection interface {
	// Execute runs a command.
	Execute(command string) (string, error)
}

// Connector is used to create connections to the Minecraft server.
type Connector interface {
	// Connect establishes a connection to the server.
	Connect() (Connection, error)
}

func newConnector(hostname, port, password string) *connector {
	return &connector{
		hostname: hostname,
		port:     port,
		password: password,
	}
}

type connector struct {
	hostname string
	port     string
	password string
}

func (c *connector) Connect() (Connection, error) {
	p, err := strconv.Atoi(c.port)
	if err != nil {
		return nil, err
	}

	client, err := rcon.NewClient(c.hostname, p)
	if err != nil {
		return nil, err
	}

	_, err = client.Authorize(c.password)
	if err != nil {
		return nil, err
	}

	return &connection{client: client}, nil
}

func newClient(connector Connector) *client {
	return &client{connector: connector}
}

type client struct {
	connector Connector
	conn      Connection
}

func (c *client) Connect() error {
	conn, err := c.connector.Connect()
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *client) Players() ([]string, error) {
	if c.conn == nil {
		err := c.Connect()
		if err != nil {
			return nil, err
		}
	}

	resp, err := c.conn.Execute("scoreboard players list")
	if err != nil {
		c.conn = nil
		return nil, err
	}

	return parsePlayers(resp), nil
}

func (c *client) Scores(player string) ([]Score, error) {
	if c.conn == nil {
		err := c.Connect()
		if err != nil {
			return nil, err
		}
	}

	resp, err := c.conn.Execute("scoreboard players list " + player)
	if err != nil {
		c.conn = nil
		return nil, err
	}

	return parseScores(resp), nil
}

type connection struct {
	client *rcon.Client
}

func (c *connection) Execute(command string) (string, error) {
	packet, err := c.client.Execute(command)
	if err != nil {
		return "", err
	}
	return packet.Body, nil
}

func parsePlayers(input string) []string {
	parts := strings.SplitAfterN(input, ":", 2)
	if len(parts) != 2 {
		return []string{}
	}

	names := strings.Split(parts[1], ",")

	// Detect Minecraft <= 1.12
	if strings.Contains(parts[0], "players on the scoreboard") && len(names) > 0 {
		// Split the last two player names: ex: "notch and dinnerbone"
		head := names[:len(names)-1]
		tail := names[len(names)-1]
		names = append(head, strings.SplitN(tail, " and ", 2)...)
	}

	var players []string
	for _, name := range names {
		name := strings.TrimSpace(name)
		if name == "" {
			continue
		}
		players = append(players, name)
	}
	return players
}

// Score is an individual tracked scoreboard stat.
type Score struct {
	Name  string
	Value int64
}

func parseScores(input string) []Score {
	if strings.Contains(input, "has no scores") {
		return []Score{}
	}

	// Detect Minecraft <= 1.12
	var re *regexp.Regexp
	if strings.Contains(input, "tracked objective") {
		re = scoreboardRegexLegacy
	} else {
		re = scoreboardRegex
	}

	var scores []Score
	matches := re.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		score := Score{}
		for i, subexp := range re.SubexpNames() {
			switch subexp {
			case "name":
				score.Name = match[i]
			case "value":
				value, err := strconv.ParseInt(match[i], 10, 64)
				if err != nil {
					continue
				}
				score.Value = value
			default:
				continue
			}
		}
		scores = append(scores, score)
	}

	return scores
}
