package reql_tests

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/stretchr/testify/suite"
	r "gopkg.in/gorethink/gorethink.v3"
	"gopkg.in/gorethink/gorethink.v3/internal/compare"
)

func maybeRun(query interface{}, session *r.Session, opts r.RunOpts) interface{} {
	switch query := query.(type) {
	case r.Term:
		cursor, err := query.Run(session, opts)
		if err != nil {
			return err
		}

		switch cursor.Type() {
		case "Cursor":
			results, err := cursor.Interface()
			if err != nil {
				return err
			}

			return results
		default:
			// If this is a changefeed then return the cursor without attempting
			// to read any documents
			return cursor
		}
	default:
		return query
	}
}

func runAndAssert(suite suite.Suite, expected, v interface{}, session *r.Session, opts r.RunOpts) {
	var cursor *r.Cursor
	var err error

	switch v := v.(type) {
	case r.Term:
		cursor, err = v.Run(session, opts)
	case *r.Cursor:
		cursor = v
	case error:
		err = v
	}

	assertExpected(suite, expected, cursor, err)
}

func fetchAndAssert(suite suite.Suite, expected, result interface{}, count int) {
	switch v := expected.(type) {
	case Expected:
		v.Fetch = true
		v.FetchCount = count

		expected = v
	default:
		expected = Expected(compare.Expected{
			Val:        v,
			Fetch:      true,
			FetchCount: count,
		})
	}

	var cursor *r.Cursor
	var err error

	switch result := result.(type) {
	case *r.Cursor:
		cursor = result
	case error:
		err = result
	}

	assertExpected(suite, expected, cursor, err)
}

func maybeLen(v interface{}) interface{} {
	switch v := v.(type) {
	case *r.Cursor:
		results := []interface{}{}
		v.All(&results)
		return len(results)
	case []interface{}:
		return len(v)
	default:
		return v
	}
}

func assertExpected(suite suite.Suite, expected interface{}, obtainedCursor *r.Cursor, obtainedErr error) {
	if expected == compare.AnythingIsFine {
		suite.NoError(obtainedErr, "Query returned unexpected error")
		return
	}

	switch expected := expected.(type) {
	case Err:
		expected.assert(suite, obtainedCursor, obtainedErr)
	case Expected:
		expected.assert(suite, obtainedCursor, obtainedErr)
	default:
		Expected(compare.Expected{Val: expected}).assert(suite, obtainedCursor, obtainedErr)
	}
}

type Expected compare.Expected

func (expected Expected) assert(suite suite.Suite, obtainedCursor *r.Cursor, obtainedErr error) {
	if suite.NoError(obtainedErr, "Query returned unexpected error") {
		return
	}

	expectedVal := reflect.ValueOf(expected.Val)

	// If expected value is nil then ensure cursor is nil (assume that an
	// invalid reflect value is because expected value is nil)
	if !expectedVal.IsValid() || (expectedVal.Kind() == reflect.Ptr && expectedVal.IsNil()) {
		suite.True(obtainedCursor.IsNil(), "Expected nil cursor")
		return
	}

	expectedType := expectedVal.Type()
	expectedKind := expectedType.Kind()

	if expectedKind == reflect.Array || expectedKind == reflect.Slice || expected.Fetch {
		if expectedType.Elem().Kind() == reflect.Uint8 {
			// Decode byte slices slightly differently
			var obtained = []byte{}
			err := obtainedCursor.One(&obtained)
			suite.NoError(err, "Error returned when reading query response")
			compare.Assert(suite.T(), expected, obtained)
		} else {
			var obtained = []interface{}{}
			if expected.Fetch {
				var v interface{}
				for obtainedCursor.Next(&v) {
					obtained = append(obtained, v)

					if expected.FetchCount != 0 && len(obtained) >= expected.FetchCount {
						break
					}
				}
				suite.NoError(obtainedCursor.Err(), "Error returned when reading query response")
			} else {
				err := obtainedCursor.All(&obtained)
				suite.NoError(err, "Error returned when reading query response")
			}

			compare.Assert(suite.T(), expected, obtained)
		}
	} else if expectedKind == reflect.Map {
		var obtained map[string]interface{}
		err := obtainedCursor.One(&obtained)
		suite.NoError(err, "Error returned when reading query response")
		compare.Assert(suite.T(), expected, obtained)
	} else {
		var obtained interface{}
		err := obtainedCursor.One(&obtained)
		suite.NoError(err, "Error returned when reading query response")
		compare.Assert(suite.T(), expected, obtained)
	}
}

func int_cmp(i int) int {
	return i
}

func float_cmp(i float64) float64 {
	return i
}

func arrlen(length int, vs ...interface{}) []interface{} {
	var v interface{} = compare.AnythingIsFine
	if len(vs) == 1 {
		v = vs[0]
	}

	arr := make([]interface{}, length)
	for i := 0; i < length; i++ {
		arr[i] = v
	}
	return arr
}

func str(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func wait(s int) interface{} {
	time.Sleep(time.Duration(s) * time.Second)

	return nil
}

type Err struct {
	Type    string
	Message string
	Regex   string
}

var exceptionRegex = regexp.MustCompile("^(?P<message>[^\n]*?)((?: in:)?\n|\nFailed assertion:)(?s).*$")

func (expected Err) assert(suite suite.Suite, obtainerCursor *r.Cursor, obtainedErr error) {
	// If the error is nil then attempt to read from the cursor and see if an
	// error is returned
	if obtainedErr == nil {
		var res []interface{}
		obtainedErr = obtainerCursor.All(&res)
	}

	if suite.Error(obtainedErr) {
		return
	}

	obtainedType := reflect.TypeOf(obtainedErr).String()
	obtainedMessage := strings.TrimPrefix(obtainedErr.Error(), "gorethink: ")
	obtainedMessage = exceptionRegex.ReplaceAllString(obtainedMessage, "${message}")

	suite.Equal(expected.Type, obtainedType)
	if expected.Regex != "" {
		suite.Regexp(expected.Regex, obtainedMessage)
	}
	if expected.Message != "" {
		suite.Equal(expected.Message, obtainedMessage)
	}
}

func err(errType, message string) Err {
	return Err{
		Type:    "gorethink.RQL" + errType[4:],
		Message: message,
	}
}

func err_regex(errType, expr string) Err {
	return Err{
		Type:  "gorethink.RQL" + errType[4:],
		Regex: expr,
	}
}

var Ast = struct {
	RqlTzinfo     func(tz string) *time.Location
	Fromtimestamp func(ts float64, loc *time.Location) time.Time
	Now           func() time.Time
}{
	func(tz string) *time.Location {
		t, _ := time.Parse("-07:00 UTC", tz+" UTC")

		return t.Location()
	},
	func(ts float64, loc *time.Location) time.Time {
		sec, nsec := math.Modf(ts)

		return time.Unix(int64(sec), int64(nsec*1000)*1000000).In(loc)
	},
	time.Now,
}

func UTCTimeZone() *time.Location {
	return time.UTC
}

func PacificTimeZone() *time.Location {
	return Ast.RqlTzinfo("-07:00")
}

var FloatInfo = struct {
	Min, Max float64
}{math.SmallestNonzeroFloat64, math.MaxFloat64}

var sys = struct {
	FloatInfo struct {
		Min, Max float64
	}
}{FloatInfo}
