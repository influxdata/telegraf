package chess

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
	if c.Profiles == nil {
		return fmt.Errorf("no profiles listed in the config")
	}
	return nil
}

func (c *Chess) Gather(acc telegraf.Accumulator) error {
	// if c.Ok {
	// 	acc.AddFields("state", map[string]interface{}{"value": "pretty good"}, nil)
	// } else {
	// 	acc.AddFields("state", map[string]interface{}{"value": "not great"}, nil)
	// }

	// check if profiles is not included
	if c.Profiles == nil {
		var responseLeaderData ResponseLeaderboards
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
		//fmt.Println(data)
		//unmarshall the data
		err = json.Unmarshal(data, &responseLeaderData)
		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}
		c.Log.Info(responseLeaderData)
		//fmt.Print(responseLeaderData)
	}
	return nil
}

func init() {
	inputs.Add("chess", func() telegraf.Input { return &Chess{} })
}
