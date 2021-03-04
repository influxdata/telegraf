package cloud_pubsub

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"encoding/base64"

	"cloud.google.com/go/pubsub"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/serializers"
	"google.golang.org/api/support/bundler"
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

		stopped bool
		pLock   sync.Mutex

		published map[string]*pubsub.Message

		bundler     *bundler.Bundler
		bLock       sync.Mutex
		bundleCount int
	}
)

func getTestResources(tT *testing.T, settings pubsub.PublishSettings, testM []testMetric) (*PubSub, *stubTopic, []telegraf.Metric) {
	s, _ := serializers.NewInfluxSerializer()

	metrics := make([]telegraf.Metric, len(testM))
	t := &stubTopic{
		T:         tT,
		ReturnErr: make(map[string]bool),
		published: make(map[string]*pubsub.Message),
	}

	for i, tm := range testM {
		metrics[i] = tm.m
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
	}
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
	p, _ := parsers.NewInfluxParser()
	metrics, err := p.Parse(msg.Data)
	if err != nil {
		// Just attempt to base64-decode first before returning error.
		d, err := base64.StdEncoding.DecodeString(string(msg.Data))
		if err != nil {
			t.Errorf("unable to base64-decode potential test message: %v", err)
		}
		metrics, err = p.Parse(d)
		if err != nil {
			t.Fatalf("unexpected parsing error: %v", err)
		}
	}

	ids := make([]string, len(metrics))
	for i, met := range metrics {
		id, _ := met.GetField("value")
		ids[i] = id.(string)
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
