package fakes

import (
	"time"

	sigar "github.com/cloudfoundry/gosigar"
)

type FakeSigar struct {
	LoadAverage    sigar.LoadAverage
	LoadAverageErr error

	Mem    sigar.Mem
	MemErr error

	Swap    sigar.Swap
	SwapErr error

	FileSystemUsage     sigar.FileSystemUsage
	FileSystemUsageErr  error
	FileSystemUsagePath string

	CollectCpuStatsCpuCh  chan sigar.Cpu
	CollectCpuStatsStopCh chan struct{}
}

func NewFakeSigar() *FakeSigar {
	return &FakeSigar{
		CollectCpuStatsCpuCh:  make(chan sigar.Cpu, 1),
		CollectCpuStatsStopCh: make(chan struct{}),
	}
}

func (f *FakeSigar) CollectCpuStats(collectionInterval time.Duration) (<-chan sigar.Cpu, chan<- struct{}) {
	samplesCh := make(chan sigar.Cpu, 1)
	stopCh := make(chan struct{})

	go func() {
		for {
			select {
			case cpuStat := <-f.CollectCpuStatsCpuCh:
				select {
				case samplesCh <- cpuStat:
				default:
					// Include default to avoid channel blocking
				}

			case <-f.CollectCpuStatsStopCh:
				return
			}
		}
	}()

	return samplesCh, stopCh
}

func (f *FakeSigar) GetLoadAverage() (sigar.LoadAverage, error) {
	return f.LoadAverage, f.LoadAverageErr
}

func (f *FakeSigar) GetMem() (sigar.Mem, error) {
	return f.Mem, f.MemErr
}

func (f *FakeSigar) GetSwap() (sigar.Swap, error) {
	return f.Swap, f.SwapErr
}

func (f *FakeSigar) GetFileSystemUsage(path string) (sigar.FileSystemUsage, error) {
	f.FileSystemUsagePath = path
	return f.FileSystemUsage, f.FileSystemUsageErr
}
