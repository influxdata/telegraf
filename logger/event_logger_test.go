//go:build windows

package logger

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/svc/eventlog"
)

type Levels int

const (
	Info Levels = iota + 1
	Warning
	Error
)

type Event struct {
	Message string `xml:"EventData>Data"`
	Level   Levels `xml:"System>EventID"`
}

func getEventLog(t *testing.T, since time.Time) []Event {
	timeStr := since.UTC().Format(time.RFC3339)
	timeStr = timeStr[:19]
	args := []string{
		"qe",
		"Application",
		"/rd:true",
		fmt.Sprintf("/q:Event[System[TimeCreated[@SystemTime >= %q] and Provider[@Name='telegraf']]]", timeStr)}
	cmd := exec.Command("wevtutil", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	require.NoError(t, err)
	xmlStr := "<events>" + out.String() + "</events>"
	var events struct {
		Events []Event `xml:"Event"`
	}
	err = xml.Unmarshal([]byte(xmlStr), &events)
	require.NoError(t, err)
	return events.Events
}

func TestEventLogIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testPrepareLogger(t)

	config := LogConfig{
		LogTarget: LogTargetEventlog,
		Logfile:   "",
	}

	err := SetupLogging(config)
	require.NoError(t, err)

	now := time.Now()
	log.Println("I! Info message")
	log.Println("W! Warn message")
	log.Println("E! Err message")
	events := getEventLog(t, now)
	require.Len(t, events, 3)
	require.Contains(t, events, Event{Message: "Info message", Level: Info})
	require.Contains(t, events, Event{Message: "Warn message", Level: Warning})
	require.Contains(t, events, Event{Message: "Err message", Level: Error})
}

func TestRestrictedEventLogIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in  short mode")
	}
	testPrepareLogger(t)

	config := LogConfig{
		LogTarget: LogTargetEventlog,
		Quiet:     true,
	}

	err := SetupLogging(config)
	require.NoError(t, err)
	//separate previous log messages by small delay
	time.Sleep(time.Second)
	now := time.Now()
	log.Println("I! Info message")
	log.Println("W! Warning message")
	log.Println("E! Error message")
	events := getEventLog(t, now)
	require.Len(t, events, 1)
	require.Contains(t, events, Event{Message: "Error message", Level: Error})
}

func testPrepareLogger(tb testing.TB) {
	eventLog, err := eventlog.Open("telegraf")
	require.NoError(tb, err)
	require.NotNil(tb, eventLog)
	registerLogger(LogTargetEventlog, &eventLoggerCreator{logger: eventLog})
}
