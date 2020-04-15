package ts3

import (
	"strings"
)

// ScanLines is a split function for a bytes.Scanner that returns each line of
// text, stripped of any trailing end-of-line marker. The returned line may
// be empty. The end-of-line marker is one newline followed by a carriage return.
// In regular expression notation, it is `\n\r`.
// The last non-empty line of input will be returned even if it has no newline.
func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := strings.Index(string(data), "\n\r"); i >= 0 {
		// We have a full end-of-line terminated line.
		return i + 2, data[0:i], nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}
