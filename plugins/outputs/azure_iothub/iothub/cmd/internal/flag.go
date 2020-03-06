package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/amenzhinsky/iothub/logger"
)

type JSONMapFlag map[string]interface{}

func (f *JSONMapFlag) Set(s string) error {
	if len(*f) == 0 {
		*f = JSONMapFlag{}
	}
	c := strings.SplitN(s, "=", 2)
	if len(c) != 2 {
		return errors.New("malformed key-value flag")
	}
	var v interface{}
	if c[1] != "" {
		if err := json.Unmarshal([]byte(c[1]), &v); err != nil {
			return err
		}
	}
	(*f)[c[0]] = v
	return nil
}

func (f *JSONMapFlag) String() string {
	return fmt.Sprintf("%v", map[string]interface{}(*f))
}

type StringsMapFlag map[string]string

func (f *StringsMapFlag) Set(s string) error {
	if len(*f) == 0 {
		*f = StringsMapFlag{}
	}
	c := strings.SplitN(s, "=", 2)
	if len(c) != 2 {
		return errors.New("malformed key-value flag")
	}
	(*f)[c[0]] = c[1]
	return nil
}

func (f *StringsMapFlag) String() string {
	return fmt.Sprintf("%v", map[string]string(*f))
}

type LogLevelFlag logger.Level

func (f *LogLevelFlag) Set(s string) error {
	var lvl logger.Level
	switch strings.ToLower(s) {
	case "off":
		lvl = logger.LevelOff
	case "e", "err", "error":
		lvl = logger.LevelError
	case "w", "warn", "warning":
		lvl = logger.LevelWarn
	case "i", "info":
		lvl = logger.LevelInfo
	case "d", "dbg", "debug":
		lvl = logger.LevelDebug
	default:
		return fmt.Errorf("cannot parse %q level", s)
	}
	*(*logger.Level)(f) = lvl
	return nil
}

func (f *LogLevelFlag) String() string {
	return fmt.Sprintf("%q", strings.ToLower(logger.Level(*f).String()))
}

type TimeFlag time.Time

func (f *TimeFlag) Set(s string) error {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*f = TimeFlag(t)
	return nil
}

func (f *TimeFlag) String() string {
	return (*time.Time)(f).Format(time.RFC3339)
}
