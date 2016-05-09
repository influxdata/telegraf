package request_aggregates

import (
	"encoding/csv"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Timestamp time.Time
	Time      float64
	Failure   bool
}

type RequestParser struct {
	TimestampPosition int
	TimestampFormat   string
	IsTimeEpoch       bool
	TimePosition      int
	ResultPosition    int
	SuccessRegexp     *regexp.Regexp
}

// Parses a CSV line and generates a Request
func (rp *RequestParser) ParseLine(line string) (*Request, error) {
	var request Request

	// Split fields and assign values
	reader := strings.NewReader(line)
	fields, err := csv.NewReader(reader).Read()
	if err != nil {
		return nil, fmt.Errorf("ERROR: could not pass CSV line, Error: %s", err)
	}
	if rp.ResultPosition < 0 || len(fields) <= rp.ResultPosition ||
		rp.TimePosition < 0 || len(fields) <= rp.TimePosition ||
		rp.TimestampPosition < 0 || len(fields) <= rp.TimestampPosition {
		return nil, fmt.Errorf("ERROR: column position out of range")
	}

	if rp.IsTimeEpoch {
		var dur time.Duration
		dur, err = time.ParseDuration(fields[rp.TimestampPosition] + rp.TimestampFormat)
		if err != nil {
			return nil, fmt.Errorf("ERROR: could not parse epoch date, Error: %s", err)
		}
		request.Timestamp = time.Unix(0, dur.Nanoseconds())
	} else {
		request.Timestamp, err = time.Parse(rp.TimestampFormat, fields[rp.TimestampPosition])
		if err != nil {
			return nil, fmt.Errorf("ERROR: could not parse date, Error: %s", err)
		}
	}

	request.Time, err = strconv.ParseFloat(fields[rp.TimePosition], 64)
	if err != nil {
		return nil, fmt.Errorf("ERROR: could not parse time value, Error: %s", err)
	}

	if rp.SuccessRegexp != nil {
		request.Failure = !rp.SuccessRegexp.MatchString(fields[rp.ResultPosition])
	}

	return &request, nil
}
