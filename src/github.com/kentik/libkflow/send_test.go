package libkflow

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"zombiezen.com/go/capnproto2"

	"github.com/kentik/go-metrics"
	"github.com/kentik/libkflow/agg"
	"github.com/kentik/libkflow/api/test"
	"github.com/kentik/libkflow/chf"
	"github.com/kentik/libkflow/flow"
	"github.com/stretchr/testify/assert"
)

func TestSender(t *testing.T) {
	sender, server, assert := setup(t)

	expected := flow.Flow{
		DeviceId:  uint32(sender.Device.ID),
		SrcAs:     rand.Uint32(),
		DstAs:     rand.Uint32(),
		SampleAdj: true,
	}

	sender.Send(&expected)

	msgs, err := receive(server)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(flowToCHF(expected, t).String(), msgs.At(0).String())
}

func TestSenderStop(t *testing.T) {
	sender, _, assert := setup(t)
	stopped := sender.Stop(100 * time.Millisecond)
	assert.True(stopped)
}

func BenchmarkSenderSend(b *testing.B) {
	sender, _, _ := setup(b)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		sender.Send(&flow.Flow{
			SrcAs: uint32(b.N),
			DstAs: uint32(b.N),
		})
	}
}

func setup(t testing.TB) (*Sender, *test.Server, *assert.Assertions) {
	metrics := &agg.Metrics{
		TotalFlowsIn:   metrics.NewMeter(),
		TotalFlowsOut:  metrics.NewMeter(),
		OrigSampleRate: metrics.NewHistogram(metrics.NewUniformSample(100)),
		NewSampleRate:  metrics.NewHistogram(metrics.NewUniformSample(100)),
		RateLimitDrops: metrics.NewMeter(),
	}

	agg, err := agg.NewAgg(10*time.Millisecond, 100, metrics)
	if err != nil {
		t.Fatal(err)
	}

	client, server, device, err := test.NewClientServer()
	if err != nil {
		t.Fatal(err)
	}

	server.Log.SetOutput(ioutil.Discard)

	url := server.URL(test.FLOW)
	sender := newSender(url, 1*time.Second, 0)
	sender.start(agg, client, device, 1)

	return sender, server, assert.New(t)
}

func receive(s *test.Server) (*chf.CHF_List, error) {
	interval := 100 * time.Millisecond
	select {
	case flow := <-s.Flows():
		msgs, err := flow.Msgs()
		return &msgs, err
	case <-time.After(interval):
		return nil, fmt.Errorf("failed to receive flow within %s", interval)
	}
}

func flowToCHF(flow flow.Flow, t testing.TB) chf.CHF {
	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	kflow, err := chf.NewCHF(seg)
	if err != nil {
		t.Fatal(err)
	}

	list, err := chf.NewCustom_List(seg, int32(len(flow.Customs)))
	if err != nil {
		t.Fatal(err)
	}

	flow.FillCHF(kflow, list)

	return kflow
}
