package cloud_pubsub

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/support/bundler"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	serializer "github.com/influxdata/telegraf/plugins/serializers/influx"
)

const (
	errMockFail = "this is an error"
)

type (
	testMetric struct {
		m         telegraf.Metric
		returnErr bool
	}

	bundledMsg struct {
		*pubsub.Message
		*stubResult
	}

	stubResult struct {
		metricIds []string

		sendError bool
		err       chan error
		done      chan struct{}
	}

	stubTopic struct {
		Settings  pubsub.PublishSettings
		ReturnErr map[string]bool
		parsers.Parser
		*testing.T
		Base64Data           bool
		ContentEncoding      string
		MaxDecompressionSize int64

		stopped bool
		pLock   sync.Mutex

		published map[string]*pubsub.Message

		bundler     *bundler.Bundler
		bLock       sync.Mutex
		bundleCount int
	}
)

func getTestResources(tT *testing.T, settings pubsub.PublishSettings, testM []testMetric) (*PubSub, *stubTopic, []telegraf.Metric) {
	// Instantiate a Influx line-protocol serializer
	s := &serializer.Serializer{}
	_ = s.Init() // We can ignore the error as the Init will never fail

	metrics := make([]telegraf.Metric, 0, len(testM))
	t := &stubTopic{
		T:                    tT,
		ReturnErr:            make(map[string]bool),
		published:            make(map[string]*pubsub.Message),
		ContentEncoding:      "identity",
		MaxDecompressionSize: internal.DefaultMaxDecompressionSize,
	}

	for _, tm := range testM {
		metrics = append(metrics, tm.m)
		if tm.returnErr {
			v, _ := tm.m.GetField("value")
			t.ReturnErr[v.(string)] = true
		}
	}

	ps := &PubSub{
		Project:               "test-project",
		Topic:                 "test-topic",
		stubTopic:             func(string) topic { return t },
		PublishCountThreshold: settings.CountThreshold,
		PublishByteThreshold:  settings.ByteThreshold,
		PublishNumGoroutines:  settings.NumGoroutines,
		PublishTimeout:        config.Duration(settings.Timeout),
		ContentEncoding:       "identity",
	}

	require.NoError(tT, ps.Init())
	ps.encoder, _ = internal.NewContentEncoder(ps.ContentEncoding)
	ps.SetSerializer(s)

	return ps, t, metrics
}

func (t *stubTopic) ID() string {
	return "test-topic"
}

func (t *stubTopic) Stop() {
	t.pLock.Lock()
	defer t.pLock.Unlock()

	t.stopped = true
	t.bundler.Flush()
}

func (t *stubTopic) Publish(ctx context.Context, msg *pubsub.Message) publishResult {
	t.pLock.Lock()
	defer t.pLock.Unlock()

	if t.stopped || ctx.Err() != nil {
		t.Fatalf("publish called after stop")
	}

	ids := t.parseIDs(msg)
	r := &stubResult{
		metricIds: ids,
		err:       make(chan error, 1),
		done:      make(chan struct{}, 1),
	}

	for _, id := range ids {
		_, ok := t.ReturnErr[id]
		r.sendError = r.sendError || ok
	}

	bundled := &bundledMsg{msg, r}
	if err := t.bundler.Add(bundled, len(msg.Data)); err != nil {
		t.Fatalf("unexpected error while adding to bundle: %v", err)
	}
	return r
}

func (t *stubTopic) PublishSettings() pubsub.PublishSettings {
	return t.Settings
}

func (t *stubTopic) SetPublishSettings(settings pubsub.PublishSettings) {
	t.Settings = settings
	t.initBundler()
}

func (t *stubTopic) initBundler() *stubTopic {
	t.bundler = bundler.NewBundler(&bundledMsg{}, t.sendBundle())
	t.bundler.DelayThreshold = 10 * time.Second
	t.bundler.BundleCountThreshold = t.Settings.CountThreshold
	if t.bundler.BundleCountThreshold > pubsub.MaxPublishRequestCount {
		t.bundler.BundleCountThreshold = pubsub.MaxPublishRequestCount
	}
	t.bundler.BundleByteThreshold = t.Settings.ByteThreshold
	t.bundler.BundleByteLimit = pubsub.MaxPublishRequestBytes
	t.bundler.HandlerLimit = 25 * runtime.GOMAXPROCS(0)

	return t
}

func (t *stubTopic) sendBundle() func(items interface{}) {
	return func(items interface{}) {
		t.bLock.Lock()
		defer t.bLock.Unlock()

		bundled := items.([]*bundledMsg)

		for _, msg := range bundled {
			r := msg.stubResult
			for _, id := range r.metricIds {
				t.published[id] = msg.Message
			}

			if r.sendError {
				r.err <- errors.New(errMockFail)
			} else {
				r.done <- struct{}{}
			}
		}

		t.bundleCount++
	}
}

func (t *stubTopic) parseIDs(msg *pubsub.Message) []string {
	p := influx.Parser{}
	err := p.Init()
	require.NoError(t, err)

	decoder, _ := internal.NewContentDecoder(t.ContentEncoding)
	d, err := decoder.Decode(msg.Data, t.MaxDecompressionSize)
	if err != nil {
		t.Errorf("unable to decode message: %v", err)
	}
	if t.Base64Data {
		strData, err := base64.StdEncoding.DecodeString(string(d))
		if err != nil {
			t.Errorf("unable to base64 decode message: %v", err)
		}
		d = strData
	}
	metrics, err := p.Parse(d)
	if err != nil {
		t.Fatalf("unexpected parsing error: %v", err)
	}

	ids := make([]string, 0, len(metrics))
	for _, met := range metrics {
		id, _ := met.GetField("value")
		ids = append(ids, id.(string))
	}
	return ids
}

func (r *stubResult) Get(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-r.err:
		return "", err
	case <-r.done:
		return fmt.Sprintf("id-%s", r.metricIds[0]), nil
	}
}

func (t *stubTopic) getBundleCount() int {
	t.bLock.Lock()
	defer t.bLock.Unlock()
	return t.bundleCount
}
