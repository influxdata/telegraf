package kairosdb

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"reflect"
)

const tcpWaitBeforeReconnects = 10 * time.Second

type tcpOutput struct {
	address string
	timeout time.Duration

	connectionLost chan struct{}
	shutdown       chan chan error

	connectionLock sync.RWMutex
	conn           tcpConn
}

var _ innerOutput = (*tcpOutput)(nil)

// tcpConn is a partial interface of net.Conn to allow for easier mocking
type tcpConn interface {
	Close() error
	Write(b []byte) (n int, err error)
}

func (t *tcpOutput) Connect() error {
	if err := t.attemptConnect(); err != nil {
		return err
	}

	t.connectionLost = make(chan struct{}, 1)
	t.shutdown = make(chan chan error)
	t.startConnectionMonitor()

	return nil
}

func (t *tcpOutput) Close() error {
	shutdownDone := make(chan error)
	t.shutdown <- shutdownDone
	r := <-shutdownDone
	return r
}

func (t *tcpOutput) Write(metrics []telegraf.Metric) error {
	buf := bytes.NewBufferString("")
	for _, metric := range metrics {
		for fieldName, fieldVal := range metric.Fields() {
			buf.Reset()
			buf.WriteString("put ")
			buf.WriteString(postedName(metric, fieldName))
			buf.WriteString(" ")
			buf.WriteString(strconv.FormatInt(toMilliseconds(metric.Time()), 10))
			buf.WriteString(" ")
			if s, err := format(fieldVal); err != nil {
				log.Println("kairosdb: skipping metric: ", err)
				continue
			} else {
				buf.WriteString(s)
			}
			for tagKey, tagVal := range metric.Tags() {
				buf.WriteString(" ")
				buf.WriteString(tagKey)
				buf.WriteString("=")
				buf.WriteString(tagVal)
			}
			buf.WriteString("\n")

			conn := t.connection()
			if conn == nil {
				return errors.New("kairosdb: tcp connection to kairosdb not established")
			}

			if _, err := conn.Write(([]byte)(buf.String())); err != nil {
				t.invalidateConnection(conn)
				return fmt.Errorf("kairosdb: failed to write metric, %s\n", err.Error())
			}
		}
	}
	return nil
}

func (t *tcpOutput) startConnectionMonitor() {
	go func() {
		var shutdownDone chan error

		for {
			select {
			case <-t.connectionLost:
				shutdownDone = t.reconnect()
			case shutdownDone = <-t.shutdown:
			}

			if shutdownDone != nil {
				break
			}
		}

		t.doShutdown(shutdownDone)
	}()
}

func (t *tcpOutput) doShutdown(shutdownDone chan error) {
	var err error
	hijacked := t.hijackConnection(t.connection())
	if hijacked != nil {
		err = hijacked.Close()
	}
	shutdownDone <- err
}

func (t *tcpOutput) reconnect() (shutdownDone chan error) {
	waitBeforeReconnect := time.NewTicker(tcpWaitBeforeReconnects)
	defer waitBeforeReconnect.Stop()

	for {
		err := t.attemptConnect()
		if err == nil {
			return nil
		}

		log.Println("kairosdb: failed to reconnect. Will reattempt in ", tcpWaitBeforeReconnects.String(), ": ", err)

		select {
		case <-waitBeforeReconnect.C:
			continue
		case shutdownDone = <-t.shutdown:
			return shutdownDone
		}
	}
}

func (t *tcpOutput) attemptConnect() error {
	conn, err := net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		return errors.New("failed to connect to " + t.address + ": " + err.Error())
	}

	t.connectionLock.Lock()
	defer t.connectionLock.Unlock()
	t.conn = conn

	return nil
}

func (t *tcpOutput) connection() tcpConn {
	t.connectionLock.Lock()
	defer t.connectionLock.Unlock()

	return t.conn
}

func (t *tcpOutput) invalidateConnection(conn tcpConn) {
	hijacked := t.hijackConnection(conn)
	if hijacked == nil {
		return
	}

	t.connectionLost <- struct{}{}

	if err := hijacked.Close(); err != nil {
		log.Println("kairosdb: failed to close tcp connection: ", err)
	}
}

func (t *tcpOutput) hijackConnection(conn tcpConn) tcpConn {
	t.connectionLock.Lock()
	defer t.connectionLock.Unlock()

	if t.conn != conn {
		return nil
	}

	t.conn = nil

	return conn
}

func populateDatapoint(metric telegraf.Metric, fieldName string, fieldVal interface{}) (datapoint, error) {
	if !isValidType(fieldVal) {
		return datapoint{}, fmt.Errorf("unsupported type: %v", reflect.TypeOf(fieldVal))
	}

	name := postedName(metric, fieldName)
	ts := toMilliseconds(metric.Time())

	res := datapoint{
		Name:      name,
		Timestamp: ts,
		Value:     fieldVal,
		Tags:      metric.Tags(),
	}

	return res, nil
}

func isValidType(v interface{}) bool {
	switch v.(type) {
	case int, int32, int64, float32, float64:
		return true
	}
	return false
}

func postedName(metric telegraf.Metric, fieldName string) string {
	if fieldName == "value" {
		return metric.Name()
	}

	return metric.Name() + "." + fieldName
}

// toMilliseconds returns time in milliseconds since epoch
func toMilliseconds(t time.Time) int64 {
	return t.UnixNano() / time.Millisecond.Nanoseconds()
}

func format(fieldVal interface{}) (string, error) {
	switch fieldVal.(type) {
	case int:
		return strconv.FormatInt(int64(fieldVal.(int)), 10), nil
	case int32:
		return strconv.FormatInt(int64(fieldVal.(int32)), 10), nil
	case int64:
		return strconv.FormatInt(int64(fieldVal.(int64)), 10), nil
	case float32:
		return strconv.FormatFloat(float64(fieldVal.(float32)), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(float64(fieldVal.(float64)), 'f', -1, 64), nil
	}
	return "", fmt.Errorf("unsupported type: %v", reflect.TypeOf(fieldVal))
}
