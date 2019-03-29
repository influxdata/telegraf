package internal

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/alecthomas/units"
)

const alphanum string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var (
	TimeoutErr = errors.New("Command timed out.")

	NotImplementedError = errors.New("not implemented yet")

	VersionAlreadySetError = errors.New("version has already been set")
)

// Set via the main module
var version string

// Duration just wraps time.Duration
type Duration struct {
	Duration time.Duration
}

// Size just wraps an int64
type Size struct {
	Size int64
}

// SetVersion sets the telegraf agent version
func SetVersion(v string) error {
	if version != "" {
		return VersionAlreadySetError
	}
	version = v
	return nil
}

// Version returns the telegraf agent version
func Version() string {
	return version
}

// ProductToken returns a tag for Telegraf that can be used in user agents.
func ProductToken() string {
	return fmt.Sprintf("Telegraf/%s Go/%s", Version(), runtime.Version())
}

// UnmarshalTOML parses the duration from the TOML config file
func (d *Duration) UnmarshalTOML(b []byte) error {
	var err error
	b = bytes.Trim(b, `'`)

	// see if we can directly convert it
	d.Duration, err = time.ParseDuration(string(b))
	if err == nil {
		return nil
	}

	// Parse string duration, ie, "1s"
	if uq, err := strconv.Unquote(string(b)); err == nil && len(uq) > 0 {
		d.Duration, err = time.ParseDuration(uq)
		if err == nil {
			return nil
		}
	}

	// First try parsing as integer seconds
	sI, err := strconv.ParseInt(string(b), 10, 64)
	if err == nil {
		d.Duration = time.Second * time.Duration(sI)
		return nil
	}
	// Second try parsing as float seconds
	sF, err := strconv.ParseFloat(string(b), 64)
	if err == nil {
		d.Duration = time.Second * time.Duration(sF)
		return nil
	}

	return nil
}

func (s *Size) UnmarshalTOML(b []byte) error {
	var err error
	b = bytes.Trim(b, `'`)

	val, err := strconv.ParseInt(string(b), 10, 64)
	if err == nil {
		s.Size = val
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
	s.Size = val
	return nil
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

// CombinedOutputTimeout runs the given command with the given timeout and
// returns the combined output of stdout and stderr.
// If the command times out, it attempts to kill the process.
func CombinedOutputTimeout(c *exec.Cmd, timeout time.Duration) ([]byte, error) {
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	if err := c.Start(); err != nil {
		return nil, err
	}
	err := WaitTimeout(c, timeout)
	return b.Bytes(), err
}

// RunTimeout runs the given command with the given timeout.
// If the command times out, it attempts to kill the process.
func RunTimeout(c *exec.Cmd, timeout time.Duration) error {
	if err := c.Start(); err != nil {
		return err
	}
	return WaitTimeout(c, timeout)
}

// WaitTimeout waits for the given command to finish with a timeout.
// It assumes the command has already been started.
// If the command times out, it attempts to kill the process.
func WaitTimeout(c *exec.Cmd, timeout time.Duration) error {
	timer := time.AfterFunc(timeout, func() {
		err := c.Process.Kill()
		if err != nil {
			log.Printf("E! FATAL error killing process: %s", err)
			return
		}
	})

	err := c.Wait()
	isTimeout := timer.Stop()

	if err != nil {
		return err
	} else if isTimeout == false {
		return TimeoutErr
	}

	return err
}

// RandomSleep will sleep for a random amount of time up to max.
// If the shutdown channel is closed, it will return before it has finished
// sleeping.
func RandomSleep(max time.Duration, shutdown chan struct{}) {
	if max == 0 {
		return
	}
	maxSleep := big.NewInt(max.Nanoseconds())

	var sleepns int64
	if j, err := rand.Int(rand.Reader, maxSleep); err == nil {
		sleepns = j.Int64()
	}

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

	var sleepns int64
	maxSleep := big.NewInt(max.Nanoseconds())
	if j, err := rand.Int(rand.Reader, maxSleep); err == nil {
		sleepns = j.Int64()
	}

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

// Exit status takes the error from exec.Command
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

// CompressWithGzip takes an io.Reader as input and pipes
// it through a gzip.Writer returning an io.Reader containing
// the gzipped data.
// An error is returned if passing data to the gzip.Writer fails
func CompressWithGzip(data io.Reader) (io.Reader, error) {
	pipeReader, pipeWriter := io.Pipe()
	gzipWriter := gzip.NewWriter(pipeWriter)

	var err error
	go func() {
		_, err = io.Copy(gzipWriter, data)
		gzipWriter.Close()
		// subsequent reads from the read half of the pipe will
		// return no bytes and the error err, or EOF if err is nil.
		pipeWriter.CloseWithError(err)
	}()

	return pipeReader, err
}

// ParseTimestamp with no location provided parses a timestamp value as UTC
func ParseTimestamp(timestamp interface{}, format string) (time.Time, error) {
	return ParseTimestampWithLocation(timestamp, format, "UTC")
}

// ParseTimestamp parses a timestamp value as a unix epoch of various precision.
//
// format = "unix": epoch is assumed to be in seconds and can come as number or string. Can have a decimal part.
// format = "unix_ms": epoch is assumed to be in milliseconds and can come as number or string. Cannot have a decimal part.
// format = "unix_us": epoch is assumed to be in microseconds and can come as number or string. Cannot have a decimal part.
// format = "unix_ns": epoch is assumed to be in nanoseconds and can come as number or string. Cannot have a decimal part.
func ParseTimestampWithLocation(timestamp interface{}, format string, location string) (time.Time, error) {
	timeInt, timeFractional := int64(0), int64(0)
	timeEpochStr, ok := timestamp.(string)
	var err error

	if !ok {
		timeEpochFloat, ok := timestamp.(float64)
		if !ok {
			return time.Time{}, fmt.Errorf("time: %v could not be converted to string nor float64", timestamp)
		}
		intPart, frac := math.Modf(timeEpochFloat)
		timeInt, timeFractional = int64(intPart), int64(frac*1e9)
	} else {
		splitted := regexp.MustCompile("[.,]").Split(timeEpochStr, 2)
		timeInt, err = strconv.ParseInt(splitted[0], 10, 64)
		if err != nil {
			loc, err := time.LoadLocation(location)
			if err != nil {
				return time.Time{}, fmt.Errorf("location: %s could not be loaded as a location", location)
			}
			return time.ParseInLocation(format, timeEpochStr, loc)
		}

		if len(splitted) == 2 {
			if len(splitted[1]) > 9 {
				splitted[1] = splitted[1][:9] //truncates decimal part to nanoseconds precision
			}
			nanosecStr := splitted[1] + strings.Repeat("0", 9-len(splitted[1])) //adds 0's to the right to obtain a valid number of nanoseconds

			timeFractional, err = strconv.ParseInt(nanosecStr, 10, 64)
			if err != nil {
				return time.Time{}, err
			}
		}
	}
	if strings.EqualFold(format, "unix") {
		return time.Unix(timeInt, timeFractional).UTC(), nil
	} else if strings.EqualFold(format, "unix_ms") {
		return time.Unix(timeInt/1000, (timeInt%1000)*1e6).UTC(), nil
	} else if strings.EqualFold(format, "unix_us") {
		return time.Unix(0, timeInt*1e3).UTC(), nil
	} else if strings.EqualFold(format, "unix_ns") {
		return time.Unix(0, timeInt).UTC(), nil
	} else {
		return time.Time{}, errors.New("Invalid unix format")
	}
}
