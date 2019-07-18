package postgresql

import (
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/columns"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/db"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestPostgresqlMetricsFromMeasure(t *testing.T) {
	postgreSQL, metrics, metricIndices := prepareAllColumnsInOnePlaceNoJSON()
	err := postgreSQL.writeMetricsFromMeasure(metrics[0].Name(), metricIndices["m"], metrics)
	assert.NoError(t, err)
	postgreSQL, metrics, metricIndices = prepareAllColumnsInOnePlaceTagsAndFieldsJSON()
	err = postgreSQL.writeMetricsFromMeasure(metrics[0].Name(), metricIndices["m"], metrics)
	assert.NoError(t, err)
}

func TestPostgresqlIsAliveCalledOnWrite(t *testing.T) {
	postgreSQL, metrics, _ := prepareAllColumnsInOnePlaceNoJSON()
	mockedDb := postgreSQL.db.(*mockDb)
	mockedDb.isAliveResponses = []bool{true}
	err := postgreSQL.Write(metrics[:1])
	assert.NoError(t, err)
	assert.Equal(t, 1, mockedDb.currentIsAliveResponse)
}

func TestPostgresqlDbAssignmentLock(t *testing.T) {
	postgreSQL, metrics, _ := prepareAllColumnsInOnePlaceNoJSON()
	mockedDb := postgreSQL.db.(*mockDb)
	mockedDb.isAliveResponses = []bool{true}
	mockedDb.secondsToSleepInIsAlive = 3
	var endOfWrite, startOfWrite, startOfReset, endOfReset time.Time
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		startOfWrite = time.Now()
		err := postgreSQL.Write(metrics[:1])
		assert.NoError(t, err)
		endOfWrite = time.Now()
		wg.Done()
	}()
	time.Sleep(time.Second)

	go func() {
		startOfReset = time.Now()
		postgreSQL.dbConnLock.Lock()
		time.Sleep(time.Second)
		postgreSQL.dbConnLock.Unlock()
		endOfReset = time.Now()
		wg.Done()
	}()
	wg.Wait()
	assert.True(t, startOfWrite.Before(startOfReset))
	assert.True(t, startOfReset.Before(endOfWrite))
	assert.True(t, endOfWrite.Before(endOfReset))
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
			dbConnLock:      sync.Mutex{},
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
			dbConnLock:        sync.Mutex{},
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
func (m *mockTables) SetConnection(db db.Wrapper) {}

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
	doCopyErr               error
	isAliveResponses        []bool
	currentIsAliveResponse  int
	secondsToSleepInIsAlive int64
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

func (m *mockDb) IsAlive() bool {
	if m.secondsToSleepInIsAlive > 0 {
		time.Sleep(time.Duration(m.secondsToSleepInIsAlive) * time.Second)
	}
	if m.isAliveResponses == nil {
		return true
	}
	if m.currentIsAliveResponse >= len(m.isAliveResponses) {
		return m.isAliveResponses[len(m.isAliveResponses)]
	}
	which := m.currentIsAliveResponse
	m.currentIsAliveResponse++
	return m.isAliveResponses[which]
}
