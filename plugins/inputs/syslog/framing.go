package syslog

import (
	"fmt"
	"strings"
)

type framing int

const (
	// OctetCounting indicates the transparent framing technique for syslog transport.
	OctetCounting framing = iota
	// NonTransparent indicates the non-transparent framing technique for syslog transport.
	NonTransparent
)

func (f framing) String() string {
	switch f {
	case OctetCounting:
		return "OCTET-COUNTING"
	case NonTransparent:
		return "NON-TRANSPARENT"
	}
	return ""
}

// UnmarshalText implements encoding.TextUnmarshaler
func (f *framing) UnmarshalText(data []byte) (err error) {
	s := string(data)
	switch strings.ToUpper(s) {
	case "OCTET-COUNTING":
		*f = OctetCounting
	case "NON-TRANSPARENT":
		*f = NonTransparent
	}
	return err
}

// MarshalText implements encoding.TextMarshaler
func (f framing) MarshalText() ([]byte, error) {
	s := f.String()
	if s != "" {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("unknown framing")
}
