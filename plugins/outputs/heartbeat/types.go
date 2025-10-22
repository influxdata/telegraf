package heartbeat

type message struct {
	ID            string       `json:"id"`
	Version       string       `json:"version"`
	Schema        int          `json:"schema"`
	Hostname      string       `json:"hostname,omitempty"`
	Metrics       *uint64      `json:"metrics,omitempty"`
	ConfigSources *[]string    `json:"configurations,omitempty"`
	Logs          *logsMessage `json:"logs,omitempty"`
	Status        string       `json:"status,omitempty"`
}

type logsMessage struct {
	Errors   *uint64     `json:"errors,omitempty"`
	Warnings *uint64     `json:"warnings,omitempty"`
	Entries  *[]logEntry `json:"entries,omitempty"`
}

type logEntry struct {
	Timestamp  string                 `json:"time"`
	Level      string                 `json:"level"`
	Source     string                 `json:"source"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Messsage   string                 `json:"message"`

	// Internal
	index int
}
