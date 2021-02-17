package websocket

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/stretchr/testify/require"
	"testing"

	"nhooyr.io/websocket"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
)

func TestReadingJSON(t *testing.T) {
	numberCases := 5
	timestamp, err := time.Parse(time.RFC3339, "2021-02-16T12:48:35Z")
	require.NoError(t, err)

	// Construct the input
	message := `
		{
			"time": %d,
			"field1": 1.345677,
			"field2": true,
			"field3": 42,
			"field4": "websockets",
			"seqno": %d,
			"server": "test",
			"user": "ender"
		}
	`
	input := make([]string, 0, numberCases)
	for i := 0; i < numberCases; i++ {
		msg := fmt.Sprintf(message, timestamp.AddDate(0, i, i).Unix(), i)
		input = append(input, msg)
	}

	// Setup the test server
	options := websocket.AcceptOptions{CompressionMode: websocket.CompressionDisabled}
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := websocket.Accept(w, r, &options)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		for _, msg := range input {
			if err := connection.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
				return
			}
		}
	}))
	defer fakeServer.Close()

	// Construct the expected metrics
	expected := make([]telegraf.Metric, 0, numberCases)
	for i := 0; i < numberCases; i++ {
		m := testutil.MustMetric(
			"websocket",
			map[string]string{
				"server": "test",
				"user":   "ender",
				"url":    fakeServer.URL,
			},
			map[string]interface{}{
				"field1": float64(1.345677),
				"field2": true,
				"field3": float64(42),
				"field4": "websockets",
				"seqno":  float64(i),
			},
			timestamp.AddDate(0, i, i),
		)
		expected = append(expected, m)
	}

	// Setup and start the plugin
	plugin := &Websocket{
		URL:     fakeServer.URL,
		Timeout: internal.Duration{Duration: 1 * time.Second},
		Log:     testutil.Logger{Name: "websocket"},
	}

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat:       "json",
		MetricName:       "websocket",
		TagKeys:          []string{"server", "user"},
		JSONStringFields: []string{"field4", "field2"},
		JSONTimeKey:      "time",
		JSONTimeFormat:   "unix",
	})
	plugin.SetParser(p)

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start(acc))

	acc.Wait(len(expected))

	plugin.Stop()

	require.Len(t, acc.Metrics, len(expected))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestReadingJSONWithCompression(t *testing.T) {
	numberCases := 5
	timestamp, err := time.Parse(time.RFC3339, "2021-02-16T12:48:35Z")
	require.NoError(t, err)

	// Construct the input
	message := `
		{
			"time": %d,
			"field1": 1.345677,
			"field2": true,
			"field3": 42,
			"field4": "websockets",
			"seqno": %d,
			"server": "test",
			"user": "ender"
		}
	`
	input := make([]string, 0, numberCases)
	for i := 0; i < numberCases; i++ {
		msg := fmt.Sprintf(message, timestamp.AddDate(0, i, i).Unix(), i)
		input = append(input, msg)
	}

	// Setup the test server
	options := websocket.AcceptOptions{CompressionMode: websocket.CompressionContextTakeover}
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := websocket.Accept(w, r, &options)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		for _, msg := range input {
			if err := connection.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
				return
			}
		}
	}))
	defer fakeServer.Close()

	// Construct the expected metrics
	expected := make([]telegraf.Metric, 0, numberCases)
	for i := 0; i < numberCases; i++ {
		m := testutil.MustMetric(
			"websocket",
			map[string]string{
				"server": "test",
				"user":   "ender",
				"url":    fakeServer.URL,
			},
			map[string]interface{}{
				"field1": float64(1.345677),
				"field2": true,
				"field3": float64(42),
				"field4": "websockets",
				"seqno":  float64(i),
			},
			timestamp.AddDate(0, i, i),
		)
		expected = append(expected, m)
	}

	// Setup and start the plugin
	plugin := &Websocket{
		URL:     fakeServer.URL,
		Timeout: internal.Duration{Duration: 1 * time.Second},
		Log:     testutil.Logger{Name: "websocket"},
	}

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat:       "json",
		MetricName:       "websocket",
		TagKeys:          []string{"server", "user"},
		JSONStringFields: []string{"field4", "field2"},
		JSONTimeKey:      "time",
		JSONTimeFormat:   "unix",
	})
	plugin.SetParser(p)

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start(acc))

	acc.Wait(len(expected))

	plugin.Stop()

	require.Len(t, acc.Metrics, len(expected))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestReadingWSAddress(t *testing.T) {
	timestamp, err := time.Parse(time.RFC3339, "2021-02-16T12:48:35Z")
	require.NoError(t, err)

	// Construct the input
	message := `
		{
			"time": %d,
			"field1": 1.345677,
			"field2": true,
			"field3": 42,
			"field4": "websockets",
			"seqno": 0,
			"server": "test",
			"user": "ender"
		}
	`
	input := []string{fmt.Sprintf(message, timestamp.Unix())}

	// Setup the test server
	options := websocket.AcceptOptions{}
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := websocket.Accept(w, r, &options)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		for _, msg := range input {
			if err := connection.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
				return
			}
		}
	}))
	defer fakeServer.Close()
	url := strings.Replace(fakeServer.URL, "http://", "ws://", 1)

	// Construct the expected metrics
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"websocket",
			map[string]string{
				"server": "test",
				"user":   "ender",
				"url":    url,
			},
			map[string]interface{}{
				"field1": float64(1.345677),
				"field2": true,
				"field3": float64(42),
				"field4": "websockets",
				"seqno":  float64(0),
			},
			timestamp,
		),
	}

	// Setup and start the plugin
	plugin := &Websocket{
		URL:     url,
		Timeout: internal.Duration{Duration: 1 * time.Second},
		Log:     testutil.Logger{Name: "websocket"},
	}

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat:       "json",
		MetricName:       "websocket",
		TagKeys:          []string{"server", "user"},
		JSONStringFields: []string{"field4", "field2"},
		JSONTimeKey:      "time",
		JSONTimeFormat:   "unix",
	})
	plugin.SetParser(p)

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start(acc))

	acc.Wait(len(expected))

	plugin.Stop()

	require.Len(t, acc.Metrics, len(expected))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestHandshake(t *testing.T) {
	timestamp, err := time.Parse(time.RFC3339, "2021-02-16T12:48:35Z")
	require.NoError(t, err)

	// Construct the input
	message := `
		{
			"time": %d,
			"field1": 1.345677,
			"field2": true,
			"field3": 42,
			"field4": "websockets",
			"seqno": 0,
			"server": "test",
			"user": "ender"
		}
	`
	input := []string{fmt.Sprintf(message, timestamp.Unix())}

	handshakeMessage := "Hello Captain!"

	// Setup the test server
	options := websocket.AcceptOptions{}
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := websocket.Accept(w, r, &options)
		require.NoError(t, err)
		defer connection.Close(websocket.StatusInternalError, "abnormal exit")

		for {
			ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
			defer cancel()

			_, buf, err := connection.Read(ctx)
			if err != nil {
				return
			}

			if string(buf) == handshakeMessage {
				break
			}
		}

		for _, msg := range input {
			if err := connection.Write(r.Context(), websocket.MessageText, []byte(msg)); err != nil {
				return
			}
		}
		connection.Close(websocket.StatusNormalClosure, "server shutdown")
	}))
	defer fakeServer.Close()

	// Construct the expected metrics
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"websocket",
			map[string]string{
				"server": "test",
				"user":   "ender",
				"url":    fakeServer.URL,
			},
			map[string]interface{}{
				"field1": float64(1.345677),
				"field2": true,
				"field3": float64(42),
				"field4": "websockets",
				"seqno":  float64(0),
			},
			timestamp,
		),
	}

	// Setup and start the plugin
	plugin := &Websocket{
		URL:           fakeServer.URL,
		Timeout:       internal.Duration{Duration: 1 * time.Second},
		HandshakeBody: handshakeMessage,
		Log:           testutil.Logger{Name: "websocket"},
	}

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat:       "json",
		MetricName:       "websocket",
		TagKeys:          []string{"server", "user"},
		JSONStringFields: []string{"field4", "field2"},
		JSONTimeKey:      "time",
		JSONTimeFormat:   "unix",
	})
	plugin.SetParser(p)

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start(acc))

	timeout := time.After(3 * time.Second)
	done := make(chan bool)
	go func() {
		acc.Wait(1)
		done <- true
	}()

	select {
	case <-timeout:
		t.Fatal("timeout while waiting for handshake")
	case <-done:
	}

	plugin.Stop()

	require.Len(t, acc.Metrics, 1)
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestGather(t *testing.T) {
	numberCases := 5

	timestamp, err := time.Parse(time.RFC3339, "2021-02-16T12:48:35Z")
	require.NoError(t, err)

	// Construct the input
	message := `
		{
			"time": %d,
			"field1": 1.345677,
			"field2": true,
			"field3": 42,
			"field4": "websockets",
			"seqno": %d,
			"server": "test",
			"user": "ender"
		}
	`

	triggerMessage := "Gimme data!"

	// Setup the test server
	options := websocket.AcceptOptions{}
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := websocket.Accept(w, r, &options)
		require.NoError(t, err)
		defer connection.Close(websocket.StatusInternalError, "abnormal exit")

		for i := 0; ; i++ {
			ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
			defer cancel()

			_, buf, err := connection.Read(ctx)
			if err != nil {
				return
			}

			if string(buf) == triggerMessage {
				msg := fmt.Sprintf(message, timestamp.AddDate(0, 0, i).Unix(), i)
				if err := connection.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
					return
				}
			}
		}
	}))
	defer fakeServer.Close()

	// Construct the expected metrics
	expected := make([]telegraf.Metric, 0, numberCases)
	for i := 0; i < numberCases; i++ {
		m := testutil.MustMetric(
			"websocket",
			map[string]string{
				"server": "test",
				"user":   "ender",
				"url":    fakeServer.URL,
			},
			map[string]interface{}{
				"field1": float64(1.345677),
				"field2": true,
				"field3": float64(42),
				"field4": "websockets",
				"seqno":  float64(i),
			},
			timestamp.AddDate(0, 0, i),
		)
		expected = append(expected, m)
	}

	// Setup and start the plugin
	plugin := &Websocket{
		URL:         fakeServer.URL,
		Timeout:     internal.Duration{Duration: 1 * time.Second},
		TriggerBody: triggerMessage,
		Log:         testutil.Logger{Name: "websocket"},
	}

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat:       "json",
		MetricName:       "websocket",
		TagKeys:          []string{"server", "user"},
		JSONStringFields: []string{"field4", "field2"},
		JSONTimeKey:      "time",
		JSONTimeFormat:   "unix",
	})
	plugin.SetParser(p)

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start(acc))

	for i := 0; i < numberCases; i++ {
		err := plugin.Gather(acc)
		require.NoError(t, err)
	}

	timeout := time.After(3 * time.Second)
	done := make(chan bool)
	go func() {
		acc.Wait(len(expected))
		done <- true
	}()

	select {
	case <-timeout:
		t.Fatal("timeout while waiting for trigger")
	case <-done:
	}

	plugin.Stop()

	require.Len(t, acc.Metrics, len(expected))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestServerDropout(t *testing.T) {
	numberCases := 5

	timestamp, err := time.Parse(time.RFC3339, "2021-02-16T12:48:35Z")
	require.NoError(t, err)

	// Construct the input
	message := `
		{
			"time": %d,
			"field1": 1.345677,
			"field2": true,
			"field3": 42,
			"field4": "websockets",
			"seqno": %d,
			"server": "test",
			"user": "ender"
		}
	`

	// Setup the test server
	request := 0
	options := websocket.AcceptOptions{}
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := websocket.Accept(w, r, &options)
		require.NoError(t, err)
		defer connection.Close(websocket.StatusInternalError, "abnormal exit")

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		msg := fmt.Sprintf(message, timestamp.AddDate(0, 0, request).Unix(), request)
		if err := connection.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		request++
	}))
	defer fakeServer.Close()

	// Construct the expected metrics
	expected := make([]telegraf.Metric, 0)
	for i := 0; i < numberCases; i++ {
		m := testutil.MustMetric(
			"websocket",
			map[string]string{
				"server": "test",
				"user":   "ender",
				"url":    fakeServer.URL,
			},
			map[string]interface{}{
				"field1": float64(1.345677),
				"field2": true,
				"field3": float64(42),
				"field4": "websockets",
				"seqno":  float64(i),
			},
			timestamp.AddDate(0, 0, i),
		)
		expected = append(expected, m)
	}

	// Setup and start the plugin
	plugin := &Websocket{
		URL:     fakeServer.URL,
		Timeout: internal.Duration{Duration: 1 * time.Second},
		Log:     testutil.Logger{Name: "websocket"},
	}

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat:       "json",
		MetricName:       "websocket",
		TagKeys:          []string{"server", "user"},
		JSONStringFields: []string{"field4", "field2"},
		JSONTimeKey:      "time",
		JSONTimeFormat:   "unix",
	})
	plugin.SetParser(p)

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start(acc))

	timeout := time.After(5 * time.Second)
	done := make(chan bool)
	go func() {
		acc.Wait(len(expected))
		done <- true
	}()

	select {
	case <-timeout:
		t.Fatalf("timeout after %d metrics", acc.NMetrics())
	case <-done:
	}

	plugin.Stop()

	require.Len(t, acc.Metrics, len(expected))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestServerDropoutGather(t *testing.T) {
	numberCases := 5

	timestamp, err := time.Parse(time.RFC3339, "2021-02-16T12:48:35Z")
	require.NoError(t, err)

	// Construct the input
	message := `
		{
			"time": %d,
			"field1": 1.345677,
			"field2": true,
			"field3": 42,
			"field4": "websockets",
			"seqno": %d,
			"server": "test",
			"user": "ender"
		}
	`

	triggerMessage := "Gimme data!"

	// Setup the test server
	request := 0
	options := websocket.AcceptOptions{}
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := websocket.Accept(w, r, &options)
		defer connection.Close(websocket.StatusInternalError, "abnormal exit")
		if err != nil {
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		_, buf, err := connection.Read(ctx)
		if err != nil {
			return
		}

		if string(buf) == triggerMessage {
			msg := fmt.Sprintf(message, timestamp.AddDate(0, 0, request).Unix(), request)
			if err := connection.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
				return
			}
			request++
		}
		connection.Close(websocket.StatusNormalClosure, "server shutdown")
	}))
	defer fakeServer.Close()

	// Construct the expected metrics
	expected := make([]telegraf.Metric, 0)
	for i := 0; i < numberCases; i++ {
		m := testutil.MustMetric(
			"websocket",
			map[string]string{
				"server": "test",
				"user":   "ender",
				"url":    fakeServer.URL,
			},
			map[string]interface{}{
				"field1": float64(1.345677),
				"field2": true,
				"field3": float64(42),
				"field4": "websockets",
				"seqno":  float64(i),
			},
			timestamp.AddDate(0, 0, i),
		)
		expected = append(expected, m)
	}

	// Setup and start the plugin
	plugin := &Websocket{
		URL:         fakeServer.URL,
		Timeout:     internal.Duration{Duration: 1 * time.Second},
		TriggerBody: triggerMessage,
		Log:         testutil.Logger{Name: "websocket"},
	}

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat:       "json",
		MetricName:       "websocket",
		TagKeys:          []string{"server", "user"},
		JSONStringFields: []string{"field4", "field2"},
		JSONTimeKey:      "time",
		JSONTimeFormat:   "unix",
	})
	plugin.SetParser(p)

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start(acc))

	for i := 0; i < 4*len(expected); i++ {
		plugin.Gather(acc)
		if acc.NMetrics() >= uint64(len(expected)) {
			break
		}
	}
	plugin.Stop()

	require.Len(t, acc.Metrics, len(expected))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}
