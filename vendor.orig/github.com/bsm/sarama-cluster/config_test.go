package cluster

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var subject *Config

	BeforeEach(func() {
		subject = NewConfig()
	})

	It("should init", func() {
		Expect(subject.Group.Session.Timeout).To(Equal(30 * time.Second))
		Expect(subject.Group.Heartbeat.Interval).To(Equal(3 * time.Second))
		Expect(subject.Group.Return.Notifications).To(BeFalse())
		Expect(subject.Metadata.Retry.Max).To(Equal(3))
		Expect(subject.Group.Offsets.Synchronization.DwellTime).NotTo(BeZero())
		// Expect(subject.Config.Version).To(Equal(sarama.V0_9_0_0))
	})

})
