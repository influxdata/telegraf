package kazoo

import (
	"sort"
	"testing"
)

func TestPartition(t *testing.T) {
	topic := &Topic{Name: "test"}
	partition := topic.Partition(1, []int32{1, 2, 3})

	if key := partition.Key(); key != "test/1" {
		t.Error("Unexpected partition key", key)
	}

	if partition.Topic() != topic {
		t.Error("Expected Topic() to return the topic the partition was created from.")
	}

	if pr := partition.PreferredReplica(); pr != 1 {
		t.Error("Expected 1 to be the preferred replica, but found", pr)
	}

	partitionWithoutReplicas := topic.Partition(1, nil)
	if pr := partitionWithoutReplicas.PreferredReplica(); pr != -1 {
		t.Error("Expected -1 to be returned if the partition does not have replicas, but found", pr)
	}
}

func TestTopicList(t *testing.T) {
	topics := TopicList{
		&Topic{Name: "foo"},
		&Topic{Name: "bar"},
		&Topic{Name: "baz"},
	}

	sort.Sort(topics)

	if topics[0].Name != "bar" || topics[1].Name != "baz" || topics[2].Name != "foo" {
		t.Error("Unexpected order after sorting topic list", topics)
	}

	topic := topics.Find("foo")
	if topic != topics[2] {
		t.Error("Should have found foo topic from the list")
	}
}

func TestPartitionList(t *testing.T) {
	var (
		topic1 = &Topic{Name: "1"}
		topic2 = &Topic{Name: "2"}
	)

	var (
		partition21 = topic2.Partition(1, nil)
		partition12 = topic1.Partition(2, nil)
		partition11 = topic1.Partition(1, nil)
	)

	partitions := PartitionList{partition21, partition12, partition11}
	sort.Sort(partitions)

	if partitions[0] != partition11 || partitions[1] != partition12 || partitions[2] != partition21 {
		t.Error("Unexpected order after sorting topic list", partitions)
	}
}

func TestGeneratePartitionAssignments(t *testing.T) {
	// check for errors
	tests := []struct {
		brokers           []int32
		partitionCount    int
		replicationFactor int
		err               error
	}{
		{[]int32{1, 2}, -1, 1, ErrInvalidPartitionCount},
		{[]int32{1, 2}, 0, 1, ErrInvalidPartitionCount},
		{[]int32{}, 1, 1, ErrInvalidReplicationFactor},
		{[]int32{1, 2}, 1, -1, ErrInvalidReplicationFactor},
		{[]int32{1, 2}, 2, 0, ErrInvalidReplicationFactor},
		{[]int32{1, 2}, 3, 3, ErrInvalidReplicationFactor},
		{[]int32{1, 2}, 2, 1, nil},
		{[]int32{1, 2}, 10, 2, nil},
		{[]int32{1}, 10, 1, nil},
		{[]int32{1, 2, 3, 4, 5}, 1, 1, nil},
		{[]int32{1, 2, 3, 4, 5}, 1, 3, nil},
		{[]int32{1, 2, 3, 4, 5}, 10, 2, nil},
	}

	for testIdx, test := range tests {
		topic := &Topic{Name: "t"}

		res, err := topic.generatePartitionAssignments(test.brokers, test.partitionCount, test.replicationFactor)
		if err != test.err {
			t.Errorf("Incorrect error for test %d. Expected (%v) got (%v)", testIdx, test.err, err)
		} else if err == nil {
			// proper number of paritions
			if len(res) != test.partitionCount {
				t.Errorf("Wrong number of partitions assigned in test %d. Expected %d got %d", testIdx, test.partitionCount, len(res))
			}
			// ensure all petitions are assigned and that they have
			// the right number of non-overlapping brokers
			for i, part := range res {
				if part == nil {
					t.Errorf("Partition %d is nil in test %d", i, testIdx)
					continue
				}
				if len(part.Replicas) != test.replicationFactor {
					t.Errorf("Partition %d does not have the correct number of brokers in test %d. Expected %d got %d", i, testIdx, test.replicationFactor, len(part.Replicas))
				}
				replicaMap := make(map[int32]bool, test.replicationFactor)
				for _, r := range part.Replicas {
					// ensure broker is in initial broker list
					found := false
					for _, broker := range test.brokers {
						if broker == r {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Partition %d has an invalid broker id %d in test %d", i, r, testIdx)
					}
					replicaMap[r] = true
				}
				if len(replicaMap) != len(part.Replicas) {
					t.Errorf("Partition %d has overlapping broker assignments (%v) in test %d", i, part.Replicas, testIdx)
				}
			}
		}
	}
}

func TestValidatePartitionAssignments(t *testing.T) {
	// check for errors
	tests := []struct {
		brokers    []int32
		partitions PartitionList
		err        error
	}{
		{[]int32{1}, PartitionList{}, ErrInvalidPartitionCount},

		{[]int32{1}, PartitionList{
			{ID: 0, Replicas: []int32{}},
		}, ErrInvalidReplicationFactor},

		{[]int32{1, 2}, PartitionList{
			{ID: 0, Replicas: []int32{1}},
			{ID: 1, Replicas: []int32{1, 2}},
		}, ErrInvalidReplicaCount},

		{[]int32{1, 2}, PartitionList{
			{ID: 0, Replicas: []int32{1, 2}},
			{ID: 1, Replicas: []int32{1}},
		}, ErrInvalidReplicaCount},

		{[]int32{1, 2}, PartitionList{
			{ID: 0, Replicas: []int32{1, 2}},
			{ID: 1, Replicas: []int32{2, 2}},
		}, ErrReplicaBrokerOverlap},

		{[]int32{1, 2}, PartitionList{
			{ID: 0, Replicas: []int32{1, 3}},
			{ID: 1, Replicas: []int32{2, 1}},
		}, ErrInvalidBroker},

		{[]int32{1, 2, 3}, PartitionList{
			{ID: 1, Replicas: []int32{1, 3}},
			{ID: 2, Replicas: []int32{2, 1}},
		}, ErrMissingPartitionID},

		{[]int32{1, 2, 3}, PartitionList{
			{ID: 0, Replicas: []int32{1, 3}},
			{ID: 2, Replicas: []int32{2, 1}},
		}, ErrMissingPartitionID},

		{[]int32{1, 2, 3}, PartitionList{
			{ID: 0, Replicas: []int32{1, 3}},
			{ID: 0, Replicas: []int32{1, 3}},
			{ID: 2, Replicas: []int32{2, 1}},
		}, ErrDuplicatePartitionID},

		{[]int32{1}, PartitionList{
			{ID: 0, Replicas: []int32{1}},
		}, nil},

		{[]int32{1}, PartitionList{
			{ID: 0, Replicas: []int32{1}},
			{ID: 1, Replicas: []int32{1}},
			{ID: 2, Replicas: []int32{1}},
		}, nil},

		{[]int32{1, 2, 3}, PartitionList{
			{ID: 0, Replicas: []int32{1, 2}},
			{ID: 1, Replicas: []int32{2, 3}},
			{ID: 2, Replicas: []int32{3, 1}},
		}, nil},
	}

	for testIdx, test := range tests {
		topic := &Topic{Name: "t"}

		err := topic.validatePartitionAssignments(test.brokers, test.partitions)
		if err != test.err {
			t.Errorf("Incorrect error for test %d. Expected (%v) got (%v)", testIdx, test.err, err)
		} else if err == nil {
		}
	}
}
