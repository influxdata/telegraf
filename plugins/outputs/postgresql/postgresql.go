package postgresql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/jackc/pgx"
	"github.com/lib/pq"
)

type MetricGroup struct {
	Metrics   []telegraf.Metric
	Timestamp time.Time
}

type Postgresql struct {
	db                *sql.DB
	Address           string
	TagsAsForeignkeys bool
	TagsAsJsonb       bool
	FieldsAsJsonb     bool
	Timestamp         string
	TableTemplate     string
	TagTableSuffix    string
	Tables            map[string]bool
	QueueDataDir      string
	Connections       int
	MaxItems          uint64
	TagKey            []string
	Strategy          string
	Exclude           []string

	inputMutex sync.RWMutex
	inputQueue chan []telegraf.Metric
	//inputQueue  *goque.Queue
	insertTypes map[string]map[string]string
	exclude     []*regexp.Regexp
}

type InsertKey struct {
	timestamp time.Time
	tags      map[string]string
}

type InsertItem struct {
	TableName string
	Columns   []string
	Values    []interface{}
	Types     map[string]string
}

func (p *Postgresql) Connect() error {
	var err error
	var db *sql.DB

	p.insertTypes = make(map[string]map[string]string)
	//gob.Register(time.Time{})
	p.inputQueue = make(chan []telegraf.Metric, p.MaxItems*uint64(p.Connections))
	// p.inputQueue, err = goque.OpenQueue(p.QueueDataDir)
	// if err != nil {
	// 	return err
	// }
	maxconn := int(p.MaxItems * uint64(p.Connections) * 3)

	//if p.Strategy == "batch" {
	//	db, err = sql.Open("pgx", p.Address)
	//} else if p.Strategy == "copy" {
	db, err = sql.Open("postgres", p.Address)
	//}

	if err != nil {
		return err
	}
	p.db = db
	p.db.SetMaxOpenConns(maxconn)
	p.Tables = make(map[string]bool)

	for i := 0; i < p.Connections; i++ {
		// db, err := sql.Open("pgx", p.Address)
		// if err != nil {
		// 	return err
		// }
		//go p.HandleInserts(db)
		go p.WriteMetrics(i)
	}

	return nil
}

func PostgreSQL_Copy(txn *sql.Tx, insertItem InsertItem) error {
	query := pq.CopyIn(insertItem.TableName, insertItem.Columns...)
	stmt, err := txn.Prepare(query)
	if err != nil {
		log.Println("ERROR: [txn.Prepare]: ", err)
		return err
	}

	_, err = stmt.Exec(insertItem.Values...)
	if err != nil {
		log.Println("ERROR: [insert.Values]: ", err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Println("ERROR: [Exec]: ", err)
		return err
	}

	err = stmt.Close()
	if err != nil {
		log.Println("ERROR: [stmt.Close]: ", err)
		return err
	}

	return nil
}

func (p *Postgresql) HandleInserts_Copy(i int, insertItem InsertItem, pwg *sync.WaitGroup) error {
	//defer pwg.Done()
	txn, err := p.db.Begin()
	if err != nil {
		log.Println("ERROR: [db.Begin]: ", err)
		return err
	}

	err = PostgreSQL_Copy(txn, insertItem)
	if err != nil {
		err = txn.Rollback()
		if err != nil {
			log.Println("ERROR: [txn.Rollback]: ", err)
			return err
		}
	} else {
		err = txn.Commit()
		if err != nil {
			log.Println("ERROR: [txn.Commit]: ", err)
			return err
		}
	}

	return nil
}

func (p *Postgresql) PostgreSQL_Batch(txn *sql.Tx, insertItems map[string]InsertItem) error {
	//var insertSqlArray []string
	//defer pwg.Done()
	for _, insert := range insertItems {
		sql := p.generateInsertWithValues(insert.TableName, insert.Columns, insert.Values)
		_, err := txn.Exec(sql)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Postgresql) HandleInserts_Batch(i int, insertItems map[string]InsertItem) error {
	//defer pwg.Done()
	txn, err := p.db.Begin()
	if err != nil {
		return err
	}

	err = p.PostgreSQL_Batch(txn, insertItems)
	if err != nil {
		err2 := txn.Rollback()
		if err2 != nil {
			log.Println("ERROR: [txn.Rollback]: ", err2)
		}

		exists, table, column := p.ColumnExists(err)
		if !exists {
			err2 = p.AddColumn(table, column)
			if err2 != nil {
				log.Println("ERROR [batch.AddColumn]: ", err2)
			}
		} else {
			log.Println("ERROR [batch.Write]: ", err)
		}
	} else {
		err = txn.Commit()
		if err != nil {
			log.Println("ERROR: [txn.Commit]: ", err)
		}
	}

	return err
}

func (p *Postgresql) ColumnExists(err error) (bool, string, string) {
	if p.FieldsAsJsonb == false {
		missingColumnRegex := regexp.MustCompile("pq: column \"(.*?)\" of relation \"(.*?)\" does not exist.*$")
		//dpInvalidInput := regexp.MustCompile("ERROR: invalid input syntax for type (.*?): \"(.*?)\".*$")
		matches := missingColumnRegex.FindStringSubmatch(err.Error())
		if matches != nil && len(matches) > 2 {
			table := matches[2]
			column := matches[1]

			return false, table, column
		}

		return true, "", ""
	}

	return true, "", ""
}

func (p *Postgresql) AddColumn(table string, column string) error {
	query := "ALTER TABLE %s.%s ADD COLUMN %s %s;"
	dbquery := fmt.Sprintf(query, quoteIdent("public"), quoteIdent(table), quoteIdent(column), "double precision")
	log.Println(dbquery)
	_, err := p.db.Exec(dbquery)
	if err != nil {
		return err
	}

	log.Println("Added Column", column, "to table", table)
	return nil
}

func (p *Postgresql) Close() error {
	return p.db.Close()
}

func contains(haystack []string, needle string) bool {
	for _, key := range haystack {
		if found, _ := regexp.MatchString(needle, key); found {
			return true
		}
	}
	return false
}

func quoteIdent(name string) string {
	return pgx.Identifier{name}.Sanitize()
}

func quoteLiteral(name string) string {
	return "'" + strings.Replace(name, "'", "''", -1) + "'"
}

func deriveDatatype(value interface{}) string {
	var datatype string

	switch value.(type) {
	case int64:
		datatype = "int8"
	case float64:
		datatype = "float8"
	case string:
		datatype = "text"
	default:
		datatype = "text"
		log.Printf("E! Unknown datatype %v", value)
	}
	return datatype
}

var sampleConfig = `
  ## specify address via a url matching:
  ##   postgres://[pqgotest[:password]]@localhost[/dbname]\
  ##       ?sslmode=[disable|verify-ca|verify-full]
  ## or a simple string:
  ##   host=localhost user=pqotest password=... sslmode=... dbname=app_production
  ##
  ## All connection parameters are optional.
  ##
  ## Without the dbname parameter, the driver will default to a database
  ## with the same name as the user. This dbname is just for instantiating a
  ## connection with the server and doesn't restrict the databases we are trying
  ## to grab metrics for.
  ##
  address = "host=localhost user=postgres sslmode=verify-full"

  ## Store tags as foreign keys in the metrics table. Default is false.
  # tags_as_foreignkeys = false

  ## Template to use for generating tables
  ## Available Variables: 
  ##   {TABLE} - tablename as identifier
  ##   {TABLELITERAL} - tablename as string literal
  ##   {COLUMNS} - column definitions
  ##   {KEY_COLUMNS} - comma-separated list of key columns (time + tags)

  ## Default template
  # table_template = "CREATE TABLE {TABLE}({COLUMNS})"
  ## Example for timescale
  # table_template = "CREATE TABLE {TABLE}({COLUMNS}); SELECT create_hypertable({TABLELITERAL},'time',chunk_time_interval := '1 week'::interval);"

  ## Use jsonb datatype for tags
  # tags_as_jsonb = true

  ## Use jsonb datatype for fields
  # fields_as_jsonb = true

`

func (p *Postgresql) SampleConfig() string { return sampleConfig }
func (p *Postgresql) Description() string  { return "Send metrics to PostgreSQL" }

func (p *Postgresql) generateCreateTable(metric telegraf.Metric) string {
	var columns []string
	var pk []string
	var sql []string

	pk = append(pk, quoteIdent("time"))
	columns = append(columns, "time timestamp")

	// handle tags if necessary
	if len(metric.Tags()) > 0 {
		if p.TagsAsForeignkeys {
			// tags in separate table
			var tag_columns []string
			var tag_columndefs []string
			columns = append(columns, "tag_id int")

			if p.TagsAsJsonb {
				tag_columns = append(tag_columns, "tags")
				tag_columndefs = append(tag_columndefs, "tags jsonb")
			} else {
				for column, _ := range metric.Tags() {
					tag_columns = append(tag_columns, quoteIdent(column))
					tag_columndefs = append(tag_columndefs, fmt.Sprintf("%s text", quoteIdent(column)))
				}
			}
			table := quoteIdent(metric.Name() + p.TagTableSuffix)
			sql = append(sql, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(tag_id serial primary key,%s,UNIQUE(%s))", table, strings.Join(tag_columndefs, ","), strings.Join(tag_columns, ",")))
		} else {
			// tags in measurement table
			if p.TagsAsJsonb {
				columns = append(columns, "tags jsonb")
			} else {
				for column, _ := range metric.Tags() {
					pk = append(pk, quoteIdent(column))
					columns = append(columns, fmt.Sprintf("%s text", quoteIdent(column)))
				}
			}
		}
	}

	if p.FieldsAsJsonb {
		columns = append(columns, "fields jsonb")
	} else {
		var datatype string
		for column, v := range metric.Fields() {
			datatype = deriveDatatype(v)
			columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(column), datatype))
		}
	}

	query := strings.Replace(p.TableTemplate, "{TABLE}", quoteIdent(metric.Name()), -1)
	query = strings.Replace(query, "{TABLELITERAL}", quoteLiteral("\""+metric.Name()+"\""), -1)
	query = strings.Replace(query, "{COLUMNS}", strings.Join(columns, ","), -1)
	query = strings.Replace(query, "{KEY_COLUMNS}", strings.Join(pk, ","), -1)

	sql = append(sql, query)
	return strings.Join(sql, ";")
}

func (p *Postgresql) generateInsert(tablename string, columns []string) string {

	var placeholder, quoted []string
	for i, column := range columns {
		placeholder = append(placeholder, fmt.Sprintf("$%d", i+1))
		quoted = append(quoted, quoteIdent(column))
	}

	sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", quoteIdent(tablename), strings.Join(quoted, ","), strings.Join(placeholder, ","))
	return sql
}

func (p *Postgresql) generateInsertWithValues(tablename string, columns []string, values []interface{}) string {
	var qvals, quoted []string
	for i, column := range columns {
		qval := values[i]
		switch qval.(type) {
		case int64:
			if val, ok := qval.(int64); ok {
				sval := fmt.Sprintf("%d", val)
				qvals = append(qvals, sval)
			} else {
				fmt.Println("Could not convert: ", qval)
			}
			break
		case float64:
			if val, ok := qval.(float64); ok {
				sval := fmt.Sprintf("%f", val)
				qvals = append(qvals, sval)
			} else {
				fmt.Println("Could not convert: ", qval)
			}
			break
		case time.Time:
			sval := fmt.Sprintf("'%s'", qval.(time.Time).Format("2006-01-02 15:04:05"))
			qvals = append(qvals, sval)
			break
		case []uint8:
			var sval []string
			for _, v := range qval.([]uint8) {
				sval = append(sval, fmt.Sprintf("%d", v))
			}
			qvals = append(qvals, strings.Join(sval, ","))
		case string:
			qvals = append(qvals, quoteLiteral(qval.(string)))
			break
		default:
			break
		}
		quoted = append(quoted, quoteIdent(column))
	}

	sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", quoteIdent(tablename), strings.Join(quoted, ","), strings.Join(qvals, ","))
	return sql
}

func (p *Postgresql) tableExists(tableName string) bool {
	stmt := "SELECT tablename FROM pg_tables WHERE tablename = $1 AND schemaname NOT IN ('information_schema','pg_catalog');"
	result, err := p.db.Exec(stmt, tableName)
	if err != nil {
		log.Printf("E! Error checking for existence of metric table %s: %v", tableName, err)
		return false
	}
	if count, _ := result.RowsAffected(); count == 1 {
		p.inputMutex.Lock()
		p.Tables[tableName] = true
		p.inputMutex.Unlock()
		return true
	}
	return false
}

func (p *Postgresql) getInsertKey(timestamp time.Time, tags map[string]string) string {
	ret := fmt.Sprintf("%d", timestamp.UTC().UnixNano())
	tagArray := make([]string, len(p.TagKey))
	i := 0
	for _, key := range p.TagKey {
		kvStr := key + "=" + tags[key]
		tagArray[i] = kvStr
		i++
	}
	sort.Strings(tagArray)
	for _, val := range tagArray {
		ret += "," + val
	}
	return ret
}

func (p *Postgresql) WriteMetrics(id int) {
	var wg sync.WaitGroup
	for true {
		var i, maxItems uint64
		var metrics []telegraf.Metric
		tableItems := make(map[string]map[string]InsertItem)

		//p.inputMutex.RLock()
		queueLength := uint64(len(p.inputQueue))
		maxItems = queueLength

		if maxItems > p.MaxItems {
			maxItems = p.MaxItems
		}

		for i = 0; i < maxItems; i++ {
			m := <-p.inputQueue
			metrics = append(metrics, m...)
		}

		for _, metric := range metrics {
			tablename := metric.Name()

			// Don't process items that are in the exclude list
			if contains(p.Exclude, tablename) {
				continue
			}

			if _, ok := tableItems[tablename]; !ok {
				tableItems[tablename] = make(map[string]InsertItem)
			}
			insertItems := tableItems[tablename]

			// // create table if needed
			// if p.Tables[tablename] == false && p.tableExists(tablename) == false {
			// 	createStmt := p.generateCreateTable(metric)
			// 	_, err := p.db.Exec(createStmt)
			// 	if err != nil {
			// 		log.Println("ERROR: ", err)
			// 	}
			// 	p.inputMutex.Lock()
			// 	p.Tables[tablename] = true
			// 	p.inputMutex.Unlock()
			// }

			var timestamp time.Time
			if p.Timestamp == "utc" {
				timestamp = metric.Time().UTC()
			} else {
				timestamp = metric.Time().Local()
			}

			var js map[string]interface{}
			insertKey := p.getInsertKey(timestamp, metric.Tags())
			if _, ok := insertItems[insertKey]; !ok {
				var newItem InsertItem
				newItem.Columns = append(newItem.Columns, "time")
				newItem.Values = append(newItem.Values, timestamp)
				newItem.Types = make(map[string]string)

				if len(metric.Tags()) > 0 {
					if p.TagsAsForeignkeys {
						// tags in separate table
						var tag_id int
						var where_columns []string
						var where_values []interface{}

						if p.TagsAsJsonb {
							js = make(map[string]interface{})
							for column, value := range metric.Tags() {
								js[column] = value
							}

							if len(js) > 0 {
								d, err := json.Marshal(js)
								if err != nil {
									log.Println("ERROR: ", err)
								}

								where_columns = append(where_columns, "tags")
								where_values = append(where_values, d)
							}
						} else {
							for column, value := range metric.Tags() {
								where_columns = append(where_columns, column)
								where_values = append(where_values, value)
								newItem.Types[column] = "text"
							}
						}

						var where_parts []string
						for i, column := range where_columns {
							where_parts = append(where_parts, fmt.Sprintf("%s = $%d", quoteIdent(column), i+1))
						}
						query := fmt.Sprintf("SELECT tag_id FROM %s WHERE %s", quoteIdent(tablename+p.TagTableSuffix), strings.Join(where_parts, " AND "))

						err := p.db.QueryRow(query, where_values...).Scan(&tag_id)
						if err != nil {
							// log.Printf("I! Foreign key reference not found %s: %v", tablename, err)
							query := p.generateInsert(tablename+p.TagTableSuffix, where_columns) + " RETURNING tag_id"
							err := p.db.QueryRow(query, where_values...).Scan(&tag_id)
							if err != nil {
								log.Println("ERROR: ", err)
							}
						}

						newItem.Columns = append(newItem.Columns, "tag_id")
						newItem.Values = append(newItem.Values, tag_id)
					} else {
						// tags in measurement table
						if p.TagsAsJsonb {
							js = make(map[string]interface{})
							for column, value := range metric.Tags() {
								js[column] = value
							}

							if len(js) > 0 {
								d, err := json.Marshal(js)
								if err != nil {
									log.Println("ERROR: ", err)
								}

								newItem.Columns = append(newItem.Columns, "tags")
								newItem.Values = append(newItem.Values, d)
							}
						} else {
							for column, value := range metric.Tags() {
								newItem.Columns = append(newItem.Columns, column)
								newItem.Values = append(newItem.Values, value)
								newItem.Types[column] = "text"
							}
						}
					}
				}

				insertItems[insertKey] = newItem
			}

			insertItem := insertItems[insertKey]

			if p.FieldsAsJsonb {
				js = make(map[string]interface{})
				for column, value := range metric.Fields() {
					js[column] = value
				}

				d, err := json.Marshal(js)
				if err != nil {
					log.Println("ERROR: ", err)
				}

				insertItem.Columns = append(insertItem.Columns, "fields")
				insertItem.Values = append(insertItem.Values, d)
			} else {
				for column, value := range metric.Fields() {
					if !contains(insertItem.Columns, column) {
						insertItem.Columns = append(insertItem.Columns, column)
						insertItem.Values = append(insertItem.Values, value)
						insertItem.Types[column] = deriveDatatype(value)
					}
				}
			}

			insertItem.TableName = tablename
			tableItems[tablename][insertKey] = insertItem
		}

		for _, insertItems := range tableItems {
			if p.Strategy == "copy" {
				for _, insertItem := range insertItems {
					//wg.Add(1)
					p.HandleInserts_Copy(id, insertItem, &wg)
				}
			} else if p.Strategy == "batch" {
				//wg.Add(1)
				p.HandleInserts_Batch(id, insertItems)
			}
		}
		//wg.Wait()

		time.Sleep(100 * time.Millisecond)
	}
}

func (p *Postgresql) Write(metrics []telegraf.Metric) error {
	//p.inputMutex.Lock()
	p.inputQueue <- metrics
	//p.inputMutex.Unlock()
	return nil
}

func init() {
	outputs.Add("postgresql", func() telegraf.Output { return newPostgresql() })
}

func newPostgresql() *Postgresql {
	return &Postgresql{
		TableTemplate:  "CREATE TABLE {TABLE}({COLUMNS})",
		TagsAsJsonb:    true,
		TagTableSuffix: "_tag",
		FieldsAsJsonb:  true,
	}
}
