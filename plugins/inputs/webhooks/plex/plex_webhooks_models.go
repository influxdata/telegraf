package plex

import (
	"fmt"
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

const meas = "plex_webhooks"

type Event interface {
	NewMetric() telegraf.Metric
}

type PlexWebhookEvent struct {
	Event    string   `json:"event"`
	User     bool     `json:"user"`
	Owner    bool     `json:"owner"`
	Account  Account  `json:"Account"`
	Server   Server   `json:"Server"`
	Player   Player   `json:"Player"`
	Metadata Metadata `json:"Metadata"`
}

type Account struct {
	ID    int    `json:"id"`
	Thumb string `json:"thumb"`
	Title string `json:"title"`
}

type Server struct {
	Title string `json:"title"`
	UUID  string `json:"uuid"`
}

type Player struct {
	Local         bool   `json:"local"`
	PublicAddress string `json:"PublicAddress"`
	Title         string `json:"title"`
	UUID          string `json:"uuid"`
}

type Metadata struct {
	LibrarySectionType   string `json:"librarySectionType"`
	RatingKey            string `json:"ratingKey"`
	Key                  string `json:"key"`
	ParentRatingKey      string `json:"parentRatingKey"`
	GrandparentRatingKey string `json:"grandparentRatingKey"`
	GUID                 string `json:"guid"`
	LibrarySectionID     int    `json:"librarySectionID"`
	MediaType            string `json:"type"`
	Title                string `json:"title"`
	GrandparentKey       string `json:"grandparentKey"`
	ParentKey            string `json:"parentKey"`
	GrandparentTitle     string `json:"grandparentTitle"`
	ParentTitle          string `json:"parentTitle"`
	Summary              string `json:"summary"`
	Index                int    `json:"index"`
	ParentIndex          int    `json:"parentIndex"`
	RatingCount          int    `json:"ratingCount"`
	Thumb                string `json:"thumb"`
	Art                  string `json:"art"`
	ParentThumb          string `json:"parentThumb"`
	GrandparentThumb     string `json:"grandparentThumb"`
	GrandparentArt       string `json:"grandparentArt"`
	AddedAt              int    `json:"addedAt"`
	UpdatedAt            int    `json:"updatedAt"`
}

func (s PlexWebhookEvent) NewMetric() telegraf.Metric {
	t := map[string]string{
		"event":                  s.Event,
		"is_user_webhook":        fmt.Sprintf("%v", s.User),
		"is_owner_webhook":       fmt.Sprintf("%v", s.Owner),
		"user_id":                fmt.Sprintf("%v", s.Account.ID),
		"user_name":              s.Account.Title,
		"server_title":           s.Server.Title,
		"server_uuid":            s.Server.UUID,
		"is_player_local":        fmt.Sprintf("%v", s.Player.Local),
		"player_public_ip":       s.Player.PublicAddress,
		"player_title":           s.Player.Title,
		"player_uuid":            s.Player.UUID,
		"library_selection_type": s.Metadata.LibrarySectionType,
		"media_type":             s.Metadata.MediaType,
		"grandparent_key":        s.Metadata.GrandparentKey,
		"parent_key":             s.Metadata.ParentKey,
		"grandparent_title":      s.Metadata.GrandparentTitle,
		"parent_title":           s.Metadata.ParentTitle,
		"parent_index":           fmt.Sprintf("%v", s.Metadata.ParentIndex),
	}
	f := map[string]interface{}{
		"rating_count":         s.Metadata.RatingCount,
		"added_at":             s.Metadata.AddedAt,
		"updated_at":           s.Metadata.UpdatedAt,
		"summary":              s.Metadata.Summary,
		"title":                s.Metadata.Title,
		"index":                s.Metadata.Index,
		"library_selection_id": s.Metadata.LibrarySectionID,
		"guid":                 s.Metadata.GUID,
	}
	m, err := metric.New(meas, t, f, time.Now())
	if err != nil {
		log.Fatalf("Failed to create %v event", s.Event)
	}
	return m
}
