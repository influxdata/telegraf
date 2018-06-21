package internal

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
	"log"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
)

const alphanum string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var (
	TimeoutErr = errors.New("Command timed out.")

	NotImplementedError = errors.New("not implemented yet")
)

// Duration just wraps time.Duration
type Duration struct {
	Duration time.Duration
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
	timer := time.NewTimer(timeout)
	done := make(chan error)
	go func() { done <- c.Wait() }()
	select {
	case err := <-done:
		timer.Stop()
		return err
	case <-timer.C:
		if err := c.Process.Kill(); err != nil {
			log.Printf("E! FATAL error killing process: %s", err)
			return err
		}
		// wait for the command to return after killing it
		<-done
		return TimeoutErr
	}
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
