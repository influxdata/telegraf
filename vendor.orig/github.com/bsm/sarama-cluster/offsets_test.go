package cluster

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OffsetStash", func() {
	var subject *OffsetStash

	BeforeEach(func() {
		subject = NewOffsetStash()
	})

	It("should update", func() {
		Expect(subject.offsets).To(HaveLen(0))

		subject.MarkPartitionOffset("topic", 0, 0, "m3ta")
		Expect(subject.offsets).To(HaveLen(1))
		Expect(subject.offsets).To(HaveKeyWithValue(
			topicPartition{Topic: "topic", Partition: 0},
			offsetInfo{Offset: 0, Metadata: "m3ta"},
		))

		subject.MarkPartitionOffset("topic", 0, 200, "m3ta")
		Expect(subject.offsets).To(HaveLen(1))
		Expect(subject.offsets).To(HaveKeyWithValue(
			topicPartition{Topic: "topic", Partition: 0},
			offsetInfo{Offset: 200, Metadata: "m3ta"},
		))

		subject.MarkPartitionOffset("topic", 0, 199, "m3t@")
		Expect(subject.offsets).To(HaveLen(1))
		Expect(subject.offsets).To(HaveKeyWithValue(
			topicPartition{Topic: "topic", Partition: 0},
			offsetInfo{Offset: 200, Metadata: "m3ta"},
		))

		subject.MarkPartitionOffset("topic", 1, 300, "")
		Expect(subject.offsets).To(HaveLen(2))
		Expect(subject.offsets).To(HaveKeyWithValue(
			topicPartition{Topic: "topic", Partition: 1},
			offsetInfo{Offset: 300, Metadata: ""},
		))
	})

	It("should reset", func() {
		Expect(subject.offsets).To(HaveLen(0))

		subject.MarkPartitionOffset("topic", 0, 0, "m3ta")
		Expect(subject.offsets).To(HaveLen(1))
		Expect(subject.offsets).To(HaveKeyWithValue(
			topicPartition{Topic: "topic", Partition: 0},
			offsetInfo{Offset: 0, Metadata: "m3ta"},
		))

		subject.MarkPartitionOffset("topic", 0, 200, "m3ta")
		Expect(subject.offsets).To(HaveLen(1))
		Expect(subject.offsets).To(HaveKeyWithValue(
			topicPartition{Topic: "topic", Partition: 0},
			offsetInfo{Offset: 200, Metadata: "m3ta"},
		))

		subject.ResetPartitionOffset("topic", 0, 199, "m3t@")
		Expect(subject.offsets).To(HaveLen(1))
		Expect(subject.offsets).To(HaveKeyWithValue(
			topicPartition{Topic: "topic", Partition: 0},
			offsetInfo{Offset: 199, Metadata: "m3t@"},
		))

		subject.MarkPartitionOffset("topic", 1, 300, "")
		Expect(subject.offsets).To(HaveLen(2))
		Expect(subject.offsets).To(HaveKeyWithValue(
			topicPartition{Topic: "topic", Partition: 1},
			offsetInfo{Offset: 300, Metadata: ""},
		))

		subject.ResetPartitionOffset("topic", 1, 200, "m3t@")
		Expect(subject.offsets).To(HaveLen(2))
		Expect(subject.offsets).To(HaveKeyWithValue(
			topicPartition{Topic: "topic", Partition: 1},
			offsetInfo{Offset: 200, Metadata: "m3t@"},
		))

	})

})
