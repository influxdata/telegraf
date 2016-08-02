package dockerhub

import (
	"fmt"
	"log"
	"time"

	"github.com/influxdata/telegraf"
)

const meas = "dockerhub"

type Event interface {
	NewMetric() telegraf.Metric
}

type PushData struct {
	Images   []string `json:"images"`
	PushedAt int64    `json:"pushed_at"`
	Pusher   string   `json:"pusher"`
	Tag      string   `json:"tag"`
}

type Repository struct {
	CommentCount    int    `json:"comment_count"`
	DateCreated     int64  `json:"date_created"`
	Description     string `json:"description"`
	Dockerfile      string `json:"dockerfile"`
	FullDescription string `json:"full_description"`
	IsOfficial      bool   `json:"is__official"`
	IsPrivate       bool   `json:"is_private"`
	IsTrusted       bool   `json:"is_trusted"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	Owner           string `json:"owner"`
	RepoName        string `json:"repo_name"`
	RepoURL         string `json:"repo_url"`
	StarCount       int    `json:"star_count"`
	Status          string `json:"status"`
}

type DockerhubEvent struct {
	CallbackURL string     `json:"callback_url"`
	PushData    PushData   `json:"push_data"`
	Repository  Repository `json:"repository"`
}

func (dhe DockerhubEvent) String() string {
	return fmt.Sprintf(`{
  callback_url: %v
}`,
		dhe.CallbackURL)
}

func (dhe DockerhubEvent) NewMetric() telegraf.Metric {
	tags := map[string]string{
		"description": dhe.Repository.Description,
		"name":        dhe.Repository.Name,
		"namespace":   dhe.Repository.Namespace,
		"owner":       dhe.Repository.Owner,
		"pusher":      dhe.PushData.Pusher,
		"repo_name":   dhe.Repository.RepoName,
		"repo_url":    dhe.Repository.RepoURL,
		"status":      dhe.Repository.Status,
		"tag":         dhe.PushData.Tag,
	}
	fields := map[string]interface{}{
		"comment_count": dhe.Repository.CommentCount,
		"date_created":  dhe.Repository.DateCreated,
		"is_official":   dhe.Repository.IsOfficial,
		"is_private":    dhe.Repository.IsPrivate,
		"is_trusted":    dhe.Repository.IsTrusted,
		"pushed_at":     dhe.PushData.PushedAt,
		"star_count":    dhe.Repository.StarCount,
	}
	metric, err := telegraf.NewMetric(meas, tags, fields, time.Unix(dhe.PushData.PushedAt, 0))
	if err != nil {
		log.Fatalf("Failed to create %v event", meas)
	}
	return metric
}
