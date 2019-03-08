package chronyc

import (
	"errors"
	"fmt"
	//	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	ExecCommand = exec.Command // execCommand is used to mock commands in tests.
)

type Chrony struct {
	UseSudo         bool     `toml:"use_sudo"`
	ChronycCommands []string `toml:"chronyc_commands"`
	ClientsSummary  bool     `toml:"clients_summary"`
	ChronycPath     string   `toml:"chronyc_path"`
}

func (*Chrony) Description() string {
	return "Get standard chrony metrics, requires chronyc executable."
}

func (*Chrony) SampleConfig() string {
	return `
  ## You need chronyc 2.4 or newer to use this input. 
  ## Invokes "chronyc -c <command>" for each command in the list, collecting everything in output.

  ## Path to chronyc executable, if you need to use specific one.
  # chronyc_path = "/usr/bin/chronyc"
  
  ## chronyc command list to run. Possible elements:
  ##  - tracking
  ##  - serverstats
  ##  - sources
  ##  - sourcestats
  ##  - ntpdata
  ##  - rtcdata
  ##  - clients
  ##  - activity
  ##  - smoothing
  #
  # chronyc_commands = ["tracking", "sources", "sourcestats"]

  ## chronyc requires root access to unix domain socket to perform some commands:
  ##  - serverstats
  ##  - ntpdata
  ##  - rtcdata
  ##  - clients
  ##
  ## sudo must be configured to allow telegraf user to run chronyc as root to use this setting.
  # use_sudo = false

  ## "clients" command may report too many metrics, one line per client host. 
  ## When the following option is True, only summary metric is added to the result.
  # clients_summary = false
`
}

func (c *Chrony) Gather(acc telegraf.Accumulator) error {
	if len(c.ChronycPath) == 0 {
		return errors.New("chronyc not found: verify that chrony is installed and that chronyc is in your PATH")
	}

	name := c.ChronycPath
	argv := []string{}
	if c.UseSudo {
		name = "sudo"
		argv = append(argv, "-n", c.ChronycPath)
	}

	argv = append(argv, "-c")
	for _, command := range c.ChronycCommands {
		//		fmt.Fprintf(os.Stderr, "sending command: %s\n", command)
		argvCmd := append(argv, command)
		cmd := ExecCommand(name, argvCmd...)
		out, err := internal.CombinedOutputTimeout(cmd, time.Second*5)
		if err != nil {
			return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
		}
		//		fmt.Fprintf(os.Stderr, "Got output: %s\n", out)
		err = c.parseChronycOutput([]string{command}, string(out), acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Chrony) MultiGather(acc telegraf.Accumulator) error {
	if len(c.ChronycPath) == 0 {
		return errors.New("chronyc not found: verify that chrony is installed and that chronyc is in your PATH")
	}

	name := c.ChronycPath
	argv := []string{}
	if c.UseSudo {
		name = "sudo"
		argv = append(argv, "-n", c.ChronycPath)
	}

	argv = append(argv, "-c", "-m")
	argv = append(argv, c.ChronycCommands...)

	cmd := ExecCommand(name, argv...)
	out, err := internal.CombinedOutputTimeout(cmd, time.Second*5)
	if err != nil {
		return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
	}
	//fmt.Fprintf(os.Stderr, "Got output: %s\n", out)
	err = c.parseChronycOutput(c.ChronycCommands, string(out), acc)
	if err != nil {
		return err
	}

	return nil
}

type formatError struct {
	error
}

type fieldCountError struct {
	error
}

func parseSources(fields []string) (map[string]interface{}, map[string]string, error) {
	//fmt.Printf("Input >>>%v<<<\n", fields)

	var err error
	var found bool
	var clockRef string
	var clockMode, clockState int
	var stratum, poll, reach, lastRx int64
	var offset, rawOffset, errorMargin float64

	n := len(fields)
	if n != 10 {
		return nil, nil, fieldCountError{fmt.Errorf("Got %d instead of 10 fields in source line", n)}
	}

	modes := map[string]int{
		"^": 0,
		"=": 1,
		"#": 2,
		" ": -1,
	}
	states := map[string]int{
		"*": 0,
		"?": 1,
		"x": 2,
		"~": 3,
		"+": 4,
		"-": 5,
		" ": -1,
	}

	for i, field := range fields {
		switch i {
		case 0:
			clockMode, found = modes[field]
			if !found {
				err = fmt.Errorf("Unknown clock mode %q", field)
			}
		case 1:
			clockState, found = states[field]
			if !found {
				err = fmt.Errorf("Unknown clock state %q", field)
			}
		case 2:
			clockRef = field
		case 3:
			stratum, err = strconv.ParseInt(field, 10, 64)
		case 4:
			poll, err = strconv.ParseInt(field, 10, 64)
		case 5:
			reach, err = strconv.ParseInt(field, 8, 0)
		case 6:
			lastRx, err = strconv.ParseInt(field, 10, 64)
		case 7:
			offset, err = strconv.ParseFloat(field, 64)
		case 8:
			rawOffset, err = strconv.ParseFloat(field, 64)
		case 9:
			errorMargin, err = strconv.ParseFloat(field, 64)
		}
		if err != nil {
			return nil, nil, formatError{err}
		}
	}

	tFields := map[string]interface{}{
		"clockMode":  clockMode,
		"clockState": clockState,
		"poll":       poll,
		"reach":      reach,
	}
	if lastRx != 4294967295 {
		tFields["stratum"] = stratum
		tFields["lastRx"] = lastRx
		tFields["offset"] = offset
		tFields["rawOffset"] = rawOffset
		tFields["errorMargin"] = errorMargin
	}

	tTags := map[string]string{
		"command": "sources",
		"clockId": clockRef,
	}

	//	fmt.Printf("Source mode: %d, state: %d, ref: %s, stratum: %d, poll: %d, reach: %d, last rx: %d, offset: %e, raw offset: %e, error margin: %e\n",
	//		clockMode, clockState, clockRef, stratum, poll, reach, lastRx, offset, rawOffset, errorMargin)
	return tFields, tTags, nil
}

func parseSourceStats(fields []string) (map[string]interface{}, map[string]string, error) {
	//fmt.Printf("Input >>>%v<<<\n", fields)

	var err error
	var clockRef string
	var np, nr, span int64
	var frequency, freqSkew, offset, stdDev float64

	n := len(fields)
	if n != 8 {
		return nil, nil, fieldCountError{fmt.Errorf("Got %d instead of 8 fields in sourcestats line", n)}
	}

	for i, field := range fields {
		switch i {
		case 0:
			clockRef = field
		case 1:
			np, err = strconv.ParseInt(field, 10, 64)
		case 2:
			nr, err = strconv.ParseInt(field, 10, 64)
		case 3:
			span, err = strconv.ParseInt(field, 10, 64)
		case 4:
			frequency, err = strconv.ParseFloat(field, 64)
		case 5:
			freqSkew, err = strconv.ParseFloat(field, 64)
		case 6:
			offset, err = strconv.ParseFloat(field, 64)
		case 7:
			stdDev, err = strconv.ParseFloat(field, 64)
		}
		if err != nil {
			return nil, nil, formatError{err}
		}
	}

	tFields := map[string]interface{}{
		"np":        np,
		"nr":        nr,
		"span":      span,
		"frequency": frequency,
		"freqSkew":  freqSkew,
		"offset":    offset,
		"stdDev":    stdDev,
	}
	tTags := map[string]string{
		"command": "sourcestats",
		"clockId": clockRef,
	}

	//	fmt.Printf("SourceStats ref: %s, np: %d, nr: %d, span: %d, frequency: %f, freqSkew: %f, offset: %e, stdDev: %e\n",
	//		clockRef, np, nr, span, frequency, freqSkew, offset, stdDev)
	return tFields, tTags, nil
}

func parseNtpData(fields []string) (map[string]interface{}, map[string]string, error) {
	//fmt.Printf("Input >>>%v<<<\n", fields)

	var err error
	var remoteAddress, remoteAddressHex, localAddress, localAddressHex string
	var leapStatusStr, clockModeStr, refIdHex, refId string
	var remotePort, version, stratum, pollInterval, precision int64
	var pollIntervalSec, precisionSec, rootDelay, rootDispersion float64
	var refTime, offset, peerDelay, peerDispersion, responseTime float64
	var jitterAsymmetry float64
	var ntpTestsA, ntpTestsB, ntpTestsC string
	var interleaved, authenticated, txTimestamping, rxTimestamping string
	var totalTX, totalRX, totalValidRX int64

	n := len(fields)
	if n != 33 {
		return nil, nil, fieldCountError{fmt.Errorf("Got %d instead of 33 fields in ntpdata line", n)}
	}

	for i, field := range fields {
		switch i {
		case 0:
			remoteAddress = field
		case 1:
			remoteAddressHex = field
		case 2:
			remotePort, err = strconv.ParseInt(field, 10, 64)
		case 3:
			localAddress = field
		case 4:
			localAddressHex = field
		case 5:
			leapStatusStr = field
		case 6:
			version, err = strconv.ParseInt(field, 10, 64)
		case 7:
			clockModeStr = field
		case 8:
			stratum, err = strconv.ParseInt(field, 10, 64)
		case 9:
			pollInterval, err = strconv.ParseInt(field, 10, 64)
		case 10:
			pollIntervalSec, err = strconv.ParseFloat(field, 64)
		case 11:
			precision, err = strconv.ParseInt(field, 10, 64)
		case 12:
			precisionSec, err = strconv.ParseFloat(field, 64)
		case 13:
			rootDelay, err = strconv.ParseFloat(field, 64)
		case 14:
			rootDispersion, err = strconv.ParseFloat(field, 64)
		case 15:
			refIdHex = field
		case 16:
			refId = field
		case 17:
			refTime, err = strconv.ParseFloat(field, 64)
		case 18:
			offset, err = strconv.ParseFloat(field, 64)
		case 19:
			peerDelay, err = strconv.ParseFloat(field, 64)
		case 20:
			peerDispersion, err = strconv.ParseFloat(field, 64)
		case 21:
			responseTime, err = strconv.ParseFloat(field, 64)
		case 22:
			jitterAsymmetry, err = strconv.ParseFloat(field, 64)
		case 23:
			ntpTestsA = field
		case 24:
			ntpTestsB = field
		case 25:
			ntpTestsC = field
		case 26:
			interleaved = field
		case 27:
			authenticated = field
		case 28:
			txTimestamping = field
		case 29:
			rxTimestamping = field
		case 30:
			totalTX, err = strconv.ParseInt(field, 10, 64)
		case 31:
			totalRX, err = strconv.ParseInt(field, 10, 64)
		case 32:
			totalValidRX, err = strconv.ParseInt(field, 10, 64)
		}
		if err != nil {
			return nil, nil, formatError{err}
		}
	}

	if remoteAddress == "[UNSPEC]" {
		return nil, nil, nil
	}

	tFields := map[string]interface{}{
		"remoteAddress":    remoteAddress,
		"remoteAddressHex": remoteAddressHex,
		"remotePort":       remotePort,
		"localAddress":     localAddress,
		"localAddressHex":  localAddressHex,
		"leapStatus":       leapStatusStr,
		"version":          version,
		"clockModeStr":     clockModeStr,
		"stratum":          stratum,
		"pollInterval":     pollInterval,
		"pollIntervalSec":  pollIntervalSec,
		"precision":        precision,
		"precisionSec":     precisionSec,
		"rootDelay":        rootDelay,
		"rootDispersion":   rootDispersion,
		"refIdHex":         refIdHex,
		"refId":            refId,
		"refTime":          refTime,
		"offset":           offset,
		"peerDelay":        peerDelay,
		"peerDispersion":   peerDispersion,
		"responseTime":     responseTime,
		"jitterAsymmetry":  jitterAsymmetry,
		"ntpTestsA":        ntpTestsA,
		"ntpTestsB":        ntpTestsB,
		"ntpTestsC":        ntpTestsC,
		"interleaved":      interleaved,
		"authenticated":    authenticated,
		"txTimestamping":   txTimestamping,
		"rxTimestamping":   rxTimestamping,
		"totalTX":          totalTX,
		"totalRX":          totalRX,
		"totalValidRX":     totalValidRX,
	}
	tTags := map[string]string{
		"command":    "ntpdata",
		"clockId":    remoteAddress,
		"clockIdHex": remoteAddressHex,
	}

	//	fmt.Printf("NtpData remoteAddress: %s, remoteAddressHex: %s, remotePort: %d, localAddress: %s, localAddressHex: %s, leapStatusStr: %s, ",
	//		remoteAddress, remoteAddressHex, remotePort, localAddress, localAddressHex, leapStatusStr)
	//	fmt.Printf("version: %d, clockMode: %s, stratum: %d, pollInterval: %d, pollIntervalSec: %f, precision: %d, precisionSec: %f, ",
	//		version, clockMode, stratum, pollInterval, pollIntervalSec, precision, precisionSec)
	//	fmt.Printf("rootDelay: %f, rootDispersion: %f, refIdHex: %s, refId: %s, refTime: %f, offset: %f, peerDelay: %f, peerDispersion: %f, ",
	//		rootDelay, rootDispersion, refIdHex, refId, refTime, offset, peerDelay, peerDispersion)
	//	fmt.Printf("responseTime: %f, jitterAsymmetry: %f, ntpTestsA: %s, ntpTestsB: %s, ntpTestsC: %s, interleaved: %s, authenticated: %s, ",
	//		responseTime, jitterAsymmetry, ntpTestsA, ntpTestsB, ntpTestsC, interleaved, authenticated)
	//	fmt.Printf("txTimestamping: %s, rxTimestamping: %s, totalTX: %d, totalRX: %d, totalValidRX: %d\n",
	//		txTimestamping, rxTimestamping, totalTX, totalRX, totalValidRX)
	return tFields, tTags, nil
}

func parseTracking(fields []string) (map[string]interface{}, map[string]string, error) {
	//fmt.Printf("Input >>>%v<<<\n", fields)

	var err error
	n := len(fields)
	if n != 14 {
		return nil, nil, fieldCountError{fmt.Errorf("Got %d instead of 14 fields in tracking line", n)}
	}
	var refId, refIdHex, leapStatusStr string
	var stratum int64
	var refTime, systemTime, lastOffset, rMSOffset, frequency, freqResidual, freqSkew, rootDelay, rootDispersion, updateInterval float64

	for i, field := range fields {
		switch i {
		case 0:
			refIdHex = field
		case 1:
			refId = field
		case 2:
			stratum, err = strconv.ParseInt(field, 10, 64)
		case 3:
			refTime, err = strconv.ParseFloat(field, 64)
		case 4:
			systemTime, err = strconv.ParseFloat(field, 64)
		case 5:
			lastOffset, err = strconv.ParseFloat(field, 64)
		case 6:
			rMSOffset, err = strconv.ParseFloat(field, 64)
		case 7:
			frequency, err = strconv.ParseFloat(field, 64)
		case 8:
			freqResidual, err = strconv.ParseFloat(field, 64)
		case 9:
			freqSkew, err = strconv.ParseFloat(field, 64)
		case 10:
			rootDelay, err = strconv.ParseFloat(field, 64)
		case 11:
			rootDispersion, err = strconv.ParseFloat(field, 64)
		case 12:
			updateInterval, err = strconv.ParseFloat(field, 64)
		case 13:
			leapStatusStr = field
		}
		if err != nil {
			return nil, nil, formatError{err}
		}
	}

	tFields := map[string]interface{}{
		"refId":            refId,
		"refIdHex":         refIdHex,
		"stratum":          stratum,
		"refTime":          refTime,
		"systemTimeOffset": systemTime,
		"lastOffset":       lastOffset,
		"rmsOffset":        rMSOffset,
		"frequency":        frequency,
		"freqResidual":     freqResidual,
		"freqSkew":         freqSkew,
		"rootDelay":        rootDelay,
		"rootDispersion":   rootDispersion,
		"updateInterval":   updateInterval,
		"leapStatus":       leapStatusStr,
	}
	tTags := map[string]string{
		"command": "tracking",
		"clockId": "chrony",
	}
	//	fmt.Printf("Tracking refIdHex: %s, refId: %s, stratum: %d, refTime: %f, systemTime: %f, lastOffset: %f, rMSOffset: %f, frequency: %f, freqResidual: %f, freqSkew: %f, rootDelay: %f, rootDispersion: %f, updateInterval: %f, leapStatus: %s\n",
	//		refIdHex, refId, stratum, refTime, systemTime, lastOffset, rMSOffset, frequency, freqResidual, freqSkew, rootDelay, rootDispersion, updateInterval, leapStatusStr)
	return tFields, tTags, nil
}

func parseServerStats(fields []string) (map[string]interface{}, map[string]string, error) {
	//fmt.Printf("Input >>>%v<<<\n", fields)

	var err error
	n := len(fields)
	if n != 5 {
		return nil, nil, fieldCountError{fmt.Errorf("Got %d instead of 5 fields in serverstats line", n)}
	}
	var ntpPacketsReceived, ntpPacketsDropped, commandPacketsReceived, commandPacketsDropped, clientLogRecordsDropped int64

	for i, field := range fields {
		switch i {
		case 0:
			ntpPacketsReceived, err = strconv.ParseInt(field, 10, 64)
		case 1:
			ntpPacketsDropped, err = strconv.ParseInt(field, 10, 64)
		case 2:
			commandPacketsReceived, err = strconv.ParseInt(field, 10, 64)
		case 3:
			commandPacketsDropped, err = strconv.ParseInt(field, 10, 64)
		case 4:
			clientLogRecordsDropped, err = strconv.ParseInt(field, 10, 64)
		}
		if err != nil {
			return nil, nil, formatError{err}
		}
	}

	tFields := map[string]interface{}{
		"ntpPacketsReceived":      ntpPacketsReceived,
		"ntpPacketsDropped":       ntpPacketsDropped,
		"commandPacketsReceived":  commandPacketsReceived,
		"commandPacketsDropped":   commandPacketsDropped,
		"clientLogRecordsDropped": clientLogRecordsDropped,
	}
	tTags := map[string]string{
		"command": "serverstats",
		"clockId": "chrony",
	}

	//	fmt.Printf("ServerStats ntpPacketsReceived: %d, ntpPacketsDropped: %d, commandPacketsReceived: %d, commandPacketsDropped: %d, clientLogRecordsDropped: %d\n",
	//		ntpPacketsReceived, ntpPacketsDropped, commandPacketsReceived, commandPacketsDropped, clientLogRecordsDropped)
	return tFields, tTags, nil
}

func parseRtcData(fields []string) (map[string]interface{}, map[string]string, error) {
	//fmt.Printf("Input >>>%v<<<\n", fields)

	var err error
	n := len(fields)
	if n != 6 {
		return nil, nil, fieldCountError{fmt.Errorf("Got %d instead of 6 fields in rtcdata line", n)}
	}
	var rtcSamples, rtcRuns, rtcSampleSpan int64
	var rtcRefTime, rtcOffset, rtcFreq float64

	for i, field := range fields {
		switch i {
		case 0:
			rtcRefTime, err = strconv.ParseFloat(field, 64)
		case 1:
			rtcSamples, err = strconv.ParseInt(field, 10, 64)
		case 2:
			rtcRuns, err = strconv.ParseInt(field, 10, 64)
		case 3:
			rtcSampleSpan, err = strconv.ParseInt(field, 10, 64)
		case 4:
			rtcOffset, err = strconv.ParseFloat(field, 64)
		case 5:
			rtcFreq, err = strconv.ParseFloat(field, 64)
		}
		if err != nil {
			return nil, nil, formatError{err}
		}
	}

	tFields := map[string]interface{}{
		"rtcRefTime":    rtcRefTime,
		"rtcSamples":    rtcSamples,
		"rtcRuns":       rtcRuns,
		"rtcSampleSpan": rtcSampleSpan,
		"rtcOffset":     rtcOffset,
		"rtcFreq":       rtcFreq,
	}
	tTags := map[string]string{
		"command": "rtcdata",
		"clockId": "chrony",
	}

	return tFields, tTags, nil
}

type clients struct {
	summaryOnly      bool
	totalClients     int64
	ntpClients       int64
	activeNtpClients int64
}

func (cl *clients) parseClients(fields []string) (map[string]interface{}, map[string]string, error) {
	//fmt.Printf("Input >>>%v<<<\n", fields)

	var err error
	var clientAddress string
	var ntpRequests, ntpDropped, ntpInterval, ntpIntervalLimited, ntpLastRequest int64
	var cmdRequests, cmdDropped, cmdInterval, cmdLastRequest int64

	n := len(fields)
	if n != 10 {
		return nil, nil, fieldCountError{fmt.Errorf("Got %d instead of 10 fields in clients line", n)}
	}

	for i, field := range fields {
		switch i {
		case 0:
			clientAddress = field
		case 1:
			ntpRequests, err = strconv.ParseInt(field, 10, 64)
		case 2:
			ntpDropped, err = strconv.ParseInt(field, 10, 64)
		case 3:
			ntpInterval, err = strconv.ParseInt(field, 10, 64)
		case 4:
			ntpIntervalLimited, err = strconv.ParseInt(field, 10, 64)
		case 5:
			ntpLastRequest, err = strconv.ParseInt(field, 10, 64)
		case 6:
			cmdRequests, err = strconv.ParseInt(field, 10, 64)
		case 7:
			cmdDropped, err = strconv.ParseInt(field, 10, 64)
		case 8:
			cmdInterval, err = strconv.ParseInt(field, 10, 64)
		case 9:
			cmdLastRequest, err = strconv.ParseInt(field, 10, 64)
		}
		if err != nil {
			return nil, nil, formatError{err}
		}
	}

	cl.totalClients += 1
	if ntpRequests != 0 {
		cl.ntpClients += 1
	}
	tFields := map[string]interface{}{
		"ntpRequests": ntpRequests,
		"ntpDropped":  ntpDropped,
		"cmdRequests": cmdRequests,
		"cmdDropped":  cmdDropped,
	}
	if ntpInterval != 127 {
		tFields["ntpInterval"] = ntpInterval
	}
	if ntpIntervalLimited != 127 {
		tFields["ntpIntervalLimited"] = ntpIntervalLimited
	}
	if ntpLastRequest != 4294967295 {
		tFields["ntpLastRequest"] = ntpLastRequest
		if ntpLastRequest <= 3600 {
			cl.activeNtpClients += 1
		}
	}
	if cmdInterval != 127 {
		tFields["cmdInterval"] = cmdInterval
	}
	if cmdLastRequest != 4294967295 {
		tFields["cmdLastRequest"] = cmdLastRequest
	}

	tTags := map[string]string{
		"command":       "clients",
		"clientAddress": clientAddress,
	}

	if cl.summaryOnly {
		return nil, nil, nil
	}
	return tFields, tTags, nil
}

func (cl *clients) summarizeClients() (map[string]interface{}, map[string]string, error) {

	tFields := map[string]interface{}{
		"totalClients":     cl.totalClients,
		"ntpClients":       cl.ntpClients,
		"activeNtpClients": cl.activeNtpClients,
	}

	tTags := map[string]string{
		"command": "clients",
	}
	return tFields, tTags, nil
}

func parseActivity(fields []string) (map[string]interface{}, map[string]string, error) {
	//fmt.Printf("Input >>>%v<<<\n", fields)

	var err error
	n := len(fields)
	if n != 5 {
		return nil, nil, fieldCountError{fmt.Errorf("Got %d instead of 5 fields in activity line", n)}
	}
	var sourcesOnline, sourcesOffline, sourcesBurstToOnline, sourcesBurstToOffline, sourcesUnknownAddress int64

	for i, field := range fields {
		switch i {
		case 0:
			sourcesOnline, err = strconv.ParseInt(field, 10, 64)
		case 1:
			sourcesOffline, err = strconv.ParseInt(field, 10, 64)
		case 2:
			sourcesBurstToOnline, err = strconv.ParseInt(field, 10, 64)
		case 3:
			sourcesBurstToOffline, err = strconv.ParseInt(field, 10, 64)
		case 4:
			sourcesUnknownAddress, err = strconv.ParseInt(field, 10, 64)
		}
		if err != nil {
			return nil, nil, formatError{err}
		}
	}

	tFields := map[string]interface{}{
		"sourcesOnline":         sourcesOnline,
		"sourcesOffline":        sourcesOffline,
		"sourcesBurstToOnline":  sourcesBurstToOnline,
		"sourcesBurstToOffline": sourcesBurstToOffline,
		"sourcesUnknownAddress": sourcesUnknownAddress,
	}
	tTags := map[string]string{
		"command": "activity",
		"clockId": "chrony",
	}

	return tFields, tTags, nil
}

func parseSmoothing(fields []string) (map[string]interface{}, map[string]string, error) {
	//fmt.Printf("Input >>>%v<<<\n", fields)

	var err error
	n := len(fields)
	if n != 7 {
		return nil, nil, fieldCountError{fmt.Errorf("Got %d instead of 7 fields in smoothing line", n)}
	}
	var smoothingActive, smoothingLeapOnly bool
	var smoothingOffset, smoothingFreq, smoothingFreqWander, smoothingLastUpdate, smoothingRemainingTime float64

	for i, field := range fields {
		switch i {
		case 0:
			smoothingActive = (field != "No")
		case 1:
			smoothingLeapOnly = (field != "")
		case 2:
			smoothingOffset, err = strconv.ParseFloat(field, 64)
		case 3:
			smoothingFreq, err = strconv.ParseFloat(field, 64)
		case 4:
			smoothingFreqWander, err = strconv.ParseFloat(field, 64)
		case 5:
			smoothingLastUpdate, err = strconv.ParseFloat(field, 64)
		case 6:
			smoothingRemainingTime, err = strconv.ParseFloat(field, 64)
		}
		if err != nil {
			return nil, nil, formatError{err}
		}
	}

	tFields := map[string]interface{}{
		"smoothingActive":        smoothingActive,
		"smoothingLeapOnly":      smoothingLeapOnly,
		"smoothingOffset":        smoothingOffset,
		"smoothingFreq":          smoothingFreq,
		"smoothingFreqWander":    smoothingFreqWander,
		"smoothingLastUpdate":    smoothingLastUpdate,
		"smoothingRemainingTime": smoothingRemainingTime,
	}
	tTags := map[string]string{
		"command": "smoothing",
		"clockId": "chrony",
	}

	return tFields, tTags, nil
}

// This function can potentially parse output from multiple commands,
// but there is a problem: whenever field count in two successive commands is the same,
// there is no solid way to differentiate between them.
func (c *Chrony) parseChronycOutput(commandList []string, out string, acc telegraf.Accumulator) error {

	var tFields map[string]interface{}
	var tTags map[string]string

	type commandRef struct {
		// A singleline command returns exactly 1 line of output.
		// All other commands return 0 or more lines.
		singleLine bool
		// field count in the output
		fields int
		// parser function
		lineParser func(fields []string) (map[string]interface{}, map[string]string, error)
		// summarizing function
		summary func() (map[string]interface{}, map[string]string, error)
	}

	cl := clients{
		summaryOnly: c.ClientsSummary,
	}
	command := map[string]*commandRef{
		"tracking":    {true, 14, parseTracking, nil},
		"serverstats": {true, 5, parseServerStats, nil},
		"sources":     {false, 10, parseSources, nil},
		"sourcestats": {false, 8, parseSourceStats, nil},
		"ntpdata":     {false, 33, parseNtpData, nil},
		"rtcdata":     {true, 6, parseRtcData, nil},
		"clients":     {false, 10, cl.parseClients, cl.summarizeClients},
		"activity":    {true, 5, parseActivity, nil},
		"smoothing":   {true, 7, parseSmoothing, nil},
	}

	var cmd *commandRef
	var commandName string

	lines := strings.Split(out, "\n")
	// There are always >=1 elements in the slice. Of them, last line is always empty.
	// But we better test it before throwaway.
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	lineDone := -1
	cmdList := commandList

	for currentLine, line := range lines {

		fields := strings.Split(line, ",")
		// empty line maps to zero fields
		if line == "" {
			fields = []string{}
		}

		var err error
	LoopCmd:
		for {
			if cmd == nil {
				if len(cmdList) == 0 {
					return fmt.Errorf("Commands done, but there is more output: %#v\n", lines[lineDone+1:])
				}
				commandName = cmdList[0]
				var found bool
				cmd, found = command[commandName]
				if !found {
					return fmt.Errorf("Unknown command '%s'", commandName)
				}
			}
			err = nil
			// When got wrong number of fields in output,
			// this may mean that it is time to switch to next command.
			if len(fields) != cmd.fields {
				if cmd.singleLine {
					return fmt.Errorf("Wrong field count for mandatory command '%s': %d, must be %d",
						commandName, len(fields), cmd.fields)
				} else {
					// try next command
					cmd = nil
					cmdList = cmdList[1:]
					continue LoopCmd
				}
			}
			// process the line with current cmd
			tFields, tTags, err = cmd.lineParser(fields)
			if err != nil {
				return err
			}
			if tFields != nil || tTags != nil {
				acc.AddFields("chronyc", tFields, tTags)
			}
			if cmd.singleLine {
				// done with it, what's next
				cmd = nil
				// Next line will belong to the next command
				cmdList = cmdList[1:]
			}
			break LoopCmd
		}
		lineDone = currentLine
	}

	if len(lines) == lineDone+1 {
		for _, cmdName := range cmdList {
			cmd, found := command[cmdName]
			if !found {
				return fmt.Errorf("Unknown cmd '%s'", cmdName)
			}
			if cmd.singleLine {
				return fmt.Errorf("Not enough output for the command: %s\n", cmdName)
			}
		}
	}

	// Add summary metrics for each of processed commands
	for _, commandName := range commandList {
		//fmt.Printf("Debug summary for command %s\n", commandName)
		cmd, found := command[commandName]
		if found && cmd.summary != nil {
			//fmt.Printf("Execute summary for command %s\n", commandName)
			tFields, tTags, err := cmd.summary()
			if err != nil {
				return err
			}
			if tFields != nil || tTags != nil {
				acc.AddFields("chronyc", tFields, tTags)
			}
		}
	}
	return nil
}

func init() {
	c := Chrony{
		ChronycCommands: []string{"tracking", "sources", "sourcestats"},
	}
	path, _ := exec.LookPath("chronyc")
	if len(path) > 0 {
		c.ChronycPath = path
	}
	inputs.Add("chronyc", func() telegraf.Input {
		return &c
	})
}
