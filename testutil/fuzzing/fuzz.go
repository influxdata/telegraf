package fuzz

// This file implements fuzzers using go-fuzz.
// More information about go-fuzz can be found here:
// https://github.com/dvyukov/go-fuzz

// To run the fuzzer locally, follow these steps:
// 1) go get github.com/influxdata/telegraf
// 2) go get -u github.com/dvyukov/go-fuzz/go-fuzz
// 3) go get -u github.com/dvyukov/go-fuzz/go-fuzz-build
// 4) cd into dir of fuzz.go
// 5) $GOPATH/bin/go-fuzz-build
// 6) $GOPATH/bin/go-fuzz

import (
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"time"
)

var DefaultTime = func() time.Time {
	return time.Unix(3600, 0)
}

func FuzzCsvParse(data []byte) int {
	parser, err := csv.NewParser(
		&csv.Config{
			ColumnNames: []string{"first", "second", "third"},
			TagColumns:  []string{"third"},
			TimeFunc:    DefaultTime,
		},
	)
	if err != nil {
		return -1
	}
	_, err = parser.Parse(data)
	if err != nil {
		return 0
	}
	return 1
}

func FuzzCsvParseLine(data []byte) int {
	parser, err := csv.NewParser(
		&csv.Config{
			ColumnNames: []string{"first", "second", "third"},
			TagColumns:  []string{"third"},
			TimeFunc:    DefaultTime,
		},
	)
	if err != nil {
		return -1
	}
	_, err = parser.ParseLine(string(data))
	if err != nil {
		return 0
	}
	return 1
}

func FuzzJSONParse(data []byte) int {
	parser, err := json.New(&json.Config{
		MetricName: "fuzz_test",
	})
	if err != nil {
		return -1
	}
	_, err = parser.Parse(data)
	if err != nil {
		return 0
	}
	return 1
}
