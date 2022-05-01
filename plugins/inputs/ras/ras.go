//go:build linux && (386 || amd64 || arm || arm64)
// +build linux
// +build 386 amd64 arm arm64

package ras

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite" //to register SQLite driver

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Ras plugin gathers and counts errors provided by RASDaemon
type Ras struct {
	DBPath string `toml:"db_path"`

	Log telegraf.Logger `toml:"-"`

	db                *sql.DB
	latestTimestamp   time.Time
	cpuSocketCounters map[int]metricCounters
	serverCounters    metricCounters
}

type machineCheckError struct {
	ID           int
	Timestamp    string
	SocketID     int
	ErrorMsg     string
	MciStatusMsg string
}

type metricCounters map[string]int64

const (
	mceQuery = `
		SELECT
			id, timestamp, error_msg, mcistatus_msg, socketid
		FROM mce_record
		WHERE timestamp > ?
		`
	defaultDbPath          = "/var/lib/rasdaemon/ras-mc_event.db"
	dateLayout             = "2006-01-02 15:04:05 -0700"
	memoryReadCorrected    = "memory_read_corrected_errors"
	memoryReadUncorrected  = "memory_read_uncorrectable_errors"
	memoryWriteCorrected   = "memory_write_corrected_errors"
	memoryWriteUncorrected = "memory_write_uncorrectable_errors"
	instructionCache       = "cache_l0_l1_errors"
	instructionTLB         = "tlb_instruction_errors"
	levelTwoCache          = "cache_l2_errors"
	upi                    = "upi_errors"
	processorBase          = "processor_base_errors"
	processorBus           = "processor_bus_errors"
	internalTimer          = "internal_timer_errors"
	smmHandlerCode         = "smm_handler_code_access_violation_errors"
	internalParity         = "internal_parity_errors"
	frc                    = "frc_errors"
	externalMCEBase        = "external_mce_errors"
	microcodeROMParity     = "microcode_rom_parity_errors"
	unclassifiedMCEBase    = "unclassified_mce_errors"
)

// Start initializes connection to DB, metrics are gathered in Gather
func (r *Ras) Start(telegraf.Accumulator) error {
	err := validateDbPath(r.DBPath)
	if err != nil {
		return err
	}

	r.db, err = connectToDB(r.DBPath)
	if err != nil {
		return err
	}

	return nil
}

// Stop closes any existing DB connection
func (r *Ras) Stop() {
	if r.db != nil {
		err := r.db.Close()
		if err != nil {
			r.Log.Errorf("Error appeared during closing DB (%s): %v", r.DBPath, err)
		}
	}
}

// Gather reads the stats provided by RASDaemon and writes it to the Accumulator.
func (r *Ras) Gather(acc telegraf.Accumulator) error {
	rows, err := r.db.Query(mceQuery, r.latestTimestamp)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		mcError, err := fetchMachineCheckError(rows)
		if err != nil {
			return err
		}
		tsErr := r.updateLatestTimestamp(mcError.Timestamp)
		if tsErr != nil {
			return err
		}
		r.updateCounters(mcError)
	}

	addCPUSocketMetrics(acc, r.cpuSocketCounters)
	addServerMetrics(acc, r.serverCounters)

	return nil
}

func (r *Ras) updateLatestTimestamp(timestamp string) error {
	ts, err := parseDate(timestamp)
	if err != nil {
		return err
	}
	if ts.After(r.latestTimestamp) {
		r.latestTimestamp = ts
	}

	return nil
}

func (r *Ras) updateCounters(mcError *machineCheckError) {
	if strings.Contains(mcError.ErrorMsg, "No Error") {
		return
	}

	r.initializeCPUMetricDataIfRequired(mcError.SocketID)
	r.updateSocketCounters(mcError)
	r.updateServerCounters(mcError)
}

func newMetricCounters() *metricCounters {
	return &metricCounters{
		memoryReadCorrected:    0,
		memoryReadUncorrected:  0,
		memoryWriteCorrected:   0,
		memoryWriteUncorrected: 0,
		instructionCache:       0,
		instructionTLB:         0,
		processorBase:          0,
		processorBus:           0,
		internalTimer:          0,
		smmHandlerCode:         0,
		internalParity:         0,
		frc:                    0,
		externalMCEBase:        0,
		microcodeROMParity:     0,
		unclassifiedMCEBase:    0,
	}
}

func (r *Ras) updateServerCounters(mcError *machineCheckError) {
	if strings.Contains(mcError.ErrorMsg, "CACHE Level-2") && strings.Contains(mcError.ErrorMsg, "Error") {
		r.serverCounters[levelTwoCache]++
	}

	if strings.Contains(mcError.ErrorMsg, "UPI:") {
		r.serverCounters[upi]++
	}
}

func validateDbPath(dbPath string) error {
	pathInfo, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("provided db_path does not exist: [%s]", dbPath)
	}

	if err != nil {
		return fmt.Errorf("cannot get system information for db_path file: [%s] - %v", dbPath, err)
	}

	if mode := pathInfo.Mode(); !mode.IsRegular() {
		return fmt.Errorf("provided db_path does not point to a regular file: [%s]", dbPath)
	}

	return nil
}

func connectToDB(dbPath string) (*sql.DB, error) {
	return sql.Open("sqlite", dbPath)
}

func (r *Ras) initializeCPUMetricDataIfRequired(socketID int) {
	if _, ok := r.cpuSocketCounters[socketID]; !ok {
		r.cpuSocketCounters[socketID] = *newMetricCounters()
	}
}

func (r *Ras) updateSocketCounters(mcError *machineCheckError) {
	r.updateMemoryCounters(mcError)
	r.updateProcessorBaseCounters(mcError)

	if strings.Contains(mcError.ErrorMsg, "Instruction TLB") && strings.Contains(mcError.ErrorMsg, "Error") {
		r.cpuSocketCounters[mcError.SocketID][instructionTLB]++
	}

	if strings.Contains(mcError.ErrorMsg, "BUS") && strings.Contains(mcError.ErrorMsg, "Error") {
		r.cpuSocketCounters[mcError.SocketID][processorBus]++
	}

	if (strings.Contains(mcError.ErrorMsg, "CACHE Level-0") ||
		strings.Contains(mcError.ErrorMsg, "CACHE Level-1")) &&
		strings.Contains(mcError.ErrorMsg, "Error") {
		r.cpuSocketCounters[mcError.SocketID][instructionCache]++
	}
}

func (r *Ras) updateProcessorBaseCounters(mcError *machineCheckError) {
	if strings.Contains(mcError.ErrorMsg, "Internal Timer error") {
		r.cpuSocketCounters[mcError.SocketID][internalTimer]++
		r.cpuSocketCounters[mcError.SocketID][processorBase]++
	}

	if strings.Contains(mcError.ErrorMsg, "SMM Handler Code Access Violation") {
		r.cpuSocketCounters[mcError.SocketID][smmHandlerCode]++
		r.cpuSocketCounters[mcError.SocketID][processorBase]++
	}

	if strings.Contains(mcError.ErrorMsg, "Internal parity error") {
		r.cpuSocketCounters[mcError.SocketID][internalParity]++
		r.cpuSocketCounters[mcError.SocketID][processorBase]++
	}

	if strings.Contains(mcError.ErrorMsg, "FRC error") {
		r.cpuSocketCounters[mcError.SocketID][frc]++
		r.cpuSocketCounters[mcError.SocketID][processorBase]++
	}

	if strings.Contains(mcError.ErrorMsg, "External error") {
		r.cpuSocketCounters[mcError.SocketID][externalMCEBase]++
		r.cpuSocketCounters[mcError.SocketID][processorBase]++
	}

	if strings.Contains(mcError.ErrorMsg, "Microcode ROM parity error") {
		r.cpuSocketCounters[mcError.SocketID][microcodeROMParity]++
		r.cpuSocketCounters[mcError.SocketID][processorBase]++
	}

	if strings.Contains(mcError.ErrorMsg, "Unclassified") || strings.Contains(mcError.ErrorMsg, "Internal unclassified") {
		r.cpuSocketCounters[mcError.SocketID][unclassifiedMCEBase]++
		r.cpuSocketCounters[mcError.SocketID][processorBase]++
	}
}

func (r *Ras) updateMemoryCounters(mcError *machineCheckError) {
	if strings.Contains(mcError.ErrorMsg, "Memory read error") {
		if strings.Contains(mcError.MciStatusMsg, "Corrected_error") {
			r.cpuSocketCounters[mcError.SocketID][memoryReadCorrected]++
		} else {
			r.cpuSocketCounters[mcError.SocketID][memoryReadUncorrected]++
		}
	}
	if strings.Contains(mcError.ErrorMsg, "Memory write error") {
		if strings.Contains(mcError.MciStatusMsg, "Corrected_error") {
			r.cpuSocketCounters[mcError.SocketID][memoryWriteCorrected]++
		} else {
			r.cpuSocketCounters[mcError.SocketID][memoryWriteUncorrected]++
		}
	}
}

func addCPUSocketMetrics(acc telegraf.Accumulator, cpuSocketCounters map[int]metricCounters) {
	for socketID, data := range cpuSocketCounters {
		tags := map[string]string{
			"socket_id": strconv.Itoa(socketID),
		}
		fields := make(map[string]interface{})

		for errorName, count := range data {
			fields[errorName] = count
		}

		acc.AddCounter("ras", fields, tags)
	}
}

func addServerMetrics(acc telegraf.Accumulator, counters map[string]int64) {
	fields := make(map[string]interface{})
	for errorName, count := range counters {
		fields[errorName] = count
	}

	acc.AddCounter("ras", fields, map[string]string{})
}

func fetchMachineCheckError(rows *sql.Rows) (*machineCheckError, error) {
	mcError := &machineCheckError{}
	err := rows.Scan(&mcError.ID, &mcError.Timestamp, &mcError.ErrorMsg, &mcError.MciStatusMsg, &mcError.SocketID)

	if err != nil {
		return nil, err
	}

	return mcError, nil
}

func parseDate(date string) (time.Time, error) {
	return time.Parse(dateLayout, date)
}

func init() {
	inputs.Add("ras", func() telegraf.Input {
		defaultTimestamp, _ := parseDate("1970-01-01 00:00:01 -0700")
		return &Ras{
			DBPath:          defaultDbPath,
			latestTimestamp: defaultTimestamp,
			cpuSocketCounters: map[int]metricCounters{
				0: *newMetricCounters(),
			},
			serverCounters: map[string]int64{
				levelTwoCache: 0,
				upi:           0,
			},
		}
	})
}
