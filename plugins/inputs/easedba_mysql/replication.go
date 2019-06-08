package easedba_mysql

import (
	"database/sql"
	"fmt"
	"github.com/influxdata/telegraf/plugins/easedbautil"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/easedba_mysql/v1"
)

var (
	showSalveStatus  = `SHOW SLAVE STATUS`
	showMasterStatus = `SHOW MASTER STATUS`
)

func (m *Mysql) gatherReplication(db *sql.DB, serv string, accumulator telegraf.Accumulator, servtag string) error {
	tags := map[string]string{"server": servtag}

	fields := make(map[string]interface{})
	// run query
	rows, err := db.Query(showSalveStatus)
	if err != nil {
		return err
	}
	defer rows.Close()

	err = m.getReplicationFields(fields, rows)
	if err != nil {
		return err
	}

	// run query
	rows, err = db.Query(showMasterStatus)
	if err != nil {
		return err
	}
	defer rows.Close()

	err = m.getReplicationFields(fields, rows)
	if err != nil {
		return err
	}

	accumulator.AddGauge(easedbautl.SchemaReplication, fields, tags)
	return nil
}

func (m *Mysql) getReplicationFields(fields map[string]interface{}, rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = &sql.RawBytes{}
	}

	if rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return err
		}
	}

	for i, val := range values {
		if convertedName, ok := easedba_v1.ReplicationMappings[columns[i]]; ok {
			switch columns[i] {
			case "Slave_IO_Running", "Slave_SQL_Running":
				if string(*(val.(*sql.RawBytes))) == "Yes" {
					fields[convertedName] = 1
				} else {
					fields[convertedName] = 0
				}

			case "Seconds_Behind_Master", "Read_Master_Log_Pos", "Exec_Master_Log_Pos",
				"SQL_Delay", "Last_SQL_Errno", "Last_IO_Errno", "Master_position":
				if num, err := strconv.ParseInt(string(*(val.(*sql.RawBytes))), 10, 64); err == nil {
					fields[convertedName] = num
				} else {
					fields[convertedName] = 0
				}
			case "Last_SQL_Error", "Last_IO_Error":
				fields[convertedName] = string(*(val.(*sql.RawBytes)))
			default:
				return fmt.Errorf("unknown key wthen scanning replication attributes: %s", columns[i])
			}
		}
	}

	return nil
}
