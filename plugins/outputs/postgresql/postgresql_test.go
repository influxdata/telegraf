package postgresql

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/columns"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestWriteAllInOnePlace(t *testing.T) {
	timestamp := time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)
	oneMetric, _ := metric.New("m", map[string]string{"t": "tv"}, map[string]interface{}{"f": 1}, timestamp)
	twoMetric, _ := metric.New("m", map[string]string{"t2": "tv2"}, map[string]interface{}{"f2": 2}, timestamp)
	threeMetric, _ := metric.New("m", map[string]string{"t": "tv", "t2": "tv2"}, map[string]interface{}{"f": 3, "f2": 4}, timestamp)
	fourMetric, _ := metric.New("m2", map[string]string{"t": "tv", "t2": "tv2"}, map[string]interface{}{"f": 5, "f2": 6}, timestamp)

	p := &Postgresql{
		Schema:          "public",
		TableTemplate:   "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS})",
		TagTableSuffix:  "_tag",
		DoSchemaUpdates: true,
		Address:         "host=localhost user=postgres password=postgres sslmode=disable dbname=postgres",
	}
	p.Connect()
	err := p.Write([]telegraf.Metric{oneMetric, twoMetric, fourMetric, threeMetric})
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
	fiveMetric, _ := metric.New("m", map[string]string{"t": "tv", "t3": "tv3"}, map[string]interface{}{"f": 7, "f3": 8}, timestamp)
	err = p.Write([]telegraf.Metric{fiveMetric})
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
}

func TestPostgresqlMetricsFromMeasure(t *testing.T) {
	postgreSQL, metrics, metricIndices := prepareAllColumnsInOnePlaceNoJSON()
	err := postgreSQL.writeMetricsFromMeasure(metrics[0].Name(), metricIndices["m"], metrics)
	assert.NoError(t, err)
	postgreSQL, metrics, metricIndices = prepareAllColumnsInOnePlaceTagsAndFieldsJSON()
	err = postgreSQL.writeMetricsFromMeasure(metrics[0].Name(), metricIndices["m"], metrics)
	assert.NoError(t, err)
}

func prepareAllColumnsInOnePlaceNoJSON() (*Postgresql, []telegraf.Metric, map[string][]int) {
	oneMetric, _ := metric.New("m", map[string]string{"t": "tv"}, map[string]interface{}{"f": 1}, time.Now())
	twoMetric, _ := metric.New("m", map[string]string{"t2": "tv2"}, map[string]interface{}{"f2": 2}, time.Now())
	threeMetric, _ := metric.New("m", map[string]string{"t": "tv", "t2": "tv2"}, map[string]interface{}{"f": 3, "f2": 4}, time.Now())

	return &Postgresql{
			TagTableSuffix:  "_tag",
			DoSchemaUpdates: true,
			tables:          &mockTables{t: map[string]bool{"m": true}, missingCols: []int{}},
			rows:            &mockTransformer{rows: [][]interface{}{nil, nil, nil}},
			columns:         columns.NewMapper(false, false, false),
			db:              &mockDb{},
		}, []telegraf.Metric{
			oneMetric, twoMetric, threeMetric,
		}, map[string][]int{
			"m": []int{0, 1, 2},
		}
}

func prepareAllColumnsInOnePlaceTagsAndFieldsJSON() (*Postgresql, []telegraf.Metric, map[string][]int) {
	oneMetric, _ := metric.New("m", map[string]string{"t": "tv"}, map[string]interface{}{"f": 1}, time.Now())
	twoMetric, _ := metric.New("m", map[string]string{"t2": "tv2"}, map[string]interface{}{"f2": 2}, time.Now())
	threeMetric, _ := metric.New("m", map[string]string{"t": "tv", "t2": "tv2"}, map[string]interface{}{"f": 3, "f2": 4}, time.Now())

	return &Postgresql{
			TagTableSuffix:    "_tag",
			DoSchemaUpdates:   true,
			TagsAsForeignkeys: false,
			TagsAsJsonb:       true,
			FieldsAsJsonb:     true,
			tables:            &mockTables{t: map[string]bool{"m": true}, missingCols: []int{}},
			columns:           columns.NewMapper(false, true, true),
			rows:              &mockTransformer{rows: [][]interface{}{nil, nil, nil}},
			db:                &mockDb{},
		}, []telegraf.Metric{
			oneMetric, twoMetric, threeMetric,
		}, map[string][]int{
			"m": []int{0, 1, 2},
		}
}

type mockTables struct {
	t           map[string]bool
	createErr   error
	missingCols []int
	mismatchErr error
	addColsErr  error
}

func (m *mockTables) Exists(tableName string) bool {
	return m.t[tableName]
}
func (m *mockTables) CreateTable(tableName string, colDetails *utils.TargetColumns) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.t[tableName] = true
	return nil
}
func (m *mockTables) FindColumnMismatch(tableName string, colDetails *utils.TargetColumns) ([]int, error) {
	return m.missingCols, m.mismatchErr
}
func (m *mockTables) AddColumnsToTable(tableName string, columnIndices []int, colDetails *utils.TargetColumns) error {
	return m.addColsErr
}

type mockTransformer struct {
	rows    [][]interface{}
	current int
	rowErr  error
}

func (mt *mockTransformer) createRowFromMetric(numColumns int, metric telegraf.Metric, targetColumns, targetTagColumns *utils.TargetColumns) ([]interface{}, error) {
	if mt.rowErr != nil {
		return nil, mt.rowErr
	}
	row := mt.rows[mt.current]
	mt.current++
	return row, nil
}

type mockDb struct {
	doCopyErr error
}

func (m *mockDb) Exec(query string, args ...interface{}) (pgx.CommandTag, error) {
	return "", nil
}
func (m *mockDb) DoCopy(fullTableName *pgx.Identifier, colNames []string, batch [][]interface{}) error {
	return m.doCopyErr
}
func (m *mockDb) Query(query string, args ...interface{}) (*pgx.Rows, error) {
	return nil, nil
}
func (m *mockDb) QueryRow(query string, args ...interface{}) *pgx.Row {
	return nil
}
func (m *mockDb) Close() error {
	return nil
}
