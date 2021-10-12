package addons

import (
	"bytes"
	"strconv"
	"time"
)

// Timestamp aliases time.Time and adapts MarshalJSON and UnmarshalJSON
type Timestamp time.Time

// MarshalJSON implements json.Marshaler.
// Overrides nested time.Time MarshalJSON.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	i := time.Time(t).UnixMilli()
	buf := make([]byte, 0, 16)
	buf = append(buf, '"')
	buf = strconv.AppendInt(buf, i, 10)
	buf = append(buf, '"')
	return buf, nil
}

// UnmarshalJSON implements json.Unmarshaler.
// Overrides nested time.Time UnmarshalJSON.
func (t *Timestamp) UnmarshalJSON(input []byte) error {
	strInput := string(bytes.Trim(input, `"`))

	i, err := strconv.ParseInt(strInput, 10, 64)
	if err != nil {
		return err
	}

	i *= int64(time.Millisecond)
	*t = Timestamp(time.Unix(0, i).UTC())
	return nil
}

func Now() *Timestamp {
	now := Timestamp(time.Now())
	return &now
}

func TimestampRef(t time.Time) *Timestamp {
	ref := Timestamp(t)
	return &ref
}
