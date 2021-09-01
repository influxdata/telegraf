//go:build linux
// +build linux

package intel_powerstat

import (
	"math"
	"strconv"
)

const (
	microJouleToJoule    = 1.0 / 1000000
	microWattToWatt      = 1.0 / 1000000
	kiloHertzToMegaHertz = 1.0 / 1000
	nanoSecondsToSeconds = 1.0 / 1000000000
	cyclesToHertz        = 1.0 / 1000000
)

func convertMicroJoulesToJoules(mJ float64) float64 {
	return mJ * microJouleToJoule
}

func convertMicroWattToWatt(mW float64) float64 {
	return mW * microWattToWatt
}

func convertKiloHertzToMegaHertz(kiloHertz float64) float64 {
	return kiloHertz * kiloHertzToMegaHertz
}

func convertNanoSecondsToSeconds(ns int64) float64 {
	return float64(ns) * nanoSecondsToSeconds
}

func convertProcessorCyclesToHertz(pc uint64) float64 {
	return float64(pc) * cyclesToHertz
}

func roundFloatToNearestTwoDecimalPlaces(n float64) float64 {
	return math.Round(n*100) / 100
}

func convertIntegerArrayToStringArray(array []int64) []string {
	stringArray := make([]string, 0)
	for _, value := range array {
		stringArray = append(stringArray, strconv.FormatInt(value, 10))
	}

	return stringArray
}
