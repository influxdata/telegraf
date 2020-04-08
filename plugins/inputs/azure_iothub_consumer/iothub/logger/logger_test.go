package logger

import (
	"bytes"
	"fmt"
	"testing"
)

func TestLogger(t *testing.T) {
	buf := bytes.Buffer{}
	l := New(LevelInfo, func(lvl Level, s string) {
		buf.WriteString(fmt.Sprint(lvl.String(), ": ", s, "\n"))
	})
	l.Errorf("error")
	l.Warnf("warn")
	l.Infof("info")
	l.Debugf("debug")

	want := "ERROR: error\nWARN: warn\nINFO: info\n"
	if have := buf.String(); have != want {
		t.Fatalf("logger output = %q, want %q", have, want)
	}
}
