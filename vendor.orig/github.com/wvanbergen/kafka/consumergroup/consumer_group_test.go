package consumergroup

import (
	"fmt"
	"github.com/wvanbergen/kazoo-go"
	"testing"
)

func createTestConsumerGroupInstanceList(size int) kazoo.ConsumergroupInstanceList {
	k := make(kazoo.ConsumergroupInstanceList, size)
	for i := range k {
		k[i] = &kazoo.ConsumergroupInstance{ID: fmt.Sprintf("consumer%d", i)}
	}
	return k
}

func createTestPartitions(count int) []partitionLeader {
	p := make([]partitionLeader, count)
	for i := range p {
		p[i] = partitionLeader{id: int32(i), leader: 1, partition: &kazoo.Partition{ID: int32(i)}}
	}
	return p
}

func Test_PartitionDivision(t *testing.T) {
	consumerPartitionTestCases := [][2]int{
		// {number of Consumers, number of Partitions}
		[2]int{2, 5},
		[2]int{5, 2},
		[2]int{9, 32},
		[2]int{10, 50},
	}
	for _, v := range consumerPartitionTestCases {
		consumers := createTestConsumerGroupInstanceList(v[0])
		partitions := createTestPartitions(v[1])
		division := dividePartitionsBetweenConsumers(consumers, partitions)

		// make sure every partition is used once
		grouping := make(map[int32]struct{})
		maxConsumed := 0
		minConsumed := len(partitions) + 1
		for _, v := range division {
			if len(v) > maxConsumed {
				maxConsumed = len(v)
			}
			if len(v) < minConsumed {
				minConsumed = len(v)
			}
			for _, partition := range v {
				if _, ok := grouping[partition.ID]; ok {
					t.Errorf("PartitionDivision: Partition %v was assigned more than once!", partition.ID)
				} else {
					grouping[partition.ID] = struct{}{}
				}
			}
		}
		if len(grouping) != len(partitions) {
			t.Errorf("PartitionDivision: Expected to divide %d partitions among consumers, but only %d partitions were consumed.", len(partitions), len(grouping))
		}
		if (maxConsumed - minConsumed) > 1 {
			t.Errorf("PartitionDivision: Partitions weren't divided evenly, consumers shouldn't have a difference of more than 1 in the number of partitions consumed (was %d).", maxConsumed-minConsumed)
		}
		if minConsumed > 1 && len(consumers) != len(division) {
			t.Errorf("PartitionDivision: Partitions weren't divided evenly, some consumers didn't get any paritions even though there were %d partitions and %d consumers.", len(partitions), len(consumers))
		}
	}
}
