package easedba_mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/easedbautil"
	"github.com/influxdata/telegraf/plugins/inputs/easedba_mysql/v1"
)

var (
	queryRunningSQL = `
SELECT
    id process_id,
    user,
    host,
    db,
    time,
    info sql_text,
    state 
FROM
    information_schema.processlist 
WHERE
    info IS NOT NULL;

`
	queryRunningTransactions = `

SELECT
    a.trx_mysql_thread_id process_id,
    d.thread_id,
    a.trx_id,
    a.trx_state,
    a.trx_started,
    a.trx_wait_started,
    a.trx_query,
    a.trx_isolation_level,
    b.blocking_trx_id,
    e.thread_id blocking_thread_id,
    c.trx_mysql_thread_id blocking_process_id,
    d.processlist_user user,
    d.processlist_host client,
    d.processlist_db db 
FROM
    information_schema.innodb_trx a 
    LEFT JOIN
        information_schema.innodb_lock_waits b 
        ON a.trx_id = b.requesting_trx_id 
    LEFT JOIN
        information_schema.innodb_trx c 
        ON b.blocking_trx_id = c.trx_id 
    LEFT JOIN
        performance_schema.threads d 
        ON a.trx_mysql_thread_id = d.processlist_id 
    LEFT JOIN
        performance_schema.threads e 
        ON c.trx_mysql_thread_id = e.processlist_id

`
	queryBlockingTransactions = `
SELECT
    b.processlist_id process_id,
    a.thread_id,
    a.sql_text,
    b.processlist_user USER,
    b.processlist_host client,
    b.processlist_db db 
FROM
    performance_schema.events_statements_history a 
    LEFT JOIN
        performance_schema.threads b 
        ON a.thread_id = b.thread_id 
WHERE
    a.thread_id IN ( %s )
ORDER BY
    a.event_id DESC LIMIT 20;
`
)

func (m *Mysql) gatherSnapshot(db *sql.DB, serv string, accumulator telegraf.Accumulator, servtag string) error {
	tags := map[string]string{"server": servtag}
	fields := map[string]interface{}{}

	// fetch sql snapshot
	rows, err := db.Query(queryRunningSQL)
	if err != nil {
		return fmt.Errorf("error querying running sql: %s", err)
	}
	defer rows.Close()

	runningSqls := easedba_v1.RunningSqls{}
	for rows.Next() {
		val := easedba_v1.RunningSql{}
		err = rows.Scan(&val.ProcessId, &val.User, &val.Host, &val.Db,
			&val.Time, &val.SqlText, &val.State)
		if err != nil {
			return fmt.Errorf("error scaning running sql %s", err)
		}
		runningSqls.RunningSqlList = append(runningSqls.RunningSqlList, val)
	}

	text := []byte{'{', '}'}
	if len(runningSqls.RunningSqlList) > 0 {
		text, err = json.Marshal(runningSqls)
		if err != nil {
			return fmt.Errorf("error marshaling running sql: %s", err)
		}
	}
	fields["sql_snapshot"] = text

	// fetch transaction snapshot
	rows, err = db.Query(queryRunningTransactions)
	if err != nil {
		return fmt.Errorf("error querying running sql: %s", err)
	}
	defer rows.Close()

	// use a map to filter out the duplicated thread ids
	blockingThreadIds := make(map[int64]bool, 2)
	runningTransactions := easedba_v1.RunningTransactions{}
	for rows.Next() {
		val := easedba_v1.RunningTransaction{}
		err = rows.Scan(&val.ProcessId, &val.ThreadId, &val.TrxId, &val.TrxState, &val.TrxStarted,
			&val.TrxWaitStarted, &val.TrxQuery, &val.TrxIsolationLevel, &val.Blocking_trx_id, &val.Blocking_thread_id, &val.Blocking_process_id, &val.User, &val.Client, &val.Db)
		if err != nil {
			return fmt.Errorf("error scaning running transaction %s", err)
		}
		if val.Blocking_thread_id.Valid {
			blockingThreadIds[val.Blocking_thread_id.Int64] = true
		}

		runningTransactions.RunningTransactionList =
			append(runningTransactions.RunningTransactionList, val)
	}

	text = []byte{'{', '}'}
	if len(runningTransactions.RunningTransactionList) > 0 {
		text, err = json.Marshal(runningTransactions)
		if err != nil {
			return fmt.Errorf("error marshaling running transactions: %s", err)
		}
	}
	fields["trx_snapshot"] = text

	// if a transaction is blocking others, try to fetch the history sql of this transaction
	if len(blockingThreadIds) > 0 {
		ids := ""
		addComma := false
		for k := range blockingThreadIds {
			if !addComma {
				ids += fmt.Sprintf("%d", k)
				addComma = true
			} else {
				ids += fmt.Sprintf(",%d", k)
			}
		}

		rows, err = db.Query(fmt.Sprintf(queryBlockingTransactions, ids))
		if err != nil {
			return fmt.Errorf("error querying running sql: %s", err)
		}
		defer rows.Close()

		transactionHistories := easedba_v1.TransactionHistories{}
		for rows.Next() {
			val := easedba_v1.TransactionHistory{}
			err = rows.Scan(&val.ProcessId, &val.ThreadId, &val.SqlText, &val.User,
				&val.Client, &val.Db)
			if err != nil {
				return fmt.Errorf("error scaning transaction histroy%s", err)
			}

			transactionHistories.TransactionHistoryList = append(
				transactionHistories.TransactionHistoryList, val)
		}

		text = []byte{'{', '}'}
		if len(transactionHistories.TransactionHistoryList) > 0 {
			text, err = json.Marshal(transactionHistories)
			if err != nil {
				return fmt.Errorf("error marshaling transaction history: %s", err)
			}
		}
		fields["trx_history"] = text
	}

	accumulator.AddFields(easedbautl.SchemaSnapshot, fields, tags)

	return nil
}
