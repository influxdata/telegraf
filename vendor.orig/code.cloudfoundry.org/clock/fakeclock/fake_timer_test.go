package fakeclock_test

import (
	"os"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/clock/fakeclock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("FakeTimer", func() {
	const Δ = 10 * time.Millisecond

	var (
		fakeClock   *fakeclock.FakeClock
		initialTime time.Time
	)

	BeforeEach(func() {
		initialTime = time.Date(2014, 1, 1, 3, 0, 30, 0, time.UTC)
		fakeClock = fakeclock.NewFakeClock(initialTime)
	})

	It("proivdes a channel that receives after the given interval has elapsed", func() {
		timer := fakeClock.NewTimer(10 * time.Second)
		timeChan := timer.C()
		Consistently(timeChan, Δ).ShouldNot(Receive())

		fakeClock.Increment(5 * time.Second)
		Consistently(timeChan, Δ).ShouldNot(Receive())

		fakeClock.Increment(4 * time.Second)
		Consistently(timeChan, Δ).ShouldNot(Receive())

		fakeClock.Increment(1 * time.Second)
		Eventually(timeChan).Should(Receive(Equal(initialTime.Add(10 * time.Second))))

		fakeClock.Increment(10 * time.Second)
		Consistently(timeChan, Δ).ShouldNot(Receive())
	})

	Describe("Stop", func() {
		It("is idempotent", func() {
			timer := fakeClock.NewTimer(time.Second)
			timer.Stop()
			timer.Stop()
			fakeClock.Increment(time.Second)
			Consistently(timer.C()).ShouldNot(Receive())
		})
	})

	Describe("WaitForWatcherAndIncrement", func() {
		const (
			duration = 10 * time.Second
		)

		var (
			process  ifrit.Process
			runner   ifrit.Runner
			received chan time.Time
		)

		BeforeEach(func() {
			received = make(chan time.Time, 100)
		})

		AfterEach(func() {
			ginkgomon.Interrupt(process)
		})

		JustBeforeEach(func() {
			process = ginkgomon.Invoke(runner)
		})

		Context("when timers are added asynchronously", func() {
			BeforeEach(func() {
				runner = ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
					close(ready)

					for {
						timer := fakeClock.NewTimer(duration)

						select {
						case ticked := <-timer.C():
							received <- ticked
						case <-signals:
							return nil
						}
					}
				})
			})

			It("consistently fires the new timers", func() {
				for i := 0; i < 100; i++ {
					fakeClock.WaitForWatcherAndIncrement(duration)
					Expect((<-received).Sub(initialTime)).To(Equal(duration * time.Duration(i+1)))
				}
			})
		})

		Context("when a timer is reset asynchronously", func() {
			var (
				timer clock.Timer
			)

			BeforeEach(func() {
				timer = fakeClock.NewTimer(duration)

				runner = ifrit.RunFunc(func(signals <-chan os.Signal, ready chan<- struct{}) error {
					close(ready)

					for {
						select {
						case ticked := <-timer.C():
							received <- ticked
							timer.Reset(duration)
						case <-signals:
							return nil
						}
					}
				})
			})

			It("consistently fires timers that reset asynchronously", func() {
				incrementClock := make(chan struct{})

				go func() {
					for {
						<-incrementClock
						fakeClock.WaitForWatcherAndIncrement(duration)
					}
				}()

				for i := 0; i < 100; i++ {
					Eventually(incrementClock).Should(BeSent(struct{}{}))
					var timestamp time.Time
					Eventually(received, 5*time.Second).Should(Receive(&timestamp))
				}
			})
		})

	})
})
