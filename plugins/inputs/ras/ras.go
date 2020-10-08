// +build !windows

package ras

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Ras struct {
	DbPath            string
	latestTimestamp   time.Time
	cpuSocketCounters map[int]metricCounters
	serverCounters    metricCounters
}

type machineCheckError struct {
	Id           int
	Timestamp    string
	SocketId     int
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

func (r *Ras) SampleConfig() string {
	return `
  ## Optional path to RASDaemon sqlite3 database.
  ## Default: /var/lib/rasdaemon/ras-mc_event.db
  # db_path = ""
`
}

func (r *Ras) Description() string {
	return "RAS plugin exposes counter metrics for Machine Check Errors provided by RASDaemon (sqlite3 output is required)."
}

func (r *Ras) Gather(acc telegraf.Accumulator) error {
	db, err := connectToDB(r.DbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query(mceQuery, r.latestTimestamp)
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

	addCpuSocketMetrics(acc, r.cpuSocketCounters)
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

	r.initializeCpuMetricDataIfRequired(mcError.SocketId)
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
		r.serverCounters[levelTwoCache] += 1
	}

	if strings.Contains(mcError.ErrorMsg, "UPI:") {
		r.serverCounters[upi] += 1
	}
}

func connectToDB(server string) (*sql.DB, error) {
	return sql.Open("sqlite3", server)
}

func (r *Ras) initializeCpuMetricDataIfRequired(socketId int) {
	if _, ok := r.cpuSocketCounters[socketId]; !ok {
		r.cpuSocketCounters[socketId] = *newMetricCounters()
	}
}

func (r *Ras) updateSocketCounters(mcError *machineCheckError) {
	r.updateMemoryCounters(mcError)
	r.updateProcessorBaseCounters(mcError)

	if strings.Contains(mcError.ErrorMsg, "Instruction TLB") && strings.Contains(mcError.ErrorMsg, "Error") {
		r.cpuSocketCounters[mcError.SocketId][instructionTLB] += 1
	}

	if strings.Contains(mcError.ErrorMsg, "BUS") && strings.Contains(mcError.ErrorMsg, "Error") {
		r.cpuSocketCounters[mcError.SocketId][processorBus] += 1
	}

	if (strings.Contains(mcError.ErrorMsg, "CACHE Level-0") ||
		strings.Contains(mcError.ErrorMsg, "CACHE Level-1")) &&
		strings.Contains(mcError.ErrorMsg, "Error") {
		r.cpuSocketCounters[mcError.SocketId][instructionCache] += 1
	}
}

func (r *Ras) updateProcessorBaseCounters(mcError *machineCheckError) {
	if strings.Contains(mcError.ErrorMsg, "Internal Timer error") {
		r.cpuSocketCounters[mcError.SocketId][internalTimer] += 1
		r.cpuSocketCounters[mcError.SocketId][processorBase] += 1
	}

	if strings.Contains(mcError.ErrorMsg, "SMM Handler Code Access Violation") {
		r.cpuSocketCounters[mcError.SocketId][smmHandlerCode] += 1
		r.cpuSocketCounters[mcError.SocketId][processorBase] += 1
	}

	if strings.Contains(mcError.ErrorMsg, "Internal parity error") {
		r.cpuSocketCounters[mcError.SocketId][internalParity] += 1
		r.cpuSocketCounters[mcError.SocketId][processorBase] += 1
	}

	if strings.Contains(mcError.ErrorMsg, "FRC error") {
		r.cpuSocketCounters[mcError.SocketId][frc] += 1
		r.cpuSocketCounters[mcError.SocketId][processorBase] += 1
	}

	if strings.Contains(mcError.ErrorMsg, "External error") {
		r.cpuSocketCounters[mcError.SocketId][externalMCEBase] += 1
		r.cpuSocketCounters[mcError.SocketId][processorBase] += 1
	}

	if strings.Contains(mcError.ErrorMsg, "Microcode ROM parity error") {
		r.cpuSocketCounters[mcError.SocketId][microcodeROMParity] += 1
		r.cpuSocketCounters[mcError.SocketId][processorBase] += 1
	}

	if strings.Contains(mcError.ErrorMsg, "Unclassified") || strings.Contains(mcError.ErrorMsg, "Internal unclassified") {
		r.cpuSocketCounters[mcError.SocketId][unclassifiedMCEBase] += 1
		r.cpuSocketCounters[mcError.SocketId][processorBase] += 1
	}
}

func (r *Ras) updateMemoryCounters(mcError *machineCheckError) {
	if strings.Contains(mcError.ErrorMsg, "Memory read error") {
		if strings.Contains(mcError.MciStatusMsg, "Corrected_error") {
			r.cpuSocketCounters[mcError.SocketId][memoryReadCorrected] += 1
		} else {
			r.cpuSocketCounters[mcError.SocketId][memoryReadUncorrected] += 1
		}
	}
	if strings.Contains(mcError.ErrorMsg, "Memory write error") {
		if strings.Contains(mcError.MciStatusMsg, "Corrected_error") {
			r.cpuSocketCounters[mcError.SocketId][memoryWriteCorrected] += 1
		} else {
			r.cpuSocketCounters[mcError.SocketId][memoryWriteUncorrected] += 1
		}
	}
}

func addCpuSocketMetrics(acc telegraf.Accumulator, cpuSocketCounters map[int]metricCounters) {
	for socketId, data := range cpuSocketCounters {
		tags := map[string]string{
			"socket_id": strconv.Itoa(socketId),
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
	err := rows.Scan(&mcError.Id, &mcError.Timestamp, &mcError.ErrorMsg, &mcError.MciStatusMsg, &mcError.SocketId)

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
			DbPath:          defaultDbPath,
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
