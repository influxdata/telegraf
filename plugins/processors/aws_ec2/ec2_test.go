package aws_ec2

import (
	"sync"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/parallel"
	"github.com/influxdata/telegraf/testutil"
)

func TestBasicStartup(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{"accountId", "instanceId"}
	acc := &testutil.Accumulator{}
	require.NoError(t, p.Init())

	require.Empty(t, acc.GetTelegrafMetrics())
	require.Empty(t, acc.Errors)
}

func TestBasicStartupWithEC2Tags(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{"accountId", "instanceId"}
	p.EC2Tags = []string{"Name"}
	acc := &testutil.Accumulator{}
	require.NoError(t, p.Init())

	require.Empty(t, acc.GetTelegrafMetrics())
	require.Empty(t, acc.Errors)
}

func TestBasicStartupWithCacheTTL(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{"accountId", "instanceId"}
	p.CacheTTL = config.Duration(12 * time.Hour)
	acc := &testutil.Accumulator{}
	require.NoError(t, p.Init())

	require.Empty(t, acc.GetTelegrafMetrics())
	require.Empty(t, acc.Errors)
}

func TestBasicStartupWithTagCacheSize(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{"accountId", "instanceId"}
	p.TagCacheSize = 100
	acc := &testutil.Accumulator{}
	require.NoError(t, p.Init())

	require.Empty(t, acc.GetTelegrafMetrics())
	require.Empty(t, acc.Errors)
}

func TestBasicInitNoTagsReturnAnError(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{}
	err := p.Init()
	require.Error(t, err)
}

func TestBasicInitInvalidTagsReturnAnError(t *testing.T) {
	p := newAwsEc2Processor()
	p.Log = &testutil.Logger{}
	p.ImdsTags = []string{"dummy", "qwerty"}
	err := p.Init()
	require.Error(t, err)
}

func TestTracking(t *testing.T) {
	// Setup raw input and expected output
	inputRaw := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{
				"metric_tag": "from_metric",
			},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m2",
			map[string]string{
				"metric_tag": "foo_metric",
			},
			map[string]interface{}{"value": int64(2)},
			time.Unix(0, 0),
		),
	}
	expected := []telegraf.Metric{
		metric.New(
			"m1",
			map[string]string{
				"metric_tag": "from_metric",
				"accountId":  "123456789",
				"instanceId": "i-123456789123",
			},
			map[string]interface{}{"value": int64(1)},
			time.Unix(0, 0),
		),
		metric.New(
			"m2",
			map[string]string{
				"metric_tag": "foo_metric",
				"accountId":  "123456789",
				"instanceId": "i-123456789123",
			},
			map[string]interface{}{"value": int64(2)},
			time.Unix(0, 0),
		),
	}

	// Create fake notification for testing
	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	// Convert raw input to tracking metric
	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	// Prepare and start the plugin
	plugin := &AwsEc2Processor{
		MaxParallelCalls: DefaultMaxParallelCalls,
		TagCacheSize:     DefaultCacheSize,
		Timeout:          config.Duration(DefaultTimeout),
		CacheTTL:         config.Duration(DefaultCacheTTL),
		ImdsTags:         []string{"accountId", "instanceId"},
		Log:              &testutil.Logger{},
		imdsTagsMap:      make(map[string]struct{}),
	}
	require.NoError(t, plugin.Init())

	// Instead of starting the plugin which tries to connect to the remote
	// service, we just fill the cache and start the minimum mechanics to
	// process the metrics.
	plugin.tagCache = freecache.NewCache(DefaultCacheSize)
	require.NoError(t, plugin.tagCache.Set([]byte("accountId"), []byte("123456789"), -1))
	require.NoError(t, plugin.tagCache.Set([]byte("instanceId"), []byte("i-123456789123"), -1))

	var acc testutil.Accumulator
	plugin.parallel = parallel.NewOrdered(&acc, plugin.asyncAdd, plugin.TagCacheSize, plugin.MaxParallelCalls)

	// Schedule the metrics and wait until they are ready to perform the
	// comparison
	for _, in := range input {
		require.NoError(t, plugin.Add(in, &acc))
	}

	require.Eventually(t, func() bool {
		return int(acc.NMetrics()) >= len(expected)
	}, 3*time.Second, 100*time.Millisecond)

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual)

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(input) == len(delivered)
	}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(expected))
}
