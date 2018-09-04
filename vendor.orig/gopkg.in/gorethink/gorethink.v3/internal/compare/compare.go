package compare

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var AnythingIsFine = "reql_test.AnythingIsFine"

func Assert(t *testing.T, expected, actual interface{}) {
	expectedVal := expected
	if e, ok := expected.(Expected); ok {
		expectedVal = e.Val
	}

	ok, msg := Compare(expected, actual)
	if !ok {
		assert.Fail(t, fmt.Sprintf("Not equal: %#v (expected)\n           != %#v (actual)", expectedVal, actual), msg)
	}
}

func AssertFalse(t *testing.T, expected, actual interface{}) {
	expectedVal := expected
	if e, ok := expected.(Expected); ok {
		expectedVal = e.Val
	}

	ok, msg := Compare(expected, actual)
	if ok {
		assert.Fail(t, fmt.Sprintf("Should not be equal: %#v (expected)\n           == %#v (actual)", expectedVal, actual), msg)
	}
}

func AssertPrecision(t *testing.T, expected, actual interface{}, precision float64) {
	expectedVal := expected
	if e, ok := expected.(Expected); ok {
		expectedVal = e.Val
	}

	ok, msg := ComparePrecision(expected, actual, precision)
	if !ok {
		assert.Fail(t, fmt.Sprintf("Not equal: %#v (expected)\n           != %#v (actual)", expectedVal, actual), msg)
	}
}

func AssertPrecisionFalse(t *testing.T, expected, actual interface{}, precision float64) {
	expectedVal := expected
	if e, ok := expected.(Expected); ok {
		expectedVal = e.Val
	}

	ok, msg := ComparePrecision(expected, actual, precision)
	if ok {
		assert.Fail(t, fmt.Sprintf("Should not be equal: %#v (expected)\n           == %#v (actual)", expectedVal, actual), msg)
	}
}

func Compare(expected, actual interface{}) (bool, string) {
	return ComparePrecision(expected, actual, 0.00000000001)
}

func ComparePrecision(expected, actual interface{}, precision float64) (bool, string) {
	return compare(expected, actual, true, false, precision)
}

func compare(expected, actual interface{}, ordered, partial bool, precision float64) (bool, string) {
	if e, ok := expected.(Expected); ok {
		partial = e.Partial
		ordered = e.Ordered
		expected = e.Val
	}

	// Anything
	if expected == AnythingIsFine {
		return true, ""
	}

	expectedVal := reflect.ValueOf(expected)
	actualVal := reflect.ValueOf(actual)

	// Nil
	if expected == nil {
		switch actualVal.Kind() {
		case reflect.Bool:
			expected = false
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			expected = 0.0
		case reflect.String:
			expected = ""
		}

		if expected == actual {
			return true, ""
		}
	}

	// Regex
	if expr, ok := expected.(Regex); ok {
		re, err := regexp.Compile(string(expr))
		if err != nil {
			return false, fmt.Sprintf("Failed to compile regexp: %s", err)
		}

		if actualVal.Kind() != reflect.String {
			return false, fmt.Sprintf("Expected string, got %t (%T)", actual, actual)
		}

		if !re.MatchString(actualVal.String()) {
			return false, fmt.Sprintf("Value %v did not match regexp '%s'", actual, expr)
		}

		return true, ""
	}

	switch expectedVal.Kind() {

	// Bool
	case reflect.Bool:
		if expected == actual {
			return true, ""
		}
	// Number
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		switch actualVal.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String:
			diff := math.Abs(reflectNumber(expectedVal) - reflectNumber(actualVal))
			if diff <= precision {
				return true, ""
			}

			if precision != 0 {
				return false, fmt.Sprintf("Value %v was not within %f of %v", expected, precision, actual)
			}

			return false, fmt.Sprintf("Expected %v but got %v", expected, actual)
		}

	// String
	case reflect.String:
		actualStr := fmt.Sprintf("%v", actual)
		if expected == actualStr {
			return true, ""
		}
	// Struct
	case reflect.Struct:
		// Convert expected struct to map and compare with actual value
		return compare(reflectMap(expectedVal), actual, ordered, partial, precision)
	// Map
	case reflect.Map:
		switch actualVal.Kind() {
		case reflect.Struct:
			// Convert actual struct to map and compare with expected map
			return compare(expected, reflectMap(actualVal), ordered, partial, precision)
		case reflect.Map:
			expectedKeys := expectedVal.MapKeys()
			actualKeys := actualVal.MapKeys()

			for _, expectedKey := range expectedKeys {
				keyFound := false
				for _, actualKey := range actualKeys {
					if ok, _ := Compare(expectedKey.Interface(), actualKey.Interface()); ok {
						keyFound = true
						break
					}
				}
				if !keyFound {
					return false, fmt.Sprintf("Expected field %v but not found", expectedKey)
				}
			}

			if !partial {
				expectedKeyVals := reflectMapKeys(expectedKeys)
				actualKeyVals := reflectMapKeys(actualKeys)
				if ok, _ := compare(expectedKeyVals, actualKeyVals, false, false, 0.0); !ok {
					return false, fmt.Sprintf(
						"Unmatched keys from either side: expected fields %v, got %v",
						expectedKeyVals, actualKeyVals,
					)
				}
			}

			expectedMap := reflectMap(expectedVal)
			actualMap := reflectMap(actualVal)

			for k, v := range expectedMap {
				if ok, reason := compare(v, actualMap[k], ordered, partial, precision); !ok {
					return false, reason
				}
			}

			return true, ""
		default:
			return false, fmt.Sprintf("Expected map, got %v (%T)", actual, actual)
		}
	// Slice/Array
	case reflect.Slice, reflect.Array:
		switch actualVal.Kind() {
		case reflect.Slice, reflect.Array:
			if ordered {
				expectedArr := reflectSlice(expectedVal)
				actualArr := reflectSlice(actualVal)

				j := 0
				for i := 0; i < len(expectedArr); i++ {
					expectedArrVal := expectedArr[i]
					for {
						if j >= len(actualArr) {
							return false, fmt.Sprintf("Ran out of results before finding %v", expectedArrVal)
						}

						actualArrVal := actualArr[j]
						j++

						if ok, _ := compare(expectedArrVal, actualArrVal, ordered, partial, precision); ok {
							break
						} else if !partial {
							return false, fmt.Sprintf("Unexpected item %v while looking for %v", actualArrVal, expectedArrVal)
						}
					}
				}
				if !partial && j < len(actualArr) {
					return false, fmt.Sprintf("Unexpected extra results: %v", actualArr[j:])
				}
			} else {
				expectedArr := reflectSlice(expectedVal)
				actualArr := reflectSlice(actualVal)

				for _, expectedArrVal := range expectedArr {
					found := false
					for j, actualArrVal := range actualArr {
						if ok, _ := compare(expectedArrVal, actualArrVal, ordered, partial, precision); ok {
							found = true
							actualArr = append(actualArr[:j], actualArr[j+1:]...)
							break
						}
					}
					if !found {
						return false, fmt.Sprintf("Missing expected item %v", expectedArrVal)
					}
				}

				if !partial && len(actualArr) > 0 {
					return false, fmt.Sprintf("Extra items returned: %v", expectedArr)
				}
			}

			return true, ""
		}
	// Other
	default:
		if expected == actual {
			return true, ""
		}
	}

	return false, fmt.Sprintf("Expected %v (%T) but got %v (%T)", expected, expected, actual, actual)
}

func reflectNumber(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.String:
		f, _ := strconv.ParseFloat(v.String(), 64)
		return f
	default:
		return float64(0)
	}
}

func reflectMap(v reflect.Value) map[interface{}]interface{} {
	switch v.Kind() {
	case reflect.Struct:
		m := map[interface{}]interface{}{}
		for i := 0; i < v.NumField(); i++ {
			sf := v.Type().Field(i)
			if sf.PkgPath != "" && !sf.Anonymous {
				continue // unexported
			}

			k := sf.Name
			v := v.Field(i).Interface()

			m[k] = v
		}
		return m
	case reflect.Map:
		m := map[interface{}]interface{}{}
		for _, mk := range v.MapKeys() {
			k := ""
			if mk.Interface() != nil {
				k = fmt.Sprintf("%v", mk.Interface())
			}
			v := v.MapIndex(mk).Interface()

			m[k] = v
		}
		return m
	default:
		return nil
	}
}

func reflectSlice(v reflect.Value) []interface{} {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		s := []interface{}{}
		for i := 0; i < v.Len(); i++ {
			s = append(s, v.Index(i).Interface())
		}
		return s
	default:
		return nil
	}
}

func reflectMapKeys(keys []reflect.Value) []interface{} {
	s := []interface{}{}
	for _, key := range keys {
		s = append(s, key.Interface())
	}
	return s
}

func reflectInterfaces(vals []reflect.Value) []interface{} {
	ret := []interface{}{}
	for _, val := range vals {
		ret = append(ret, val.Interface())
	}
	return ret
}
