/*
	Provides an io.Writer that filters log messages based on a log level.

	Valid log levels are: DEBUG, INFO, WARN, ERROR.

	Log messages need to begin with a L! where L is one of D, I, W, or E.

	Examples:
		log.Println("D! this is a debug log")
		log.Println("I! this is an info log")
		log.Println("W! this is a warn log")
		log.Println("E! this is an error log")

	Simply pass a instance of wlog.Writer to log.New or use the helper wlog.New function.

	The log level can be changed via the SetLevel or the SetLevelFromName functions.
*/
package wlog

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

type Level int

const (
	_ Level = iota
	DEBUG
	INFO
	WARN
	ERROR
	OFF
)

const Delimiter = '!'

var invalidMSG = []byte("log messages must have 'L!' prefix where L is one of 'D', 'I', 'W', 'E'")

var Levels = map[byte]Level{
	'D': DEBUG,
	'I': INFO,
	'W': WARN,
	'E': ERROR,
}
var ReverseLevels map[Level]byte

func init() {
	ReverseLevels = make(map[Level]byte, len(Levels))
	for k, l := range Levels {
		ReverseLevels[l] = k
	}
}

// The global and only log level. Log levels are not implemented per writer.
var logLevel = INFO

var mu sync.RWMutex

// Set the current logging Level.
func SetLevel(l Level) {
	mu.Lock()
	defer mu.Unlock()
	logLevel = l
}

// Retrieve the current logging Level.
func LogLevel() Level {
	mu.RLock()
	defer mu.RUnlock()
	return logLevel
}

// name to Level mappings
var StringToLevel = map[string]Level{
	"DEBUG": DEBUG,
	"INFO":  INFO,
	"WARN":  WARN,
	"ERROR": ERROR,
	"OFF":   OFF,
}

// Set the log level via a string name. To set it directly use 'logLevel'.
func SetLevelFromName(level string) error {
	l := StringToLevel[strings.ToUpper(level)]
	if l > 0 {
		SetLevel(l)
	} else {
		return fmt.Errorf("invalid log level: %q", level)
	}
	return nil
}

// Implements io.Writer. Checks first byte of write for log level
// and drops the log if necessary
type Writer struct {
	start int
	w     io.Writer
}

// Create a new *log.Logger wrapping w in a wlog.Writer
func New(w io.Writer, prefix string, flag int) *log.Logger {
	return log.New(NewWriter(w), prefix, flag)
}

// Create a new wlog.Writer wrapping w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{-1, w}
}

// Implements the io.Writer method.
func (w *Writer) Write(buf []byte) (int, error) {
	if len(buf) > 0 {
		if w.start == -1 {
			// Find start of message index
			for i, c := range buf {
				if c == Delimiter && i > 0 {
					l := buf[i-1]
					level := Levels[l]
					if level > 0 {
						w.start = i - 1
						break
					}
				}
			}
			if w.start == -1 {
				buf = append(invalidMSG, buf...)
				return w.w.Write(buf)
			}
		}
		l := Levels[buf[w.start]]
		if l >= LogLevel() {
			return w.w.Write(buf)
		} else if l == 0 {
			buf = append(invalidMSG, buf...)
			return w.w.Write(buf)
		}
	}
	return 0, nil
}

// StaticLevelWriter prefixes all log messages
// with a static log level.
type StaticLevelWriter struct {
	levelPrefix []byte
	w           io.Writer
}

// Create a writer that always append a static log prefix to all messages.
// Usefult for supplying a *log.Logger to a package that doesn't
// prefix log messages itself.
func NewStaticLevelWriter(w io.Writer, level Level) *StaticLevelWriter {
	levelPrefix := []byte{ReverseLevels[level], '!', ' '}
	return &StaticLevelWriter{
		levelPrefix: levelPrefix,
		w:           w,
	}
}

func (w *StaticLevelWriter) Write(buf []byte) (int, error) {
	buf = append(w.levelPrefix, buf...)
	return w.w.Write(buf)
}
