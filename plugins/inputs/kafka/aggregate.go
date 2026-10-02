package kafka

// topicPartition identifies a single partition of a topic.
type topicPartition struct {
	topic     string
	partition int32
}

// lagAggregate accumulates per-partition lag into sums and maxima.
type lagAggregate struct {
	sum        int64
	max        int64
	partitions int64
}

func (a *lagAggregate) add(lag int64) {
	a.sum += lag
	a.partitions++
	a.max = max(a.max, lag)
}
