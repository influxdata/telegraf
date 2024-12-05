package mikrotik

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

func parse(data common) (points []parsedPoint, err error) {
	var errorList []error

	for i := range data {
		if !ignoreCommentsFunction(data[i]) {
			continue
		}
		tags := make(map[string]string)
		fields := make(map[string]interface{})
		for _, tagName := range tagFields {
			if data[i][tagName] != "" {
				tags[tagName] = data[i][tagName]
			}
		}
		for _, fieldName := range valueFields {
			if data[i][fieldName] != "" {
				if slices.Contains(durationParseFieldNames, fieldName) {
					fields[fieldName], err = parseUptimeIntoDuration(data[i][fieldName])
					if err != nil {
						errorList = append(errorList, err)
					}
				} else if strings.Contains(data[i][fieldName], ",") {
					rxTxValues := strings.Split(data[i][fieldName], ",")
					fields[fieldName+"_tx"], err = strconv.ParseInt(rxTxValues[0], 10, 64)
					if err != nil {
						errorList = append(errorList, err)
					}
					fields[fieldName+"_rx"], err = strconv.ParseInt(rxTxValues[1], 10, 64)
					if err != nil {
						errorList = append(errorList, err)
					}
				} else {
					fields[fieldName], err = strconv.ParseInt(data[i][fieldName], 10, 64)
					if err != nil {
						errorList = append(errorList, err)
					}
				}
			}
		}

		points = append(points, parsedPoint{Tags: tags, Fields: fields})
	}

	return points, errors.Join(errorList...)
}

func parseUptimeIntoDuration(uptime string) (int64, error) {
	re := regexp.MustCompile(`^(?:(\d+)w)?(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?$`)
	matches := re.FindStringSubmatch(uptime)
	if matches == nil {
		return 0, nil
	}
	var duration time.Duration
	if matches[1] != "" {
		weeks, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parseUptimeIntoDuration -> %w", err)
		}
		duration += time.Duration(weeks*7*24) * time.Hour
	}
	if matches[2] != "" {
		days, err := strconv.ParseInt(matches[2], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parseUptimeIntoDuration -> %w", err)
		}
		duration += time.Duration(days*24) * time.Hour
	}
	if matches[3] != "" {
		hours, err := strconv.ParseInt(matches[3], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parseUptimeIntoDuration -> %w", err)
		}
		duration += time.Duration(hours) * time.Hour
	}
	if matches[4] != "" {
		minutes, err := strconv.ParseInt(matches[4], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parseUptimeIntoDuration -> %w", err)
		}
		duration += time.Duration(minutes) * time.Minute
	}
	if matches[5] != "" {
		seconds, err := strconv.ParseInt(matches[5], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parseUptimeIntoDuration -> %w", err)
		}
		duration += time.Duration(seconds) * time.Second
	}

	return int64(duration / time.Second), nil
}
