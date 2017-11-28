package twitter

import (
	"context"
	"net/url"
	"strings"
	"sync"

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

	api    *anaconda.TwitterApi
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// Description - returns a description of the plugin
func (s *Twitter) Description() string {
	return "a plugin for pulling metrics about tweets"
}

// SampleConfig - returns a sample config block for the plugin
func (s *Twitter) SampleConfig() string {
	return `
  ## These values can be obtained from apps.twitter.com
  consumer_key = ""
  consumer_secret = ""
  access_token = ""
  access_token_secret = ""
  keywords_to_track = ""
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
	s.api = anaconda.NewTwitterApi(s.AccessToken, s.AccessTokenSecret)

	// Store the cancel function so we can call it on Stop
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.wg = &sync.WaitGroup{}
	s.wg.Add(1)
	go s.fetchTweets(acc)

	return nil
}

// Stop - Called once when stopping the plugin
func (s *Twitter) Stop() {
	s.cancel()
	s.wg.Wait()
	return
}

func (s *Twitter) fetchTweets(acc telegraf.Accumulator) {
	defer s.wg.Done()
	// We will use this a little later for finding keywords in tweets
	keywordsList := strings.Split(s.KeywordsToTrack, ",")
	// Setting the keywords we want to track
	v := url.Values{}
	v.Set("track", s.KeywordsToTrack)
	stream := s.api.PublicStreamFilter(v)
	for item := range stream.C {
		select {
		// Listen for the call to cancel so we know it's time to stop
		case <-s.ctx.Done():
			stream.Stop()
			return
		default:
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
				time, err := tweet.CreatedAtTime()
				if err != nil {
					acc.AddError(err)
					continue
				}
				for _, keyword := range keywordsList {
					if strings.Contains(strings.ToLower(tweet.Text), strings.ToLower(keyword)) {
						tags["keyword"] = strings.ToLower(keyword)
						acc.AddFields("tweets", fields, tags, time)
					}
				}
			}
		}
	}
}

func init() {
	inputs.Add("twitter", func() telegraf.Input { return &Twitter{} })
}
