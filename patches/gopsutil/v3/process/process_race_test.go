//go:build race
// +build race

package process

import (
	"sync"
	"testing"
)

func Test_Process_Ppid_Race(t *testing.T) {
	wg := sync.WaitGroup{}
	testCount := 10
	p := testGetProcess()
	wg.Add(testCount)
	for i := 0; i < testCount; i++ {
		go func(j int) {
			ppid, err := p.Ppid()
			wg.Done()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("Ppid() failed, %v", err)
			}

			if j == 9 {
				t.Logf("Ppid(): %d", ppid)
			}
		}(i)
	}
	wg.Wait()
}
