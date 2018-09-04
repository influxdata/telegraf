// +build windows

package wmi

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestWbemQuery(t *testing.T) {
	s, err := InitializeSWbemServices(DefaultClient)
	if err != nil {
		t.Fatalf("InitializeSWbemServices: %s", err)
	}

	var dst []Win32_Process
	q := CreateQuery(&dst, "WHERE name='lsass.exe'")
	errQuery := s.Query(q, &dst)
	if errQuery != nil {
		t.Fatalf("Query1: %s", errQuery)
	}
	count := len(dst)
	if count < 1 {
		t.Fatal("Query1: no results found for lsass.exe")
	}
	//fmt.Printf("dst[0].ProcessID=%d\n", dst[0].ProcessId)

	q2 := CreateQuery(&dst, "WHERE name='svchost.exe'")
	errQuery = s.Query(q2, &dst)
	if errQuery != nil {
		t.Fatalf("Query2: %s", errQuery)
	}
	count = len(dst)
	if count < 1 {
		t.Fatal("Query2: no results found for svchost.exe")
	}
	//for index, item := range dst {
	//	fmt.Printf("dst[%d].ProcessID=%d\n", index, item.ProcessId)
	//}
	errClose := s.Close()
	if errClose != nil {
		t.Fatalf("Close: %s", errClose)
	}
}

func TestWbemQueryNamespace(t *testing.T) {
	s, err := InitializeSWbemServices(DefaultClient)
	if err != nil {
		t.Fatalf("InitializeSWbemServices: %s", err)
	}
	var dst []MSFT_NetAdapter
	q := CreateQuery(&dst, "")
	errQuery := s.Query(q, &dst, nil, "root\\StandardCimv2")
	if errQuery != nil {
		t.Fatalf("Query: %s", errQuery)
	}
	count := len(dst)
	if count < 1 {
		t.Fatal("Query: no results found for MSFT_NetAdapter in root\\StandardCimv2")
	}
	errClose := s.Close()
	if errClose != nil {
		t.Fatalf("Close: %s", errClose)
	}
}

// Run using: go test -run TestWbemMemory -timeout 60m
func TestWbemMemory(t *testing.T) {
	s, err := InitializeSWbemServices(DefaultClient)
	if err != nil {
		t.Fatalf("InitializeSWbemServices: %s", err)
	}
	start := time.Now()
	limit := 500000
	fmt.Printf("Benchmark Iterations: %d (Memory should stabilize around 7MB after ~3000)\n", limit)
	var privateMB, allocMB, allocTotalMB float64
	for i := 0; i < limit; i++ {
		privateMB, allocMB, allocTotalMB = WbemGetMemoryUsageMB(s)
		if i%100 == 0 {
			privateMB, allocMB, allocTotalMB = WbemGetMemoryUsageMB(s)
			fmt.Printf("Time: %4ds  Count: %5d  Private Memory: %5.1fMB  MemStats.Alloc: %4.1fMB  MemStats.TotalAlloc: %5.1fMB\n", time.Now().Sub(start)/time.Second, i, privateMB, allocMB, allocTotalMB)
		}
	}
	errClose := s.Close()
	if errClose != nil {
		t.Fatalf("Close: %s", err)
	}
	fmt.Printf("Final Time: %4ds  Private Memory: %5.1fMB  MemStats.Alloc: %4.1fMB  MemStats.TotalAlloc: %5.1fMB\n", time.Now().Sub(start)/time.Second, privateMB, allocMB, allocTotalMB)
}

func WbemGetMemoryUsageMB(s *SWbemServices) (float64, float64, float64) {
	runtime.ReadMemStats(&mMemoryUsageMB)
	errGetMemoryUsageMB = s.Query(qGetMemoryUsageMB, &dstGetMemoryUsageMB)
	if errGetMemoryUsageMB != nil {
		fmt.Println("ERROR GetMemoryUsage", errGetMemoryUsageMB)
		return 0, 0, 0
	}
	return float64(dstGetMemoryUsageMB[0].WorkingSetPrivate) / MB, float64(mMemoryUsageMB.Alloc) / MB, float64(mMemoryUsageMB.TotalAlloc) / MB
}

//Run all benchmarks (should run for at least 60s to get a stable number):
//go test -run=NONE -bench=Version -benchtime=120s

//Individual benchmarks:
//go test -run=NONE -bench=NewVersion -benchtime=120s
func BenchmarkNewVersion(b *testing.B) {
	s, err := InitializeSWbemServices(DefaultClient)
	if err != nil {
		b.Fatalf("InitializeSWbemServices: %s", err)
	}
	var dst []Win32_OperatingSystem
	q := CreateQuery(&dst, "")
	for n := 0; n < b.N; n++ {
		errQuery := s.Query(q, &dst)
		if errQuery != nil {
			b.Fatalf("Query%d: %s", n, errQuery)
		}
		count := len(dst)
		if count < 1 {
			b.Fatalf("Query%d: no results found for Win32_OperatingSystem", n)
		}
	}
	errClose := s.Close()
	if errClose != nil {
		b.Fatalf("Close: %s", errClose)
	}
}

//go test -run=NONE -bench=OldVersion -benchtime=120s
func BenchmarkOldVersion(b *testing.B) {
	var dst []Win32_OperatingSystem
	q := CreateQuery(&dst, "")
	for n := 0; n < b.N; n++ {
		errQuery := Query(q, &dst)
		if errQuery != nil {
			b.Fatalf("Query%d: %s", n, errQuery)
		}
		count := len(dst)
		if count < 1 {
			b.Fatalf("Query%d: no results found for Win32_OperatingSystem", n)
		}
	}
}

type MSFT_NetAdapter struct {
	Name              string
	InterfaceIndex    int
	DriverDescription string
}
