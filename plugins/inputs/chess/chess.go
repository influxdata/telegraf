package chess

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// chess.go

// Chess is the plugin type
type Chess struct {
	Profiles    []string        `toml:"profiles"`
	Leaderboard bool            `toml:"leaderboard"`
	Log         telegraf.Logger `toml:"-"`
}

const SampleConfig = `
  # A list of profiles for monotoring 
  profiles = ["username1", "username2"]
  leaderboard = false
`

func (c *Chess) Description() string {
	return "Monitor profiles from chess.com"
}

func (c *Chess) SampleConfig() string {
	return SampleConfig
}

// Init is a method that sets up and validates the config
func (c *Chess) Init() error {
	// if c.Profiles == nil && len(c.Profiles) <= 0 {
	// 	return fmt.Errorf("no profiles listed in the config")
	// }
	return nil
}

func (c *Chess) Gather(acc telegraf.Accumulator) error {

	// check if profiles is not included
	if c.Leaderboard {
		var leaderboards Leaderboards
		// request and unmarshall leaderboard information
		// and add it to the accumulator
		resp, err := http.Get("https://api.chess.com/pub/leaderboards")
		if err != nil {
			c.Log.Errorf("failed to GET leaderboards json: %w", err)
			return err
		}

		data, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			c.Log.Errorf("failed to read leaderboards json response body: %w", err)
			return err
		}

		//unmarshall the data
		err = json.Unmarshal(data, &leaderboards)
		if err != nil {
			c.Log.Errorf("failed to unmarshall leaderboards json: %w", err)
			return err
		}

		for _, stat := range leaderboards.Daily {
			var fields = make(map[string]interface{}, len(leaderboards.Daily))
			var tags = map[string]string{
				"playerId": strconv.Itoa(stat.PlayerID),
			}
			fields["username"] = stat.Username
			fields["rank"] = stat.Rank
			fields["score"] = stat.Score
			acc.AddFields("leaderboards", fields, tags)
		}
	}
	return nil
}

func init() {
	inputs.Add("chess", func() telegraf.Input { return &Chess{} })
}
