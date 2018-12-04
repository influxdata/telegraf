package sqlquery

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/srclosson/telegraf/metric"
	// This is my justification
	_ "github.com/denisenkom/go-mssqldb"
)

type Query struct {
	Name  string
	Query string

	stmt *sql.Stmt
}

type Sqlquery struct {
	ConnectionUrl string
	Query         []Query

	connUrl    *url.URL
	conn       *sql.DB
	configured bool
	acc        telegraf.Accumulator
	wg         sync.WaitGroup
}

var sampleConfig = `
  ##
`

func (m *Sqlquery) SampleConfig() string {
	return sampleConfig
}

func (m *Sqlquery) Description() string {
	return "Read metrics from various sql databases using sql queries"
}

func (m *Sqlquery) Configure() error {
	var err error
	m.connUrl, err = url.Parse(m.ConnectionUrl)
	if err != nil {
		return err
	}

	m.conn, err = sql.Open("mssql", m.ConnectionUrl)
	if err != nil {
		return err
	}

	// for _, query := range m.Query {
	// 	query.stmt, err = m.conn.Prepare(query.Query)

	// 	if err != nil {
	// 		return err
	// 	}
	// }

	m.configured = true
	return nil
}

func (q *Query) Get(conn *sql.DB, pwg *sync.WaitGroup) ([]telegraf.Metric, error) {
	defer pwg.Done()

	var ret []telegraf.Metric
	rows, err := conn.Query(q.Query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	timestamp := time.Now()

	for rows.Next() {

		for i, _ := range columns {
			valuePtrs[i] = &values[i]
		}

		rows.Scan(valuePtrs...)

		for i, col := range columns {
			var v interface{}

			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}

			switch col {
			case "timestamp":
				//2018-07-01 14:16:26.3991292 +0000 UTC
				//t, err := time.Parse("2006-01-02 15-04-05.0000000 -0700 MST", v.(string))
				// if err != nil {
				// 	return nil, err
				// }
				timestamp = v.(time.Time)
			case "metric":
				q.Name = v.(string)
			default:
				if strings.HasPrefix(col, "tag_") {
					// We are using this value as a tag value.
					tagname := strings.TrimPrefix(col, "tag_")
					stag := fmt.Sprintf("%v", v)
					tags[tagname] = stag
				} else {
					fields[col] = v
				}
			}
		}

		m, err := metric.New(q.Name, tags, fields, timestamp)
		if err != nil {
			return nil, err
		}

		ret = append(ret, m)
	}

	return ret, nil
}

func (m *Sqlquery) Gather(acc telegraf.Accumulator) error {
	if !m.configured {
		err := m.Configure()
		if err != nil {
			log.Fatalf("Could not configure the mssql input plugin: %s", err)
		}
	}

	for _, query := range m.Query {
		m.wg.Add(1)
		metrics, err := query.Get(m.conn, &m.wg)
		if err != nil {
			fmt.Println("ERROR [mssql.get]", err)
		}

		for _, metric := range metrics {
			acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		}
	}
	m.wg.Wait()

	return nil
}

func init() {
	inputs.Add("sqlquery", func() telegraf.Input {
		return &Sqlquery{}
	})
}
