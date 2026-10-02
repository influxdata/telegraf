package fxmacrodata

// series is one of the series requested in every gather cycle. The list of
// series is built once in Init from the configured currencies, indicators and
// currency pairs.
type series struct {
	endpoint string
	name     string
	field    string
	tags     map[string]string
}

type seriesResponse struct {
	Source string      `json:"source"`
	Data   []dataPoint `json:"data"`
}

type dataPoint struct {
	Date                 string   `json:"date"`
	Value                *float64 `json:"val"`
	AnnouncementDatetime *int64   `json:"announcement_datetime"`
}
