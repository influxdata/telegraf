package papertrail

import (
	"time"
)

type event struct {
	ID                int64     `json:"id"`
	ReceivedAt        time.Time `json:"received_at"`
	DisplayReceivedAt string    `json:"display_received_at"`
	SourceIP          string    `json:"source_ip"`
	SourceName        string    `json:"source_name"`
	SourceID          int       `json:"source_id"`
	Hostname          string    `json:"hostname"`
	Program           string    `json:"program"`
	Severity          string    `json:"severity"`
	Facility          string    `json:"facility"`
	Message           string    `json:"message"`
}

type count struct {
	SourceName string            `json:"source_name"`
	SourceID   int64             `json:"source_id"`
	TimeSeries *map[int64]uint64 `json:"timeseries"`
}

type savedSearch struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Query     string `json:"query"`
	EditURL   string `json:"html_edit_url"`
	SearchURL string `json:"html_search_url"`
}

type payload struct {
	Events      []*event     `json:"events"`
	Counts      []*count     `json:"counts"`
	SavedSearch *savedSearch `json:"saved_search"`
	MaxID       string       `json:"max_id"`
	MinID       string       `json:"min_id"`
}
