package cluster

import (
	"fmt"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Consumer", func() {

	var newConsumerOf = func(group string, topics ...string) (*Consumer, error) {
		config := NewConfig()
		config.Consumer.Return.Errors = true
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
		return NewConsumer(testKafkaAddrs, group, topics, config)
	}

	var subscriptionsOf = func(c *Consumer) GomegaAsyncAssertion {
		return Eventually(func() map[string][]int32 {
			return c.Subscriptions()
		}, "10s", "100ms")
	}

	It("should init and share", func() {
		// start CS1
		cs1, err := newConsumerOf(testGroup, testTopics...)
		Expect(err).NotTo(HaveOccurred())

		// CS1 should consume all 8 partitions
		subscriptionsOf(cs1).Should(Equal(map[string][]int32{
			"topic-a": {0, 1, 2, 3},
			"topic-b": {0, 1, 2, 3},
		}))

		// start CS2
		cs2, err := newConsumerOf(testGroup, testTopics...)
		Expect(err).NotTo(HaveOccurred())
		defer cs2.Close()

		// CS1 and CS2 should consume 4 partitions each
		subscriptionsOf(cs1).Should(HaveLen(2))
		subscriptionsOf(cs1).Should(HaveKeyWithValue("topic-a", HaveLen(2)))
		subscriptionsOf(cs1).Should(HaveKeyWithValue("topic-b", HaveLen(2)))

		subscriptionsOf(cs2).Should(HaveLen(2))
		subscriptionsOf(cs2).Should(HaveKeyWithValue("topic-a", HaveLen(2)))
		subscriptionsOf(cs2).Should(HaveKeyWithValue("topic-b", HaveLen(2)))

		// shutdown CS1, now CS2 should consume all 8 partitions
		Expect(cs1.Close()).NotTo(HaveOccurred())
		subscriptionsOf(cs2).Should(Equal(map[string][]int32{
			"topic-a": {0, 1, 2, 3},
			"topic-b": {0, 1, 2, 3},
		}))
	})

	It("should allow more consumers than partitions", func() {
		cs1, err := newConsumerOf(testGroup, "topic-a")
		Expect(err).NotTo(HaveOccurred())
		defer cs1.Close()
		cs2, err := newConsumerOf(testGroup, "topic-a")
		Expect(err).NotTo(HaveOccurred())
		defer cs2.Close()
		cs3, err := newConsumerOf(testGroup, "topic-a")
		Expect(err).NotTo(HaveOccurred())
		defer cs3.Close()
		cs4, err := newConsumerOf(testGroup, "topic-a")
		Expect(err).NotTo(HaveOccurred())

		// start 4 consumers, one for each partition
		subscriptionsOf(cs1).Should(HaveKeyWithValue("topic-a", HaveLen(1)))
		subscriptionsOf(cs2).Should(HaveKeyWithValue("topic-a", HaveLen(1)))
		subscriptionsOf(cs3).Should(HaveKeyWithValue("topic-a", HaveLen(1)))
		subscriptionsOf(cs4).Should(HaveKeyWithValue("topic-a", HaveLen(1)))

		// add a 5th consumer
		cs5, err := newConsumerOf(testGroup, "topic-a")
		Expect(err).NotTo(HaveOccurred())
		defer cs5.Close()

		// make sure no errors occurred
		Expect(cs1.Errors()).ShouldNot(Receive())
		Expect(cs2.Errors()).ShouldNot(Receive())
		Expect(cs3.Errors()).ShouldNot(Receive())
		Expect(cs4.Errors()).ShouldNot(Receive())
		Expect(cs5.Errors()).ShouldNot(Receive())

		// close 4th, make sure the 5th takes over
		Expect(cs4.Close()).To(Succeed())
		subscriptionsOf(cs1).Should(HaveKeyWithValue("topic-a", HaveLen(1)))
		subscriptionsOf(cs2).Should(HaveKeyWithValue("topic-a", HaveLen(1)))
		subscriptionsOf(cs3).Should(HaveKeyWithValue("topic-a", HaveLen(1)))
		subscriptionsOf(cs4).Should(BeEmpty())
		subscriptionsOf(cs5).Should(HaveKeyWithValue("topic-a", HaveLen(1)))

		// there should still be no errors
		Expect(cs1.Errors()).ShouldNot(Receive())
		Expect(cs2.Errors()).ShouldNot(Receive())
		Expect(cs3.Errors()).ShouldNot(Receive())
		Expect(cs4.Errors()).ShouldNot(Receive())
		Expect(cs5.Errors()).ShouldNot(Receive())
	})

	It("should be allowed to subscribe to partitions via white/black-lists", func() {
		config := NewConfig()
		config.Consumer.Return.Errors = true
		config.Group.Topics.Whitelist = regexp.MustCompile(`topic-\w+`)
		config.Group.Topics.Blacklist = regexp.MustCompile(`[bcd]$`)

		cs, err := NewConsumer(testKafkaAddrs, testGroup, nil, config)
		Expect(err).NotTo(HaveOccurred())
		defer cs.Close()

		subscriptionsOf(cs).Should(Equal(map[string][]int32{
			"topic-a": {0, 1, 2, 3},
		}))
	})

	It("should receive rebalance notifications", func() {
		config := NewConfig()
		config.Consumer.Return.Errors = true
		config.Group.Return.Notifications = true

		cs, err := NewConsumer(testKafkaAddrs, testGroup, testTopics, config)
		Expect(err).NotTo(HaveOccurred())
		defer cs.Close()

		select {
		case n := <-cs.Notifications():
			Expect(n).To(Equal(&Notification{
				Type:    RebalanceStart,
				Current: map[string][]int32{},
			}))
		case err := <-cs.Errors():
			Expect(err).NotTo(HaveOccurred())
		case <-cs.Messages():
			Fail("expected notification to arrive before message")
		}

		select {
		case n := <-cs.Notifications():
			Expect(n).To(Equal(&Notification{
				Type: RebalanceOK,
				Claimed: map[string][]int32{
					"topic-a": {0, 1, 2, 3},
					"topic-b": {0, 1, 2, 3},
				},
				Released: map[string][]int32{},
				Current: map[string][]int32{
					"topic-a": {0, 1, 2, 3},
					"topic-b": {0, 1, 2, 3},
				},
			}))
		case err := <-cs.Errors():
			Expect(err).NotTo(HaveOccurred())
		case <-cs.Messages():
			Fail("expected notification to arrive before message")
		}
	})

	It("should support manual mark/commit", func() {
		cs, err := newConsumerOf(testGroup, "topic-a")
		Expect(err).NotTo(HaveOccurred())
		defer cs.Close()

		subscriptionsOf(cs).Should(Equal(map[string][]int32{
			"topic-a": {0, 1, 2, 3}},
		))

		cs.MarkPartitionOffset("topic-a", 1, 3, "")
		cs.MarkPartitionOffset("topic-a", 2, 4, "")
		Expect(cs.CommitOffsets()).NotTo(HaveOccurred())

		offsets, err := cs.fetchOffsets(cs.Subscriptions())
		Expect(err).NotTo(HaveOccurred())
		Expect(offsets).To(Equal(map[string]map[int32]offsetInfo{
			"topic-a": {0: {Offset: -1}, 1: {Offset: 4}, 2: {Offset: 5}, 3: {Offset: -1}},
		}))
	})

	It("should support manual mark/commit, reset/commit", func() {
		cs, err := newConsumerOf(testGroup, "topic-a")
		Expect(err).NotTo(HaveOccurred())
		defer cs.Close()

		subscriptionsOf(cs).Should(Equal(map[string][]int32{
			"topic-a": {0, 1, 2, 3}},
		))

		cs.MarkPartitionOffset("topic-a", 1, 3, "")
		cs.MarkPartitionOffset("topic-a", 2, 4, "")
		cs.MarkPartitionOffset("topic-b", 1, 2, "") // should not throw NPE
		Expect(cs.CommitOffsets()).NotTo(HaveOccurred())

		cs.ResetPartitionOffset("topic-a", 1, 2, "")
		cs.ResetPartitionOffset("topic-a", 2, 3, "")
		cs.ResetPartitionOffset("topic-b", 1, 2, "") // should not throw NPE
		Expect(cs.CommitOffsets()).NotTo(HaveOccurred())

		offsets, err := cs.fetchOffsets(cs.Subscriptions())
		Expect(err).NotTo(HaveOccurred())
		Expect(offsets).To(Equal(map[string]map[int32]offsetInfo{
			"topic-a": {0: {Offset: -1}, 1: {Offset: 3}, 2: {Offset: 4}, 3: {Offset: -1}},
		}))
	})

	It("should not commit unprocessed offsets", func() {
		const groupID = "panicking"

		cs, err := newConsumerOf(groupID, "topic-a")
		Expect(err).NotTo(HaveOccurred())

		subscriptionsOf(cs).Should(Equal(map[string][]int32{
			"topic-a": {0, 1, 2, 3},
		}))

		n := 0
		Expect(func() {
			for range cs.Messages() {
				n++
				panic("stop here!")
			}
		}).To(Panic())
		Expect(cs.Close()).To(Succeed())
		Expect(n).To(Equal(1))

		bk, err := testClient.Coordinator(groupID)
		Expect(err).NotTo(HaveOccurred())

		req := &sarama.OffsetFetchRequest{
			Version:       1,
			ConsumerGroup: groupID,
		}
		req.AddPartition("topic-a", 0)
		req.AddPartition("topic-a", 1)
		req.AddPartition("topic-a", 2)
		req.AddPartition("topic-a", 3)
		Expect(bk.FetchOffset(req)).To(Equal(&sarama.OffsetFetchResponse{
			Blocks: map[string]map[int32]*sarama.OffsetFetchResponseBlock{
				"topic-a": {0: {Offset: -1}, 1: {Offset: -1}, 2: {Offset: -1}, 3: {Offset: -1}},
			},
		}))
	})

	It("should consume partitions", func() {
		count := int32(0)
		consume := func(consumerID string) {
			defer GinkgoRecover()

			config := NewConfig()
			config.Group.Mode = ConsumerModePartitions
			config.Consumer.Offsets.Initial = sarama.OffsetOldest

			cs, err := NewConsumer(testKafkaAddrs, "partitions", testTopics, config)
			Expect(err).NotTo(HaveOccurred())
			defer cs.Close()

			for pc := range cs.Partitions() {
				go func(pc PartitionConsumer) {
					defer pc.Close()

					for msg := range pc.Messages() {
						atomic.AddInt32(&count, 1)
						cs.MarkOffset(msg, "")
					}
				}(pc)
			}
		}

		go consume("A")
		go consume("B")
		go consume("C")

		Eventually(func() int32 {
			return atomic.LoadInt32(&count)
		}, "30s", "100ms").Should(BeNumerically(">=", 2000))
	})

	It("should consume/commit/resume", func() {
		acc := make(chan *testConsumerMessage, 20000)
		consume := func(consumerID string, max int32) {
			defer GinkgoRecover()

			cs, err := NewConsumer(testKafkaAddrs, "fuzzing", testTopics, nil)
			Expect(err).NotTo(HaveOccurred())
			defer cs.Close()
			cs.consumerID = consumerID

			for msg := range cs.Messages() {
				acc <- &testConsumerMessage{*msg, consumerID}
				cs.MarkOffset(msg, "")

				if atomic.AddInt32(&max, -1) <= 0 {
					return
				}
			}
		}

		go consume("A", 1500)
		go consume("B", 2000)
		go consume("C", 1500)
		go consume("D", 200)
		go consume("E", 100)
		time.Sleep(10 * time.Second) // wait for consumers to subscribe to topics
		Expect(testSeed(5000, testTopics)).NotTo(HaveOccurred())
		Eventually(func() int { return len(acc) }, "30s", "100ms").Should(BeNumerically(">=", 5000))

		go consume("F", 300)
		go consume("G", 400)
		go consume("H", 1000)
		go consume("I", 2000)
		Expect(testSeed(5000, testTopics)).NotTo(HaveOccurred())
		Eventually(func() int { return len(acc) }, "30s", "100ms").Should(BeNumerically(">=", 8000))

		go consume("J", 1000)
		Expect(testSeed(5000, testTopics)).NotTo(HaveOccurred())
		Eventually(func() int { return len(acc) }, "30s", "100ms").Should(BeNumerically(">=", 9000))

		go consume("K", 1000)
		go consume("L", 3000)
		Expect(testSeed(5000, testTopics)).NotTo(HaveOccurred())
		Eventually(func() int { return len(acc) }, "30s", "100ms").Should(BeNumerically(">=", 12000))

		go consume("M", 1000)
		Expect(testSeed(5000, testTopics)).NotTo(HaveOccurred())
		Eventually(func() int { return len(acc) }, "30s", "100ms").Should(BeNumerically(">=", 15000))

		close(acc)

		uniques := make(map[string][]string)
		for msg := range acc {
			key := fmt.Sprintf("%s/%d/%d", msg.Topic, msg.Partition, msg.Offset)
			uniques[key] = append(uniques[key], msg.ConsumerID)
		}
		Expect(uniques).To(HaveLen(15000))
	})

	It("should allow close to be called multiple times", func() {
		cs, err := newConsumerOf(testGroup, testTopics...)
		Expect(err).NotTo(HaveOccurred())
		Expect(cs.Close()).NotTo(HaveOccurred())
		Expect(cs.Close()).NotTo(HaveOccurred())
	})

})
