package sigar_test

import (
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	sigar "github.com/cloudfoundry/gosigar"
)

var _ = Describe("sigarLinux", func() {
	var procd string

	BeforeEach(func() {
		var err error
		procd, err = ioutil.TempDir("", "sigarTests")
		Expect(err).ToNot(HaveOccurred())
		sigar.Procd = procd
	})

	AfterEach(func() {
		sigar.Procd = "/proc"
	})

	Describe("CPU", func() {
		var (
			statFile string
			cpu      sigar.Cpu
		)

		BeforeEach(func() {
			statFile = procd + "/stat"
			cpu = sigar.Cpu{}
		})

		Describe("Get", func() {
			It("gets CPU usage", func() {
				statContents := []byte("cpu 25 1 2 3 4 5 6 7")
				err := ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				err = cpu.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(cpu.User).To(Equal(uint64(25)))
			})

			It("ignores empty lines", func() {
				statContents := []byte("cpu ")
				err := ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				err = cpu.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(cpu.User).To(Equal(uint64(0)))
			})
		})

		Describe("CollectCpuStats", func() {
			It("collects CPU usage over time", func() {
				statContents := []byte("cpu 25 1 2 3 4 5 6 7")
				err := ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				concreteSigar := &sigar.ConcreteSigar{}
				cpuUsages, stop := concreteSigar.CollectCpuStats(500 * time.Millisecond)

				Expect(<-cpuUsages).To(Equal(sigar.Cpu{
					User:    uint64(25),
					Nice:    uint64(1),
					Sys:     uint64(2),
					Idle:    uint64(3),
					Wait:    uint64(4),
					Irq:     uint64(5),
					SoftIrq: uint64(6),
					Stolen:  uint64(7),
				}))

				statContents = []byte("cpu 30 3 7 10 25 55 36 65")
				err = ioutil.WriteFile(statFile, statContents, 0644)
				Expect(err).ToNot(HaveOccurred())

				Expect(<-cpuUsages).To(Equal(sigar.Cpu{
					User:    uint64(5),
					Nice:    uint64(2),
					Sys:     uint64(5),
					Idle:    uint64(7),
					Wait:    uint64(21),
					Irq:     uint64(50),
					SoftIrq: uint64(30),
					Stolen:  uint64(58),
				}))

				stop <- struct{}{}
			})
		})
	})

	Describe("Mem", func() {
		var meminfoFile string
		BeforeEach(func() {
			meminfoFile = procd + "/meminfo"

			meminfoContents := `
MemTotal:         374256 kB
MemFree:          274460 kB
Buffers:            9764 kB
Cached:            38648 kB
SwapCached:            0 kB
Active:            33772 kB
Inactive:          31184 kB
Active(anon):      16572 kB
Inactive(anon):      552 kB
Active(file):      17200 kB
Inactive(file):    30632 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:        786428 kB
SwapFree:         786428 kB
Dirty:                 0 kB
Writeback:             0 kB
AnonPages:         16564 kB
Mapped:             6612 kB
Shmem:               584 kB
Slab:              19092 kB
SReclaimable:       9128 kB
SUnreclaim:         9964 kB
KernelStack:         672 kB
PageTables:         1864 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:      973556 kB
Committed_AS:      55880 kB
VmallocTotal:   34359738367 kB
VmallocUsed:       21428 kB
VmallocChunk:   34359713596 kB
HardwareCorrupted:     0 kB
AnonHugePages:         0 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
DirectMap4k:       59328 kB
DirectMap2M:      333824 kB
`
			err := ioutil.WriteFile(meminfoFile, []byte(meminfoContents), 0444)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns correct memory info", func() {
			mem := sigar.Mem{}
			err := mem.Get()
			Expect(err).ToNot(HaveOccurred())

			Expect(mem.Total).To(BeNumerically("==", 374256*1024))
			Expect(mem.Free).To(BeNumerically("==", 274460*1024))
		})
	})

	Describe("Swap", func() {
		var meminfoFile string
		BeforeEach(func() {
			meminfoFile = procd + "/meminfo"

			meminfoContents := `
MemTotal:         374256 kB
MemFree:          274460 kB
Buffers:            9764 kB
Cached:            38648 kB
SwapCached:            0 kB
Active:            33772 kB
Inactive:          31184 kB
Active(anon):      16572 kB
Inactive(anon):      552 kB
Active(file):      17200 kB
Inactive(file):    30632 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:        786428 kB
SwapFree:         786428 kB
Dirty:                 0 kB
Writeback:             0 kB
AnonPages:         16564 kB
Mapped:             6612 kB
Shmem:               584 kB
Slab:              19092 kB
SReclaimable:       9128 kB
SUnreclaim:         9964 kB
KernelStack:         672 kB
PageTables:         1864 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:      973556 kB
Committed_AS:      55880 kB
VmallocTotal:   34359738367 kB
VmallocUsed:       21428 kB
VmallocChunk:   34359713596 kB
HardwareCorrupted:     0 kB
AnonHugePages:         0 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
DirectMap4k:       59328 kB
DirectMap2M:      333824 kB
`
			err := ioutil.WriteFile(meminfoFile, []byte(meminfoContents), 0444)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns correct memory info", func() {
			swap := sigar.Swap{}
			err := swap.Get()
			Expect(err).ToNot(HaveOccurred())

			Expect(swap.Total).To(BeNumerically("==", 786428*1024))
			Expect(swap.Free).To(BeNumerically("==", 786428*1024))
		})
	})
})
