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
	Sql        string
	Tags       []string
	Fields     []string
	Pivot      bool
	PivotKey   string
	SeriesName string
}

type Goldilocks struct {
	OdbcDriverPath string `toml:"goldilocks_odbc_driver_path"`
	Host           string `toml:"goldilocks_host"`
	Port           int    `toml:"goldilocks_port"`
	User           string `toml:"goldilocks_user"`
	Password       string `toml:"goldilocks_password"`
	AddPostfix     bool   `toml:"goldilocks_add_group_postfix_to_series_name"`
	Elements       []MonitorElement
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

	if connectionString == "" {
		return fmt.Errorf("ConnectionString is empty")
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

func (m *Goldilocks) runSQL(acc telegraf.Accumulator, db *sql.DB, clusterMode int) error {

	sSeriesName := ""

	for _, element := range m.Elements {

		tags := make(map[string]string)

		sSeriesName = element.SeriesName

		fields := make(map[string]interface{})

		r, err := m.getSQLResult(db, element.Sql)
		if err != nil {
			return err
		}

		if element.Pivot {

			for _, v := range r {
				for _, v2 := range element.Tags {
					if value, ok := v[v2].(string); ok {
						tags[v2] = value
					} else {
						fmt.Printf("[%s:%d] tag key [%s] not in metrics series(%s)\n", m.Host, m.Port, v2, element.SeriesName)
						continue
					}

					tags[v2] = v[v2].(string)
				}

				key := v[element.PivotKey].(string)
				data := v[element.Fields[0]]
				fields[key] = data
			}

			acc.AddFields(sSeriesName, fields, tags)

		} else {

			for _, v := range r {
				for _, v2 := range element.Tags {
					if value, ok := v[v2].(string); ok {
						tags[v2] = value
					} else {
						fmt.Printf("[%s:%d] tag key [%s] not in metrics series(%s)\n", m.Host, m.Port, v2, element.SeriesName)
						continue
					}
				}

				for _, v2 := range element.Fields {
					if value, ok := v[v2].(interface{}); ok {
						fields[v2] = value
					} else {
						fmt.Printf("[%s:%d] field key [%s] not in metrics or value is null, series(%s)\n", m.Host, m.Port, v2, element.SeriesName)
						continue
					}

				}
				acc.AddFields(sSeriesName, fields, tags)

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

func (m *Goldilocks) getConfig(db *sql.DB) error {

	var (
		sSeriesName string
		sQuery      string
		sTags       sql.NullString
		sFields     sql.NullString
		sPivotKey   sql.NullString
		sPivot      int
	)

	m.Elements = m.Elements[:0]

	metricSQL := "SELECT *  FROM TELEGRAF_METRIC_SETTINGS"

	rows, err := db.Query(metricSQL)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&sSeriesName, &sQuery, &sTags, &sFields, &sPivotKey, &sPivot)
		if err != nil {
			return err
		}
		element := MonitorElement{}

		element.SeriesName = sSeriesName
		element.Sql = sQuery

		if sFields.Valid {
			element.Fields = strings.Split(sFields.String, "|")
		} else {
			element.Fields = nil
		}

		if sTags.Valid {
			element.Tags = strings.Split(sTags.String, "|")
		} else {
			element.Tags = nil
		}

		if sPivotKey.Valid {
			element.PivotKey = sPivotKey.String
		} else {
			element.PivotKey = ""
		}

		if sPivot == 0 {
			element.Pivot = false
		} else {
			element.Pivot = true
		}
		m.Elements = append(m.Elements, element)
	}

	return nil
}

func (m *Goldilocks) getClusterMode(db *sql.DB) int {
	var sClusterMode int

	sql := `SELECT SETTING_VALUE FROM TELEGRAF_GLOBAL_SETTINGS WHERE SETTING_KEY='CLUSTER_MODE'`
	rows, err := db.Query(sql)

	if err != nil {
		return 0
	}

	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(&sClusterMode)

		if err != nil {
			return 0
		}

	}
	return sClusterMode
}

func (m *Goldilocks) gatherServer(serv string, acc telegraf.Accumulator) error {

	db, err := sql.Open("odbc", serv)
	if err != nil {
		return fmt.Errorf("[%s:%d] %s", m.Host, m.Port, err.Error())
	}
	defer db.Close()

	sClusterMode := m.getClusterMode(db)

	err = m.getConfig(db)
	if err != nil {
		return fmt.Errorf("[%s:%d] %s", m.Host, m.Port, err.Error())
	}

	err = m.runSQL(acc, db, sClusterMode)
	if err != nil {
		return fmt.Errorf("[%s:%d] %s", m.Host, m.Port, err.Error())
	}

	return nil
}
