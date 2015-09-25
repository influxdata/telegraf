package sigar_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	sigar "github.com/cloudfoundry/gosigar"
)

var _ = Describe("ConcreteSigar", func() {
	var concreteSigar *sigar.ConcreteSigar

	BeforeEach(func() {
		concreteSigar = &sigar.ConcreteSigar{}
	})

	Describe("CollectCpuStats", func() {
		It("immediately makes first CPU usage available even though it's not very accurate", func() {
			samplesCh, stop := concreteSigar.CollectCpuStats(500 * time.Millisecond)

			firstValue := <-samplesCh
			Expect(firstValue.User).To(BeNumerically(">", 0))

			stop <- struct{}{}
		})

		It("makes CPU usage delta values available", func() {
			samplesCh, stop := concreteSigar.CollectCpuStats(500 * time.Millisecond)

			firstValue := <-samplesCh

			secondValue := <-samplesCh
			Expect(secondValue.User).To(BeNumerically("<", firstValue.User))

			stop <- struct{}{}
		})

		It("does not block", func() {
			_, stop := concreteSigar.CollectCpuStats(10 * time.Millisecond)

			// Sleep long enough for samplesCh to fill at least 2 values
			time.Sleep(20 * time.Millisecond)

			stop <- struct{}{}

			// If CollectCpuStats blocks it will never get here
			Expect(true).To(BeTrue())
		})
	})

	It("GetLoadAverage", func() {
		avg, err := concreteSigar.GetLoadAverage()
		Expect(avg.One).ToNot(BeNil())
		Expect(avg.Five).ToNot(BeNil())
		Expect(avg.Fifteen).ToNot(BeNil())

		Expect(err).ToNot(HaveOccurred())
	})

	It("GetMem", func() {
		mem, err := concreteSigar.GetMem()
		Expect(err).ToNot(HaveOccurred())

		Expect(mem.Total).To(BeNumerically(">", 0))
		Expect(mem.Used + mem.Free).To(BeNumerically("<=", mem.Total))
	})

	It("GetSwap", func() {
		swap, err := concreteSigar.GetSwap()
		Expect(err).ToNot(HaveOccurred())
		Expect(swap.Used + swap.Free).To(BeNumerically("<=", swap.Total))
	})

	It("GetSwap", func() {
		fsusage, err := concreteSigar.GetFileSystemUsage("/")
		Expect(err).ToNot(HaveOccurred())
		Expect(fsusage.Total).ToNot(BeNil())

		fsusage, err = concreteSigar.GetFileSystemUsage("T O T A L L Y B O G U S")
		Expect(err).To(HaveOccurred())
		Expect(fsusage.Total).To(Equal(uint64(0)))
	})
})
