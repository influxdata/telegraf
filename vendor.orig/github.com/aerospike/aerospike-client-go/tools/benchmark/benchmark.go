// Copyright 2013-2016 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	as "github.com/aerospike/aerospike-client-go"
	asl "github.com/aerospike/aerospike-client-go/logger"
	ast "github.com/aerospike/aerospike-client-go/types"
)

type TStats struct {
	Exit       bool
	W, R       int // write and read counts
	WE, RE     int // write and read errors
	WTO, RTO   int // write and read timeouts
	WMin, WMax int64
	RMin, RMax int64
	WLat, RLat int64
	Wn, Rn     []int64
}

var countReportChan chan *TStats

var host = flag.String("h", "127.0.0.1", "Aerospike server seed hostnames or IP addresses")
var port = flag.Int("p", 3000, "Aerospike server seed hostname or IP address port number.")
var namespace = flag.String("n", "test", "Aerospike namespace.")
var set = flag.String("s", "testset", "Aerospike set name.")
var keyCount = flag.Int("k", 1000000, "Key/record count or key/record range.")

var user = flag.String("U", "", "User name.")
var password = flag.String("P", "", "User password.")

var binDef = flag.String("o", "I", "Bin object specification.\n\tI\t: Read/write integer bin.\n\tB:200\t: Read/write byte array bin of length 200.\n\tS:50\t: Read/write string bin of length 50.")
var concurrency = flag.Int("c", 32, "Number of goroutines to generate load.")
var workloadDef = flag.String("w", "I:100", "Desired workload.\n\tI:60\t: Linear 'insert' workload initializing 60% of the keys.\n\tRU:80\t: Random read/update workload with 80% reads and 20% writes.")
var latency = flag.String("L", "", "Latency <columns>,<shift>.\n\tShow transaction latency percentages using elapsed time ranges.\n\t<columns> Number of elapsed time ranges.\n\t<shift>   Power of 2 multiple between each range starting at column 3.")
var throughput = flag.Int64("g", 0, "Throttle transactions per second to a maximum value.\n\tIf tps is zero, do not throttle throughput.")
var timeout = flag.Int("T", 0, "Read/Write timeout in milliseconds.")
var maxRetries = flag.Int("maxRetries", 2, "Maximum number of retries before aborting the current transaction.")
var connQueueSize = flag.Int("queueSize", 4096, "Maximum number of connections to pool.")

var randBinData = flag.Bool("R", false, "Use dynamically generated random bin values instead of default static fixed bin values.")
var useMarshalling = flag.Bool("M", false, "Use marshaling a struct instead of simple key/value operations")
var debugMode = flag.Bool("d", false, "Run benchmarks in debug mode.")
var profileMode = flag.Bool("profile", false, "Run benchmarks with profiler active on port 6060.")
var showUsage = flag.Bool("u", false, "Show usage information.")

// parsed data
var binDataType string
var binDataSize int
var workloadType string
var workloadPercent int
var latBase, latCols int

// group mutex to wait for all load generating go routines to finish
var wg sync.WaitGroup

// throughput counter
var currThroughput int64
var lastReport int64

// Underscores are there so that the field name is the same as key/value mode
type dataStruct struct {
	I int64
	S string
	B []byte
}

var logger *log.Logger

func main() {
	var buf bytes.Buffer
	logger = log.New(&buf, "", log.LstdFlags|log.Lshortfile)
	logger.SetOutput(os.Stdout)

	// use all cpus in the system for concurrency
	log.Printf("Setting number of CPUs to use: %d", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())
	readFlags()

	countReportChan = make(chan *TStats, 4*(*concurrency)) // async chan

	if *debugMode {
		asl.Logger.SetLogger(logger)
		asl.Logger.SetLevel(asl.DEBUG)
	}

	// launch profiler if in profile mode
	if *profileMode {
		runtime.SetBlockProfileRate(1)
		go func() {
			logger.Println(http.ListenAndServe(":6060", nil))
		}()
	}

	printBenchmarkParams()

	clientPolicy := as.NewClientPolicy()
	// cache lots  connections
	clientPolicy.ConnectionQueueSize = *connQueueSize
	clientPolicy.User = *user
	clientPolicy.Password = *password
	clientPolicy.Timeout = 10 * time.Second
	client, err := as.NewClientWithPolicy(clientPolicy, *host, *port)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Println("Nodes Found:", client.GetNodeNames())

	go reporter()

	switch workloadType {
	case "I":
		wg.Add(*concurrency)
		for i := 1; i < *concurrency; i++ {
			go runBench_I(client, i-1, *keyCount / *concurrency)
		}
		go runBench_I(client, *concurrency-1, *keyCount / *concurrency + *keyCount%*concurrency)
		wg.Wait()
	case "RU":
		for i := 1; i < *concurrency; i++ {
			go runBench_RU(client, i-1, *keyCount / *concurrency)
		}
		runBench_RU(client, *concurrency-1, *keyCount / *concurrency + *keyCount%*concurrency)
	default:
		log.Fatal("Invalid workload type " + workloadType)
	}

	// send term to reporter, and wait for it to terminate
	countReportChan <- &TStats{Exit: true}
	time.Sleep(10 * time.Millisecond)
	<-countReportChan
}

func workloadToString() string {
	switch workloadType {
	case "RU":
		return fmt.Sprintf("Read %d%%, Write %d%%", workloadPercent, 100-workloadPercent)
	default:
		return fmt.Sprintf("Initialize %d%% of records", workloadPercent)
	}
}

func throughputToString() string {
	if *throughput <= 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", *throughput)
}

func printBenchmarkParams() {
	logger.Printf("hosts:\t\t%s", *host)
	logger.Printf("port:\t\t%d", *port)
	logger.Printf("namespace:\t\t%s", *namespace)
	logger.Printf("set:\t\t%s", *set)
	logger.Printf("keys/records:\t%d", *keyCount)
	logger.Printf("object spec:\t%s, size: %d", binDataType, binDataSize)
	logger.Printf("random bin values\t%v", *randBinData)
	logger.Printf("workload:\t\t%s", workloadToString())
	logger.Printf("concurrency:\t%d", *concurrency)
	logger.Printf("max throughput\t%s", throughputToString())
	logger.Printf("timeout\t\t%v ms", *timeout)
	logger.Printf("max retries\t\t%d", *maxRetries)
	logger.Printf("debug:\t\t%v", *debugMode)
	logger.Printf("latency:\t\t%d:%d", latBase, latCols)
}

// parses an string of (key:value) type
func parseValuedParam(param string) (string, *int) {
	re := regexp.MustCompile(`(\w+)([:,](\d+))?`)
	values := re.FindStringSubmatch(param)

	parStr := strings.ToUpper(strings.Trim(values[1], " "))

	// see if the value is supplied
	if len(values) > 3 && strings.Trim(values[3], " ") != "" {
		if value, err := strconv.Atoi(strings.Trim(values[3], " ")); err == nil {
			return parStr, &value
		}
	}

	return parStr, nil
}

func parseLatency(param string) (int, int) {
	re := regexp.MustCompile(`(\d+)[:,](\d+)`)
	values := re.FindStringSubmatch(param)

	// see if the value is supplied
	if len(values) > 2 && strings.Trim(values[1], " ") != "" && strings.Trim(values[2], " ") != "" {
		if value1, err := strconv.Atoi(strings.Trim(values[1], " ")); err == nil {
			if value2, err := strconv.Atoi(strings.Trim(values[2], " ")); err == nil {
				return value1, value2
			}
		}
	}

	logger.Fatal("Wrong latency values requested.")
	return 0, 0
}

// reads input flags and interprets the complex ones
func readFlags() {
	flag.Parse()

	if *showUsage {
		flag.Usage()
		os.Exit(0)
	}

	if *debugMode {
		asl.Logger.SetLevel(asl.INFO)
	}

	if *latency != "" {
		latCols, latBase = parseLatency(*latency)
	}

	var binDataSz, workloadPct *int

	binDataType, binDataSz = parseValuedParam(*binDef)
	if binDataSz != nil {
		binDataSize = *binDataSz
	} else {
		switch binDataType {
		case "B":
			binDataSize = 200
		case "S":
			binDataSize = 50
		}
	}

	workloadType, workloadPct = parseValuedParam(*workloadDef)
	if workloadPct != nil {
		workloadPercent = *workloadPct
	} else {
		switch workloadType {
		case "I":
			workloadPercent = 100
		case "RU":
			workloadPercent = 50
		}
	}
}

// new random bin generator based on benchmark specs
func getRandValue(xr *XorRand) as.Value {
	switch binDataType {
	case "B":
		return as.NewBytesValue(randBytes(binDataSize, xr))
	case "S":
		return as.NewStringValue(string(randBytes(binDataSize, xr)))
	default:
		return as.NewLongValue(xr.Int64())
	}
}

// new random bin generator based on benchmark specs
func getBin(xr *XorRand) *as.Bin {
	var bin *as.Bin
	switch binDataType {
	case "B":
		bin = &as.Bin{Name: "B", Value: getRandValue(xr)}
	case "S":
		bin = &as.Bin{Name: "S", Value: getRandValue(xr)}
	default:
		bin = &as.Bin{Name: "I", Value: getRandValue(xr)}
	}

	return bin
}

func setBin(bin *as.Bin, xr *XorRand) {
	switch binDataType {
	case "B":
		bin.Value = getRandValue(xr)
	case "S":
		bin.Value = getRandValue(xr)
	default:
		bin.Value = getRandValue(xr)
	}
}

// new random bin generator based on benchmark specs
func getDataStruct(xr *XorRand) *dataStruct {
	var ds *dataStruct
	switch binDataType {
	case "B":
		ds = &dataStruct{B: randBytes(binDataSize, xr)}
	case "S":
		ds = &dataStruct{S: string(randBytes(binDataSize, xr))}
	default:
		ds = &dataStruct{I: xr.Int64()}
	}

	return ds
}

// new random bin generator based on benchmark specs
func setDataStruct(ds *dataStruct, xr *XorRand) {
	switch binDataType {
	case "B":
		ds.B = randBytes(binDataSize, xr)
	case "S":
		ds.S = string(randBytes(binDataSize, xr))
	default:
		ds.I = xr.Int64()
	}
}

const random_alpha_num = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const l = 62

func randBytes(size int, xr *XorRand) []byte {
	buf := make([]byte, size, size)
	xr.Read(buf)
	return buf
}

func incOnError(op, timeout *int, err error) {
	if ae, ok := err.(ast.AerospikeError); ok && ae.ResultCode() == ast.TIMEOUT {
		*timeout++
	} else {
		*op++
	}
}

func runBench_I(client *as.Client, ident int, times int) {
	defer wg.Done()

	xr := NewXorRand()

	var err error
	var forceReport bool = false

	writepolicy := as.NewWritePolicy(0, 0)
	writepolicy.Timeout = time.Duration(*timeout) * time.Millisecond
	writepolicy.MaxRetries = *maxRetries

	defaultBin := getBin(xr)
	defaultObj := getDataStruct(xr)

	t := time.Now()
	var WCount int
	var writeErr int
	var writeTOErr int

	var tm time.Time
	var wLat int64
	var wLatTotal int64
	var wMinLat int64
	var wMaxLat int64

	wLatList := make([]int64, latCols+1)

	bin := defaultBin
	obj := defaultObj
	key, _ := as.NewKey(*namespace, *set, 0)
	partition := ident * times
	for i := 1; i <= times; i++ {
		wLat = 0
		key.SetValue(as.IntegerValue(partition + (i % times)))
		WCount++
		if !*useMarshalling {
			// if randomBin data has been requested
			if *randBinData {
				setBin(bin, xr)
			}
			tm = time.Now()
			if err = client.PutBins(writepolicy, key, bin); err != nil {
				incOnError(&writeErr, &writeTOErr, err)
			}
		} else {
			// if randomBin data has been requested
			if *randBinData {
				setDataStruct(obj, xr)
			}
			tm = time.Now()
			if err = client.PutObject(writepolicy, key, obj); err != nil {
				incOnError(&writeErr, &writeTOErr, err)
			}
		}
		wLat = int64(time.Now().Sub(tm) / time.Millisecond)
		wLatTotal += wLat

		// under 1 ms
		if wLat <= int64(latBase) {
			wLatList[0]++
		}

		for i := 1; i <= latCols; i++ {
			if wLat > int64(latBase<<uint(i-1)) {
				wLatList[i]++
			}
		}

		wMinLat = min(wLat, wMinLat)
		wMaxLat = max(wLat, wMaxLat)

		// if throughput is set, check for threshold. All goroutines add a record on first iteration,
		// so take that into account as well
		if *throughput > 0 {
			forceReport = atomic.LoadInt64(&currThroughput) >= (*throughput - int64(*concurrency))
			if !forceReport {
				atomic.AddInt64(&currThroughput, 1)
			}
		}

		if forceReport || (time.Now().Sub(t) > (99 * time.Millisecond)) {
			countReportChan <- &TStats{false, WCount, 0, writeErr, 0, writeTOErr, 0, wMinLat, wMaxLat, 0, 0, wLatTotal, 0, wLatList, nil}
			WCount = 0
			writeErr = 0
			writeTOErr = 0

			// reset stats
			wLatTotal = 0
			wMinLat, wMaxLat = 0, 0

			wLatList = make([]int64, latCols+1)

			t = time.Now()
		}

		if forceReport {
			forceReport = false
			// sleep till next report
			time.Sleep(time.Second - time.Duration(time.Now().UnixNano()-atomic.LoadInt64(&lastReport)))
		}
	}
	countReportChan <- &TStats{false, WCount, 0, writeErr, 0, writeTOErr, 0, wMinLat, wMaxLat, 0, 0, wLatTotal, 0, wLatList, nil}
}

func runBench_RU(client *as.Client, ident int, times int) {
	defer wg.Done()

	xr := NewXorRand()

	// var r *as.Record

	var err error
	var forceReport bool = false

	writepolicy := as.NewWritePolicy(0, 0)
	writepolicy.Timeout = time.Duration(*timeout) * time.Millisecond
	writepolicy.MaxRetries = *maxRetries

	readpolicy := writepolicy.GetBasePolicy()

	defaultBin := getBin(xr)
	defaultObj := getDataStruct(xr)

	t := time.Now()
	var WCount, RCount int
	var writeErr, readErr int
	var writeTOErr, readTOErr int

	var tm time.Time
	var wLat, rLat int64
	var wLatTotal, rLatTotal int64
	var wMinLat, rMinLat int64
	var wMaxLat, rMaxLat int64

	wLatList := make([]int64, latCols+1)
	rLatList := make([]int64, latCols+1)

	bin := defaultBin
	obj := defaultObj
	i := 0
	key, _ := as.NewKey(*namespace, *set, 0)
	partition := ident * times
	for {
		i++
		rLat, wLat = 0, 0
		key.SetValue(as.IntegerValue(partition + (i % times)))
		// key, _ := as.NewKey(*namespace, *set, as.IntegerValue(partition+(i%times)))
		if int(xr.Uint64()%100) >= workloadPercent {
			WCount++
			if !*useMarshalling {
				// if randomBin data has been requested
				if *randBinData {
					setBin(bin, xr)
				}
				tm = time.Now()
				err = client.PutBins(writepolicy, key, bin)
			} else {
				// if randomBin data has been requested
				if *randBinData {
					setDataStruct(obj, xr)
				}
				tm = time.Now()
				err = client.PutObject(writepolicy, key, obj)
			}
			wLat = int64(time.Now().Sub(tm) / time.Millisecond)
			wLatTotal += wLat
			if err != nil {
				incOnError(&writeErr, &writeTOErr, err)
			}

			// under 1 ms
			if wLat <= int64(latBase) {
				wLatList[0]++
			}

			for i := 1; i <= latCols; i++ {
				if wLat > int64(latBase<<uint(i-1)) {
					wLatList[i]++
				}
			}

			wMinLat = min(wLat, wMinLat)
			wMaxLat = max(wLat, wMaxLat)
		} else {
			RCount++
			if !*useMarshalling {
				tm = time.Now()
				_, err = client.Get(readpolicy, key, bin.Name)
			} else {
				tm = time.Now()
				err = client.GetObject(readpolicy, key, obj)
			}
			rLat = int64(time.Now().Sub(tm) / time.Millisecond)
			rLatTotal += rLat
			if err != nil {
				incOnError(&readErr, &readTOErr, err)
			}

			// under 1 ms
			if rLat <= int64(latBase) {
				rLatList[0]++
			}

			for i := 1; i <= latCols; i++ {
				if rLat > int64(latBase<<uint(i-1)) {
					rLatList[i]++
				}
			}

			rMinLat = min(rLat, rMinLat)
			rMaxLat = max(rLat, rMaxLat)
		}

		// if throughput is set, check for threshold. All goroutines add a record on first iteration,
		// so take that into account as well
		if *throughput > 0 {
			forceReport = atomic.LoadInt64(&currThroughput) >= (*throughput - int64(*concurrency))
			if !forceReport {
				atomic.AddInt64(&currThroughput, 1)
			}
		}

		if forceReport || (time.Now().Sub(t) > (99 * time.Millisecond)) {
			countReportChan <- &TStats{false, WCount, RCount, writeErr, readErr, writeTOErr, readTOErr, wMinLat, wMaxLat, rMinLat, rMaxLat, wLatTotal, rLatTotal, wLatList, rLatList}
			WCount, RCount = 0, 0
			writeErr, readErr = 0, 0
			writeTOErr, readTOErr = 0, 0

			// reset stats
			wLatTotal, rLatTotal = 0, 0
			wMinLat, wMaxLat = 0, 0
			rMinLat, rMaxLat = 0, 0

			wLatList = make([]int64, latCols+1)
			rLatList = make([]int64, latCols+1)

			t = time.Now()
		}

		if forceReport {
			forceReport = false
			// sleep till next report
			time.Sleep(time.Second - time.Duration(time.Now().UnixNano()-atomic.LoadInt64(&lastReport)))
		}
	}
	countReportChan <- &TStats{false, WCount, RCount, writeErr, readErr, writeTOErr, readTOErr, wMinLat, wMaxLat, rMinLat, rMaxLat, wLatTotal, rLatTotal, wLatList, rLatList}
}

// calculates transactions per second
func calcTPS(count int, duration time.Duration) int {
	return int(float64(count) / (float64(duration) / float64(time.Second)))
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// listens to transaction report channel, and print them out on intervals
func reporter() {
	var totalWCount, totalRCount int
	var totalWErrCount, totalRErrCount int
	var totalWTOCount, totalRTOCount int
	var totalCount, totalTOCount, totalErrCount int
	lastReportTime := time.Now()

	var memStats = new(runtime.MemStats)
	var lastTotalAllocs, lastPauseNs uint64

	// var wLat, rLat int64
	var wTotalLat, rTotalLat int64
	var wMinLat, rMinLat int64
	var wMaxLat, rMaxLat int64
	wLatList := make([]int64, latCols+1)
	rLatList := make([]int64, latCols+1)

	var strBuff bytes.Buffer

	memProfileStr := func() string {
		var res string
		if *debugMode {
			// GC stats
			runtime.ReadMemStats(memStats)
			allocMem := (memStats.TotalAlloc - lastTotalAllocs) / (1024)
			pauseNs := (memStats.PauseTotalNs - lastPauseNs) / 1e6
			res = fmt.Sprintf(" (malloc (KiB): %d, GC pause(ms): %d)",
				allocMem,
				pauseNs,
			)
			// GC
			lastPauseNs = memStats.PauseTotalNs
			lastTotalAllocs = memStats.TotalAlloc
		}

		return res
	}

Loop:
	for {
		select {
		case stats := <-countReportChan:
			totalWCount += stats.W
			totalRCount += stats.R

			totalWErrCount += stats.WE
			totalRErrCount += stats.RE

			totalWTOCount += stats.WTO
			totalRTOCount += stats.RTO

			totalCount += (stats.W + stats.R)
			totalErrCount += (stats.WE + stats.RE)
			totalTOCount += (stats.WTO + stats.RTO)

			wTotalLat += stats.WLat
			rTotalLat += stats.RLat

			for i := 0; i <= latCols; i++ {
				if stats.Wn != nil {
					wLatList[i] += stats.Wn[i]
				}
				if stats.Rn != nil {
					rLatList[i] += stats.Rn[i]
				}
			}

			if stats.RMax > rMaxLat {
				rMaxLat = stats.RMax
			}
			if stats.RMin < rMinLat {
				rMinLat = stats.RMin
			}
			if stats.WMax > wMaxLat {
				wMaxLat = stats.WMax
			}
			if stats.WMin < wMinLat {
				wMinLat = stats.WMin
			}

			if stats.Exit || time.Now().Sub(lastReportTime) >= time.Second {
				// reset throughput
				atomic.StoreInt64(&currThroughput, 0)
				atomic.StoreInt64(&lastReport, time.Now().UnixNano())

				if workloadType == "I" {
					logger.Printf("write(tps=%d timeouts=%d errors=%d totalCount=%d)%s",
						totalWCount, totalTOCount, totalErrCount, totalCount,
						memProfileStr(),
					)
				} else {
					logger.Printf(
						"write(tps=%d timeouts=%d errors=%d) read(tps=%d timeouts=%d errors=%d) total(tps=%d timeouts=%d errors=%d, count=%d)%s",
						totalWCount, totalWTOCount, totalWErrCount,
						totalRCount, totalRTOCount, totalRErrCount,
						totalWCount+totalRCount, totalTOCount, totalErrCount, totalCount,
						memProfileStr(),
					)
				}

				if *latency != "" {
					strBuff.WriteString(fmt.Sprintf("\t\tMin(ms)\tAvg(ms)\tMax(ms)\t|<=%4d ms\t", latBase))
					for i := 0; i < latCols; i++ {
						strBuff.WriteString(fmt.Sprintf("|>%4d ms\t", latBase<<uint(i)))
					}
					logger.Println(strBuff.String())
					strBuff.Reset()

					strBuff.WriteString(fmt.Sprintf("\tREAD\t%d\t%3.3f\t%d", rMinLat, float64(rTotalLat)/float64(totalRCount+1), rMaxLat))
					for i := 0; i <= latCols; i++ {
						strBuff.WriteString(fmt.Sprintf("\t|%7d/%4.2f%%", rLatList[i], float64(rLatList[i])/float64(totalRCount+1)*100))
					}
					logger.Println(strBuff.String())
					strBuff.Reset()

					strBuff.WriteString(fmt.Sprintf("\tWRITE\t%d\t%3.3f\t%d", wMinLat, float64(wTotalLat)/float64(totalWCount+1), wMaxLat))
					for i := 0; i <= latCols; i++ {
						strBuff.WriteString(fmt.Sprintf("\t|%7d/%4.2f%%", wLatList[i], float64(wLatList[i])/float64(totalWCount+1)*100))
					}
					logger.Println(strBuff.String())
					strBuff.Reset()
				}

				// reset stats
				wTotalLat, rTotalLat = 0, 0
				wMinLat, wMaxLat = 0, 0
				rMinLat, rMaxLat = 0, 0
				for i := 0; i <= latCols; i++ {
					wLatList[i] = 0
					rLatList[i] = 0
				}

				totalWCount, totalRCount = 0, 0
				totalWErrCount, totalRErrCount = 0, 0
				totalTOCount, totalWTOCount, totalRTOCount = 0, 0, 0
				lastReportTime = time.Now()

				if stats.Exit {
					break Loop
				}
			}
		}
	}
	countReportChan <- &TStats{}
}

type XorRand struct {
	src [2]uint64
}

func NewXorRand() *XorRand {
	return &XorRand{[2]uint64{uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())}}
}

func (r *XorRand) Int64() int64 {
	return int64(r.Uint64())
}

func (r *XorRand) Uint64() uint64 {
	s1 := r.src[0]
	s0 := r.src[1]
	r.src[0] = s0
	s1 ^= s1 << 23
	r.src[1] = (s1 ^ s0 ^ (s1 >> 17) ^ (s0 >> 26))
	return r.src[1] + s0
}

func (r *XorRand) Read(p []byte) (n int, err error) {
	l := len(p) / 8
	for i := 0; i < l; i += 8 {
		binary.PutUvarint(p[i:], r.Uint64())
	}
	return len(p), nil
}
