package config

import (
	"bytes"
	"strconv"
	"time"

	"github.com/alecthomas/units"
)

// Duration is a time.Duration
type Duration time.Duration

// Size is an int64
type Size int64

// Number is a float
type Number float64

// UnmarshalTOML parses the duration from the TOML config file
func (d Duration) UnmarshalTOML(b []byte) error {
	var err error
	b = bytes.Trim(b, `'`)

	// see if we can directly convert it
	dur, err := time.ParseDuration(string(b))
	if err == nil {
		d = Duration(dur)
		return nil
	}

	// Parse string duration, ie, "1s"
	if uq, err := strconv.Unquote(string(b)); err == nil && len(uq) > 0 {
		dur, err := time.ParseDuration(uq)
		if err == nil {
			d = Duration(dur)
			return nil
		}
	}

	// First try parsing as integer seconds
	sI, err := strconv.ParseInt(string(b), 10, 64)
	if err == nil {
		dur := time.Second * time.Duration(sI)
		d = Duration(dur)
		return nil
	}
	// Second try parsing as float seconds
	sF, err := strconv.ParseFloat(string(b), 64)
	if err == nil {
		dur := time.Second * time.Duration(sF)
		d = Duration(dur)
		return nil
	}

	return nil
}

func (s Size) UnmarshalTOML(b []byte) error {
	var err error
	b = bytes.Trim(b, `'`)

	val, err := strconv.ParseInt(string(b), 10, 64)
	if err == nil {
		s = Size(val)
		return nil
	}
	uq, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	val, err = units.ParseStrictBytes(uq)
	if err != nil {
		return err
	}
	s = Size(val)
	return nil
}

func (n Number) UnmarshalTOML(b []byte) error {
	value, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}

	n = Number(value)
	return nil
}
