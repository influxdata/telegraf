package postgresql

import (
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
)

const (
	tagIDColumn             = "tag_id"
	createTagsTableTemplate = "CREATE TABLE IF NOT EXISTS %s(tag_id serial primary key,%s,UNIQUE(%s))"
)

func (p *Postgresql) generateCreateTable(metric telegraf.Metric) string {
	var columns []string
	var pk []string
	var sql []string

	pk = append(pk, quoteIdent("time"))
	columns = append(columns, "time timestamptz")

	// handle tags if necessary
	if len(metric.Tags()) > 0 {
		if p.TagsAsForeignkeys {
			// tags in separate table
			var tagColumns []string
			var tagColumndefs []string
			columns = append(columns, "tag_id int")

			if p.TagsAsJsonb {
				tagColumns = append(tagColumns, "tags")
				tagColumndefs = append(tagColumndefs, "tags jsonb")
			} else {
				for column := range metric.Tags() {
					tagColumns = append(tagColumns, quoteIdent(column))
					tagColumndefs = append(tagColumndefs, fmt.Sprintf("%s text", quoteIdent(column)))
				}
			}
			table := p.fullTableName(metric.Name() + p.TagTableSuffix)
			sql = append(sql, fmt.Sprintf(createTagsTableTemplate, table, strings.Join(tagColumndefs, ","), strings.Join(tagColumns, ",")))
		} else {
			// tags in measurement table
			if p.TagsAsJsonb {
				columns = append(columns, "tags jsonb")
			} else {
				for column := range metric.Tags() {
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

	query := strings.Replace(p.TableTemplate, "{TABLE}", p.fullTableName(metric.Name()), -1)
	query = strings.Replace(query, "{TABLELITERAL}", quoteLiteral(p.fullTableName(metric.Name())), -1)
	query = strings.Replace(query, "{COLUMNS}", strings.Join(columns, ","), -1)
	query = strings.Replace(query, "{KEY_COLUMNS}", strings.Join(pk, ","), -1)

	sql = append(sql, query)
	return strings.Join(sql, ";")
}
