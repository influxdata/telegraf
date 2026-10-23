package kafka

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	kafkacontainer "github.com/testcontainers/testcontainers-go/modules/kafka"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitDefault(t *testing.T) {
	plugin := &Kafka{
		Brokers:             []string{"localhost:9092"},
		MetricLevels:        []string{"partition", "topic", "group"},
		CoordinatorBrokerID: -1,
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *Kafka
		expected string
	}{
		{
			name: "no brokers",
			plugin: &Kafka{
				MetricLevels:        []string{"partition"},
				CoordinatorBrokerID: -1,
			},
			expected: "brokers must not be empty",
		},
		{
			name: "no metric levels",
			plugin: &Kafka{
				Brokers:             []string{"localhost:9092"},
				CoordinatorBrokerID: -1,
			},
			expected: "metric_levels must not be empty",
		},
		{
			name: "invalid metric level",
			plugin: &Kafka{
				Brokers:             []string{"localhost:9092"},
				MetricLevels:        []string{"cluster"},
				CoordinatorBrokerID: -1,
			},
			expected: `invalid metric level "cluster"`,
		},
		{
			name: "invalid group filter",
			plugin: &Kafka{
				Brokers:             []string{"localhost:9092"},
				MetricLevels:        []string{"partition"},
				GroupsInclude:       []string{"[invalid"},
				CoordinatorBrokerID: -1,
			},
			expected: "creating group filter failed",
		},
		{
			name: "invalid version",
			plugin: &Kafka{
				Brokers:             []string{"localhost:9092"},
				MetricLevels:        []string{"partition"},
				CoordinatorBrokerID: -1,
				Config:              kafka.Config{Version: "not-a-version"},
			},
			expected: "setting config failed",
		},
		{
			name: "version too old",
			plugin: &Kafka{
				Brokers:             []string{"localhost:9092"},
				MetricLevels:        []string{"partition"},
				CoordinatorBrokerID: -1,
				Config:              kafka.Config{Version: "0.10.2.0"},
			},
			expected: "0.11.0.0 or greater is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = testutil.Logger{}
			require.ErrorContains(t, tt.plugin.Init(), tt.expected)
		})
	}
}

func TestInitDisablesTopicCreation(t *testing.T) {
	plugin := &Kafka{
		Brokers:             []string{"localhost:9092"},
		MetricLevels:        []string{"partition", "topic", "group"},
		CoordinatorBrokerID: -1,
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.False(t, plugin.config.Metadata.AllowAutoTopicCreation)
}

func TestInitBatchOffsets(t *testing.T) {
	for version, expected := range map[string]bool{"2.8.0": false, "3.0.0": true, "3.6.0": true} {
		t.Run(version, func(t *testing.T) {
			plugin := &Kafka{
				Brokers:             []string{"localhost:9092"},
				MetricLevels:        []string{"partition", "topic", "group"},
				CoordinatorBrokerID: -1,
				Log:                 testutil.Logger{},
				Config:              kafka.Config{Version: version},
			}
			require.NoError(t, plugin.Init())
			require.Equal(t, expected, plugin.batchOffsets)
		})
	}
}

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Prepare the influx parser for expectations
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			configFilename := filepath.Join(testcasePath, "telegraf.conf")
			clusterFilename := filepath.Join(testcasePath, "cluster.json")
			expectedFilename := filepath.Join(testcasePath, "expected.out")
			expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Read the expected errors if any
			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				var err error
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			// Configure and initialize the plugin. Init replaces the global
			// sarama logger, so do it before the mock brokers start.
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)
			plugin := cfg.Inputs[0].Input.(*Kafka)
			require.NoError(t, plugin.Init())

			// Setup the cluster and point the plugin to it
			cluster := newMockCluster(t, clusterFilename)
			defer cluster.close()
			plugin.Brokers = cluster.addresses()

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect the data and treat the error returned by Gather like
			// the accumulated ones
			if err := plugin.Gather(&acc); err != nil {
				acc.AddError(err)
			}

			// Check the result
			actualErrors := make([]string, 0, len(acc.Errors))
			for _, err := range acc.Errors {
				actualErrors = append(actualErrors, err.Error())
			}
			require.ElementsMatch(t, expectedErrors, actualErrors)

			options := []cmp.Option{
				testutil.SortMetrics(),
				testutil.IgnoreTime(),
				testutil.IgnoreType(),
			}
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), options...)
		})
	}
}

func TestGatherSkipsDescribeWithoutGroupLevel(t *testing.T) {
	cluster := newMockCluster(t, "testdata/cluster.json")
	defer cluster.close()

	plugin := &Kafka{
		Brokers:             cluster.addresses(),
		MetricLevels:        []string{"partition", "topic"},
		CoordinatorBrokerID: -1,
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)
	require.NotEmpty(t, acc.GetTelegrafMetrics())

	// The member count is only used on the group level, so no group must
	// have been described.
	for id, broker := range cluster.brokers {
		for _, rr := range broker.History() {
			_, described := rr.Request.(*sarama.DescribeGroupsRequest)
			require.Falsef(t, described, "broker %d received a DescribeGroupsRequest", id)
		}
	}
}

func TestGatherBrokerDown(t *testing.T) {
	cluster := newMockCluster(t, "testdata/cluster.json")
	defer cluster.brokers[1].Close()

	// Take broker 2 down while broker 1 still advertises it in its metadata
	cluster.brokers[2].Close()

	plugin := &Kafka{
		Brokers:             []string{cluster.brokers[1].Addr()},
		MetricLevels:        []string{"partition"},
		CoordinatorBrokerID: -1,
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))

	// The groups of broker 1 and the partitions it leads must still be
	// collected, only the ones of broker 2 are missing.
	expected := []telegraf.Metric{
		metric.New(
			"kafka_consumer_group_partition",
			map[string]string{"group": "order-svc", "topic": "orders", "partition": "0"},
			map[string]any{"committed_offset": int64(50), "log_end_offset": int64(100), "lag": int64(50)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())

	require.NotEmpty(t, acc.Errors)
	require.ErrorContains(t, errors.Join(acc.Errors...), "listing groups on broker 2 failed")
}

func TestGatherControllerDown(t *testing.T) {
	cluster := newMockCluster(t, "testdata/cluster.json")
	defer cluster.brokers[2].Close()

	// Take the controller down while broker 2 still advertises it
	cluster.brokers[1].Close()

	plugin := &Kafka{
		Brokers:             []string{cluster.brokers[2].Addr()},
		MetricLevels:        []string{"partition"},
		CoordinatorBrokerID: -1,
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))

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
			// Broker 1 coordinates two groups
			cluster := newMockCluster(t, "testdata/cluster_two_groups.json")
			defer cluster.close()

			// Make the brokers advertise OffsetFetch up to the given version only
			for id := range cluster.brokers {
				cluster.override(id, "ApiVersionsRequest",
					sarama.NewMockApiVersionsResponse(t).SetApiKeys([]sarama.ApiVersionsResponseKey{
						{ApiKey: 2, MinVersion: 0, MaxVersion: 20},
						{ApiKey: 3, MinVersion: 0, MaxVersion: 20},
						{ApiKey: 9, MinVersion: 0, MaxVersion: tt.maxOffsetAPI},
						{ApiKey: 10, MinVersion: 0, MaxVersion: 20},
						{ApiKey: 15, MinVersion: 0, MaxVersion: 20},
						{ApiKey: 16, MinVersion: 0, MaxVersion: 20},
					}),
				)
			}

			plugin := &Kafka{
				Brokers:             cluster.addresses(),
				MetricLevels:        []string{"partition"},
				CoordinatorBrokerID: -1,
				Log:                 testutil.Logger{},
				Config:              kafka.Config{Version: tt.version},
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			require.NoError(t, plugin.Gather(&acc))
			require.Empty(t, acc.Errors)
			require.Len(t, acc.GetTelegrafMetrics(), 5)

			// Broker 1 receives a single request when batching and one per
			// group otherwise.
			require.Equal(t, tt.expectedBatch, plugin.batchOffsets)
			var requests int
			for _, rr := range cluster.brokers[1].History() {
				if _, ok := rr.Request.(*sarama.OffsetFetchRequest); ok {
					requests++
				}
			}
			if tt.expectedBatch {
				require.Equal(t, 1, requests)
			} else {
				require.Equal(t, 2, requests)
			}
		})
	}
}

func TestStartControllerUnavailable(t *testing.T) {
	cluster := newMockCluster(t, "testdata/cluster.json")
	defer cluster.close()

	// Advertise a controller which is not part of the cluster, as brokers do
	// during a controller election
	metadata := sarama.NewMockMetadataResponse(t).SetController(99)
	for id, broker := range cluster.brokers {
		metadata.SetBroker(broker.Addr(), id)
	}
	for id := range cluster.brokers {
		cluster.override(id, "MetadataRequest", metadata)
	}

	plugin := &Kafka{
		Brokers:             cluster.addresses(),
		MetricLevels:        []string{"partition", "topic", "group"},
		CoordinatorBrokerID: -1,
		Log:                 testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	err := plugin.Start(&acc)
	require.ErrorIs(t, err, sarama.ErrControllerNotAvailable)

	var serr *internal.StartupError
	require.ErrorAs(t, err, &serr)
	require.True(t, serr.Retry)
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

	// Run against both the per-group and the batched OffsetFetch protocol.
	// Init replaces the global sarama logger, so initialize the plugins before
	// any sarama goroutine starts reading it.
	plugins := make(map[string]*Kafka)
	loggers := make(map[string]*testutil.CaptureLogger)
	for _, version := range []string{"", "3.5.0"} {
		logger := &testutil.CaptureLogger{}
		plugin := &Kafka{
			Brokers:             brokers,
			GroupsInclude:       []string{group},
			MetricLevels:        []string{"partition", "topic", "group"},
			CoordinatorBrokerID: -1,
			Log:                 logger,
			Config:              kafka.Config{Version: version},
		}
		require.NoError(t, plugin.Init())
		plugins[version] = plugin
		loggers[version] = logger
	}

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
		metric.New(
			"kafka_consumer_group_partition",
			map[string]string{"group": group, "topic": topic, "partition": "0"},
			map[string]any{"committed_offset": int64(4), "log_end_offset": int64(10), "lag": int64(6)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"kafka_consumer_group_partition",
			map[string]string{"group": group, "topic": topic, "partition": "1"},
			map[string]any{"committed_offset": int64(6), "log_end_offset": int64(6), "lag": int64(0)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"kafka_consumer_group_topic",
			map[string]string{"group": group, "topic": topic},
			map[string]any{"lag_sum": int64(6), "lag_max": int64(6), "partitions": int64(2)},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"kafka_consumer_group",
			map[string]string{"group": group},
			map[string]any{
				"lag_sum":    int64(6),
				"lag_max":    int64(6),
				"partitions": int64(2),
				"topics":     int64(1),
				"members":    int64(0),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	for version, plugin := range plugins {
		t.Run("version="+version, func(t *testing.T) {
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			require.NoError(t, plugin.Gather(&acc))
			require.Empty(t, acc.Errors)
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())

			// A failed batch silently falls back to per-group requests
			require.Equal(t, version != "", plugin.batchOffsets)
			for _, entry := range loggers[version].Messages() {
				require.NotContains(t, entry.Text, "falling back")
			}
		})
	}
}

// clusterDefinition describes a mocked Kafka cluster. Topics with offsets
// committed by a group but missing from the topic list are treated as deleted.
type clusterDefinition struct {
	Brokers    []int32                                  `json:"brokers"`
	Controller int32                                    `json:"controller"`
	Topics     map[string]map[int32]partitionDefinition `json:"topics"`
	Groups     map[string]groupDefinition               `json:"groups"`
}

type partitionDefinition struct {
	Leader       int32 `json:"leader"`
	LogEndOffset int64 `json:"log_end_offset"`
}

type groupDefinition struct {
	Coordinator int32 `json:"coordinator"`
	// State reported by DescribeGroups, "Stable" if the group has members
	// and "Empty" otherwise if not set.
	State   string `json:"state"`
	Members int    `json:"members"`
	// Error names of the DescribeGroups and OffsetFetch responses. Mocked
	// brokers can only fail OffsetFetch for all groups they coordinate.
	DescribeError    string `json:"describe_error"`
	OffsetFetchError string `json:"offset_fetch_error"`
	// Committed offsets by topic and partition, -1 for never committed
	Offsets map[string]map[int32]int64 `json:"offsets"`
}

// Kafka protocol errors usable in cluster definitions by name
var kafkaErrors = map[string]sarama.KError{
	"":                             sarama.ErrNoError,
	"COORDINATOR_LOAD_IN_PROGRESS": sarama.ErrOffsetsLoadInProgress,
	"GROUP_AUTHORIZATION_FAILED":   sarama.ErrGroupAuthorizationFailed,
	"GROUP_ID_NOT_FOUND":           sarama.ErrGroupIDNotFound,
	"INVALID_GROUP_ID":             sarama.ErrInvalidGroupId,
}

// mockCluster serves the state of a cluster definition from mocked brokers.
type mockCluster struct {
	brokers  map[int32]*sarama.MockBroker
	handlers map[int32]map[string]sarama.MockResponse
}

func newMockCluster(t *testing.T, filename string) *mockCluster {
	t.Helper()

	f, err := os.Open(filename)
	require.NoError(t, err)
	defer f.Close()

	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	var def clusterDefinition
	require.NoError(t, decoder.Decode(&def))

	// Install the sarama logger before the mock brokers start reading it
	kafka.SetLogger(testutil.Logger{}.Level())

	cluster := &mockCluster{
		brokers:  make(map[int32]*sarama.MockBroker, len(def.Brokers)),
		handlers: make(map[int32]map[string]sarama.MockResponse, len(def.Brokers)),
	}
	for _, id := range def.Brokers {
		cluster.brokers[id] = sarama.NewMockBroker(t, id)
	}

	metadata := sarama.NewMockMetadataResponse(t).SetController(def.Controller)
	for _, id := range def.Brokers {
		metadata.SetBroker(cluster.brokers[id].Addr(), id)
	}
	for topic, partitions := range def.Topics {
		for partition, p := range partitions {
			metadata.SetLeader(topic, partition, p.Leader)
		}
	}

	coordinators := sarama.NewMockFindCoordinatorResponse(t)
	for name, group := range def.Groups {
		broker, ok := cluster.brokers[group.Coordinator]
		require.Truef(t, ok, "coordinator %d of group %q is not a broker", group.Coordinator, name)
		coordinators.SetCoordinator(sarama.CoordinatorGroup, name, broker)
	}

	for _, id := range def.Brokers {
		listGroups := sarama.NewMockListGroupsResponse(t)
		describeGroups := sarama.NewMockDescribeGroupsResponse(t)
		offsetFetch := sarama.NewMockOffsetFetchResponse(t)
		var offsetFetchError *sarama.KError
		for _, name := range slices.Sorted(maps.Keys(def.Groups)) {
			group := def.Groups[name]
			if group.Coordinator != id {
				continue
			}
			listGroups.AddGroup(name, "consumer")

			describeError, ok := kafkaErrors[group.DescribeError]
			require.Truef(t, ok, "unknown describe error %q of group %q", group.DescribeError, name)
			state := group.State
			if state == "" {
				state = "Empty"
				if group.Members > 0 {
					state = "Stable"
				}
			}
			members := make(map[string]*sarama.GroupMemberDescription, group.Members)
			for i := range group.Members {
				member := fmt.Sprintf("member-%d", i)
				members[member] = &sarama.GroupMemberDescription{MemberId: member, ClientId: name}
			}
			describeGroups.AddGroupDescription(name, &sarama.GroupDescription{
				GroupId:   name,
				State:     state,
				Members:   members,
				ErrorCode: int16(describeError),
			})

			fetchError, ok := kafkaErrors[group.OffsetFetchError]
			require.Truef(t, ok, "unknown offset fetch error %q of group %q", group.OffsetFetchError, name)
			if offsetFetchError != nil {
				require.Equalf(t, *offsetFetchError, fetchError, "groups coordinated by broker %d must share the offset fetch error", id)
			}
			offsetFetchError = &fetchError
			for topic, partitions := range group.Offsets {
				for partition, offset := range partitions {
					offsetFetch.SetOffset(name, topic, partition, offset, "", sarama.ErrNoError)
				}
			}
		}
		if offsetFetchError != nil {
			offsetFetch.SetError(*offsetFetchError)
		}

		listOffsets := sarama.NewMockOffsetResponse(t)
		for topic, partitions := range def.Topics {
			for partition, p := range partitions {
				if p.Leader == id {
					listOffsets.SetOffset(topic, partition, sarama.OffsetNewest, p.LogEndOffset)
				}
			}
		}

		cluster.handlers[id] = map[string]sarama.MockResponse{
			"ApiVersionsRequest": sarama.NewMockApiVersionsResponse(t).SetApiKeys([]sarama.ApiVersionsResponseKey{
				{ApiKey: 2, MinVersion: 0, MaxVersion: 20},
				{ApiKey: 3, MinVersion: 0, MaxVersion: 20},
				{ApiKey: 9, MinVersion: 0, MaxVersion: 20},
				{ApiKey: 10, MinVersion: 0, MaxVersion: 20},
				{ApiKey: 15, MinVersion: 0, MaxVersion: 20},
				{ApiKey: 16, MinVersion: 0, MaxVersion: 20},
			}),
			"MetadataRequest":        metadata,
			"FindCoordinatorRequest": coordinators,
			"ListGroupsRequest":      listGroups,
			"DescribeGroupsRequest":  describeGroups,
			"OffsetFetchRequest":     offsetFetch,
			"OffsetRequest":          listOffsets,
		}
		cluster.brokers[id].SetHandlerByMap(cluster.handlers[id])
	}

	return cluster
}

// addresses returns the addresses of all brokers ordered by broker ID.
func (c *mockCluster) addresses() []string {
	addresses := make([]string, 0, len(c.brokers))
	for _, id := range slices.Sorted(maps.Keys(c.brokers)) {
		addresses = append(addresses, c.brokers[id].Addr())
	}
	return addresses
}

// override replaces the response of a broker for one request type.
func (c *mockCluster) override(id int32, request string, response sarama.MockResponse) {
	c.handlers[id][request] = response
	c.brokers[id].SetHandlerByMap(c.handlers[id])
}

func (c *mockCluster) close() {
	for _, broker := range c.brokers {
		broker.Close()
	}
}
