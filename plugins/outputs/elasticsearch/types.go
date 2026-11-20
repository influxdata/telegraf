package elasticsearch

type document struct {
	index    string
	id       string
	opType   string
	pipeline string
	body     map[string]any
}

type bulkMetadata struct {
	Index    string `json:"_index"`
	ID       string `json:"_id,omitempty"`
	Type     string `json:"_type,omitempty"`
	Pipeline string `json:"pipeline,omitempty"`
}

type serverVersion struct {
	Number string `json:"number"`
}

type serverInfo struct {
	Version serverVersion `json:"version"`
}

type apiErrorResponse struct {
	Error errorDetails `json:"error"`
}

type errorDetails struct {
	Type     string         `json:"type"`
	Reason   string         `json:"reason"`
	CausedBy map[string]any `json:"caused_by"`
}

type bulkResponse struct {
	Errors bool                  `json:"errors"`
	Items  []map[string]bulkItem `json:"items"`
}

type bulkItem struct {
	Status int           `json:"status"`
	Error  *errorDetails `json:"error"`
}

type templateResponse struct {
	Acknowledged bool `json:"acknowledged"`
}
