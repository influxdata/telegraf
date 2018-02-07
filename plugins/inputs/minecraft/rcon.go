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

// RCONClient is a representation of RCON command authorizaiton and exectution
type RCONClient interface {
	Authorize(password string) (*rcon.Packet, error)
	Execute(command string) (*rcon.Packet, error)
}

// RCON represents a RCON server connection
type RCON struct {
	Server   string
	Port     string
	Password string
	client   RCONClient
}

// RCONClientProducer is an interface which defines how a new client will be
// produced in the event of a network disconnect. It exists mainly for
// testing purposes
type RCONClientProducer interface {
	newClient() (RCONClient, error)
}

type defaultClientProducer struct {
	Server string
	Port   string
}

func (d defaultClientProducer) newClient() (RCONClient, error) {
	return newClient(d.Server, d.Port)
}

// NewRCON creates a new RCON
func NewRCON(server, port, password string) (*RCON, error) {
	client, err := newClient(server, port)
	if err != nil {
		return nil, err
	}

	return &RCON{
		Server:   server,
		Port:     port,
		Password: password,
		client:   client,
	}, nil
}

func newClient(server, port string) (*rcon.Client, error) {
	p, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	client, err := rcon.NewClient(server, p)

	// If we've encountered any error, the contents of `client` could be corrupted,
	// so we must return nil, err
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Gather receives all player scoreboard information and returns it per player.
func (r *RCON) Gather(producer RCONClientProducer) ([]string, error) {
	if r.client == nil {
		var err error
		r.client, err = producer.newClient()
		if err != nil {
			return nil, err
		}
	}

	if _, err := r.client.Authorize(r.Password); err != nil {
		// Potentially a network problem where the client will need to be
		// re-initialized
		r.client = nil
		return nil, err
	}

	packet, err := r.client.Execute(ScoreboardPlayerList)
	if err != nil {
		// Potentially a network problem where the client will need to be
		// re-initialized
		r.client = nil
		return nil, err
	}

	if !strings.Contains(packet.Body, NoMatches) {
		players := strings.Split(packet.Body, "Showing")
		if len(players) > 1 {
			return players[1:], nil
		}
	}

	return []string{}, nil
}
