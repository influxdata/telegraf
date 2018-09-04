package cluster

import (
	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("partitionConsumer", func() {
	var subject *partitionConsumer

	BeforeEach(func() {
		var err error
		subject, err = newPartitionConsumer(&mockConsumer{}, "topic", 0, offsetInfo{2000, "m3ta"}, sarama.OffsetOldest)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		close(subject.dead)
		Expect(subject.Close()).NotTo(HaveOccurred())
	})

	It("should set state", func() {
		Expect(subject.State()).To(Equal(partitionState{
			Info: offsetInfo{2000, "m3ta"},
		}))
	})

	It("should recover from default offset if requested offset is out of bounds", func() {
		pc, err := newPartitionConsumer(&mockConsumer{}, "topic", 0, offsetInfo{200, "m3ta"}, sarama.OffsetOldest)
		Expect(err).NotTo(HaveOccurred())
		defer pc.Close()
		close(pc.dead)

		state := pc.State()
		Expect(state.Info.Offset).To(Equal(int64(-1)))
		Expect(state.Info.Metadata).To(Equal("m3ta"))
	})

	It("should update state", func() {
		subject.MarkOffset(2001, "met@") // should set state
		Expect(subject.State()).To(Equal(partitionState{
			Info:  offsetInfo{2001, "met@"},
			Dirty: true,
		}))

		subject.MarkCommitted(2001) // should reset dirty status
		Expect(subject.State()).To(Equal(partitionState{
			Info: offsetInfo{2001, "met@"},
		}))

		subject.MarkOffset(2001, "me7a") // should not update state
		Expect(subject.State()).To(Equal(partitionState{
			Info: offsetInfo{2001, "met@"},
		}))

		subject.MarkOffset(2002, "me7a") // should bump state
		Expect(subject.State()).To(Equal(partitionState{
			Info:  offsetInfo{2002, "me7a"},
			Dirty: true,
		}))

		// After committing a later offset, try rewinding back to earlier offset with new metadata.
		subject.ResetOffset(2001, "met@")
		Expect(subject.State()).To(Equal(partitionState{
			Info:  offsetInfo{2001, "met@"},
			Dirty: true,
		}))

		subject.MarkCommitted(2001) // should not unset state
		Expect(subject.State()).To(Equal(partitionState{
			Info: offsetInfo{2001, "met@"},
		}))

		subject.MarkOffset(2002, "me7a") // should bump state
		Expect(subject.State()).To(Equal(partitionState{
			Info:  offsetInfo{2002, "me7a"},
			Dirty: true,
		}))

		subject.MarkCommitted(2002)
		Expect(subject.State()).To(Equal(partitionState{
			Info: offsetInfo{2002, "me7a"},
		}))
	})

	It("should not fail when nil", func() {
		blank := (*partitionConsumer)(nil)
		Expect(func() {
			_ = blank.State()
			blank.MarkOffset(2001, "met@")
			blank.MarkCommitted(2001)
		}).NotTo(Panic())
	})

})

var _ = Describe("partitionMap", func() {
	var subject *partitionMap

	BeforeEach(func() {
		subject = newPartitionMap()
	})

	It("should fetch/store", func() {
		Expect(subject.Fetch("topic", 0)).To(BeNil())

		pc, err := newPartitionConsumer(&mockConsumer{}, "topic", 0, offsetInfo{2000, "m3ta"}, sarama.OffsetNewest)
		Expect(err).NotTo(HaveOccurred())

		subject.Store("topic", 0, pc)
		Expect(subject.Fetch("topic", 0)).To(Equal(pc))
		Expect(subject.Fetch("topic", 1)).To(BeNil())
		Expect(subject.Fetch("other", 0)).To(BeNil())
	})

	It("should return info", func() {
		pc0, err := newPartitionConsumer(&mockConsumer{}, "topic", 0, offsetInfo{2000, "m3ta"}, sarama.OffsetNewest)
		Expect(err).NotTo(HaveOccurred())
		pc1, err := newPartitionConsumer(&mockConsumer{}, "topic", 1, offsetInfo{2000, "m3ta"}, sarama.OffsetNewest)
		Expect(err).NotTo(HaveOccurred())
		subject.Store("topic", 0, pc0)
		subject.Store("topic", 1, pc1)

		info := subject.Info()
		Expect(info).To(HaveLen(1))
		Expect(info).To(HaveKeyWithValue("topic", []int32{0, 1}))
	})

	It("should create snapshots", func() {
		pc0, err := newPartitionConsumer(&mockConsumer{}, "topic", 0, offsetInfo{2000, "m3ta"}, sarama.OffsetNewest)
		Expect(err).NotTo(HaveOccurred())
		pc1, err := newPartitionConsumer(&mockConsumer{}, "topic", 1, offsetInfo{2000, "m3ta"}, sarama.OffsetNewest)
		Expect(err).NotTo(HaveOccurred())

		subject.Store("topic", 0, pc0)
		subject.Store("topic", 1, pc1)
		subject.Fetch("topic", 1).MarkOffset(2001, "met@")

		Expect(subject.Snapshot()).To(Equal(map[topicPartition]partitionState{
			{"topic", 0}: {Info: offsetInfo{2000, "m3ta"}, Dirty: false},
			{"topic", 1}: {Info: offsetInfo{2001, "met@"}, Dirty: true},
		}))
	})

})
