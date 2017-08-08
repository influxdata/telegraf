// +build linux

// Package smartctl is a collector for S.M.A.R.T data for HDD, SSD + NVMe devices, linux only
// https://www.smartmontools.org/
package smartctl

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

var bencher = &SmartCtl{
	Disks:      []string{"/dev/sda -d scsi", "/dev/bus/0 -d megaraid,4", "/dev/bus/0 -d megaraid,24"},
	Init:       true,
	DiskOutput: make(map[string]Disk, 3),
	DiskFailed: make(map[string]error),
}

// TestSmartCtl_Gather sends a naked SmartCtl function to the machine to see how data can parse. Virtually
// any output is accepted here, provided it's not an err
func TestSmartCtl_Gather(t *testing.T) {
	tester := &SmartCtl{}

	acc := new(testutil.Accumulator)
	err := tester.Gather(acc)

	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "executable") {
			t.Logf("[INFO] smartctl binary doesn't exist, but in testmode we continue: %v", err)
		} else {
			t.Errorf("[ERROR] Did not expect error, but received err: %v, data: %v\n", err, tester.DiskFailed)
		}
	} else if len(tester.DiskFailed) > 0 {
		t.Logf("[INFO] Did not receive error, but some disks failed: %v, init: %t, output: %#v\n", tester.DiskFailed, tester.Init, tester.Disks)
	} else {
		t.Logf("[INFO]  init: %t, output: %#v, len: %d\n", tester.Init, tester.Disks, len(tester.Disks))
	}
}

// TestSmartCtl_GatherInclude sends a specific disk name in the Include portion to Gather, which
// should skip smartctl --scan altogether; we fail if err != nil or the disklist returned is > 1
func TestSmartCtl_GatherInclude(t *testing.T) {
	tester := bencher
	tester.Init = false
	tester.Include = []string{"/dev/myrandomtester -d scsi"}

	acc := new(testutil.Accumulator)
	err := tester.Gather(acc)

	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "executable") {
			t.Logf("[INFO] smartctl binary doesn't exist, but in testmode we continue: %v", err)
		} else {
			t.Errorf("[ERROR] Did not expect error, but received err: %v, data: %v\n", err, tester.DiskFailed)
		}
	}

	if len(tester.DiskFailed) > 1 {
		t.Errorf("[ERROR] len of failures should be 1 (/dev/sdc only), include: %v, disks: %v\n", tester.Include, tester.Disks)
	}
}

// TestSmartCtl_GatherExclude checks whether our ability to exclude disks from scanning is possible. Here we pass
// a known-good disk in Include, and Exclude it shortly thereafter (with an extra element in our disklist). If we see
// anything > 0 in the list of disks to parse, we fail out
func TestSmartCtl_GatherExclude(t *testing.T) {
	tester := &SmartCtl{Include: []string{"/dev/sda -d scsi"}, Exclude: []string{"/dev/sda -d scsi", "/dev/bus/0 -d megaraid,4"}}

	acc := new(testutil.Accumulator)
	err := tester.Gather(acc)

	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "executable") {
			t.Logf("[INFO] smartctl binary doesn't exist, but in testmode we continue: %v", err)
		} else {
			t.Errorf("[ERROR] Did not expect error, but received err: %v, data: %v\n", err, tester.DiskFailed)
		}
	}

	if len(tester.DiskOutput) > 0 {
		t.Errorf("[ERROR] Disks should have cancelled out, include: %v, exclude: %v, disks: %v\n", tester.Include, tester.Exclude, tester.Disks)
	}
}

// TestSmartCtl_ParseDisks checks whether we can parse any of the data at all
func TestSmartCtl_ParseDisks(t *testing.T) {
	tester := bencher
	tester.SudoPath, _ = exec.LookPath("sudo")
	tester.CtlPath, _ = exec.LookPath("smartctl")

	if err := tester.ParseDisks(); err != nil {
		t.Errorf("[ERROR] Unable to parse disks from the system: %v\n", err)
	}

	if len(tester.DiskFailed) > 0 {
		t.Logf("[INFO]  Did not get a good result back from ParseDisks: %v\n", tester.DiskFailed)
	}

	if len(tester.DiskOutput) > 0 {
		t.Logf("[INFO]  We parsed the disk info:\n")

		for _, each := range tester.DiskOutput {
			t.Logf("[%s] vendor: %s, product: %s, block: %s, serial: %s, rotation: %s, transport: %s, health: %s\n",
				each.Name, each.Vendor, each.Product, each.Block, each.Serial, each.Rotation, each.Transport, each.Health)
			t.Logf("[%s] stats: %#v\n", each.Name, each.Stats)
		}
	}
}

// BenchmarkParseString will check how performant our parsing is
func BenchmarkParseString(b *testing.B) {
	var buf bytes.Buffer
	var str string
	buf.WriteString(testData)

	findVendor := regexp.MustCompile(`Vendor:\s+(\w+)`)

	for n := 0; n < b.N; n++ {
		bencher.ParseString(findVendor, &buf, &str)
	}
}

// BenchmarkParseStringSlice will check how performant our parsing is
func BenchmarkParseStringSlice(b *testing.B) {
	var buf bytes.Buffer
	var str []string
	buf.WriteString(testData)

	findVerify := regexp.MustCompile(`verify:\s+(.*)\n`)

	for n := 0; n < b.N; n++ {
		bencher.ParseStringSlice(findVerify, &buf, &str)
	}
}

// BenchmarkParseFloat will check how performant our parsing is
func BenchmarkParseFloat(b *testing.B) {
	var buf bytes.Buffer
	var val float64
	buf.WriteString(testData)

	findTemp := regexp.MustCompile(`Current Drive Temperature:\s+([0-9]+)`)

	for n := 0; n < b.N; n++ {
		bencher.ParseFloat(findTemp, &buf, &val)
	}
}

// BenchmarkParseFloatSlice will check how performant our parsing is
func BenchmarkParseFloatSlice(b *testing.B) {
	var buf bytes.Buffer
	var val []float64
	buf.WriteString(testData)

	findVerify := regexp.MustCompile(`verify:\s+(.*)\n`)

	for n := 0; n < b.N; n++ {
		bencher.ParseFloatSlice(findVerify, &buf, &val)
	}
}

// BenchmarkExcludeDisks checks how quickly we can drill down to the set of disks to parse for
func BenchmarkExcludeDisks(b *testing.B) {
	tester := &SmartCtl{Disks: []string{
		"/dev/bus/0 -d megaraid,0",
		"/dev/bus/0 -d megaraid,1",
		"/dev/bus/0 -d megaraid,2",
		"/dev/bus/0 -d megaraid,3",
		"/dev/bus/0 -d megaraid,4",
		"/dev/bus/0 -d megaraid,5",
		"/dev/bus/0 -d megaraid,6",
		"/dev/bus/0 -d megaraid,7",
		"/dev/bus/0 -d megaraid,8",
		"/dev/bus/0 -d megaraid,9",
		"/dev/bus/0 -d megaraid,10",
		"/dev/bus/0 -d megaraid,11",
		"/dev/bus/0 -d megaraid,12",
		"/dev/bus/0 -d megaraid,13",
		"/dev/bus/0 -d megaraid,14",
		"/dev/bus/0 -d megaraid,15",
		"/dev/bus/0 -d megaraid,16",
		"/dev/bus/0 -d megaraid,17",
		"/dev/bus/0 -d megaraid,18",
		"/dev/bus/0 -d megaraid,19",
		"/dev/bus/0 -d megaraid,20",
		"/dev/bus/0 -d megaraid,21",
		"/dev/bus/0 -d megaraid,22",
		"/dev/bus/0 -d megaraid,23",
		"/dev/bus/0 -d megaraid,24",
		"/dev/bus/0 -d megaraid,25",
	}, Exclude: []string{"/dev/bus/0 -d megaraid,21"}, Init: true}

	for n := 0; n < b.N; n++ {
		tester.ExcludeDisks()
	}
}

var testData = `
smartctl 6.2 2013-07-26 r3841 [x86_64-linux-3.10.0-327.10.1.el7.jump7.x86_64] (local build)
Copyright (C) 2002-13, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Vendor:               SEAGATE
Product:              ST1200MM0007
Revision:             IS06
User Capacity:        1,200,243,695,616 bytes [1.20 TB]
Logical block size:   512 bytes
Logical block provisioning type unreported, LBPME=-1, LBPRZ=0
Rotation Rate:        10000 rpm
Form Factor:          2.5 inches
Logical Unit id:      0x5000c5007f4fefff
Serial number:        S3L183CF
Device type:          disk
Transport protocol:   SAS
Local Time is:        Tue Jan 17 16:19:22 2017 CST
SMART support is:     Available - device has SMART capability.
SMART support is:     Enabled
Temperature Warning:  Disabled or Not Supported
Read Cache is:        Enabled
Writeback Cache is:   Disabled

=== START OF READ SMART DATA SECTION ===
SMART Health Status: OK

Current Drive Temperature:     26 C
Drive Trip Temperature:        60 C

Manufactured in week 03 of year 2015
Specified cycle count over device lifetime:  10000
Accumulated start-stop cycles:  52
Specified load-unload count over device lifetime:  300000
Accumulated load-unload cycles:  761
Elements in grown defect list: 0

Vendor (Seagate) cache information
  Blocks sent to initiator = 2498780339
  Blocks received from initiator = 3563773581
  Blocks read from cache and sent to initiator = 2325589923
  Number of read and write commands whose size <= segment size = 15756254
  Number of read and write commands whose size > segment size = 740

Vendor (Seagate/Hitachi) factory information
  number of hours powered up = 12809.43
  number of minutes until next internal SMART test = 38

Error counter log:
           Errors Corrected by           Total   Correction     Gigabytes    Total
               ECC          rereads/    errors   algorithm      processed    uncorrected
           fast | delayed   rewrites  corrected  invocations   [10^9 bytes]  errors
read:   1492558298        0         0  1492558298          0       9653.059           0
write:         0        0         1         1          1       4144.460           0
verify: 2501279159        0         0  2501279159          0      14578.622           0

Non-medium error count:        6

SMART Self-test log
Num  Test              Status                 segment  LifeTime  LBA_first_err [SK ASC ASQ]
     Description                              number   (hours)
# 1  Reserved(7)       Completed                  48      72                 - [-   -    -]
# 2  Background short  Completed                  64      70                 - [-   -    -]
# 3  Background short  Completed                  64      70                 - [-   -    -]
# 4  Background short  Completed                  64      68                 - [-   -    -]
# 5  Background short  Completed                  64      61                 - [-   -    -]
# 6  Background short  Completed                  64      59                 - [-   -    -]
# 7  Background short  Completed                  64      53                 - [-   -    -]
# 8  Background short  Completed                  64      47                 - [-   -    -]
# 9  Background short  Completed                  64      35                 - [-   -    -]
#10  Background short  Completed                  64      25                 - [-   -    -]
#11  Background short  Completed                  64      21                 - [-   -    -]
#12  Background short  Completed                  64      16                 - [-   -    -]
#13  Background short  Completed                  64       0                 - [-   -    -]
Long (extended) Self Test duration: 8400 seconds [140.0 minutes]

Background scan results log
  Status: waiting until BMS interval timer expires
    Accumulated power on time, hours:minutes 12809:26 [768566 minutes]
    Number of background scans performed: 105,  scan progress: 0.00%
    Number of background medium scans performed: 105

   #  when        lba(hex)    [sk,asc,ascq]    reassign_status
   1  474:13  00000000123b12c6  [1,17,1]   Recovered via rewrite in-place
   2  553:44  00000000123b12c7  [1,17,1]   Recovered via rewrite in-place
   3  666:11  00000000123b12c8  [1,17,1]   Recovered via rewrite in-place
   4 2120:54  000000001292b871  [1,17,1]   Recovered via rewrite in-place
   5 2193:52  000000001292b876  [1,17,1]   Recovered via rewrite in-place
   6 2193:52  000000001292b877  [1,17,1]   Recovered via rewrite in-place
   7 2865:53  00000000123b12c7  [1,17,1]   Recovered via rewrite in-place
   8 2865:53  000000001292b878  [1,17,1]   Recovered via rewrite in-place
   9 2865:53  000000001292b879  [1,17,1]   Recovered via rewrite in-place
  10 3033:53  00000000123b12c6  [1,17,1]   Recovered via rewrite in-place
  11 3201:54  00000000123b12c8  [1,17,1]   Recovered via rewrite in-place
  12 4156:03  00000000123b12c7  [1,17,1]   Recovered via rewrite in-place
  13 5974:27  000000001292b874  [1,17,1]   Recovered via rewrite in-place
  14 8157:21  000000001292b875  [1,17,1]   Recovered via rewrite in-place
  15 11685:33  00000000123b12c8  [1,17,1]   Recovered via rewrite in-place

Protocol Specific port log page for SAS SSP
relative target port id = 1
  generation code = 0
  number of phys = 1
  phy identifier = 0
    attached device type: expander device
    attached reason: SMP phy control function
    reason: power on
    negotiated logical link rate: phy enabled; 6 Gbps
    attached initiator port: ssp=0 stp=0 smp=0
    attached target port: ssp=0 stp=0 smp=1
    SAS address = 0x5000c5007f4feffd
    attached SAS address = 0x500056b36789abff
    attached phy identifier = 4
    Invalid DWORD count = 0
    Running disparity error count = 0
    Loss of DWORD synchronization = 0
    Phy reset problem = 0
    Phy event descriptors:
     Invalid word count: 0
     Running disparity error count: 0
     Loss of dword synchronization count: 0
     Phy reset problem count: 0
relative target port id = 2
  generation code = 0
  number of phys = 1
  phy identifier = 1
    attached device type: no device attached
    attached reason: unknown
    reason: unknown
    negotiated logical link rate: phy enabled; unknown
    attached initiator port: ssp=0 stp=0 smp=0
    attached target port: ssp=0 stp=0 smp=0
    SAS address = 0x5000c5007f4feffe
    attached SAS address = 0x0
    attached phy identifier = 0
    Invalid DWORD count = 0
    Running disparity error count = 0
    Loss of DWORD synchronization = 0
    Phy reset problem = 0
    Phy event descriptors:
     Invalid word count: 0
     Running disparity error count: 0
     Loss of dword synchronization count: 0
     Phy reset problem count: 0
`

var testFailedCollect = `
smartctl 6.2 2013-07-26 r3841 [x86_64-linux-3.10.0-327.10.1.el7.jump7.x86_64] (local build)
Copyright (C) 2002-13, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Vendor:               DELL
Product:              PERC H710P
Revision:             3.13
User Capacity:        299,439,751,168 bytes [299 GB]
Logical block size:   512 bytes
Logical Unit id:      0x6848f690ee7499001c10f9d8041a7867
Serial number:        0067781a04d8f9101c009974ee90f648
Device type:          disk
Local Time is:        Thu Jan 19 14:03:10 2017 CST
SMART support is:     Unavailable - device lacks SMART capability.
Read Cache is:        Enabled
Writeback Cache is:   Enabled

=== START OF READ SMART DATA SECTION ===

Error Counter logging not supported

Device does not support Self Test logging
Device does not support Background scan results logging
scsiPrintSasPhy Log Sense Failed [unsupported scsi opcode]
`
