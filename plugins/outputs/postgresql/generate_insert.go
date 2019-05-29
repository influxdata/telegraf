package postgresql

import (
	"fmt"
	"strings"
)

const (
	insertIntoSQLTemplate = "INSERT INTO %s(%s) VALUES(%s)"
)

func (p *Postgresql) generateInsert(tablename string, columns []string) string {
	valuePlaceholders := make([]string, len(columns))
	quotedColumns := make([]string, len(columns))
	for i, column := range columns {
		valuePlaceholders[i] = fmt.Sprintf("$%d", i+1)
		quotedColumns[i] = quoteIdent(column)
	}

	fullTableName := p.fullTableName(tablename)
	columnNames := strings.Join(quotedColumns, ",")
	values := strings.Join(valuePlaceholders, ",")
	return fmt.Sprintf(insertIntoSQLTemplate, fullTableName, columnNames, values)
}
