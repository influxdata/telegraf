package cluster

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var subject *Client

	BeforeEach(func() {
		var err error
		subject, err = NewClient(testKafkaAddrs, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should not allow to share clients across multiple consumers", func() {
		c1, err := NewConsumerFromClient(subject, testGroup, testTopics)
		Expect(err).NotTo(HaveOccurred())
		defer c1.Close()

		_, err = NewConsumerFromClient(subject, testGroup, testTopics)
		Expect(err).To(MatchError("cluster: client is already used by another consumer"))

		Expect(c1.Close()).To(Succeed())
		c2, err := NewConsumerFromClient(subject, testGroup, testTopics)
		Expect(err).NotTo(HaveOccurred())
		Expect(c2.Close()).To(Succeed())
	})

})
