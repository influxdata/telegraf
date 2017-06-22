package minecraft

import (
	"strconv"
	"strings"

	"github.com/influxdata/telegraf/plugins/inputs/minecraft/internal/rcon"
)

const (
	// NoMatches is a sentinel value returned when there are no statistics defined on the
	//minecraft server
	NoMatches = `All matches failed`
	// ScoreboardPlayerList is the command to see all player statistics
	ScoreboardPlayerList = `scoreboard players list *`
)

// RCON represents a RCON server connection
type RCON struct {
	Server   string
	Port     string
	Password string
}

// Gather recieves all player scoreboard information and returns it per user.
func (r *RCON) Gather() ([]string, error) {
	port, err := strconv.Atoi(r.Port)
	if err != nil {
		return nil, err
	}

	client, err := rcon.NewClient(r.Server, port)
	if err != nil {
		return nil, err
	}

	if _, err = client.Authorize(r.Password); err != nil {
		return nil, err
	}

	packet, err := client.Execute(ScoreboardPlayerList)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(packet.Body, NoMatches) {
		users := strings.Split(packet.Body, "Showing")
		if len(users) > 1 {
			return users[1:], nil
		}
	}

	return []string{}, nil
}
