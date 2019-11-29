package twitter

import (
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Twitter - plugin main structure
type Twitter struct {
	Accounts          []int64           `toml:"accounts"`
	ConsumerKey       string            `toml:"consumer_key"`
	ConsumerSecret    string            `toml:"consumer_secret"`
	AccessToken       string            `toml:"access_token"`
	AccessTokenSecret string            `toml:"access_token_secret"`
	HTTPTimeout       internal.Duration `toml:"http_timeout"`
	twitterClient     *anaconda.TwitterApi
}

const sampleConfig = `
  ## List of accounts to monitor.
  accounts = [
    783214,
    1967601206
  ]

  ## Twitter API consumer key.
  # consumer_key = ""
  ## Twitter API consumer secret.
  # consumer_secret = ""
  ## Twitter API access token.
  # access_token = ""
  ## Twitter API access token secret.
  # access_token_secret = ""
`

// SampleConfig returns sample configuration for this plugin.
func (t *Twitter) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (t *Twitter) Description() string {
	return "Gather account information from Twitter accounts."
}

// Create Twitter Client
func (t *Twitter) createTwitterClient() *anaconda.TwitterApi {
	twitterClient := anaconda.NewTwitterApiWithCredentials(
		t.AccessToken,
		t.AccessTokenSecret,
		t.ConsumerKey,
		t.ConsumerSecret,
	)

	return twitterClient
}

// Gather GitHub Metrics
func (t *Twitter) Gather(acc telegraf.Accumulator) error {
	if t.twitterClient == nil {
		t.twitterClient = t.createTwitterClient()
	}

	users, err := t.twitterClient.GetUsersLookupByIds(t.Accounts, nil)
	if err != nil {
		return err
	}

	now := time.Now()

	for _, userInfo := range users {
		tags := getTags(userInfo)
		fields := getFields(userInfo)

		acc.AddFields("twitter_account", fields, tags, now)
	}

	return nil
}

func getTags(userInfo anaconda.User) map[string]string {
	return map[string]string{
		"id":          userInfo.IdStr,
		"screen_name": userInfo.ScreenName,
	}
}

func getFields(userInfo anaconda.User) map[string]interface{} {
	return map[string]interface{}{
		"favourites": userInfo.FavouritesCount,
		"followers":  userInfo.FollowersCount,
		"friends":    userInfo.FriendsCount,
		"statuses":   userInfo.StatusesCount,
	}
}

func init() {
	inputs.Add("twitter", func() telegraf.Input {
		return &Twitter{}
	})
}
