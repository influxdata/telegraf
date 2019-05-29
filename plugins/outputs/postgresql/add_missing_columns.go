package postgresql

import (
	"fmt"
	"log"
	"strings"
)

func (p *Postgresql) addMissingColumns(tableName string, columns []string, values []interface{}) (bool, error) {
	columnStatuses, err := p.whichColumnsAreMissing(columns, tableName)
	if err != nil {
		return false, err
	}

	retry := false
	for currentColumn, isMissing := range columnStatuses {
		if !isMissing {
			continue
		}

		dataType := deriveDatatype(values[currentColumn])
		columnName := columns[currentColumn]
		if err := p.addColumnToTable(columnName, dataType, tableName); err != nil {
			return false, err
		}
		retry = true
	}

	return retry, nil
}

func prepareMissingColumnsQuery(columns []string) string {
	var quotedColumns = make([]string, len(columns))
	for i, column := range columns {
		quotedColumns[i] = quoteLiteral(column)
	}
	return fmt.Sprintf(missingColumnsTemplate, strings.Join(quotedColumns, ","))
}

// for a given array of columns x = [a, b, c ...] it returns an array of bools indicating
// if x[i] is missing
func (p *Postgresql) whichColumnsAreMissing(columns []string, tableName string) ([]bool, error) {
	missingColumnsQuery := prepareMissingColumnsQuery(columns)
	result, err := p.db.Query(missingColumnsQuery, p.Schema, tableName)
	if err != nil {
		return nil, err
	}
	defer result.Close()
	columnStatus := make([]bool, len(columns))
	var isMissing bool
	var columnName string
	currentColumn := 0

	for result.Next() {
		err := result.Scan(&columnName, &isMissing)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		columnStatus[currentColumn] = isMissing
		currentColumn++
	}

	return columnStatus, nil
}

func (p *Postgresql) addColumnToTable(columnName, dataType, tableName string) error {
	fullTableName := p.fullTableName(tableName)
	addColumnQuery := fmt.Sprintf(addColumnTemplate, fullTableName, quoteIdent(columnName), dataType)
	_, err := p.db.Exec(addColumnQuery)
	return err
}
