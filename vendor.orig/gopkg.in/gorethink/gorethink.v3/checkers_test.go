package gorethink

import (
	"encoding/json"
	"fmt"

	test "gopkg.in/check.v1"

	"gopkg.in/gorethink/gorethink.v3/types"
)

type jsonChecker struct {
	*test.CheckerInfo
}

func (j jsonChecker) Check(params []interface{}, names []string) (result bool, error string) {
	var jsonParams []interface{}
	for _, param := range params {
		jsonParam, err := json.Marshal(param)
		if err != nil {
			return false, err.Error()
		}
		jsonParams = append(jsonParams, jsonParam)
	}
	return test.DeepEquals.Check(jsonParams, names)
}

// jsonEquals compares two interface{} objects by converting them to JSON and
// seeing if the strings match
var jsonEquals = &jsonChecker{
	&test.CheckerInfo{Name: "jsonEquals", Params: []string{"obtained", "expected"}},
}

type geometryChecker struct {
	*test.CheckerInfo
}

func (j geometryChecker) Check(params []interface{}, names []string) (result bool, error string) {
	obtained, ok := params[0].(types.Geometry)
	if !ok {
		return false, "obtained must be a Geometry"
	}
	expectedType, ok := params[1].(string)
	if !ok {
		return false, "expectedType must be a string"
	}

	switch expectedType {
	case "Polygon":
		expectedCoords, ok := params[2].([][][]float64)
		if !ok {
			return false, "expectedCoords must be a [][][]float64"
		}

		return comparePolygon(expectedCoords, obtained)
	case "LineString":
		expectedCoords, ok := params[2].([][]float64)
		if !ok {
			return false, "expectedCoords must be a [][]float64"
		}

		return compareLineString(expectedCoords, obtained)
	case "Point":
		expectedCoords, ok := params[2].([]float64)
		if !ok {
			return false, "expectedCoords must be a []float64"
		}

		return comparePoint(expectedCoords, obtained)
	default:
		return false, "unknown expectedType"
	}
}

// geometryEquals compares two geometry values, all coordinates are compared with a small amount of tolerance
var geometryEquals = &geometryChecker{
	&test.CheckerInfo{Name: "geometryEquals", Params: []string{"obtained", "expectedType", "expectedCoords"}},
}

/* BEGIN FLOAT HELPERS */

// totally ripped off from math/all_test.go
// https://github.com/golang/go/blob/master/src/math/all_test.go#L1723-L1749
func tolerance(a, b, e float64) bool {
	d := a - b
	if d < 0 {
		d = -d

	}

	if a != 0 {
		e = e * a
		if e < 0 {
			e = -e

		}

	}
	return d < e
}

func mehclose(a, b float64) bool    { return tolerance(a, b, 1e-2) }
func kindaclose(a, b float64) bool  { return tolerance(a, b, 1e-8) }
func prettyclose(a, b float64) bool { return tolerance(a, b, 1e-14) }
func veryclose(a, b float64) bool   { return tolerance(a, b, 4e-16) }
func soclose(a, b, e float64) bool  { return tolerance(a, b, e) }

func comparePolygon(expected [][][]float64, obtained types.Geometry) (result bool, error string) {
	if obtained.Type != "Polygon" {
		return false, fmt.Sprintf("obtained geometry has incorrect type, has %s but expected Polygon", obtained.Type)
	}

	for i, line := range obtained.Lines {
		for j, point := range line {
			ok, err := assertPointsEqual(
				expected[i][j][0], point.Lon, // Lon
				expected[i][j][1], point.Lat, // Lat
			)
			if !ok {
				return false, err
			}
		}
	}

	return true, ""
}

func compareLineString(expected [][]float64, obtained types.Geometry) (result bool, error string) {
	if obtained.Type != "LineString" {
		return false, fmt.Sprintf("obtained geometry has incorrect type, has %s but expected LineString", obtained.Type)
	}

	for j, point := range obtained.Line {
		ok, err := assertPointsEqual(
			expected[j][0], point.Lon, // Lon
			expected[j][1], point.Lat, // Lat
		)
		if !ok {
			return false, err
		}
	}

	return true, ""
}

func comparePoint(expected []float64, obtained types.Geometry) (result bool, error string) {
	if obtained.Type != "Point" {
		return false, fmt.Sprintf("obtained geometry has incorrect type, has %s but expected Point", obtained.Type)
	}

	return assertPointsEqual(
		expected[0], obtained.Point.Lon, // Lon
		expected[1], obtained.Point.Lat, // Lat
	)
}

func assertPointsEqual(expectedLon, obtainedLon, expectedLat, obtainedLat float64) (result bool, error string) {
	if !kindaclose(expectedLon, obtainedLon) {
		return false, fmt.Sprintf("the deviation between the compared floats is too great [%v:%v]", expectedLon, obtainedLon)
	}
	if !kindaclose(expectedLat, obtainedLat) {
		return false, fmt.Sprintf("the deviation between the compared floats is too great [%v:%v]", expectedLat, obtainedLat)
	}

	return true, ""
}

/* END FLOAT HELPERS */
