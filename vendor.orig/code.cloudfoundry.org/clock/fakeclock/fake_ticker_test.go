package fakeclock_test

import (
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FakeTicker", func() {
	const Δ = 10 * time.Millisecond

	var (
		fakeClock   *fakeclock.FakeClock
		initialTime time.Time
	)

	BeforeEach(func() {
		initialTime = time.Date(2014, 1, 1, 3, 0, 30, 0, time.UTC)
		fakeClock = fakeclock.NewFakeClock(initialTime)
	})

	It("provides a channel that receives the time at each interval", func() {
		ticker := fakeClock.NewTicker(10 * time.Second)
		timeChan := ticker.C()
		Consistently(timeChan, Δ).ShouldNot(Receive())

		fakeClock.Increment(5 * time.Second)
		Consistently(timeChan, Δ).ShouldNot(Receive())

		fakeClock.Increment(4 * time.Second)
		Consistently(timeChan, Δ).ShouldNot(Receive())

		fakeClock.Increment(1 * time.Second)
		Eventually(timeChan).Should(Receive(Equal(initialTime.Add(10 * time.Second))))

		fakeClock.Increment(10 * time.Second)
		Eventually(timeChan).Should(Receive(Equal(initialTime.Add(20 * time.Second))))

		fakeClock.Increment(10 * time.Second)
		Eventually(timeChan).Should(Receive(Equal(initialTime.Add(30 * time.Second))))
	})

	It("when there are multiple tickers", func() {
		const period = 1 * time.Second

		ticker1 := fakeClock.NewTicker(period)
		ticker2 := fakeClock.NewTicker(period)

		// Eventually(ticker.C()).Should(Recieve) make it hard to detect this error
		// due to the polling nature of Eventually. We usually end up missing the
		// second event and it gets dropped on the floor. Use counters instead to
		// make sure we don't miss the second erroneous event
		count1 := uint32(0)
		count2 := uint32(0)

		go func() {
			for {
				select {
				case <-ticker1.C():
					atomic.AddUint32(&count1, 1)
				case <-ticker2.C():
					atomic.AddUint32(&count2, 1)
				}
			}
		}()

		fakeClock.Increment(period)

		Eventually(func() uint32 { return atomic.LoadUint32(&count1) }).Should(BeEquivalentTo(1))
		Eventually(func() uint32 { return atomic.LoadUint32(&count2) }).Should(BeEquivalentTo(1))

		Consistently(func() uint32 { return atomic.LoadUint32(&count1) }).Should(BeEquivalentTo(1))
		Consistently(func() uint32 { return atomic.LoadUint32(&count2) }).Should(BeEquivalentTo(1))
	})

	It("should not fire until a period has passed", func() {
		const period = 1 * time.Second

		ticker := fakeClock.NewTicker(period)
		Consistently(ticker.C()).ShouldNot(Receive())

		fakeClock.Increment(period)
		Eventually(ticker.C()).Should(Receive(Equal(initialTime.Add(period))))

		fakeClock.Increment(0)
		Consistently(ticker.C()).ShouldNot(Receive())
	})

	It("panics given an invalid duration", func() {
		Expect(func() { fakeClock.NewTicker(0) }).Should(Panic())
	})
})
