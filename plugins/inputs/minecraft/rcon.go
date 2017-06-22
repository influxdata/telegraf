package minecraft

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf/plugins/inputs/minecraft/internal/rcon"
)

const (
	//A sentinel value returned when there are no statistics defined on the
	//minecraft server
	NoMatches = `All matches failed`
	//Use this command to see all player statistics
	ScoreboardPlayerList = `scoreboard players list *`
)

type RCON struct {
	Server   string
	Port     string
	Password string
}

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
		fmt.Println(packet.Body)
		return strings.Split(packet.Body, "Showing"), nil
	}

	return []string{}, nil
}
