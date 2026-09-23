package kafka_consumer_lag

import (
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	kafkacontainer "github.com/testcontainers/testcontainers-go/modules/kafka"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/testutil"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *KafkaConsumerLag
		expected string
	}{
		{
			name:   "defaults",
			plugin: newPlugin(),
		},
		{
			name: "no brokers",
			plugin: &KafkaConsumerLag{
				MetricLevels:        []string{levelPartition},
				CoordinatorBrokerID: -1,
			},
			expected: "brokers must not be empty",
		},
		{
			name: "no metric levels",
			plugin: &KafkaConsumerLag{
				Brokers:             []string{"localhost:9092"},
				CoordinatorBrokerID: -1,
			},
			expected: "metric_levels must not be empty",
		},
		{
			name: "invalid metric level",
			plugin: &KafkaConsumerLag{
				Brokers:             []string{"localhost:9092"},
				MetricLevels:        []string{"cluster"},
				CoordinatorBrokerID: -1,
			},
			expected: `invalid metric level "cluster"`,
		},
		{
			name: "invalid group filter",
			plugin: &KafkaConsumerLag{
				Brokers:             []string{"localhost:9092"},
				MetricLevels:        []string{levelPartition},
				GroupsInclude:       []string{"[invalid"},
				CoordinatorBrokerID: -1,
			},
			expected: "creating group filter failed",
		},
		{
			name: "invalid version",
			plugin: &KafkaConsumerLag{
				Brokers:             []string{"localhost:9092"},
				MetricLevels:        []string{levelPartition},
				CoordinatorBrokerID: -1,
				Version:             "not-a-version",
			},
			expected: "setting config failed",
		},
		{
			name: "version too old",
			plugin: &KafkaConsumerLag{
				Brokers:             []string{"localhost:9092"},
				MetricLevels:        []string{levelPartition},
				CoordinatorBrokerID: -1,
				Version:             "0.10.2.0",
			},
			expected: "0.11.0.0 or greater is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = testutil.Logger{}
			err := tt.plugin.Init()
			if tt.expected == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestInitDisablesTopicCreation(t *testing.T) {
	plugin := newPlugin()
	require.NoError(t, plugin.Init())
	require.False(t, plugin.config.Metadata.AllowAutoTopicCreation)
}

func TestInitBatchOffsets(t *testing.T) {
	for version, expected := range map[string]bool{"2.8.0": false, "3.0.0": true, "3.6.0": true} {
		t.Run(version, func(t *testing.T) {
			plugin := newPlugin()
			plugin.Version = version
			require.NoError(t, plugin.Init())
			require.Equal(t, expected, plugin.batchOffsets)
		})
	}
}

// mockCluster wires two mock brokers into a cluster with the following layout:
//
//	broker 1: leader of orders/0 and events/0, coordinator of "order-svc"
//	broker 2: leader of orders/1 and events/1, coordinator of "analytics"
//
//	group "order-svc" (2 members): orders/0=50, orders/1=80, events/0=never committed, events/1=20
//	                               retired/0=5 (topic deleted, offset still stored)
//	group "analytics" (0 members): orders/0=10
//
//	log end offsets: orders/0=100, orders/1=100, events/0=30, events/1=30
type mockCluster struct {
	broker1 *sarama.MockBroker
	broker2 *sarama.MockBroker

	handlers map[*sarama.MockBroker]map[string]sarama.MockResponse
}

func newMockCluster(t *testing.T) *mockCluster {
	t.Helper()

	// Install the sarama logger before the mock brokers start reading it
	kafka.SetLogger(testutil.Logger{}.Level())

	broker1 := sarama.NewMockBroker(t, 1)
	broker2 := sarama.NewMockBroker(t, 2)

	metadata := sarama.NewMockMetadataResponse(t).
		SetController(broker1.BrokerID()).
		SetBroker(broker1.Addr(), broker1.BrokerID()).
		SetBroker(broker2.Addr(), broker2.BrokerID()).
		SetLeader("orders", 0, broker1.BrokerID()).
		SetLeader("orders", 1, broker2.BrokerID()).
		SetLeader("events", 0, broker1.BrokerID()).
		SetLeader("events", 1, broker2.BrokerID())

	coordinators := sarama.NewMockFindCoordinatorResponse(t).
		SetCoordinator(sarama.CoordinatorGroup, "order-svc", broker1).
		SetCoordinator(sarama.CoordinatorGroup, "analytics", broker2)

	handlers1 := map[string]sarama.MockResponse{
		"ApiVersionsRequest":     sarama.NewMockApiVersionsResponse(t),
		"MetadataRequest":        metadata,
		"FindCoordinatorRequest": coordinators,
		"ListGroupsRequest": sarama.NewMockListGroupsResponse(t).
			AddGroup("order-svc", "consumer"),
		"DescribeGroupsRequest": sarama.NewMockDescribeGroupsResponse(t).
			AddGroupDescription("order-svc", &sarama.GroupDescription{
				GroupId: "order-svc",
				State:   "Stable",
				Members: map[string]*sarama.GroupMemberDescription{
					"member-1": {MemberId: "member-1", ClientId: "order-svc-1", ClientHost: "/10.0.0.1"},
					"member-2": {MemberId: "member-2", ClientId: "order-svc-2", ClientHost: "/10.0.0.2"},
				},
			}),
		"OffsetFetchRequest": sarama.NewMockOffsetFetchResponse(t).
			SetOffset("order-svc", "orders", 0, 50, "", sarama.ErrNoError).
			SetOffset("order-svc", "orders", 1, 80, "", sarama.ErrNoError).
			SetOffset("order-svc", "events", 0, -1, "", sarama.ErrNoError).
			SetOffset("order-svc", "events", 1, 20, "", sarama.ErrNoError).
			SetOffset("order-svc", "retired", 0, 5, "", sarama.ErrNoError),
		"OffsetRequest": sarama.NewMockOffsetResponse(t).
			SetOffset("orders", 0, sarama.OffsetNewest, 100).
			SetOffset("events", 0, sarama.OffsetNewest, 30),
	}

	handlers2 := map[string]sarama.MockResponse{
		"ApiVersionsRequest":     sarama.NewMockApiVersionsResponse(t),
		"MetadataRequest":        metadata,
		"FindCoordinatorRequest": coordinators,
		"ListGroupsRequest": sarama.NewMockListGroupsResponse(t).
			AddGroup("analytics", "consumer"),
		"DescribeGroupsRequest": sarama.NewMockDescribeGroupsResponse(t).
			AddGroupDescription("analytics", &sarama.GroupDescription{
				GroupId: "analytics",
				State:   "Empty",
			}),
		"OffsetFetchRequest": sarama.NewMockOffsetFetchResponse(t).
			SetOffset("analytics", "orders", 0, 10, "", sarama.ErrNoError),
		"OffsetRequest": sarama.NewMockOffsetResponse(t).
			SetOffset("orders", 1, sarama.OffsetNewest, 100).
			SetOffset("events", 1, sarama.OffsetNewest, 30),
	}

	broker1.SetHandlerByMap(handlers1)
	broker2.SetHandlerByMap(handlers2)

	return &mockCluster{
		broker1:  broker1,
		broker2:  broker2,
		handlers: map[*sarama.MockBroker]map[string]sarama.MockResponse{broker1: handlers1, broker2: handlers2},
	}
}

// override replaces the response of a broker for one request type.
func (c *mockCluster) override(broker *sarama.MockBroker, request string, response sarama.MockResponse) {
	c.handlers[broker][request] = response
	broker.SetHandlerByMap(c.handlers[broker])
}

func (c *mockCluster) brokers() []string {
	return []string{c.broker1.Addr(), c.broker2.Addr()}
}

func (c *mockCluster) close() {
	c.broker1.Close()
	c.broker2.Close()
}

// addGroup registers a second group coordinated by broker 1 which committed
// orders/0 only, so that broker 1 coordinates two groups.
func (c *mockCluster) addGroup(t *testing.T, group string, offset int64) {
	t.Helper()

	c.override(c.broker1, "ListGroupsRequest",
		sarama.NewMockListGroupsResponse(t).
			AddGroup("order-svc", "consumer").
			AddGroup(group, "consumer"),
	)
	coordinators := sarama.NewMockFindCoordinatorResponse(t).
		SetCoordinator(sarama.CoordinatorGroup, "order-svc", c.broker1).
		SetCoordinator(sarama.CoordinatorGroup, group, c.broker1).
		SetCoordinator(sarama.CoordinatorGroup, "analytics", c.broker2)
	c.override(c.broker1, "FindCoordinatorRequest", coordinators)
	c.override(c.broker2, "FindCoordinatorRequest", coordinators)
	c.override(c.broker1, "OffsetFetchRequest",
		sarama.NewMockOffsetFetchResponse(t).
			SetOffset("order-svc", "orders", 0, 50, "", sarama.ErrNoError).
			SetOffset("order-svc", "orders", 1, 80, "", sarama.ErrNoError).
			SetOffset("order-svc", "events", 0, -1, "", sarama.ErrNoError).
			SetOffset("order-svc", "events", 1, 20, "", sarama.ErrNoError).
			SetOffset("order-svc", "retired", 0, 5, "", sarama.ErrNoError).
			SetOffset(group, "orders", 0, offset, "", sarama.ErrNoError),
	)
}

// limitOffsetFetch makes both brokers advertise OffsetFetch up to the given
// protocol version only.
func (c *mockCluster) limitOffsetFetch(t *testing.T, maxVersion int16) {
	t.Helper()

	for _, broker := range []*sarama.MockBroker{c.broker1, c.broker2} {
		c.override(broker, "ApiVersionsRequest",
			sarama.NewMockApiVersionsResponse(t).SetApiKeys([]sarama.ApiVersionsResponseKey{
				{ApiKey: 9, MinVersion: 0, MaxVersion: maxVersion},
			}),
		)
	}
}

// countRequests returns how many requests of type T the broker received.
func countRequests[T any](broker *sarama.MockBroker) int {
	var n int
	for _, rr := range broker.History() {
		if _, ok := rr.Request.(T); ok {
			n++
		}
	}
	return n
}

func newPlugin() *KafkaConsumerLag {
	return &KafkaConsumerLag{
		Brokers:             []string{"localhost:9092"},
		MetricLevels:        []string{levelPartition, levelTopic, levelGroup},
		CoordinatorBrokerID: -1,
		Log:                 testutil.Logger{},
	}
}

func gather(t *testing.T, plugin *KafkaConsumerLag) *testutil.Accumulator {
	t.Helper()

	acc := gatherWithErrors(t, plugin)
	require.Empty(t, acc.Errors)

	return acc
}

// gatherWithErrors runs one collection cycle and leaves checking the
// accumulated errors to the caller.
func gatherWithErrors(t *testing.T, plugin *KafkaConsumerLag) *testutil.Accumulator {
	t.Helper()

	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))

	return &acc
}

func partitionMetric(group, topic, partition string, committed, logEnd int64) telegraf.Metric {
	return metric.New(
		"kafka_consumer_group_partition",
		map[string]string{
			"group":     group,
			"topic":     topic,
			"partition": partition,
		},
		map[string]any{
			"committed_offset": committed,
			"log_end_offset":   logEnd,
			"lag":              logEnd - committed,
		},
		time.Unix(0, 0),
		telegraf.Gauge,
	)
}

func topicMetric(group, topic string, sum, maximum, partitions int64) telegraf.Metric {
	return metric.New(
		"kafka_consumer_group_topic",
		map[string]string{
			"group": group,
			"topic": topic,
		},
		map[string]any{
			"lag_sum":    sum,
			"lag_max":    maximum,
			"partitions": partitions,
		},
		time.Unix(0, 0),
		telegraf.Gauge,
	)
}

func groupMetric(group string, sum, maximum, partitions, topics, members int64) telegraf.Metric {
	return metric.New(
		"kafka_consumer_group",
		map[string]string{"group": group},
		map[string]any{
			"lag_sum":    sum,
			"lag_max":    maximum,
			"partitions": partitions,
			"topics":     topics,
			"members":    members,
		},
		time.Unix(0, 0),
		telegraf.Gauge,
	)
}

func TestGather(t *testing.T) {
	// Run against both the legacy and the multi-group OffsetFetch protocol
	for _, version := range []string{"2.1.0", "3.6.0"} {
		t.Run(version, func(t *testing.T) {
			cluster := newMockCluster(t)
			defer cluster.close()

			plugin := newPlugin()
			plugin.Brokers = cluster.brokers()
			plugin.Version = version

			acc := gather(t, plugin)

			expected := []telegraf.Metric{
				partitionMetric("order-svc", "orders", "0", 50, 100),
				partitionMetric("order-svc", "orders", "1", 80, 100),
				partitionMetric("order-svc", "events", "1", 20, 30),
				partitionMetric("analytics", "orders", "0", 10, 100),
				topicMetric("order-svc", "orders", 70, 50, 2),
				topicMetric("order-svc", "events", 10, 10, 1),
				topicMetric("analytics", "orders", 90, 90, 1),
				groupMetric("order-svc", 80, 50, 3, 2, 2),
				groupMetric("analytics", 90, 90, 1, 1, 0),
			}
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}

func TestGatherMetadataFullDisabled(t *testing.T) {
	cluster := newMockCluster(t)
	defer cluster.close()

	plugin := newPlugin()
	plugin.Brokers = cluster.brokers()
	plugin.MetricLevels = []string{levelPartition}
	full := false
	plugin.MetadataFull = &full

	acc := gather(t, plugin)

	expected := []telegraf.Metric{
		partitionMetric("order-svc", "orders", "0", 50, 100),
		partitionMetric("order-svc", "orders", "1", 80, 100),
		partitionMetric("order-svc", "events", "1", 20, 30),
		partitionMetric("analytics", "orders", "0", 10, 100),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherCoordinatorFilter(t *testing.T) {
	cluster := newMockCluster(t)
	defer cluster.close()

	plugin := newPlugin()
	plugin.Brokers = cluster.brokers()
	plugin.MetricLevels = []string{levelPartition}
	plugin.CoordinatorBrokerID = cluster.broker2.BrokerID()

	acc := gather(t, plugin)

	expected := []telegraf.Metric{
		partitionMetric("analytics", "orders", "0", 10, 100),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherGroupAndTopicFilter(t *testing.T) {
	cluster := newMockCluster(t)
	defer cluster.close()

	plugin := newPlugin()
	plugin.Brokers = cluster.brokers()
	plugin.MetricLevels = []string{levelPartition}
	plugin.GroupsInclude = []string{"order-*"}
	plugin.TopicsExclude = []string{"events"}

	acc := gather(t, plugin)

	expected := []telegraf.Metric{
		partitionMetric("order-svc", "orders", "0", 50, 100),
		partitionMetric("order-svc", "orders", "1", 80, 100),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherMetricLevels(t *testing.T) {
	cluster := newMockCluster(t)
	defer cluster.close()

	plugin := newPlugin()
	plugin.Brokers = cluster.brokers()
	plugin.MetricLevels = []string{levelGroup}

	acc := gather(t, plugin)

	expected := []telegraf.Metric{
		groupMetric("order-svc", 80, 50, 3, 2, 2),
		groupMetric("analytics", 90, 90, 1, 1, 0),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherSkipsDescribeWithoutGroupLevel(t *testing.T) {
	cluster := newMockCluster(t)
	defer cluster.close()

	plugin := newPlugin()
	plugin.Brokers = cluster.brokers()
	plugin.MetricLevels = []string{levelPartition, levelTopic}

	acc := gather(t, plugin)
	require.NotEmpty(t, acc.GetTelegrafMetrics())

	// The member count is only used on the group level, so no group must
	// have been described.
	for _, broker := range []*sarama.MockBroker{cluster.broker1, cluster.broker2} {
		for _, rr := range broker.History() {
			_, described := rr.Request.(*sarama.DescribeGroupsRequest)
			require.Falsef(t, described, "broker %d received a DescribeGroupsRequest", broker.BrokerID())
		}
	}
}

func TestGatherDescribeFailure(t *testing.T) {
	cluster := newMockCluster(t)
	defer cluster.close()

	// Broker 2 refuses to describe its group
	cluster.override(cluster.broker2, "DescribeGroupsRequest",
		sarama.NewMockDescribeGroupsResponse(t).
			AddGroupDescription("analytics", &sarama.GroupDescription{
				GroupId:   "analytics",
				ErrorCode: int16(sarama.ErrGroupAuthorizationFailed),
			}),
	)

	plugin := newPlugin()
	plugin.Brokers = cluster.brokers()
	plugin.MetricLevels = []string{levelGroup}

	acc := gatherWithErrors(t, plugin)

	// The lag is still reported, only the members field is missing
	expected := []telegraf.Metric{
		groupMetric("order-svc", 80, 50, 3, 2, 2),
		metric.New(
			"kafka_consumer_group",
			map[string]string{"group": "analytics"},
			map[string]any{
				"lag_sum":    int64(90),
				"lag_max":    int64(90),
				"partitions": int64(1),
				"topics":     int64(1),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())

	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], `describing group "analytics" failed`)
}

func TestGatherDescribeNonClassicGroup(t *testing.T) {
	// Kafka 4.0 answers DescribeGroups for a group using the new consumer
	// protocol with an empty "Dead" group up to v5, and with an error from v6
	tests := []struct {
		name        string
		description *sarama.GroupDescription
	}{
		{
			name:        "dead group",
			description: &sarama.GroupDescription{GroupId: "analytics", State: "Dead"},
		},
		{
			name: "group not found",
			description: &sarama.GroupDescription{
				GroupId:   "analytics",
				State:     "Dead",
				ErrorCode: int16(sarama.ErrGroupIDNotFound),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := newMockCluster(t)
			defer cluster.close()

			cluster.override(cluster.broker2, "DescribeGroupsRequest",
				sarama.NewMockDescribeGroupsResponse(t).AddGroupDescription("analytics", tt.description),
			)

			plugin := newPlugin()
			plugin.Brokers = cluster.brokers()
			plugin.MetricLevels = []string{levelGroup}

			// The lag is still reported without the members field and without error
			acc := gather(t, plugin)

			expected := []telegraf.Metric{
				groupMetric("order-svc", 80, 50, 3, 2, 2),
				metric.New(
					"kafka_consumer_group",
					map[string]string{"group": "analytics"},
					map[string]any{
						"lag_sum":    int64(90),
						"lag_max":    int64(90),
						"partitions": int64(1),
						"topics":     int64(1),
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
			}
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}

func TestGatherCoordinatorBrokerUnknown(t *testing.T) {
	cluster := newMockCluster(t)
	defer cluster.close()

	plugin := newPlugin()
	plugin.Brokers = cluster.brokers()
	plugin.CoordinatorBrokerID = 99

	acc := gatherWithErrors(t, plugin)

	require.Empty(t, acc.GetTelegrafMetrics())
	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], "finding broker 99 failed")
}

func TestGatherBrokerDown(t *testing.T) {
	cluster := newMockCluster(t)
	defer cluster.broker1.Close()

	// Take broker 2 down while broker 1 still advertises it in its metadata
	cluster.broker2.Close()

	plugin := newPlugin()
	plugin.Brokers = []string{cluster.broker1.Addr()}
	plugin.MetricLevels = []string{levelPartition}

	acc := gatherWithErrors(t, plugin)

	// The groups of broker 1 and the partitions it leads must still be
	// collected, only the ones of broker 2 are missing.
	expected := []telegraf.Metric{
		partitionMetric("order-svc", "orders", "0", 50, 100),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())

	require.NotEmpty(t, acc.Errors)
	require.ErrorContains(t, errors.Join(acc.Errors...), "listing groups on broker 2 failed")
}

func TestGatherControllerDown(t *testing.T) {
	cluster := newMockCluster(t)
	defer cluster.broker2.Close()

	// Take the controller down while broker 2 still advertises it
	cluster.broker1.Close()

	plugin := newPlugin()
	plugin.Brokers = []string{cluster.broker2.Addr()}
	plugin.MetricLevels = []string{levelPartition}

	acc := gatherWithErrors(t, plugin)

	// Checking for deleted topics must not depend on the controller, so the
	// only errors are the groups and partitions of the dead broker.
	err := errors.Join(acc.Errors...)
	require.ErrorContains(t, err, "listing groups on broker 1 failed")
	require.NotContains(t, err.Error(), "describing topics failed")
}

func TestGatherOffsetFetchBatching(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		maxOffsetAPI  int16
		expectedBatch bool
	}{
		{name: "batched", version: "3.6.0", maxOffsetAPI: 8, expectedBatch: true},
		{name: "broker too old", version: "3.6.0", maxOffsetAPI: 7},
		{name: "configured version too old", version: "2.1.0", maxOffsetAPI: 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := newMockCluster(t)
			defer cluster.close()
			cluster.addGroup(t, "order-audit", 90)
			cluster.limitOffsetFetch(t, tt.maxOffsetAPI)

			plugin := newPlugin()
			plugin.Brokers = cluster.brokers()
			plugin.Version = tt.version
			plugin.MetricLevels = []string{levelPartition}

			acc := gather(t, plugin)

			expected := []telegraf.Metric{
				partitionMetric("order-svc", "orders", "0", 50, 100),
				partitionMetric("order-svc", "orders", "1", 80, 100),
				partitionMetric("order-svc", "events", "1", 20, 30),
				partitionMetric("order-audit", "orders", "0", 90, 100),
				partitionMetric("analytics", "orders", "0", 10, 100),
			}
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())

			// Broker 1 coordinates two groups, so it receives a single
			// request when batching and one per group otherwise.
			require.Equal(t, tt.expectedBatch, plugin.batchOffsets)
			if tt.expectedBatch {
				require.Equal(t, 1, countRequests[*sarama.OffsetFetchRequest](cluster.broker1))
			} else {
				require.Equal(t, 2, countRequests[*sarama.OffsetFetchRequest](cluster.broker1))
			}
		})
	}
}

func TestGatherOffsetFetchGroupError(t *testing.T) {
	// Run against both the batched and the per-group OffsetFetch protocol
	for _, version := range []string{"2.1.0", "3.6.0"} {
		t.Run(version, func(t *testing.T) {
			cluster := newMockCluster(t)
			defer cluster.close()

			// Broker 2 refuses to hand out the offsets of its group
			cluster.override(cluster.broker2, "OffsetFetchRequest",
				sarama.NewMockOffsetFetchResponse(t).SetError(sarama.ErrGroupAuthorizationFailed),
			)

			plugin := newPlugin()
			plugin.Brokers = cluster.brokers()
			plugin.Version = version
			plugin.MetricLevels = []string{levelPartition}

			acc := gatherWithErrors(t, plugin)

			// The other group is unaffected
			expected := []telegraf.Metric{
				partitionMetric("order-svc", "orders", "0", 50, 100),
				partitionMetric("order-svc", "orders", "1", 80, 100),
				partitionMetric("order-svc", "events", "1", 20, 30),
			}
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())

			require.Len(t, acc.Errors, 1)
			require.ErrorContains(t, acc.Errors[0], `listing offsets of group "analytics" failed`)
		})
	}
}

func TestGatherIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	const (
		topic = "lag-test"
		group = "lag-test-group"
	)
	// Messages produced into and offsets committed on each partition
	produced := []int64{10, 6}
	committed := []int64{4, 6}

	container, err := kafkacontainer.Run(t.Context(), "confluentinc/confluent-local:7.5.0")
	require.NoError(t, err)
	defer container.Terminate(t.Context()) //nolint:errcheck // ignored

	brokers, err := container.Brokers(t.Context())
	require.NoError(t, err)

	// Install the sarama logger before any sarama goroutine starts reading it
	kafka.SetLogger(testutil.Logger{}.Level())

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Partitioner = sarama.NewManualPartitioner
	client, err := sarama.NewClient(brokers, cfg)
	require.NoError(t, err)

	// Closing the admin also closes the client
	admin, err := sarama.NewClusterAdminFromClient(client)
	require.NoError(t, err)
	defer admin.Close()

	require.NoError(t, admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     int32(len(produced)),
		ReplicationFactor: 1,
	}, false))

	producer, err := sarama.NewSyncProducerFromClient(client)
	require.NoError(t, err)
	for partition, n := range produced {
		for range n {
			_, _, err := producer.SendMessage(&sarama.ProducerMessage{
				Topic:     topic,
				Partition: int32(partition),
				Value:     sarama.StringEncoder("message"),
			})
			require.NoError(t, err)
		}
	}
	require.NoError(t, producer.Close())

	// Commit offsets for the group without joining it
	offsets, err := sarama.NewOffsetManagerFromClient(group, client)
	require.NoError(t, err)
	for partition, offset := range committed {
		pom, err := offsets.ManagePartition(topic, int32(partition))
		require.NoError(t, err)
		pom.MarkOffset(offset, "")
		offsets.Commit()
		require.NoError(t, pom.Close())
	}
	require.NoError(t, offsets.Close())

	// The group never joined, so it has no members
	expected := []telegraf.Metric{
		partitionMetric(group, topic, "0", committed[0], produced[0]),
		partitionMetric(group, topic, "1", committed[1], produced[1]),
		topicMetric(group, topic, 6, 6, 2),
		groupMetric(group, 6, 6, 2, 1, 0),
	}

	// Run against both the per-group and the batched OffsetFetch protocol
	for _, version := range []string{"", "3.5.0"} {
		t.Run("version="+version, func(t *testing.T) {
			plugin := newPlugin()
			plugin.Brokers = brokers
			plugin.Version = version
			plugin.GroupsInclude = []string{group}
			logger := &testutil.CaptureLogger{}
			plugin.Log = logger

			acc := gather(t, plugin)

			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())

			// A failed batch silently falls back to per-group requests
			require.Equal(t, version != "", plugin.batchOffsets)
			for _, entry := range logger.Messages() {
				require.NotContains(t, entry.Text, "falling back")
			}
		})
	}
}
