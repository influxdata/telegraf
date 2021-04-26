package postgresql

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/influxdata/telegraf/plugins/outputs/postgresql/template"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
)

const (
	refreshTableStructureStatement = `
		SELECT column_name, data_type, col_description(format('%I.%I', table_schema, table_name)::regclass::oid, ordinal_position)
		FROM information_schema.columns
		WHERE table_schema = $1 and table_name = $2
	`
)

type tableState struct {
	name string
	// The atomic.Value protects columns from simple data race corruption as columns can be read while the mutex is
	// locked.
	columns atomic.Value
	// The mutex protects columns when doing a check-and-set operation. It prevents 2 goroutines from independently
	// checking the table's schema, and both trying to modify it, whether inconsistently, or to the same result.
	sync.Mutex
}
func (ts *tableState) Columns() map[string]utils.Column {
	cols := ts.columns.Load()
	if cols == nil {
		return nil
	}
	return cols.(map[string]utils.Column)
}
func (ts *tableState) SetColumns(cols map[string]utils.Column) {
	ts.columns.Store(cols)
}

type TableManager struct {
	*Postgresql

	// map[tableName]map[columnName]utils.Column
	tables      map[string]*tableState
	tablesMutex sync.Mutex
	// schemaMutex is used to prevent parallel table creations/alters in Postgres.
	schemaMutex sync.Mutex
}

// NewTableManager returns an instance of the tables.Manager interface
// that can handle checking and updating the state of tables in the PG database.
func NewTableManager(postgresql *Postgresql) *TableManager {
	return &TableManager{
		Postgresql: postgresql,
		tables:     make(map[string]*tableState),
	}
}

// ClearTableCache clear the table structure cache.
func (tm *TableManager) ClearTableCache() {
	tm.tablesMutex.Lock()
	for _, tbl := range tm.tables {
		tbl.SetColumns(nil)
	}
	tm.tablesMutex.Unlock()

	if tm.tagsCache != nil {
		tm.tagsCache.Clear()
	}
}

func (tm *TableManager) table(name string) *tableState {
	tm.tablesMutex.Lock()
	tbl := tm.tables[name]
	if tbl == nil {
		tbl = &tableState{name: name}
		tm.tables[name] = tbl
	}
	tm.tablesMutex.Unlock()
	return tbl
}

func (tm *TableManager) refreshTableStructure(ctx context.Context, db dbh, tbl *tableState) error {
	rows, err := db.Query(ctx, refreshTableStructureStatement, tm.Schema, tbl.name)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols := make(map[string]utils.Column)
	for rows.Next() {
		var colName, colTypeStr string
		desc := new(string)
		err := rows.Scan(&colName, &colTypeStr, &desc)
		if err != nil {
			return err
		}

		role := utils.FieldColType
		switch colName {
		case TimeColumnName:
			role = utils.TimeColType
		case TagIDColumnName:
			role = utils.TagsIDColType
		case TagsJSONColumnName:
			role = utils.TagColType
		case FieldsJSONColumnName:
			role = utils.FieldColType
		default:
			// We don't want to monopolize the column comment (preventing user from storing other information there), so just look at the first word
			if desc != nil {
				descWords := strings.Split(*desc, " ")
				if descWords[0] == "tag" {
					role = utils.TagColType
				}
			}
		}

		cols[colName] = utils.Column{
			Name: colName,
			Type: utils.PgDataType(colTypeStr),
			Role: role,
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(cols) > 0 {
		tbl.SetColumns(cols)
	}

	return nil
}

// EnsureStructure ensures that the table identified by tableName contains the provided columns.
//
// createTemplates and addColumnTemplates are the templates which are executed in the event of table create or alter
// (respectively).
// metricsTableName and tagsTableName are passed to the templates.
//
// If the table cannot be modified, the returned column list is the columns which are missing from the table.
func (tm *TableManager) EnsureStructure(
	ctx context.Context,
	db dbh,
	tbl *tableState,
	columns []utils.Column,
	createTemplates []*template.Template,
	addColumnsTemplates []*template.Template,
	metricsTable *tableState,
	tagsTable *tableState,
) ([]utils.Column, error) {
	// Sort so that:
	//   * When we create/alter the table the columns are in a sane order (telegraf gives us the fields in random order)
	//   * When we display errors about missing columns, the order is also sane, and consistent
	utils.ColumnList(columns).Sort()

	tbl.Lock()
	tblColumns := tbl.Columns()
	if tblColumns == nil {
		// We don't know about the table. First try to query it.
		if err := tm.refreshTableStructure(ctx, db, tbl); err != nil {
			tbl.Unlock()
			return nil, fmt.Errorf("querying table structure: %w", err)
		}
		tblColumns = tbl.Columns()

		if tblColumns == nil {
			// Ok, table doesn't exist, now we can create it.
			if err := tm.executeTemplates(ctx, db, createTemplates, tbl, columns, metricsTable, tagsTable); err != nil {
				tbl.Unlock()
				return nil, fmt.Errorf("creating table: %w", err)
			}

			tblColumns = tbl.Columns()
		}
	}
	tbl.Unlock()

	missingColumns, err := tm.checkColumns(tblColumns, columns)
	if err != nil {
		return nil, fmt.Errorf("column validation: %w", err)
	}
	if len(missingColumns) == 0 {
		return nil, nil
	}

	if len(addColumnsTemplates) == 0 {
		return missingColumns, nil
	}

	tbl.Lock()
	// Check again in case someone else got it while table was unlocked.
	tblColumns = tbl.Columns()
	missingColumns, _ = tm.checkColumns(tblColumns, columns)
	if len(missingColumns) == 0 {
		tbl.Unlock()
		return nil, nil
	}

	if err := tm.executeTemplates(ctx, db, addColumnsTemplates, tbl, missingColumns, metricsTable, tagsTable); err != nil {
		tbl.Unlock()
		return nil, fmt.Errorf("adding columns: %w", err)
	}
	tbl.Unlock()
	return tm.checkColumns(tbl.Columns(), columns)
}

func (tm *TableManager) checkColumns(dbColumns map[string]utils.Column, srcColumns []utils.Column) ([]utils.Column, error) {
	var missingColumns []utils.Column
	for _, srcCol := range srcColumns {
		dbCol, ok := dbColumns[srcCol.Name]
		if !ok {
			missingColumns = append(missingColumns, srcCol)
			continue
		}
		if !utils.PgTypeCanContain(dbCol.Type, srcCol.Type) {
			return nil, fmt.Errorf("column type '%s' cannot store '%s'", dbCol.Type, srcCol.Type)
		}
	}
	return missingColumns, nil
}

func (tm *TableManager) executeTemplates(
	ctx context.Context,
	db dbh,
	tmpls []*template.Template,
	tbl *tableState,
	newColumns []utils.Column,
	metricsTable *tableState,
	tagsTable *tableState,
) error {
	tmplTable := template.NewTable(tm.Schema, tbl.name, colMapToSlice(tbl.Columns()))
	metricsTmplTable := template.NewTable(tm.Schema, metricsTable.name, colMapToSlice(metricsTable.Columns()))
	var tagsTmplTable *template.Table
	if tagsTable != nil {
		tagsTmplTable = template.NewTable(tm.Schema, tagsTable.name, colMapToSlice(tagsTable.Columns()))
	} else {
		tagsTmplTable = template.NewTable("", "", nil)
	}

	/* https://github.com/jackc/pgx/issues/872
	stmts := make([]string, len(tmpls))
	batch := &pgx.Batch{}
	for i, tmpl := range tmpls {
		sql, err := tmpl.Render(tmplTable, newColumns, metricsTmplTable, tagsTmplTable)
		if err != nil {
			return err
		}
		stmts[i] = string(sql)
		batch.Queue(stmts[i])
	}

	batch.Queue(refreshTableStructureStatement, tm.Schema, tableName)

	batchResult := tm.db.SendBatch(ctx, batch)
	defer batchResult.Close()

	for i := 0; i < len(tmpls); i++ {
		if x, err := batchResult.Exec(); err != nil {
			return fmt.Errorf("executing `%.40s...`: %v %w", stmts[i], x, err)
		}
	}

	rows, err := batchResult.Query()
	if err != nil {
		return fmt.Errorf("refreshing table: %w", err)
	}
	tm.refreshTableStructureResponse(tableName, rows)
	*/

	// Lock to prevent concurrency issues in postgres (pg_type_typname_nsp_index unique constraint; SQLSTATE 23505)
	tm.schemaMutex.Lock()
	defer tm.schemaMutex.Unlock()

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, tmpl := range tmpls {
		sql, err := tmpl.Render(tmplTable, newColumns, metricsTmplTable, tagsTmplTable)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("executing `%s`: %w", sql, err)
		}
	}

	// We need to be able to determine the role of the column when reading the structure back (because of the templates).
	// For some columns we can determine this by the column name (time, tag_id, etc). However tags and fields can have any
	// name, and look the same. So we add a comment to tag columns, and through process of elimination what remains are
	// field columns.
	for _, col := range newColumns {
		if col.Role != utils.TagColType {
			continue
		}
		if _, err := tx.Exec(ctx, "COMMENT ON COLUMN "+tmplTable.String()+"."+template.QuoteIdentifier(col.Name)+" IS 'tag'"); err != nil {
			return fmt.Errorf("setting column role comment: %s", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return tm.refreshTableStructure(ctx, db, tbl)
}

func colMapToSlice(colMap map[string]utils.Column) []utils.Column {
	if colMap == nil {
		return nil
	}
	cols := make([]utils.Column, 0, len(colMap))
	for _, col := range colMap {
		cols = append(cols, col)
	}
	return cols
}

// MatchSource scans through the metrics, determining what columns are needed for inserting, and ensuring the DB schema matches.
//
// If the schema does not match, and schema updates are disabled:
// If a field missing from the DB, the field is omitted.
// If a tag is missing from the DB, the metric is dropped.
func (tm *TableManager) MatchSource(ctx context.Context, db dbh, rowSource *TableSource) error {
	metricTable := tm.table(rowSource.Name())
	var tagTable *tableState
	if tm.TagsAsForeignKeys {
		tagTable = tm.table(metricTable.name + tm.TagTableSuffix)

		missingCols, err := tm.EnsureStructure(
			ctx,
			db,
			tagTable,
			rowSource.TagTableColumns(),
			tm.TagTableCreateTemplates,
			tm.TagTableAddColumnTemplates,
			metricTable,
			tagTable,
		)
		if err != nil {
			return err
		}

		if len(missingCols) > 0 {
			colDefs := make([]string, len(missingCols))
			for i, col := range missingCols {
				if err := rowSource.DropColumn(col); err != nil {
					return fmt.Errorf("metric/table mismatch: Unable to omit field/column from \"%s\": %w", tagTable.name, err)
				}
				colDefs[i] = col.Name + " " + string(col.Type)
			}
			tm.Logger.Errorf("table '%s' is missing tag columns (dropping metrics): %s", tagTable.name, strings.Join(colDefs, ", "))
		}
	}

	missingCols, err := tm.EnsureStructure(
		ctx,
		db,
		metricTable,
		rowSource.MetricTableColumns(),
		tm.CreateTemplates,
		tm.AddColumnTemplates,
		metricTable,
		tagTable,
	)
	if err != nil {
		return err
	}

	if len(missingCols) > 0 {
		colDefs := make([]string, len(missingCols))
		for i, col := range missingCols {
			if err := rowSource.DropColumn(col); err != nil {
				return fmt.Errorf("metric/table mismatch: Unable to omit field/column from \"%s\": %w", metricTable.name, err)
			}
			colDefs[i] = col.Name + " " + string(col.Type)
		}
		tm.Logger.Errorf("table \"%s\" is missing columns (omitting fields): %s", metricTable.name, strings.Join(colDefs, ", "))
	}

	return nil
}
