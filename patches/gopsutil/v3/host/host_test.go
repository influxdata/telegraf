package host

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/shirou/gopsutil/v3/internal/common"
)

func skipIfNotImplementedErr(t *testing.T, err error) {
	if errors.Is(err, common.ErrNotImplementedError) {
		t.Skip("not implemented")
	}
}

func TestHostInfo(t *testing.T) {
	v, err := Info()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	empty := &InfoStat{}
	if v == empty {
		t.Errorf("Could not get hostinfo %v", v)
	}
	if v.Procs == 0 {
		t.Errorf("Could not determine the number of host processes")
	}
}

func TestUptime(t *testing.T) {
	if os.Getenv("CIRCLECI") == "true" {
		t.Skip("Skip CI")
	}

	v, err := Uptime()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v == 0 {
		t.Errorf("Could not get up time %v", v)
	}
}

func TestBoot_time(t *testing.T) {
	if os.Getenv("CIRCLECI") == "true" {
		t.Skip("Skip CI")
	}
	v, err := BootTime()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v == 0 {
		t.Errorf("Could not get boot time %v", v)
	}
	if v < 946652400 {
		t.Errorf("Invalid Boottime, older than 2000-01-01")
	}
	t.Logf("first boot time: %d", v)

	v2, err := BootTime()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	if v != v2 {
		t.Errorf("cached boot time is different")
	}
	t.Logf("second boot time: %d", v2)
}

func TestUsers(t *testing.T) {
	v, err := Users()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("error %v", err)
	}
	empty := UserStat{}
	if len(v) == 0 {
		t.Skip("Users is empty")
	}
	for _, u := range v {
		if u == empty {
			t.Errorf("Could not Users %v", v)
		}
	}
}

func TestHostInfoStat_String(t *testing.T) {
	v := InfoStat{
		Hostname:   "test",
		Uptime:     3000,
		Procs:      100,
		OS:         "linux",
		Platform:   "ubuntu",
		BootTime:   1447040000,
		HostID:     "edfd25ff-3c9c-b1a4-e660-bd826495ad35",
		KernelArch: "x86_64",
	}
	e := `{"hostname":"test","uptime":3000,"bootTime":1447040000,"procs":100,"os":"linux","platform":"ubuntu","platformFamily":"","platformVersion":"","kernelVersion":"","kernelArch":"x86_64","virtualizationSystem":"","virtualizationRole":"","hostId":"edfd25ff-3c9c-b1a4-e660-bd826495ad35"}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("HostInfoStat string is invalid:\ngot  %v\nwant %v", v, e)
	}
}

func TestUserStat_String(t *testing.T) {
	v := UserStat{
		User:     "user",
		Terminal: "term",
		Host:     "host",
		Started:  100,
	}
	e := `{"user":"user","terminal":"term","host":"host","started":100}`
	if e != fmt.Sprintf("%v", v) {
		t.Errorf("UserStat string is invalid: %v", v)
	}
}

func TestHostGuid(t *testing.T) {
	id, err := HostID()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Error(err)
	}
	if id == "" {
		t.Error("Host id is empty")
	} else {
		t.Logf("Host id value: %v", id)
	}
}

func TestTemperatureStat_String(t *testing.T) {
	v := TemperatureStat{
		SensorKey:   "CPU",
		Temperature: 1.1,
		High:        30.1,
		Critical:    0.1,
	}
	s := `{"sensorKey":"CPU","temperature":1.1,"sensorHigh":30.1,"sensorCritical":0.1}`
	if s != fmt.Sprintf("%v", v) {
		t.Errorf("TemperatureStat string is invalid, %v", fmt.Sprintf("%v", v))
	}
}

func TestVirtualization(t *testing.T) {
	wg := sync.WaitGroup{}
	testCount := 10
	wg.Add(testCount)
	for i := 0; i < testCount; i++ {
		go func(j int) {
			system, role, err := Virtualization()
			wg.Done()
			skipIfNotImplementedErr(t, err)
			if err != nil {
				t.Errorf("Virtualization() failed, %v", err)
			}

			if j == 9 {
				t.Logf("Virtualization(): %s, %s", system, role)
			}
		}(i)
	}
	wg.Wait()
}

func TestKernelVersion(t *testing.T) {
	version, err := KernelVersion()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("KernelVersion() failed, %v", err)
	}
	if version == "" {
		t.Errorf("KernelVersion() returns empty: %s", version)
	}

	t.Logf("KernelVersion(): %s", version)
}

func TestPlatformInformation(t *testing.T) {
	platform, family, version, err := PlatformInformation()
	skipIfNotImplementedErr(t, err)
	if err != nil {
		t.Errorf("PlatformInformation() failed, %v", err)
	}
	if platform == "" {
		t.Errorf("PlatformInformation() returns empty: %v", platform)
	}

	t.Logf("PlatformInformation(): %v, %v, %v", platform, family, version)
}
