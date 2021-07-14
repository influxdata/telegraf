package chess

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// chess.go

// Chess is the plugin type
type Chess struct {
	Profiles []string        `toml:"profiles"`
	Log      telegraf.Logger `toml:"-"`
}

const SampleConfig = `
  # A list of profiles for monotoring 
  profiles = ["username1", "username2"]
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
	// if c.Ok {
	// 	acc.AddFields("state", map[string]interface{}{"value": "pretty good"}, nil)
	// } else {
	// 	acc.AddFields("state", map[string]interface{}{"value": "not great"}, nil)
	// }

	// check if profiles is not included
	if c.Profiles == nil && len(c.Profiles) == 0 {
		var Leaderboards Leaderboards
		// request and unmarshall leaderboard information
		// and add it to the accumulator
		resp, err := http.Get("https://api.chess.com/pub/leaderboards")
		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
		data, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Println(string(data))
		//unmarshall the data
		err = json.Unmarshal(data, &Leaderboards)
		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
		//fmt.Printf("%+v\n", Leaderboards)

		for _, stat := range Leaderboards.Daily {
			var fields = make(map[string]interface{}, len(Leaderboards.Daily))
			var tags = map[string]string{
				"playerId": strconv.Itoa(stat.PlayerId),
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
