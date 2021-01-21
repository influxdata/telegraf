package papertrail

import (
	"time"
)

type Event struct {
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

type Count struct {
	SourceName string            `json:"source_name"`
	SourceID   int64             `json:"source_id"`
	TimeSeries *map[int64]uint64 `json:"timeseries"`
}

type SavedSearch struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Query     string `json:"query"`
	EditURL   string `json:"html_edit_url"`
	SearchURL string `json:"html_search_url"`
}

type Payload struct {
	Events      []*Event     `json:"events"`
	Counts      []*Count     `json:"counts"`
	SavedSearch *SavedSearch `json:"saved_search"`
	MaxID       string       `json:"max_id"`
	MinID       string       `json:"min_id"`
}
