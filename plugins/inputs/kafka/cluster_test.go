package kafka

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/testutil"
)

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
			"ApiVersionsRequest":     sarama.NewMockApiVersionsResponse(t),
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
