package goldilocks_cluster

import (
	"database/sql"
	"fmt"
	_ "github.com/alexbrainman/odbc"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"os"
	"strings"
	"sync"
)

type MonitorElement struct {
	Sql        string   `toml:"sql"`
	Tags       []string `toml:"tags"`
	Fields     []string `toml:"fields"`
	Pivot      bool     `toml:"pivot"`
	PivotKey   string   `toml:"pivot_key"`
	SeriesName string   `toml:"series_name"`
}
type Goldilocks struct {
	OdbcDriverPath string           `toml:"goldilocks_odbc_driver_path"`
	Host           string           `toml:"goldilocks_host"`
	Port           int              `toml:"goldilocks_port"`
	User           string           `toml:"goldilocks_user"`
	Password       string           `toml:"goldilocks_password"`
	Elements       []MonitorElement `toml:"elements"`
}

func init() {
	inputs.Add("goldilocks", func() telegraf.Input {
		return &Goldilocks{}
	})
}

var sampleConfig = `

## specify connection string
goldilocks_odbc_driver_path = "?/lib/libgoldilockscs-ul64.so" 
goldilocks_host = "127.0.0.1" 
goldilocks_port = 37562
goldilocks_user = "test"
goldilocks_password = "test"


###### DO NOT EDIT : Start  ########################################
[[ inputs.goldilocks.elements ]]
series_name = "goldilocks_default_tags"
sql = """

SELECT
        Y.GROUP_NAME,
        X.LOCAL_MEMBER_NAME MEMBER_NAME
FROM X$INSTANCE@LOCAL X INNER JOIN CLUSTER_GROUP@LOCAL Y ON X.LOCAL_GROUP_ID = Y.GROUP_ID;

"""
###### DO NOT EDIT : End   ########################################


[[ inputs.goldilocks.elements]]
series_name="session_stat"
sql = """
SELECT NVL( CLIENT_ADDRESS, 'DA') CLIENT_ADDRESS,
       COUNT(*) CNT
FROM V$SESSION
WHERE USER_NAME IS NOT NULL
AND   PROGRAM_NAME != 'gmaster'
GROUP BY CLIENT_ADDRESS
"""
tags = ["CLIENT_ADDRESS"]
fields = ["CNT"]
pivot = false

[[ inputs.goldilocks.elements ]]
series_name = "goldilocks_statement_stat"
sql = """
SELECT COUNT(*) TOTAL_COUNT,
       SUM(CASE WHEN X.ELAPSED > 5 THEN 1 ELSE 0 END ) LONG_RUNNING_COUNT
FROM
(
 SELECT
       DATEDIFF ( SECOND, START_TIME, SYSTIMESTAMP  ) ELAPSED
 FROM V$STATEMENT
) X
"""
tags = []
fields = ["TOTAL_COUNT", "LONG_RUNNING_COUNT"]
pivot = false

[[ inputs.goldilocks.elements ]]

series_name = "goldilocks_sql_execution_stat"
sql = """

SELECT STAT_NAME , CAST ( STAT_VALUE AS NATIVE_BIGINT )  VALUE
FROM   V$SYSTEM_SQL_STAT;

"""
tags = []
fields = ["VALUE"]
pivot_key = "STAT_NAME"
pivot = true

[[ inputs.goldilocks.elements ]]
series_name = "goldilocks_transaction_stat"
sql = """
SELECT
 (SELECT COUNT(*) FROM V$TRANSACTION) ACTIVE_TRANSACTIONS,
 (SELECT COUNT(*) FROM V$LOCK_WAIT) WAIT_TRANSACTIONS
FROM DUAL ;
"""

tags = []
fields = ["ACTIVE_TRANSACTIONS", "WAIT_TRANSACTIONS"]
pivot_key = ""
pivot = false

[[ inputs.goldilocks.elements ]]
series_name = "goldilocks_cluster_net_stat"
sql = """
SELECT DECODE(IS_SYNC,FALSE, 'SYNC', 'ASYNC' ) TYPE ,
        SUM(RX_BYTES) RX_BYTES,
        SUM(TX_BYTES) TX_BYTES,
        SUM(RX_JOBS) RX_JOBS,
        SUM(TX_JOBS) TX_JOBS
 FROM X$CLUSTER_DISPATCHER@LOCAL
 GROUP BY IS_SYNC
"""
tags = ["TYPE"]
fields = ["RX_BYTES", "TX_BYTES", "RX_JOBS", "TX_JOBS"]
pivot_key = ""
pivot = false


[[ inputs.goldilocks.elements ]]
series_name = "goldilocks_tablespaces"
sql = """
SELECT X.NAME, 
       X.TOTAL TOTAL_BYTES, 
       X.PAGE_SIZE * NVL ( Y.PAGE_COUNT, 0)  USED_BYTES, 
       ROUND ( X.PAGE_SIZE * NVL ( Y.PAGE_COUNT, 0) * 100  / X.TOTAL , 2 ) USED_PCT    
  FROM (
       SELECT TABLESPACE_ID ID
            , NAME
            , SUM(SIZE) TOTAL
            , PAGE_SIZE
         FROM X$DATAFILE@LOCAL XX INNER JOIN X$TABLESPACE@LOCAL YY ON XX.TABLESPACE_ID = YY.ID
        WHERE XX.STATE != 'DROPPED'
        GROUP BY TABLESPACE_ID, NAME, PAGE_SIZE
     ) X LEFT OUTER JOIN (
       SELECT 1 TBS_ID
            , SUM(CASE WHEN REAL_COUNT < PROP_VALUE THEN PROP_VALUE ELSE REAL_COUNT END) PAGE_COUNT
         FROM (
              SELECT ALLOC_PAGE_COUNT - AGABLE_PAGE_COUNT REAL_COUNT
                   , (
                     SELECT TO_NUMBER(PROPERTY_VALUE)
                       FROM V$PROPERTY
                      WHERE PROPERTY_NAME = 'MINIMUM_UNDO_PAGE_COUNT'
                   ) PROP_VALUE
                FROM X$UNDO_SEGMENT@LOCAL
            )
        UNION ALL
       SELECT TBS_ID
            , SUM(ALLOC_PAGE_COUNT) ALLOC
         FROM X$SEGMENT@LOCAL
        WHERE TBS_ID != 1
        GROUP BY TBS_ID
     ) Y ON X.ID = Y.TBS_ID 
"""
tags = ["NAME"]
fields = ["TOTAL_BYTES", "USED_BYTES", "USED_PCT"]
pivot_key = ""
pivot = false


`

func (m *Goldilocks) BuildConnectionString() string {

	sGoldilocksHome := os.Getenv("GOLDILOCKS_HOME")
	sDriverPath := strings.Replace(m.OdbcDriverPath, "?", sGoldilocksHome, 1)

	sConnectionString := fmt.Sprintf("DRIVER=%s;HOST=%s;PORT=%d;UID=%s;PWD=%s", sDriverPath, m.Host, m.Port, m.User, m.Password)
	return sConnectionString
}

func (m *Goldilocks) SampleConfig() string {
	return sampleConfig
}

func (m *Goldilocks) Description() string {
	return "Read metrics from one goldilocks server ( per instance ) "
}

func (m *Goldilocks) GatherServer(acc telegraf.Accumulator) error {
	return nil
}

func (m *Goldilocks) Gather(acc telegraf.Accumulator) error {

	var wg sync.WaitGroup
	connectionString := m.BuildConnectionString()

	if m.OdbcDriverPath == "" {
		return nil
	}

	// Loop through each server and collect metrics
	wg.Add(1)
	go func(s string) {
		defer wg.Done()
		acc.AddError(m.gatherServer(s, acc))
	}(connectionString)

	wg.Wait()

	return nil
}

func (m *Goldilocks) getCommonTags(db *sql.DB) map[string]string {

	v := make(map[string]string)
	for _, element := range m.Elements {

		if element.SeriesName == "goldilocks_default_tags" {
			q, err := m.getSQLResult(db, element.Sql)
			if err != nil {
				return nil
			}

			for k, _ := range q[0] {
				v[k] = q[0][k].(string)
			}
			break
		}
	}

	return v
}

func (m *Goldilocks) runSQL(acc telegraf.Accumulator, db *sql.DB) error {

	for _, element := range m.Elements {
		tags := m.getCommonTags(db)
		fields := make(map[string]interface{})

		if element.SeriesName == "goldilocks_default_tags" {
			continue
		}

		r, err := m.getSQLResult(db, element.Sql)
		if err != nil {
			return err
		}

		if element.Pivot {

			for _, v := range r {
				for _, v2 := range element.Tags {
					tags[v2] = v[v2].(string)
				}

				key := v[element.PivotKey].(string)
				data := v[element.Fields[0]]
				fields[key] = data
			}
			acc.AddFields(element.SeriesName, fields, tags)

		} else {

			for _, v := range r {
				for _, v2 := range element.Tags {
					tags[v2] = v[v2].(string)
				}

				for _, v2 := range element.Fields {
					fields[v2] = v[v2]
				}
				acc.AddFields(element.SeriesName, fields, tags)

			}
		}
	}

	return nil
}

func (m *Goldilocks) getSQLResult(db *sql.DB, sqlText string) ([]map[string]interface{}, error) {
	rows, err := db.Query(sqlText)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	column_count := len(columns)

	result_data := make([]map[string]interface{}, 0)
	value_data := make([]interface{}, column_count)
	value_ptrs := make([]interface{}, column_count)

	for rows.Next() {

		for i := 0; i < column_count; i++ {
			value_ptrs[i] = &value_data[i]
		}

		rows.Scan(value_ptrs...)
		entry := make(map[string]interface{})

		for i, col := range columns {
			var v interface{}
			val := value_data[i]

			b, ok := val.([]byte)

			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		result_data = append(result_data, entry)
	}
	return result_data, nil

}

func (m *Goldilocks) gatherServer(serv string, acc telegraf.Accumulator) error {

	db, err := sql.Open("odbc", serv)
	if err != nil {
		return err
	}

	err = m.runSQL(acc, db)
	if err != nil {
		return err
	}

	defer db.Close()

	return nil
}
