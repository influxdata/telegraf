// +build windows

package wmi

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func TestQuery(t *testing.T) {
	var dst []Win32_Process
	q := CreateQuery(&dst, "")
	err := Query(q, &dst)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFieldMismatch(t *testing.T) {
	type s struct {
		Name        string
		HandleCount uint32
		Blah        uint32
	}
	var dst []s
	err := Query("SELECT Name, HandleCount FROM Win32_Process", &dst)
	if err == nil || err.Error() != `wmi: cannot load field "Blah" into a "uint32": no such struct field` {
		t.Error("Expected err field mismatch")
	}
}

func TestStrings(t *testing.T) {
	printed := false
	f := func() {
		var dst []Win32_Process
		zeros := 0
		q := CreateQuery(&dst, "")
		for i := 0; i < 5; i++ {
			err := Query(q, &dst)
			if err != nil {
				t.Fatal(err, q)
			}
			for _, d := range dst {
				v := reflect.ValueOf(d)
				for j := 0; j < v.NumField(); j++ {
					f := v.Field(j)
					if f.Kind() != reflect.String {
						continue
					}
					s := f.Interface().(string)
					if len(s) > 0 && s[0] == '\u0000' {
						zeros++
						if !printed {
							printed = true
							j, _ := json.MarshalIndent(&d, "", "  ")
							t.Log("Example with \\u0000:\n", string(j))
						}
					}
				}
			}
			fmt.Println("iter", i, "zeros:", zeros)
		}
		if zeros > 0 {
			t.Error("> 0 zeros")
		}
	}

	fmt.Println("Disabling GC")
	debug.SetGCPercent(-1)
	f()
	fmt.Println("Enabling GC")
	debug.SetGCPercent(100)
	f()
}

func TestNamespace(t *testing.T) {
	var dst []Win32_Process
	q := CreateQuery(&dst, "")
	err := QueryNamespace(q, &dst, `root\CIMV2`)
	if err != nil {
		t.Fatal(err)
	}
	dst = nil
	err = QueryNamespace(q, &dst, `broken\nothing`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateQuery(t *testing.T) {
	type TestStruct struct {
		Name  string
		Count int
	}
	var dst []TestStruct
	output := "SELECT Name, Count FROM TestStruct WHERE Count > 2"
	tests := []interface{}{
		&dst,
		dst,
		TestStruct{},
		&TestStruct{},
	}
	for i, test := range tests {
		if o := CreateQuery(test, "WHERE Count > 2"); o != output {
			t.Error("bad output on", i, o)
		}
	}
	if CreateQuery(3, "") != "" {
		t.Error("expected empty string")
	}
}

// Run using: go test -run TestMemoryWMISimple -timeout 60m
func _TestMemoryWMISimple(t *testing.T) {
	start := time.Now()
	limit := 500000
	fmt.Printf("Benchmark Iterations: %d (Memory should stabilize around 7MB after ~3000)\n", limit)
	var privateMB, allocMB, allocTotalMB float64
	//var dst []Win32_PerfRawData_PerfDisk_LogicalDisk
	//q := CreateQuery(&dst, "")
	for i := 0; i < limit; i++ {
		privateMB, allocMB, allocTotalMB = GetMemoryUsageMB()
		if i%1000 == 0 {
			//privateMB, allocMB, allocTotalMB = GetMemoryUsageMB()
			fmt.Printf("Time: %4ds  Count: %5d  Private Memory: %5.1fMB  MemStats.Alloc: %4.1fMB  MemStats.TotalAlloc: %5.1fMB\n", time.Now().Sub(start)/time.Second, i, privateMB, allocMB, allocTotalMB)
		}
		//Query(q, &dst)
	}
	//privateMB, allocMB, allocTotalMB = GetMemoryUsageMB()
	fmt.Printf("Final Time: %4ds  Private Memory: %5.1fMB  MemStats.Alloc: %4.1fMB  MemStats.TotalAlloc: %5.1fMB\n", time.Now().Sub(start)/time.Second, privateMB, allocMB, allocTotalMB)
}

func _TestMemoryWMIConcurrent(t *testing.T) {
	if testing.Short() {
		return
	}
	start := time.Now()
	limit := 50000
	fmt.Println("Total Iterations:", limit)
	fmt.Println("No panics mean it succeeded. Other errors are OK. Memory should stabilize after ~1500 iterations.")
	runtime.GOMAXPROCS(2)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for i := 0; i < limit; i++ {
			if i%500 == 0 {
				privateMB, allocMB, allocTotalMB := GetMemoryUsageMB()
				fmt.Printf("Time: %4ds  Count: %4d  Private Memory: %5.1fMB  MemStats.Alloc: %4.1fMB  MemStats.TotalAlloc: %5.1fMB\n", time.Now().Sub(start)/time.Second, i, privateMB, allocMB, allocTotalMB)
			}
			var dst []Win32_PerfRawData_PerfDisk_LogicalDisk
			q := CreateQuery(&dst, "")
			err := Query(q, &dst)
			if err != nil {
				fmt.Println("ERROR disk", err)
			}
		}
		wg.Done()
	}()
	go func() {
		for i := 0; i > -limit; i-- {
			//if i%500 == 0 {
			//	fmt.Println(i)
			//}
			var dst []Win32_OperatingSystem
			q := CreateQuery(&dst, "")
			err := Query(q, &dst)
			if err != nil {
				fmt.Println("ERROR OS", err)
			}
		}
		wg.Done()
	}()
	wg.Wait()
	//privateMB, allocMB, allocTotalMB := GetMemoryUsageMB()
	//fmt.Printf("Final Private Memory: %5.1fMB  MemStats.Alloc: %4.1fMB  MemStats.TotalAlloc: %5.1fMB\n", privateMB, allocMB, allocTotalMB)
}

var lockthread sync.Mutex
var refcount1 int32
var refcount2 int32
var refcount3 int32

// Test function showing memory leak in unknown.QueryInterface call on Server2016/Windows10
func getRSS(url string, xmlhttp *ole.IDispatch, MinimalTest bool) (int, error) {

	// call using url,nil to see memory leak
	if xmlhttp == nil {
		//Initialize inside loop if not passed in from outer section
		lockthread.Lock()
		defer lockthread.Unlock()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
		if err != nil {
			oleCode := err.(*ole.OleError).Code()
			if oleCode != ole.S_OK && oleCode != S_FALSE {
				return 0, err
			}
		}
		defer ole.CoUninitialize()

		//fmt.Println("CreateObject Microsoft.XMLHTTP")
		unknown, err := oleutil.CreateObject("Microsoft.XMLHTTP")
		if err != nil {
			return 0, err
		}
		defer func() { refcount1 += xmlhttp.Release() }()

		//Memory leak occurs here
		xmlhttp, err = unknown.QueryInterface(ole.IID_IDispatch)
		if err != nil {
			return 0, err
		}
		defer func() { refcount2 += xmlhttp.Release() }()
		//Nothing below this really matters. Can be removed if you want a tighter loop
	}

	//fmt.Printf("Download %s\n", url)
	openRaw, err := oleutil.CallMethod(xmlhttp, "open", "GET", url, false)
	if err != nil {
		return 0, err
	}
	defer openRaw.Clear()

	if MinimalTest {
		return 1, nil
	}

	//Initiate http request
	sendRaw, err := oleutil.CallMethod(xmlhttp, "send", nil)
	if err != nil {
		return 0, err
	}
	defer sendRaw.Clear()
	state := -1 // https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest/readyState
	for state != 4 {
		time.Sleep(5 * time.Millisecond)
		stateRaw := oleutil.MustGetProperty(xmlhttp, "readyState")
		state = int(stateRaw.Val)
		stateRaw.Clear()
	}

	responseXMLRaw := oleutil.MustGetProperty(xmlhttp, "responseXml")
	responseXML := responseXMLRaw.ToIDispatch()
	defer responseXMLRaw.Clear()
	itemsRaw := oleutil.MustCallMethod(responseXML, "selectNodes", "/rdf:RDF/item")
	items := itemsRaw.ToIDispatch()
	defer itemsRaw.Clear()
	lengthRaw := oleutil.MustGetProperty(items, "length")
	defer lengthRaw.Clear()
	length := int(lengthRaw.Val)

	/* This just bloats the TotalAlloc and slows the test down. Doesn't effect Private Working Set
	for n := 0; n < length; n++ {
		itemRaw := oleutil.MustGetProperty(items, "item", n)
		item := itemRaw.ToIDispatch()
		title := oleutil.MustCallMethod(item, "selectSingleNode", "title").ToIDispatch()

		//fmt.Println(oleutil.MustGetProperty(title, "text").ToString())
		textRaw := oleutil.MustGetProperty(title, "text")
		textRaw.ToString()

		link := oleutil.MustCallMethod(item, "selectSingleNode", "link").ToIDispatch()
		//fmt.Println("  " + oleutil.MustGetProperty(link, "text").ToString())
		textRaw2 := oleutil.MustGetProperty(link, "text")
		textRaw2.ToString()

		textRaw2.Clear()
		link.Release()
		textRaw.Clear()
		title.Release()
		itemRaw.Clear()
	}
	*/
	return length, nil
}

// Testing go-ole/oleutil
// Run using: go test -run TestMemoryOLE -timeout 60m
// Code from https://github.com/go-ole/go-ole/blob/master/example/msxml/rssreader.go
func _TestMemoryOLE(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()

	start := time.Now()
	limit := 50000000
	url := "http://localhost/slashdot.xml" //http://rss.slashdot.org/Slashdot/slashdot"
	fmt.Printf("Benchmark Iterations: %d (Memory should stabilize around 8MB to 12MB after ~2k full or 250k minimal)\n", limit)

	//On Server 2016 or Windows 10 changing leakMemory=true will cause it to leak ~1.5MB per 10000 calls to unknown.QueryInterface
	leakMemory := true

	////////////////////////////////////////
	//Start outer section
	var unknown *ole.IUnknown
	var xmlhttp *ole.IDispatch
	if !leakMemory {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
		if err != nil {
			oleCode := err.(*ole.OleError).Code()
			if oleCode != ole.S_OK && oleCode != S_FALSE {
				t.Fatal(err)
			}
		}
		defer ole.CoUninitialize()

		//fmt.Println("CreateObject Microsoft.XMLHTTP")
		unknown, err = oleutil.CreateObject("Microsoft.XMLHTTP")
		if err != nil {
			t.Fatal(err)
		}
		defer unknown.Release()

		//Memory leak starts here
		xmlhttp, err = unknown.QueryInterface(ole.IID_IDispatch)
		if err != nil {
			t.Fatal(err)
		}
		defer xmlhttp.Release()
	}
	//End outer section
	////////////////////////////////////////

	totalItems := uint64(0)
	for i := 0; i < limit; i++ {
		if i%2000 == 0 {
			privateMB, allocMB, allocTotalMB := GetMemoryUsageMB()
			fmt.Printf("Time: %4ds  Count: %7d  Private Memory: %5.1fMB  MemStats.Alloc: %4.1fMB  MemStats.TotalAlloc: %5.1fMB  %7d/%7d\n", time.Now().Sub(start)/time.Second, i, privateMB, allocMB, allocTotalMB, refcount1, refcount2)
		}
		//This should use less than 10MB for 1 million iterations if xmlhttp was initialized above
		//On Server 2016 or Windows 10 changing leakMemory=true above will cause it to leak ~1.5MB per 10000 calls to unknown.QueryInterface
		count, err := getRSS(url, xmlhttp, true) //last argument is for Minimal test. Doesn't effect leak just overall allocations/time
		if err != nil {
			t.Fatal(err)
		}
		totalItems += uint64(count)
	}
	privateMB, allocMB, allocTotalMB := GetMemoryUsageMB()
	fmt.Printf("Final totalItems: %d  Private Memory: %5.1fMB  MemStats.Alloc: %4.1fMB  MemStats.TotalAlloc: %5.1fMB\n", totalItems, privateMB, allocMB, allocTotalMB)
}

const MB = 1024 * 1024

var (
	mMemoryUsageMB      runtime.MemStats
	errGetMemoryUsageMB error
	dstGetMemoryUsageMB []Win32_PerfRawData_PerfProc_Process
	filterProcessID     = fmt.Sprintf("WHERE IDProcess = %d", os.Getpid())
	qGetMemoryUsageMB   = CreateQuery(&dstGetMemoryUsageMB, filterProcessID)
)

func GetMemoryUsageMB() (float64, float64, float64) {
	runtime.ReadMemStats(&mMemoryUsageMB)
	//errGetMemoryUsageMB = nil //Query(qGetMemoryUsageMB, &dstGetMemoryUsageMB) float64(dstGetMemoryUsageMB[0].WorkingSetPrivate)
	errGetMemoryUsageMB = Query(qGetMemoryUsageMB, &dstGetMemoryUsageMB)
	if errGetMemoryUsageMB != nil {
		fmt.Println("ERROR GetMemoryUsage", errGetMemoryUsageMB)
		return 0, 0, 0
	}
	return float64(dstGetMemoryUsageMB[0].WorkingSetPrivate) / MB, float64(mMemoryUsageMB.Alloc) / MB, float64(mMemoryUsageMB.TotalAlloc) / MB
}

type Win32_PerfRawData_PerfProc_Process struct {
	IDProcess         uint32
	WorkingSetPrivate uint64
}

type Win32_Process struct {
	CSCreationClassName        string
	CSName                     string
	Caption                    *string
	CommandLine                *string
	CreationClassName          string
	CreationDate               *time.Time
	Description                *string
	ExecutablePath             *string
	ExecutionState             *uint16
	Handle                     string
	HandleCount                uint32
	InstallDate                *time.Time
	KernelModeTime             uint64
	MaximumWorkingSetSize      *uint32
	MinimumWorkingSetSize      *uint32
	Name                       string
	OSCreationClassName        string
	OSName                     string
	OtherOperationCount        uint64
	OtherTransferCount         uint64
	PageFaults                 uint32
	PageFileUsage              uint32
	ParentProcessId            uint32
	PeakPageFileUsage          uint32
	PeakVirtualSize            uint64
	PeakWorkingSetSize         uint32
	Priority                   uint32
	PrivatePageCount           uint64
	ProcessId                  uint32
	QuotaNonPagedPoolUsage     uint32
	QuotaPagedPoolUsage        uint32
	QuotaPeakNonPagedPoolUsage uint32
	QuotaPeakPagedPoolUsage    uint32
	ReadOperationCount         uint64
	ReadTransferCount          uint64
	SessionId                  uint32
	Status                     *string
	TerminationDate            *time.Time
	ThreadCount                uint32
	UserModeTime               uint64
	VirtualSize                uint64
	WindowsVersion             string
	WorkingSetSize             uint64
	WriteOperationCount        uint64
	WriteTransferCount         uint64
}

type Win32_PerfRawData_PerfDisk_LogicalDisk struct {
	AvgDiskBytesPerRead          uint64
	AvgDiskBytesPerRead_Base     uint32
	AvgDiskBytesPerTransfer      uint64
	AvgDiskBytesPerTransfer_Base uint32
	AvgDiskBytesPerWrite         uint64
	AvgDiskBytesPerWrite_Base    uint32
	AvgDiskQueueLength           uint64
	AvgDiskReadQueueLength       uint64
	AvgDiskSecPerRead            uint32
	AvgDiskSecPerRead_Base       uint32
	AvgDiskSecPerTransfer        uint32
	AvgDiskSecPerTransfer_Base   uint32
	AvgDiskSecPerWrite           uint32
	AvgDiskSecPerWrite_Base      uint32
	AvgDiskWriteQueueLength      uint64
	Caption                      *string
	CurrentDiskQueueLength       uint32
	Description                  *string
	DiskBytesPerSec              uint64
	DiskReadBytesPerSec          uint64
	DiskReadsPerSec              uint32
	DiskTransfersPerSec          uint32
	DiskWriteBytesPerSec         uint64
	DiskWritesPerSec             uint32
	FreeMegabytes                uint32
	Frequency_Object             uint64
	Frequency_PerfTime           uint64
	Frequency_Sys100NS           uint64
	Name                         string
	PercentDiskReadTime          uint64
	PercentDiskReadTime_Base     uint64
	PercentDiskTime              uint64
	PercentDiskTime_Base         uint64
	PercentDiskWriteTime         uint64
	PercentDiskWriteTime_Base    uint64
	PercentFreeSpace             uint32
	PercentFreeSpace_Base        uint32
	PercentIdleTime              uint64
	PercentIdleTime_Base         uint64
	SplitIOPerSec                uint32
	Timestamp_Object             uint64
	Timestamp_PerfTime           uint64
	Timestamp_Sys100NS           uint64
}

type Win32_OperatingSystem struct {
	BootDevice                                string
	BuildNumber                               string
	BuildType                                 string
	Caption                                   *string
	CodeSet                                   string
	CountryCode                               string
	CreationClassName                         string
	CSCreationClassName                       string
	CSDVersion                                *string
	CSName                                    string
	CurrentTimeZone                           int16
	DataExecutionPrevention_Available         bool
	DataExecutionPrevention_32BitApplications bool
	DataExecutionPrevention_Drivers           bool
	DataExecutionPrevention_SupportPolicy     *uint8
	Debug                                     bool
	Description                               *string
	Distributed                               bool
	EncryptionLevel                           uint32
	ForegroundApplicationBoost                *uint8
	FreePhysicalMemory                        uint64
	FreeSpaceInPagingFiles                    uint64
	FreeVirtualMemory                         uint64
	InstallDate                               time.Time
	LargeSystemCache                          *uint32
	LastBootUpTime                            time.Time
	LocalDateTime                             time.Time
	Locale                                    string
	Manufacturer                              string
	MaxNumberOfProcesses                      uint32
	MaxProcessMemorySize                      uint64
	MUILanguages                              *[]string
	Name                                      string
	NumberOfLicensedUsers                     *uint32
	NumberOfProcesses                         uint32
	NumberOfUsers                             uint32
	OperatingSystemSKU                        uint32
	Organization                              string
	OSArchitecture                            string
	OSLanguage                                uint32
	OSProductSuite                            uint32
	OSType                                    uint16
	OtherTypeDescription                      *string
	PAEEnabled                                *bool
	PlusProductID                             *string
	PlusVersionNumber                         *string
	PortableOperatingSystem                   bool
	Primary                                   bool
	ProductType                               uint32
	RegisteredUser                            string
	SerialNumber                              string
	ServicePackMajorVersion                   uint16
	ServicePackMinorVersion                   uint16
	SizeStoredInPagingFiles                   uint64
	Status                                    string
	SuiteMask                                 uint32
	SystemDevice                              string
	SystemDirectory                           string
	SystemDrive                               string
	TotalSwapSpaceSize                        *uint64
	TotalVirtualMemorySize                    uint64
	TotalVisibleMemorySize                    uint64
	Version                                   string
	WindowsDirectory                          string
}
