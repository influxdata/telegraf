package docker_log

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/moby/moby/api/pkg/stdcopy"

	"github.com/influxdata/telegraf"
)

// Parse container name
func parseContainerName(containerNames []string) string {
	for _, name := range containerNames {
		trimmedName := strings.TrimPrefix(name, "/")
		if !strings.Contains(trimmedName, "/") {
			return trimmedName
		}
	}

	return ""
}

func parseLine(line []byte) (time.Time, string, error) {
	parts := bytes.SplitN(line, []byte(" "), 2)

	if len(parts) == 1 {
		parts = append(parts, []byte(""))
	}

	tsString := string(parts[0])

	// Keep any leading space, but remove whitespace from end of line.
	// This preserves space in, for example, stacktraces, while removing
	// annoying end of line characters and is similar to how other logging
	// plugins such as syslog behave.
	message := bytes.TrimRightFunc(parts[1], unicode.IsSpace)

	ts, err := time.Parse(time.RFC3339Nano, tsString)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("error parsing timestamp %q: %w", tsString, err)
	}

	return ts, string(message), nil
}

func hostnameFromID(id string) string {
	if len(id) > 12 {
		return id[0:12]
	}
	return id
}

func tailStream(
	acc telegraf.Accumulator,
	baseTags map[string]string,
	containerID string,
	reader io.ReadCloser,
	stream string,
) (time.Time, error) {
	defer reader.Close()

	tags := make(map[string]string, len(baseTags)+1)
	for k, v := range baseTags {
		tags[k] = v
	}
	tags["stream"] = stream

	r := bufio.NewReaderSize(reader, 64*1024)

	var lastTS time.Time
	for {
		line, err := r.ReadBytes('\n')

		if len(line) != 0 {
			ts, message, err := parseLine(line)
			if err != nil {
				acc.AddError(err)
			} else {
				acc.AddFields("docker_log", map[string]interface{}{
					"container_id": containerID,
					"message":      message,
				}, tags, ts)
			}

			// Store the last processed timestamp
			if ts.After(lastTS) {
				lastTS = ts
			}
		}

		if err != nil {
			if err == io.EOF {
				return lastTS, nil
			}
			return time.Time{}, err
		}
	}
}

func tailMultiplexed(acc telegraf.Accumulator, tags map[string]string, containerID string, src io.ReadCloser) (time.Time, error) {
	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()

	var tsStdout, tsStderr time.Time
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		tsStdout, err = tailStream(acc, tags, containerID, outReader, "stdout")
		if err != nil {
			acc.AddError(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		tsStderr, err = tailStream(acc, tags, containerID, errReader, "stderr")
		if err != nil {
			acc.AddError(err)
		}
	}()

	_, err := stdcopy.StdCopy(outWriter, errWriter, src)

	// Ignore the returned errors as we cannot do anything if the closing fails
	_ = outWriter.Close()
	_ = errWriter.Close()
	_ = src.Close()
	wg.Wait()

	if err != nil {
		return time.Time{}, err
	}
	if tsStdout.After(tsStderr) {
		return tsStdout, nil
	}
	return tsStderr, nil
}
