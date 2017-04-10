package metric

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
)

var (
	ErrInvalidNumber = errors.New("invalid number")
)

const (
	// the number of characters for the largest possible int64 (9223372036854775807)
	maxInt64Digits = 19

	// the number of characters for the smallest possible int64 (-9223372036854775808)
	minInt64Digits = 20

	// the number of characters required for the largest float64 before a range check
	// would occur during parsing
	maxFloat64Digits = 25

	// the number of characters required for smallest float64 before a range check occur
	// would occur during parsing
	minFloat64Digits = 27

	MaxKeyLength = 65535
)

// The following constants allow us to specify which state to move to
// next, when scanning sections of a Point.
const (
	tagKeyState = iota
	tagValueState
	fieldsState
)

func Parse(buf []byte) ([]telegraf.Metric, error) {
	return ParseWithDefaultTimePrecision(buf, time.Now(), "")
}

func ParseWithDefaultTime(buf []byte, t time.Time) ([]telegraf.Metric, error) {
	return ParseWithDefaultTimePrecision(buf, t, "")
}

func ParseWithDefaultTimePrecision(
	buf []byte,
	t time.Time,
	precision string,
) ([]telegraf.Metric, error) {
	if len(buf) == 0 {
		return []telegraf.Metric{}, nil
	}
	if len(buf) <= 6 {
		return []telegraf.Metric{}, makeError("buffer too short", buf, 0)
	}
	metrics := make([]telegraf.Metric, 0, bytes.Count(buf, []byte("\n"))+1)
	var errStr string
	i := 0
	for {
		j := bytes.IndexByte(buf[i:], '\n')
		if j == -1 {
			break
		}
		if len(buf[i:i+j]) < 2 {
			i += j + 1 // increment i past the previous newline
			continue
		}

		m, err := parseMetric(buf[i:i+j], t, precision)
		if err != nil {
			i += j + 1 // increment i past the previous newline
			errStr += " " + err.Error()
			continue
		}
		i += j + 1 // increment i past the previous newline

		metrics = append(metrics, m)
	}

	if len(errStr) > 0 {
		return metrics, fmt.Errorf(errStr)
	}
	return metrics, nil
}

func parseMetric(buf []byte,
	defaultTime time.Time,
	precision string,
) (telegraf.Metric, error) {
	var dTime string
	// scan the first block which is measurement[,tag1=value1,tag2=value=2...]
	pos, key, err := scanKey(buf, 0)
	if err != nil {
		return nil, err
	}

	// measurement name is required
	if len(key) == 0 {
		return nil, fmt.Errorf("missing measurement")
	}

	if len(key) > MaxKeyLength {
		return nil, fmt.Errorf("max key length exceeded: %v > %v", len(key), MaxKeyLength)
	}

	// scan the second block is which is field1=value1[,field2=value2,...]
	pos, fields, err := scanFields(buf, pos)
	if err != nil {
		return nil, err
	}

	// at least one field is required
	if len(fields) == 0 {
		return nil, fmt.Errorf("missing fields")
	}

	// scan the last block which is an optional integer timestamp
	pos, ts, err := scanTime(buf, pos)
	if err != nil {
		return nil, err
	}

	// apply precision multiplier
	var nsec int64
	multiplier := getPrecisionMultiplier(precision)
	if multiplier > 1 {
		tsint, err := parseIntBytes(ts, 10, 64)
		if err != nil {
			return nil, err
		}

		nsec := multiplier * tsint
		ts = []byte(strconv.FormatInt(nsec, 10))
	}

	m := &metric{
		fields: fields,
		t:      ts,
		nsec:   nsec,
	}

	// parse out the measurement name
	// namei is the index at which the "name" ends
	namei := indexUnescapedByte(key, ',')
	if namei < 1 {
		// no tags
		m.name = key
	} else {
		m.name = key[0:namei]
		m.tags = key[namei:]
	}

	if len(m.t) == 0 {
		if len(dTime) == 0 {
			dTime = fmt.Sprint(defaultTime.UnixNano())
		}
		// use default time
		m.t = []byte(dTime)
	}

	// here we copy on return because this allows us to later call
	// AddTag, AddField, RemoveTag, RemoveField, etc. without worrying about
	// modifying 'tag' bytes having an affect on 'field' bytes, for example.
	return m.Copy(), nil
}

// scanKey scans buf starting at i for the measurement and tag portion of the point.
// It returns the ending position and the byte slice of key within buf.  If there
// are tags, they will be sorted if they are not already.
func scanKey(buf []byte, i int) (int, []byte, error) {
	start := skipWhitespace(buf, i)
	i = start

	// First scan the Point's measurement.
	state, i, err := scanMeasurement(buf, i)
	if err != nil {
		return i, buf[start:i], err
	}

	// Optionally scan tags if needed.
	if state == tagKeyState {
		i, err = scanTags(buf, i)
		if err != nil {
			return i, buf[start:i], err
		}
	}

	return i, buf[start:i], nil
}

// scanMeasurement examines the measurement part of a Point, returning
// the next state to move to, and the current location in the buffer.
func scanMeasurement(buf []byte, i int) (int, int, error) {
	// Check first byte of measurement, anything except a comma is fine.
	// It can't be a space, since whitespace is stripped prior to this
	// function call.
	if i >= len(buf) || buf[i] == ',' {
		return -1, i, makeError("missing measurement", buf, i)
	}

	for {
		i++
		if i >= len(buf) {
			// cpu
			return -1, i, makeError("missing fields", buf, i)
		}

		if buf[i-1] == '\\' {
			// Skip character (it's escaped).
			continue
		}

		// Unescaped comma; move onto scanning the tags.
		if buf[i] == ',' {
			return tagKeyState, i + 1, nil
		}

		// Unescaped space; move onto scanning the fields.
		if buf[i] == ' ' {
			// cpu value=1.0
			return fieldsState, i, nil
		}
	}
}

// scanTags examines all the tags in a Point, keeping track of and
// returning the updated indices slice, number of commas and location
// in buf where to start examining the Point fields.
func scanTags(buf []byte, i int) (int, error) {
	var (
		err   error
		state = tagKeyState
	)

	for {
		switch state {
		case tagKeyState:
			i, err = scanTagsKey(buf, i)
			state = tagValueState // tag value always follows a tag key
		case tagValueState:
			state, i, err = scanTagsValue(buf, i)
		case fieldsState:
			return i, nil
		}

		if err != nil {
			return i, err
		}
	}
}

// scanTagsKey scans each character in a tag key.
func scanTagsKey(buf []byte, i int) (int, error) {
	// First character of the key.
	if i >= len(buf) || buf[i] == ' ' || buf[i] == ',' || buf[i] == '=' {
		// cpu,{'', ' ', ',', '='}
		return i, makeError("missing tag key", buf, i)
	}

	// Examine each character in the tag key until we hit an unescaped
	// equals (the tag value), or we hit an error (i.e., unescaped
	// space or comma).
	for {
		i++

		// Either we reached the end of the buffer or we hit an
		// unescaped comma or space.
		if i >= len(buf) ||
			((buf[i] == ' ' || buf[i] == ',') && buf[i-1] != '\\') {
			// cpu,tag{'', ' ', ','}
			return i, makeError("missing tag value", buf, i)
		}

		if buf[i] == '=' && buf[i-1] != '\\' {
			// cpu,tag=
			return i + 1, nil
		}
	}
}

// scanTagsValue scans each character in a tag value.
func scanTagsValue(buf []byte, i int) (int, int, error) {
	// Tag value cannot be empty.
	if i >= len(buf) || buf[i] == ',' || buf[i] == ' ' {
		// cpu,tag={',', ' '}
		return -1, i, makeError("missing tag value", buf, i)
	}

	// Examine each character in the tag value until we hit an unescaped
	// comma (move onto next tag key), an unescaped space (move onto
	// fields), or we error out.
	for {
		i++
		if i >= len(buf) {
			// cpu,tag=value
			return -1, i, makeError("missing fields", buf, i)
		}

		// An unescaped equals sign is an invalid tag value.
		if buf[i] == '=' && buf[i-1] != '\\' {
			// cpu,tag={'=', 'fo=o'}
			return -1, i, makeError("invalid tag format", buf, i)
		}

		if buf[i] == ',' && buf[i-1] != '\\' {
			// cpu,tag=foo,
			return tagKeyState, i + 1, nil
		}

		// cpu,tag=foo value=1.0
		// cpu, tag=foo\= value=1.0
		if buf[i] == ' ' && buf[i-1] != '\\' {
			return fieldsState, i, nil
		}
	}
}

// scanFields scans buf, starting at i for the fields section of a point.  It returns
// the ending position and the byte slice of the fields within buf
func scanFields(buf []byte, i int) (int, []byte, error) {
	start := skipWhitespace(buf, i)
	i = start
	quoted := false

	// tracks how many '=' we've seen
	equals := 0

	// tracks how many commas we've seen
	commas := 0

	for {
		// reached the end of buf?
		if i >= len(buf) {
			break
		}

		// escaped characters?
		if buf[i] == '\\' && i+1 < len(buf) {
			i += 2
			continue
		}

		// If the value is quoted, scan until we get to the end quote
		// Only quote values in the field value since quotes are not significant
		// in the field key
		if buf[i] == '"' && equals > commas {
			quoted = !quoted
			i++
			continue
		}

		// If we see an =, ensure that there is at least on char before and after it
		if buf[i] == '=' && !quoted {
			equals++

			// check for "... =123" but allow "a\ =123"
			if buf[i-1] == ' ' && buf[i-2] != '\\' {
				return i, buf[start:i], makeError("missing field key", buf, i)
			}

			// check for "...a=123,=456" but allow "a=123,a\,=456"
			if buf[i-1] == ',' && buf[i-2] != '\\' {
				return i, buf[start:i], makeError("missing field key", buf, i)
			}

			// check for "... value="
			if i+1 >= len(buf) {
				return i, buf[start:i], makeError("missing field value", buf, i)
			}

			// check for "... value=,value2=..."
			if buf[i+1] == ',' || buf[i+1] == ' ' {
				return i, buf[start:i], makeError("missing field value", buf, i)
			}

			if isNumeric(buf[i+1]) || buf[i+1] == '-' || buf[i+1] == 'N' || buf[i+1] == 'n' {
				var err error
				i, err = scanNumber(buf, i+1)
				if err != nil {
					return i, buf[start:i], err
				}
				continue
			}
			// If next byte is not a double-quote, the value must be a boolean
			if buf[i+1] != '"' {
				var err error
				i, _, err = scanBoolean(buf, i+1)
				if err != nil {
					return i, buf[start:i], err
				}
				continue
			}
		}

		if buf[i] == ',' && !quoted {
			commas++
		}

		// reached end of block?
		if buf[i] == ' ' && !quoted {
			break
		}
		i++
	}

	if quoted {
		return i, buf[start:i], makeError("unbalanced quotes", buf, i)
	}

	// check that all field sections had key and values (e.g. prevent "a=1,b"
	if equals == 0 || commas != equals-1 {
		return i, buf[start:i], makeError("invalid field format", buf, i)
	}

	return i, buf[start:i], nil
}

// scanTime scans buf, starting at i for the time section of a point. It
// returns the ending position and the byte slice of the timestamp within buf
// and and error if the timestamp is not in the correct numeric format.
func scanTime(buf []byte, i int) (int, []byte, error) {
	start := skipWhitespace(buf, i)
	i = start

	for {
		// reached the end of buf?
		if i >= len(buf) {
			break
		}

		// Reached end of block or trailing whitespace?
		if buf[i] == '\n' || buf[i] == ' ' {
			break
		}

		// Handle negative timestamps
		if i == start && buf[i] == '-' {
			i++
			continue
		}

		// Timestamps should be integers, make sure they are so we don't need
		// to actually  parse the timestamp until needed.
		if buf[i] < '0' || buf[i] > '9' {
			return i, buf[start:i], makeError("invalid timestamp", buf, i)
		}
		i++
	}
	return i, buf[start:i], nil
}

func isNumeric(b byte) bool {
	return (b >= '0' && b <= '9') || b == '.'
}

// scanNumber returns the end position within buf, start at i after
// scanning over buf for an integer, or float.  It returns an
// error if a invalid number is scanned.
func scanNumber(buf []byte, i int) (int, error) {
	start := i
	var isInt bool

	// Is negative number?
	if i < len(buf) && buf[i] == '-' {
		i++
		// There must be more characters now, as just '-' is illegal.
		if i == len(buf) {
			return i, ErrInvalidNumber
		}
	}

	// how many decimal points we've see
	decimal := false

	// indicates the number is float in scientific notation
	scientific := false

	for {
		if i >= len(buf) {
			break
		}

		if buf[i] == ',' || buf[i] == ' ' {
			break
		}

		if buf[i] == 'i' && i > start && !isInt {
			isInt = true
			i++
			continue
		}

		if buf[i] == '.' {
			// Can't have more than 1 decimal (e.g. 1.1.1 should fail)
			if decimal {
				return i, ErrInvalidNumber
			}
			decimal = true
		}

		// `e` is valid for floats but not as the first char
		if i > start && (buf[i] == 'e' || buf[i] == 'E') {
			scientific = true
			i++
			continue
		}

		// + and - are only valid at this point if they follow an e (scientific notation)
		if (buf[i] == '+' || buf[i] == '-') && (buf[i-1] == 'e' || buf[i-1] == 'E') {
			i++
			continue
		}

		// NaN is an unsupported value
		if i+2 < len(buf) && (buf[i] == 'N' || buf[i] == 'n') {
			return i, ErrInvalidNumber
		}

		if !isNumeric(buf[i]) {
			return i, ErrInvalidNumber
		}
		i++
	}

	if isInt && (decimal || scientific) {
		return i, ErrInvalidNumber
	}

	numericDigits := i - start
	if isInt {
		numericDigits--
	}
	if decimal {
		numericDigits--
	}
	if buf[start] == '-' {
		numericDigits--
	}

	if numericDigits == 0 {
		return i, ErrInvalidNumber
	}

	// It's more common that numbers will be within min/max range for their type but we need to prevent
	// out or range numbers from being parsed successfully.  This uses some simple heuristics to decide
	// if we should parse the number to the actual type.  It does not do it all the time because it incurs
	// extra allocations and we end up converting the type again when writing points to disk.
	if isInt {
		// Make sure the last char is an 'i' for integers (e.g. 9i10 is not valid)
		if buf[i-1] != 'i' {
			return i, ErrInvalidNumber
		}
		// Parse the int to check bounds the number of digits could be larger than the max range
		// We subtract 1 from the index to remove the `i` from our tests
		if len(buf[start:i-1]) >= maxInt64Digits || len(buf[start:i-1]) >= minInt64Digits {
			if _, err := parseIntBytes(buf[start:i-1], 10, 64); err != nil {
				return i, makeError(fmt.Sprintf("unable to parse integer %s: %s", buf[start:i-1], err), buf, i)
			}
		}
	} else {
		// Parse the float to check bounds if it's scientific or the number of digits could be larger than the max range
		if scientific || len(buf[start:i]) >= maxFloat64Digits || len(buf[start:i]) >= minFloat64Digits {
			if _, err := parseFloatBytes(buf[start:i], 10); err != nil {
				return i, makeError("invalid float", buf, i)
			}
		}
	}

	return i, nil
}

// scanBoolean returns the end position within buf, start at i after
// scanning over buf for boolean. Valid values for a boolean are
// t, T, true, TRUE, f, F, false, FALSE. It returns an error if a invalid boolean
// is scanned.
func scanBoolean(buf []byte, i int) (int, []byte, error) {
	start := i

	if i < len(buf) && (buf[i] != 't' && buf[i] != 'f' && buf[i] != 'T' && buf[i] != 'F') {
		return i, buf[start:i], makeError("invalid value", buf, i)
	}

	i++
	for {
		if i >= len(buf) {
			break
		}

		if buf[i] == ',' || buf[i] == ' ' {
			break
		}
		i++
	}

	// Single char bool (t, T, f, F) is ok
	if i-start == 1 {
		return i, buf[start:i], nil
	}

	// length must be 4 for true or TRUE
	if (buf[start] == 't' || buf[start] == 'T') && i-start != 4 {
		return i, buf[start:i], makeError("invalid boolean", buf, i)
	}

	// length must be 5 for false or FALSE
	if (buf[start] == 'f' || buf[start] == 'F') && i-start != 5 {
		return i, buf[start:i], makeError("invalid boolean", buf, i)
	}

	// Otherwise
	valid := false
	switch buf[start] {
	case 't':
		valid = bytes.Equal(buf[start:i], []byte("true"))
	case 'f':
		valid = bytes.Equal(buf[start:i], []byte("false"))
	case 'T':
		valid = bytes.Equal(buf[start:i], []byte("TRUE")) || bytes.Equal(buf[start:i], []byte("True"))
	case 'F':
		valid = bytes.Equal(buf[start:i], []byte("FALSE")) || bytes.Equal(buf[start:i], []byte("False"))
	}

	if !valid {
		return i, buf[start:i], makeError("invalid boolean", buf, i)
	}

	return i, buf[start:i], nil

}

// skipWhitespace returns the end position within buf, starting at i after
// scanning over spaces in tags
func skipWhitespace(buf []byte, i int) int {
	for i < len(buf) {
		if buf[i] != ' ' && buf[i] != '\t' && buf[i] != 0 {
			break
		}
		i++
	}
	return i
}

// makeError is a helper function for making a metric parsing error.
//   reason is the reason that the error occured.
//   buf should be the current buffer we are parsing.
//   i is the current index, to give some context on where in the buffer we are.
func makeError(reason string, buf []byte, i int) error {
	return fmt.Errorf("metric parsing error, reason: [%s], buffer: [%s], index: [%d]",
		reason, buf, i)
}

// getPrecisionMultiplier will return a multiplier for the precision specified.
func getPrecisionMultiplier(precision string) int64 {
	d := time.Nanosecond
	switch precision {
	case "u":
		d = time.Microsecond
	case "ms":
		d = time.Millisecond
	case "s":
		d = time.Second
	case "m":
		d = time.Minute
	case "h":
		d = time.Hour
	}
	return int64(d)
}
