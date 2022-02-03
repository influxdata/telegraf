package internal

import (
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
)

const alphanum string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var (
	ErrTimeout             = errors.New("command timed out")
	ErrorNotImplemented    = errors.New("not implemented yet")
	ErrorVersionAlreadySet = errors.New("version has already been set")
)

// Set via the main module
var version string

type ReadWaitCloser struct {
	pipeReader *io.PipeReader
	wg         sync.WaitGroup
}

// SetVersion sets the telegraf agent version
func SetVersion(v string) error {
	if version != "" {
		return ErrorVersionAlreadySet
	}
	version = v
	if version == "" {
		version = "unknown"
	}

	return nil
}

// Version returns the telegraf agent version
func Version() string {
	return version
}

// ProductToken returns a tag for Telegraf that can be used in user agents.
func ProductToken() string {
	return fmt.Sprintf("Telegraf/%s Go/%s",
		Version(), strings.TrimPrefix(runtime.Version(), "go"))
}

// ReadLines reads contents from a file and splits them by new lines.
// A convenience wrapper to ReadLinesOffsetN(filename, 0, -1).
func ReadLines(filename string) ([]string, error) {
	return ReadLinesOffsetN(filename, 0, -1)
}

// ReadLines reads contents from file and splits them by new line.
// The offset tells at which line number to start.
// The count determines the number of lines to read (starting from offset):
//   n >= 0: at most n lines
//   n < 0: whole file
func ReadLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}

// RandomString returns a random string of alpha-numeric characters
func RandomString(n int) string {
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// SnakeCase converts the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func SnakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

// RandomSleep will sleep for a random amount of time up to max.
// If the shutdown channel is closed, it will return before it has finished
// sleeping.
func RandomSleep(max time.Duration, shutdown chan struct{}) {
	if max == 0 {
		return
	}

	sleepns := rand.Int63n(max.Nanoseconds())

	t := time.NewTimer(time.Nanosecond * time.Duration(sleepns))
	select {
	case <-t.C:
		return
	case <-shutdown:
		t.Stop()
		return
	}
}

// RandomDuration returns a random duration between 0 and max.
func RandomDuration(max time.Duration) time.Duration {
	if max == 0 {
		return 0
	}

	sleepns := rand.Int63n(max.Nanoseconds())

	return time.Duration(sleepns)
}

// SleepContext sleeps until the context is closed or the duration is reached.
func SleepContext(ctx context.Context, duration time.Duration) error {
	if duration == 0 {
		return nil
	}

	t := time.NewTimer(duration)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		t.Stop()
		return ctx.Err()
	}
}

// AlignDuration returns the duration until next aligned interval.
// If the current time is aligned a 0 duration is returned.
func AlignDuration(tm time.Time, interval time.Duration) time.Duration {
	return AlignTime(tm, interval).Sub(tm)
}

// AlignTime returns the time of the next aligned interval.
// If the current time is aligned the current time is returned.
func AlignTime(tm time.Time, interval time.Duration) time.Time {
	truncated := tm.Truncate(interval)
	if truncated == tm {
		return tm
	}
	return truncated.Add(interval)
}

// ExitStatus takes the error from exec.Command
// and returns the exit status and true
// if error is not exit status, will return 0 and false
func ExitStatus(err error) (int, bool) {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), true
		}
	}
	return 0, false
}

func (r *ReadWaitCloser) Close() error {
	err := r.pipeReader.Close()
	r.wg.Wait() // wait for the gzip goroutine finish
	return err
}

// CompressWithGzip takes an io.Reader as input and pipes
// it through a gzip.Writer returning an io.Reader containing
// the gzipped data.
// An error is returned if passing data to the gzip.Writer fails
func CompressWithGzip(data io.Reader) (io.ReadCloser, error) {
	pipeReader, pipeWriter := io.Pipe()
	gzipWriter := gzip.NewWriter(pipeWriter)

	rc := &ReadWaitCloser{
		pipeReader: pipeReader,
	}

	rc.wg.Add(1)
	var err error
	go func() {
		_, err = io.Copy(gzipWriter, data)
		gzipWriter.Close()
		// subsequent reads from the read half of the pipe will
		// return no bytes and the error err, or EOF if err is nil.
		pipeWriter.CloseWithError(err)
		rc.wg.Done()
	}()

	return pipeReader, err
}

// ParseTimestamp parses a Time according to the standard Telegraf options.
// These are generally displayed in the toml similar to:
//   json_time_key= "timestamp"
//   json_time_format = "2006-01-02T15:04:05Z07:00"
//   json_timezone = "America/Los_Angeles"
//
// The format can be one of "unix", "unix_ms", "unix_us", "unix_ns", or a Go
// time layout suitable for time.Parse.
//
// When using the "unix" format, a optional fractional component is allowed.
// Specific unix time precisions cannot have a fractional component.
//
// Unix times may be an int64, float64, or string.  When using a Go format
// string the timestamp must be a string.
//
// The location is a location string suitable for time.LoadLocation.  Unix
// times do not use the location string, a unix time is always return in the
// UTC location.
func ParseTimestamp(format string, timestamp interface{}, location string) (time.Time, error) {
	switch format {
	case "unix", "unix_ms", "unix_us", "unix_ns":
		return parseUnix(format, timestamp)
	default:
		if location == "" {
			location = "UTC"
		}
		return parseTime(format, timestamp, location)
	}
}

func parseUnix(format string, timestamp interface{}) (time.Time, error) {
	integer, fractional, err := parseComponents(timestamp)
	if err != nil {
		return time.Unix(0, 0), err
	}

	switch strings.ToLower(format) {
	case "unix":
		return time.Unix(integer, fractional).UTC(), nil
	case "unix_ms":
		return time.Unix(0, integer*1e6).UTC(), nil
	case "unix_us":
		return time.Unix(0, integer*1e3).UTC(), nil
	case "unix_ns":
		return time.Unix(0, integer).UTC(), nil
	default:
		return time.Unix(0, 0), errors.New("unsupported type")
	}
}

// Returns the integers before and after an optional decimal point.  Both '.'
// and ',' are supported for the decimal point.  The timestamp can be an int64,
// float64, or string.
//   ex: "42.5" -> (42, 5, nil)
func parseComponents(timestamp interface{}) (int64, int64, error) {
	switch ts := timestamp.(type) {
	case string:
		parts := strings.SplitN(ts, ".", 2)
		if len(parts) == 2 {
			return parseUnixTimeComponents(parts[0], parts[1])
		}

		parts = strings.SplitN(ts, ",", 2)
		if len(parts) == 2 {
			return parseUnixTimeComponents(parts[0], parts[1])
		}

		integer, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		return integer, 0, nil
	case int8:
		return int64(ts), 0, nil
	case int16:
		return int64(ts), 0, nil
	case int32:
		return int64(ts), 0, nil
	case int64:
		return ts, 0, nil
	case uint8:
		return int64(ts), 0, nil
	case uint16:
		return int64(ts), 0, nil
	case uint32:
		return int64(ts), 0, nil
	case uint64:
		return int64(ts), 0, nil
	case float32:
		integer, fractional := math.Modf(float64(ts))
		return int64(integer), int64(fractional * 1e9), nil
	case float64:
		integer, fractional := math.Modf(ts)
		return int64(integer), int64(fractional * 1e9), nil
	default:
		return 0, 0, errors.New("unsupported type")
	}
}

func parseUnixTimeComponents(first, second string) (int64, int64, error) {
	integer, err := strconv.ParseInt(first, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	// Convert to nanoseconds, dropping any greater precision.
	buf := []byte("000000000")
	copy(buf, second)

	fractional, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return integer, fractional, nil
}

// ParseTime parses a string timestamp according to the format string.
func parseTime(format string, timestamp interface{}, location string) (time.Time, error) {
	switch ts := timestamp.(type) {
	case string:
		loc, err := time.LoadLocation(location)
		if err != nil {
			return time.Unix(0, 0), err
		}
		switch strings.ToLower(format) {
		case "ansic":
			format = time.ANSIC
		case "unixdate":
			format = time.UnixDate
		case "rubydate":
			format = time.RubyDate
		case "rfc822":
			format = time.RFC822
		case "rfc822z":
			format = time.RFC822Z
		case "rfc850":
			format = time.RFC850
		case "rfc1123":
			format = time.RFC1123
		case "rfc1123z":
			format = time.RFC1123Z
		case "rfc3339":
			format = time.RFC3339
		case "rfc3339nano":
			format = time.RFC3339Nano
		case "stamp":
			format = time.Stamp
		case "stampmilli":
			format = time.StampMilli
		case "stampmicro":
			format = time.StampMicro
		case "stampnano":
			format = time.StampNano
		}
		return time.ParseInLocation(format, ts, loc)
	default:
		return time.Unix(0, 0), errors.New("unsupported type")
	}
}
