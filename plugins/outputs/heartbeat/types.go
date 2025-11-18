package heartbeat

type message struct {
	ID                   string      `json:"id"`
	Version              string      `json:"version"`
	Schema               int         `json:"schema"`
	LastSuccessfulUpdate *int64      `json:"last,omitempty"`
	Hostname             string      `json:"hostname,omitempty"`
	ConfigSources        *[]string   `json:"configurations,omitempty"`
	Statistics           *statsEntry `json:"statistics,omitempty"`
	Logs                 *[]logEntry `json:"logs,omitempty"`
	Status               string      `json:"status,omitempty"`
}

type statsEntry struct {
	Errors   uint64 `json:"errors"`
	Warnings uint64 `json:"warnings"`
	Metrics  uint64 `json:"metrics"`
}

type logEntry struct {
	Timestamp  string                 `json:"time"`
	Level      string                 `json:"level"`
	Source     string                 `json:"source"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Message    string                 `json:"message"`
}
