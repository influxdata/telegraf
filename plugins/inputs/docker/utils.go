package docker

import (
	"fmt"
	"strconv"
	"strings"
)

func hostnameFromID(id string) string {
	if len(id) > 12 {
		return id[0:12]
	}
	return id
}

// Parse container name
func parseContainerName(containerNames []string) string {
	for _, name := range containerNames {
		trimmedName := strings.TrimPrefix(name, "/")
		if !strings.Contains(trimmedName, "/") {
			return trimmedName
		}
	}

	return ""
}

// Parses the human-readable size string into the amount it represents.
func parseSize(sizeStr string) (int64, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) != 4 {
		return -1, fmt.Errorf("invalid size: %s", sizeStr)
	}

	size, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return -1, err
	}

	unitPrefix := strings.ToLower(matches[3])
	if mul, ok := sizeUnitMap[unitPrefix]; ok {
		size *= float64(mul)
	}

	return int64(size), nil
}
