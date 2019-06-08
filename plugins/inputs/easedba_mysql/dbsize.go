package easedba_mysql

import (
	"database/sql"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/easedba_mysql/v1"
	"log"
)

var (
	logbinWarningFrequency = 0

	tableAndIndexSizeQuery = `
	SELECT TRUNCATE(SUM(data_length) , 0)  AS Table_data_size, 
		   TRUNCATE(SUM(index_length), 0) AS Table_index_size 
	FROM   information_schema.TABLES 
	`
)

func (m *Mysql) gatherDbSizes(db *sql.DB, serv string, accumulator telegraf.Accumulator, servtag string) error {
	tags := map[string]string{"server": servtag}

	// binary log size
	binLogSize, err := getBinaryLogs(db, servtag)
	if err != nil {
		return fmt.Errorf("error gathering binary log size: %s", err)
	}

	fields := map[string]interface{}{
		"binary_log_size": binLogSize,
	}

	// table data and index size
	rows, err := db.Query(tableAndIndexSizeQuery)
	if err != nil {
		return fmt.Errorf("error querying table and index size: %s", err)
	}
	defer rows.Close()

	tableDataSize, tableIndexSize := uint64(0), uint64(0)

	for rows.Next() {
		err := rows.Scan(&tableDataSize, &tableIndexSize)
		if err != nil {
			return fmt.Errorf("error scaning table and index size %s", err)
		}
	}

	fields["table_data_size"] = tableDataSize
	fields["table_index_size"] = tableIndexSize

	// disk cache and tmp table size
	log.Printf("collect disk cache and tmp table size ...")
	key, val := "", sql.RawBytes{}
	rows, err = db.Query(globalStatusQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&key, &val)
		if err != nil {
			return fmt.Errorf("error scaning for disk cache and tmp table size %s", err)
		}
		if convertedName, ok := easedba_v1.DbsizeMappings[key]; ok {
			fields[convertedName] = string(val)
		}
	}


	log.Printf("add for mysql-dbsize ...")
	accumulator.AddGauge("mysql-dbsize", fields, tags)

	return nil
}

// get the total binary log size in bytes
func getBinaryLogs(db *sql.DB, servtag string) (size int64, err error) {
	rows, err := db.Query("SHOW VARIABLES LIKE 'log_bin'")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	key, val := "", sql.RawBytes{}
	if rows.Next() {
		if rows.Scan(&key, &val); string(val) != "ON" {
			if logbinWarningFrequency%10 == 0 {
				log.Printf("INFO: [%s] binary log not open, skip metrics collection. atrr: %s, value: %s", servtag, key, string(val))
			}
			logbinWarningFrequency++

			return 0, nil
		}
	}

	rows, err = db.Query(binaryLogsQuery)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	fileSize, fileName := int64(0), ""
	// iterate over rows and count the size and count of files
	for rows.Next() {
		if err := rows.Scan(&fileName, &fileSize); err != nil {
			return 0, err
		}
		size += fileSize
	}

	return size, nil
}
