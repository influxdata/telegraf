package internal

import "os"

func GetHostname(defaultVal string) string {
	hostname, err := os.Hostname()
	if err != nil {
		return defaultVal
	}
	return hostname
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
