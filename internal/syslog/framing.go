package syslog

import (
	"fmt"
	"strings"
)

// Framing represents the framing technique we expect the messages to come.
type Framing int

const (
	// OctetCounting indicates the transparent framing technique for syslog transport.
	OctetCounting Framing = iota
	// NonTransparent indicates the non-transparent framing technique for syslog transport.
	NonTransparent
)

func (f Framing) String() string {
	switch f {
	case OctetCounting:
		return "OCTET-COUNTING"
	case NonTransparent:
		return "NON-TRANSPARENT"
	}
	return ""
}

// UnmarshalTOML implements ability to unmarshal framing from TOML files.
func (f *Framing) UnmarshalTOML(data []byte) (err error) {
	return f.UnmarshalText(data)
}

// UnmarshalText implements encoding.TextUnmarshaler
func (f *Framing) UnmarshalText(data []byte) (err error) {
	s := string(data)
	switch strings.ToUpper(s) {
	case `OCTET-COUNTING`:
		fallthrough
	case `"OCTET-COUNTING"`:
		fallthrough
	case `'OCTET-COUNTING'`:
		*f = OctetCounting
		return

	case `NON-TRANSPARENT`:
		fallthrough
	case `"NON-TRANSPARENT"`:
		fallthrough
	case `'NON-TRANSPARENT'`:
		*f = NonTransparent
		return
	}
	*f = -1
	return fmt.Errorf("unknown framing")
}

// MarshalText implements encoding.TextMarshaler
func (f Framing) MarshalText() ([]byte, error) {
	s := f.String()
	if s != "" {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("unknown framing")
}
