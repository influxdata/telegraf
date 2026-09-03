//go:generate ../../../tools/readme_config_includer/generator
//go:build linux && (386 || amd64 || arm || arm64)

package ras

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	// Required for SQL framework driver
	_ "modernc.org/sqlite"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const mceQuery = `
	SELECT
		id, timestamp, error_msg, mcistatus_msg, mcastatus_msg, socketid
	FROM mce_record
	WHERE timestamp > ?
`

type Ras struct {
	DBPath string          `toml:"db_path"`
	Log    telegraf.Logger `toml:"-"`

	db                *sql.DB
	latestTimestamp   time.Time
	cpuSocketCounters map[int]map[string]int64
	serverCounters    map[string]int64
}

type machineCheckError struct {
	id           int
	timestamp    string
	socketID     int
	errorMsg     string
	mciStatusMsg string
	mcaStatusMsg string
}

func (*Ras) SampleConfig() string {
	return sampleConfig
}

func (r *Ras) Init() error {
	// Setup defaults
	if r.DBPath == "" {
		r.DBPath = "/var/lib/rasdaemon/ras-mc_event.db"
	}
	r.cpuSocketCounters = map[int]map[string]int64{
		0: {
			"memory_ecc_corrected_errors":              0,
			"memory_ecc_uncorrectable_errors":          0,
			"memory_read_corrected_errors":             0,
			"memory_read_uncorrectable_errors":         0,
			"memory_write_corrected_errors":            0,
			"memory_write_uncorrectable_errors":        0,
			"cache_l0_l1_errors":                       0,
			"tlb_instruction_errors":                   0,
			"processor_base_errors":                    0,
			"processor_bus_errors":                     0,
			"internal_timer_errors":                    0,
			"smm_handler_code_access_violation_errors": 0,
			"internal_parity_errors":                   0,
			"frc_errors":                               0,
			"external_mce_errors":                      0,
			"microcode_rom_parity_errors":              0,
			"unclassified_mce_errors":                  0,
		},
	}
	r.serverCounters = map[string]int64{
		"cache_l2_errors": 0,
		"upi_errors":      0,
	}

	// Check the database readability
	pathInfo, err := os.Stat(r.DBPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("provided db_path %q does not exist", r.DBPath)
	}

	if err != nil {
		return fmt.Errorf("cannot get system information for db_path file %q: %w", r.DBPath, err)
	}

	if mode := pathInfo.Mode(); !mode.IsRegular() {
		return fmt.Errorf("provided db_path does not point to a regular file: %q", r.DBPath)
	}

	return nil
}

func (r *Ras) Start(telegraf.Accumulator) error {
	// Open the DB for reading the RAS events
	db, err := sql.Open("sqlite", r.DBPath)
	if err != nil {
		return fmt.Errorf("opening database at %q failed: %w", r.DBPath, err)
	}
	r.db = db

	return nil
}

func (r *Ras) Stop() {
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			r.Log.Errorf("Error appeared during closing DB (%s): %v", r.DBPath, err)
		}
	}
}

func (r *Ras) Gather(acc telegraf.Accumulator) error {
	// Execute the query
	rows, err := r.db.Query(mceQuery, r.latestTimestamp)
	if err != nil {
		return fmt.Errorf("querying failed: %w", err)
	}
	defer rows.Close()

	// Parse the data and corresponding events
	for rows.Next() {
		data := &machineCheckError{}
		if err := rows.Scan(
			&data.id,
			&data.timestamp,
			&data.errorMsg,
			&data.mciStatusMsg,
			&data.mcaStatusMsg,
			&data.socketID,
		); err != nil {
			return fmt.Errorf("scanning row failed: %w", err)
		}

		if err := r.updateLatestTimestamp(data.timestamp); err != nil {
			return fmt.Errorf("updating timestamp failed: %w", err)
		}
		r.updateCounters(data)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scanning rows failed: %w", err)
	}

	// Add CPU-socket metrics
	for socketID, data := range r.cpuSocketCounters {
		tags := map[string]string{
			"socket_id": strconv.Itoa(socketID),
		}
		fields := make(map[string]interface{}, len(data))
		for name, count := range data {
			fields[name] = count
		}
		acc.AddCounter("ras", fields, tags)
	}

	// Add the server metrics
	tags := make(map[string]string)
	fields := make(map[string]interface{}, len(r.serverCounters))
	for name, count := range r.serverCounters {
		fields[name] = count
	}
	acc.AddCounter("ras", fields, tags)

	return nil
}

func (r *Ras) updateLatestTimestamp(timestamp string) error {
	ts, err := time.Parse("2006-01-02 15:04:05 -0700", timestamp)
	if err != nil {
		return err
	}
	if ts.After(r.latestTimestamp) {
		r.latestTimestamp = ts
	}

	return nil
}

func (r *Ras) updateCounters(mcError *machineCheckError) {
	if strings.Contains(mcError.errorMsg, "No Error") {
		return
	}

	if _, ok := r.cpuSocketCounters[mcError.socketID]; !ok {
		r.cpuSocketCounters[mcError.socketID] = map[string]int64{
			"memory_ecc_corrected_errors":              0,
			"memory_ecc_uncorrectable_errors":          0,
			"memory_read_corrected_errors":             0,
			"memory_read_uncorrectable_errors":         0,
			"memory_write_corrected_errors":            0,
			"memory_write_uncorrectable_errors":        0,
			"cache_l0_l1_errors":                       0,
			"tlb_instruction_errors":                   0,
			"processor_base_errors":                    0,
			"processor_bus_errors":                     0,
			"internal_timer_errors":                    0,
			"smm_handler_code_access_violation_errors": 0,
			"internal_parity_errors":                   0,
			"frc_errors":                               0,
			"external_mce_errors":                      0,
			"microcode_rom_parity_errors":              0,
			"unclassified_mce_errors":                  0,
		}
	}
	r.updateSocketCounters(mcError)
	r.updateServerCounters(mcError)
}

func (r *Ras) updateSocketCounters(mcError *machineCheckError) {
	r.updateMemoryCounters(mcError)
	r.updateProcessorBaseCounters(mcError)

	if strings.Contains(mcError.errorMsg, "Error") {
		switch {
		case strings.Contains(mcError.errorMsg, "Instruction TLB"):
			r.cpuSocketCounters[mcError.socketID]["tlb_instruction_errors"]++
		case strings.Contains(mcError.errorMsg, "BUS"):
			r.cpuSocketCounters[mcError.socketID]["processor_bus_errors"]++
		case strings.Contains(mcError.errorMsg, "CACHE Level-0"), strings.Contains(mcError.errorMsg, "CACHE Level-1"):
			r.cpuSocketCounters[mcError.socketID]["cache_l0_l1_errors"]++
		}
	}
}

func (r *Ras) updateServerCounters(mcError *machineCheckError) {
	switch {
	case strings.Contains(mcError.errorMsg, "CACHE Level-2") && strings.Contains(mcError.errorMsg, "Error"):
		r.serverCounters["cache_l2_errors"]++
	case strings.Contains(mcError.errorMsg, "UPI:"):
		r.serverCounters["upi_errors"]++
	}
}

func (r *Ras) updateProcessorBaseCounters(mcError *machineCheckError) {
	switch {
	case strings.Contains(mcError.errorMsg, "Internal Timer error"):
		r.cpuSocketCounters[mcError.socketID]["internal_timer_errors"]++
		r.cpuSocketCounters[mcError.socketID]["processor_base_errors"]++
	case strings.Contains(mcError.errorMsg, "SMM Handler Code Access Violation"):
		r.cpuSocketCounters[mcError.socketID]["smm_handler_code_access_violation_errors"]++
		r.cpuSocketCounters[mcError.socketID]["processor_base_errors"]++
	case strings.Contains(mcError.errorMsg, "Internal parity error"):
		r.cpuSocketCounters[mcError.socketID]["internal_parity_errors"]++
		r.cpuSocketCounters[mcError.socketID]["processor_base_errors"]++
	case strings.Contains(mcError.errorMsg, "FRC error"):
		r.cpuSocketCounters[mcError.socketID]["frc_errors"]++
		r.cpuSocketCounters[mcError.socketID]["processor_base_errors"]++
	case strings.Contains(mcError.errorMsg, "External error"):
		r.cpuSocketCounters[mcError.socketID]["external_mce_errors"]++
		r.cpuSocketCounters[mcError.socketID]["processor_base_errors"]++
	case strings.Contains(mcError.errorMsg, "Microcode ROM parity error"):
		r.cpuSocketCounters[mcError.socketID]["microcode_rom_parity_errors"]++
		r.cpuSocketCounters[mcError.socketID]["processor_base_errors"]++
	case strings.Contains(mcError.errorMsg, "Unclassified"), strings.Contains(mcError.errorMsg, "Internal unclassified"):
		r.cpuSocketCounters[mcError.socketID]["unclassified_mce_errors"]++
		r.cpuSocketCounters[mcError.socketID]["processor_base_errors"]++
	}
}

func (r *Ras) updateMemoryCounters(mcError *machineCheckError) {
	switch {
	case strings.Contains(mcError.errorMsg, "Memory read error"):
		if strings.Contains(mcError.mciStatusMsg, "Corrected_error") {
			r.cpuSocketCounters[mcError.socketID]["memory_read_corrected_errors"]++
		} else {
			r.cpuSocketCounters[mcError.socketID]["memory_read_uncorrectable_errors"]++
		}
	case strings.Contains(mcError.errorMsg, "Memory write error"):
		if strings.Contains(mcError.mciStatusMsg, "Corrected_error") {
			r.cpuSocketCounters[mcError.socketID]["memory_write_corrected_errors"]++
		} else {
			r.cpuSocketCounters[mcError.socketID]["memory_write_uncorrectable_errors"]++
		}
	case strings.Contains(mcError.mcaStatusMsg, "DRAM ECC error"):
		if strings.Contains(mcError.errorMsg, "Corrected error") {
			r.cpuSocketCounters[mcError.socketID]["memory_ecc_corrected_errors"]++
		} else {
			r.cpuSocketCounters[mcError.socketID]["memory_ecc_uncorrectable_errors"]++
		}
	}
}

func init() {
	inputs.Add("ras", func() telegraf.Input {
		return &Ras{}
	})
}
