package influxdb_v3

type apiErrorBody struct {
	ErrorMsg string `json:"error"`
	Data     []struct {
		Metric  string `json:"original_line"`
		Line    int    `json:"line_number"`
		Message string `json:"error_message"`
	} `json:"data"`
}
