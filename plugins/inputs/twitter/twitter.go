package twitter

import (
	"net/url"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	measurement = "tweets"
)

// Twitter - The object used to hold the configuration options from the user.
type Twitter struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
	KeywordsToTrack   string
}

// Description - returns a description of the plugin
func (s *Twitter) Description() string {
	return "a plugin for pulling metrics about tweets"
}

// SampleConfig - returns a sample config block for the plugin
func (s *Twitter) SampleConfig() string {
	return `
    ## These values can be obtained from apps.twitter.com
    consumerKey = ""
    consumerSecret = ""
    accessToken = ""
    accessTokenSecret = ""
    keywordsToTrack = ""
  `
}

// Gather - Called every interval and used for polling inputs
func (s *Twitter) Gather(acc telegraf.Accumulator) error {
	return nil
}

// Start - Called once when starting the plugin
func (s *Twitter) Start(acc telegraf.Accumulator) error {
	anaconda.SetConsumerKey(s.ConsumerKey)
	anaconda.SetConsumerSecret(s.ConsumerSecret)
	api := anaconda.NewTwitterApi(s.AccessToken, s.AccessTokenSecret)
	go s.fetchTweets(api, acc)
	return nil
}

// Stop - Called once when stopping the plugin
func (s *Twitter) Stop() {
	return
}

func (s *Twitter) fetchTweets(api *anaconda.TwitterApi, acc telegraf.Accumulator) error {
	// We will use this a little later for finding keywords in tweets
	keywordsList := strings.Split(s.KeywordsToTrack, ",")
	// Setting the keywords we want to track
	v := url.Values{}
	v.Set("track", s.KeywordsToTrack)
	stream := api.PublicStreamFilter(v)
	for {
		item := <-stream.C
		switch tweet := item.(type) {
		case anaconda.Tweet:
			fields := make(map[string]interface{})
			tags := make(map[string]string)
			if tweet.Lang != "" {
				tags["lang"] = tweet.Lang
			}
			fields["retweet_count"] = tweet.RetweetCount
			fields["tweet_id"] = tweet.IdStr
			fields["followers_count"] = tweet.User.FollowersCount
			fields["screen_name"] = tweet.User.ScreenName
			fields["friends_count"] = tweet.User.FriendsCount
			fields["favourites_count"] = tweet.User.FavouritesCount
			fields["screen_name"] = tweet.User.ScreenName
			fields["user_verified"] = tweet.User.Verified
			fields["raw"] = tweet.Text
			time, _ := tweet.CreatedAtTime()
			for i := 0; i < len(keywordsList); i++ {
				if strings.Contains(strings.ToLower(tweet.Text), strings.ToLower(keywordsList[i])) {
					tags["keyword"] = strings.ToLower(keywordsList[i])
					acc.AddFields("tweets", fields, tags, time)
				}
			}
		default:
		}
	}
}

func init() {
	inputs.Add("twitter", func() telegraf.Input { return &Twitter{} })
}
