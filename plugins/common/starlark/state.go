package starlark

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	starlark_time "go.starlark.net/lib/time"
	"go.starlark.net/starlark"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// Prefixes for identifying the starlark type of a serialized value
const (
	prefixNone      = "n"
	prefixBool      = "b"
	prefixInt       = "i"
	prefixFloat     = "f"
	prefixString    = "s"
	prefixRawString = "r"
	prefixBytes     = "y"
	prefixTime      = "t"
	prefixDuration  = "d"
	prefixMetric    = "M"
	prefixList      = "L"
	prefixTuple     = "T"
	prefixSet       = "S"
	prefixDict      = "D"
)

// maxNestingDepth limits the recursion when (de-)serializing values to prevent
// endless recursion for self-referencing containers, e.g. a list added to
// itself, as well as stack exhaustion for untrusted input.
const maxNestingDepth = 64

// The state is serialized as a single JSON document in which each value is
// represented in a self-describing manner. Scalar values are serialized as
// strings of the format
//
//	<type-prefix>:<value>
//
// with <type-prefix> denoting the type of the serialized value according to
// the declarations above. <value> is the string representation of the
// serialized value itself. As JSON requires strings to be valid UTF-8, strings
// containing arbitrary data are base64 encoded and use a separate type-prefix.
// Metrics are serialized to their binary representation and base64 encoded.
// The tracking information of a metric is NOT preserved.
//
// Containers are serialized as JSON lists with the type-prefix as first
// element followed by the serialized values of the contained elements, e.g.
//
//	["L","i:42",["T","s:foo","n:"]]
//
// for a list containing the integer 42 and a tuple of the string "foo" and
// None. Dictionaries contain the serialized keys and values in an alternating
// fashion, preserving the insert order of the dictionary and allowing keys of
// arbitrary (hashable) type.

// marshal serializes the given state dictionary into a string preserving the
// types of all keys and values.
func marshal(state *starlark.Dict) (string, error) {
	if state == nil {
		return "", fmt.Errorf("invalid state %v", state)
	}

	encoded, err := encodeValue("state", state, 0)
	if err != nil {
		return "", err
	}

	buf, err := json.Marshal(encoded)
	if err != nil {
		return "", fmt.Errorf("marshalling state failed: %w", err)
	}

	return string(buf), nil
}

// unmarshal deserializes a state serialized by marshal back into a starlark
// dictionary restoring the original types of all keys and values.
func unmarshal(state string) (*starlark.Dict, error) {
	if state == "" {
		return nil, fmt.Errorf("invalid state %q", state)
	}

	var encoded any
	if err := json.Unmarshal([]byte(state), &encoded); err != nil {
		return nil, fmt.Errorf("unmarshalling state failed: %w", err)
	}

	value, err := decodeValue("state", encoded, 0)
	if err != nil {
		return nil, err
	}
	dict, ok := value.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("unexpected state type %T", value)
	}

	return dict, nil
}

// encodeValue serializes a starlark value into a JSON encodable representation
func encodeValue(path string, value starlark.Value, depth int) (any, error) {
	if depth > maxNestingDepth {
		return nil, fmt.Errorf("in %q: exceeding maximum nesting depth %d", path, maxNestingDepth)
	}

	switch v := value.(type) {
	case nil, starlark.NoneType:
		return prefixNone + ":", nil
	case starlark.Bool:
		return prefixBool + ":" + strconv.FormatBool(bool(v)), nil
	case starlark.Int:
		// Starlark integers are of arbitrary precision, so serialize them as
		// decimal string to not lose any information
		return prefixInt + ":" + v.String(), nil
	case starlark.Float:
		// Use the shortest representation that parses back to the exact same
		// value. This also covers the special values NaN and +/-Inf.
		return prefixFloat + ":" + strconv.FormatFloat(float64(v), 'g', -1, 64), nil
	case starlark.String:
		// Strings may contain arbitrary, non-UTF8 data which does not survive
		// being serialized to JSON, so encode those strings safely
		if !utf8.ValidString(string(v)) {
			return prefixRawString + ":" + base64.StdEncoding.EncodeToString([]byte(v)), nil
		}
		return prefixString + ":" + string(v), nil
	case starlark.Bytes:
		// Bytes may contain arbitrary, non-UTF8 data which might not survive
		// being handled as string, so encode them safely
		return prefixBytes + ":" + base64.StdEncoding.EncodeToString([]byte(v)), nil
	case starlark_time.Time:
		return prefixTime + ":" + time.Time(v).Format(time.RFC3339Nano), nil
	case starlark_time.Duration:
		return prefixDuration + ":" + strconv.FormatInt(int64(v), 10), nil
	case *Metric:
		// Drop the tracking information of tracking metrics by unwrapping the
		// underlying metric. This has to be done as we won't be able to restore
		// tracking metrics later because the delivery notification is only
		// valid at runtime and won't be available at restore time.
		m := v.Unwrap()
		if tm, ok := m.(telegraf.TrackingMetric); ok {
			m = tm.Unwrap()
		}

		buf, err := metric.ToBytes(m)
		if err != nil {
			return nil, fmt.Errorf("in %q: encoding metric %v failed: %w", path, v.Unwrap(), err)
		}
		return prefixMetric + ":" + base64.StdEncoding.EncodeToString(buf), nil
	case *starlark.List:
		entries := make([]any, 0, v.Len()+1)
		entries = append(entries, prefixList)
		idx := 0
		for se := range v.Elements() {
			e, err := encodeValue(path+"."+strconv.Itoa(idx), se, depth+1)
			if err != nil {
				return nil, err
			}
			entries = append(entries, e)
			idx++
		}
		return entries, nil
	case starlark.Tuple:
		entries := make([]any, 0, v.Len()+1)
		entries = append(entries, prefixTuple)
		idx := 0
		for se := range v.Elements() {
			e, err := encodeValue(path+"."+strconv.Itoa(idx), se, depth+1)
			if err != nil {
				return nil, err
			}
			entries = append(entries, e)
			idx++
		}
		return entries, nil
	case *starlark.Set:
		entries := make([]any, 0, v.Len()+1)
		entries = append(entries, prefixSet)
		idx := 0
		for se := range v.Elements() {
			e, err := encodeValue(path+"."+strconv.Itoa(idx), se, depth+1)
			if err != nil {
				return nil, err
			}
			entries = append(entries, e)
			idx++
		}
		return entries, nil
	case *starlark.Dict:
		// Serialize the keys and values in an alternating fashion to preserve
		// the insert order and to allow keys of arbitrary type
		entries := make([]any, 0, 2*v.Len()+1)
		entries = append(entries, prefixDict)
		for sk, sv := range v.Entries() {
			gk := toString(sk)
			key, err := encodeValue(path+".key:"+gk, sk, depth+1)
			if err != nil {
				return nil, err
			}
			val, err := encodeValue(path+".val:"+gk, sv, depth+1)
			if err != nil {
				return nil, err
			}
			entries = append(entries, key, val)
		}
		return entries, nil
	}

	return nil, fmt.Errorf("in %q: unsupported starlark type %T", path, value)
}

// decodeValue deserializes a serialized value and restores its original
// starlark type
func decodeValue(path string, encoded any, depth int) (starlark.Value, error) {
	if depth > maxNestingDepth {
		return nil, fmt.Errorf("in %q: exceeding maximum nesting depth %d", path, maxNestingDepth)
	}

	switch v := encoded.(type) {
	case string:
		prefix, value, found := strings.Cut(v, ":")
		if !found {
			return nil, fmt.Errorf("in %q: malformed serialized value %q", path, v)
		}

		switch prefix {
		case prefixNone:
			if value != "" {
				return nil, fmt.Errorf("in %q: unexpected value %q for none type", path, value)
			}
			return starlark.None, nil
		case prefixBool:
			v, err := strconv.ParseBool(value)
			if err != nil {
				return nil, fmt.Errorf("in %q: invalid boolean %q: %w", path, value, err)
			}
			return starlark.Bool(v), nil
		case prefixInt:
			// Starlark integers are of arbitrary precision, so decode them via a
			// big integer to not lose any information
			v, ok := new(big.Int).SetString(value, 10)
			if !ok {
				return nil, fmt.Errorf("in %q: invalid integer %q", path, value)
			}
			return starlark.MakeBigInt(v), nil
		case prefixFloat:
			// This also handles the special values NaN and +/-Inf
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("in %q: invalid float %q: %w", path, value, err)
			}
			return starlark.Float(v), nil
		case prefixString:
			return starlark.String(value), nil
		case prefixRawString:
			v, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				return nil, fmt.Errorf("in %q: invalid raw string %q: %w", path, value, err)
			}
			return starlark.String(v), nil
		case prefixBytes:
			v, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				return nil, fmt.Errorf("in %q: invalid bytes %q: %w", path, value, err)
			}
			return starlark.Bytes(v), nil
		case prefixTime:
			v, err := time.Parse(time.RFC3339Nano, value)
			if err != nil {
				return nil, fmt.Errorf("in %q: invalid time %q: %w", path, value, err)
			}
			return starlark_time.Time(v), nil
		case prefixDuration:
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("in %q: invalid duration %q: %w", path, value, err)
			}
			return starlark_time.Duration(v), nil
		case prefixMetric:
			buf, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				return nil, fmt.Errorf("in %q: unmarshalling metric %q failed: %w", path, value, err)
			}

			m, err := metric.FromBytes(buf)
			if err != nil {
				return nil, fmt.Errorf("in %q: decoding metric %q failed: %w", path, value, err)
			}
			return &Metric{metric: m}, nil
		}

		return nil, fmt.Errorf("in %q: unsupported type prefix %q", path, prefix)
	case []any:
		if len(v) == 0 {
			return nil, fmt.Errorf("in %q: container without type prefix", path)
		}
		prefix, ok := v[0].(string)
		if !ok {
			return nil, fmt.Errorf("in %q: invalid type %T for container prefix", path, v[0])
		}
		entries := v[1:]

		switch prefix {
		case prefixList:
			elements := make([]starlark.Value, 0, len(entries))
			for i, entry := range entries {
				se, err := decodeValue(path+"."+strconv.Itoa(i), entry, depth+1)
				if err != nil {
					return nil, err
				}
				elements = append(elements, se)
			}
			return starlark.NewList(elements), nil
		case prefixTuple:
			elements := make([]starlark.Value, 0, len(entries))
			for i, entry := range entries {
				se, err := decodeValue(path+"."+strconv.Itoa(i), entry, depth+1)
				if err != nil {
					return nil, err
				}
				elements = append(elements, se)
			}
			return starlark.Tuple(elements), nil
		case prefixSet:
			elements := make([]starlark.Value, 0, len(entries))
			for i, entry := range entries {
				se, err := decodeValue(path+"."+strconv.Itoa(i), entry, depth+1)
				if err != nil {
					return nil, err
				}
				elements = append(elements, se)
			}
			set := starlark.NewSet(len(elements))
			for i, element := range elements {
				if err := set.Insert(element); err != nil {
					return nil, fmt.Errorf("in %q: inserting set element failed: %w", path+"."+strconv.Itoa(i), err)
				}
			}
			return set, nil
		case prefixDict:
			if len(entries)%2 != 0 {
				return nil, fmt.Errorf("in %q: dictionary contains an incomplete key-value pair", path)
			}
			dict := starlark.NewDict(len(entries) / 2)
			for i := 0; i < len(entries); i += 2 {
				sk, err := decodeValue(path+".key:"+strconv.Itoa(i), entries[i], depth+1)
				if err != nil {
					return nil, err
				}
				gk := toString(sk)
				sv, err := decodeValue(path+".val:"+gk+"("+strconv.Itoa(i+1)+")", entries[i+1], depth+1)
				if err != nil {
					return nil, err
				}
				if err := dict.SetKey(sk, sv); err != nil {
					return nil, fmt.Errorf("in %q: setting dictionary key failed: %w", path+".key:"+gk+"("+strconv.Itoa(i)+")", err)
				}
			}
			return dict, nil
		}

		return nil, fmt.Errorf("in %q: unsupported container prefix %q", path, prefix)
	}

	return nil, fmt.Errorf("in %q: unsupported serialized type %T", path, encoded)
}
