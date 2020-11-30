// +build linux

package services

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	microJouleToJoule    = 1.0 / 1000000
	microWattToWatt      = 1.0 / 1000000
	kiloHertzToMegaHertz = 1.0 / 1000
	nanoSecondsToSeconds = 1.0 / 1000000000
	cyclesToHertz        = 1.0 / 1000000
)

// ConvertMicroJoulesToJoules converts MicroJoules to Joules.
func ConvertMicroJoulesToJoules(mJ float64) float64 {
	return mJ * microJouleToJoule
}

// ConvertMicroWattToWatt converts MicroWatt to Watt.
func ConvertMicroWattToWatt(mW float64) float64 {
	return mW * microWattToWatt
}

// ConvertKiloHertzToMegaHertz converts KiloHertz to MegaHertz.
func ConvertKiloHertzToMegaHertz(kHz float64) float64 {
	return kHz * kiloHertzToMegaHertz
}

// ConvertNanoSecondsToSeconds converts NanoSeconds to Seconds.
func ConvertNanoSecondsToSeconds(ns int64) float64 {
	return float64(ns) * nanoSecondsToSeconds
}

// ConvertProcessorCyclesToHertz converts processor cycles to Hz.
func ConvertProcessorCyclesToHertz(pc uint64) float64 {
	return float64(pc) * cyclesToHertz
}

// RoundFloatToNearestTwoDecimalPlaces returns the nearest float.
func RoundFloatToNearestTwoDecimalPlaces(n float64) float64 {
	return math.Round(n*100) / 100
}

// ConvertHexArrayToIntegerArray converts hex array of strings to int64 array.
func ConvertHexArrayToIntegerArray(hexArray []string) ([]int64, error) {
	convertedHex := make([]int64, 0)

	for _, hex := range hexArray {
		parsedHex, err := strconv.ParseInt(strings.Replace(hex, "0x", "", -1), 16, 64)
		if err != nil {
			return convertedHex, fmt.Errorf("error while parsing hex %s to int, err: %v", hex, err)
		}

		convertedHex = append(convertedHex, parsedHex)
	}
	return convertedHex, nil
}

// ConvertIntegerArrayToStringArray converts int64 array to string array.
func ConvertIntegerArrayToStringArray(array []int64) []string {
	stringArray := make([]string, 0)
	for _, value := range array {
		stringArray = append(stringArray, strconv.FormatInt(value, 10))
	}

	return stringArray
}
