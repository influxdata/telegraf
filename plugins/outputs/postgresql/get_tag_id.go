package postgresql

import (
	"fmt"
	"log"
	"strings"

	"github.com/influxdata/telegraf"
)

const (
	selectTagIDTemplate    = "SELECT tag_id FROM %s WHERE %s"
	missingColumnsTemplate = "WITH available AS (SELECT column_name as c FROM information_schema.columns WHERE table_schema = $1 and table_name = $2)," +
		"required AS (SELECT c FROM unnest(array [%s]) AS c) " +
		"SELECT required.c, available.c IS NULL FROM required LEFT JOIN available ON required.c = available.c;"

	addColumnTemplate = "ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s;"
)

func (p *Postgresql) getTagID(metric telegraf.Metric) (int, error) {
	var tagID int
	var whereColumns []string
	var whereValues []interface{}
	tablename := metric.Name()

	if p.TagsAsJsonb && len(metric.Tags()) > 0 {
		d, err := buildJsonbTags(metric.Tags())
		if err != nil {
			return tagID, err
		}

		whereColumns = append(whereColumns, "tags")
		whereValues = append(whereValues, d)
	} else {
		for column, value := range metric.Tags() {
			whereColumns = append(whereColumns, column)
			whereValues = append(whereValues, value)
		}
	}

	whereParts := make([]string, len(whereColumns))
	for i, column := range whereColumns {
		whereParts[i] = fmt.Sprintf("%s = $%d", quoteIdent(column), i+1)
	}

	tagsTableName := tablename + p.TagTableSuffix
	tagsTableFullName := p.fullTableName(tagsTableName)
	query := fmt.Sprintf(selectTagIDTemplate, tagsTableFullName, strings.Join(whereParts, " AND "))

	err := p.db.QueryRow(query, whereValues...).Scan(&tagID)
	if err == nil {
		return tagID, nil
	}
	query = p.generateInsert(tagsTableName, whereColumns) + " RETURNING tag_id"
	err = p.db.QueryRow(query, whereValues...).Scan(&tagID)
	if err == nil {
		return tagID, nil
	}

	// check if insert error was caused by column mismatch

	// if tags are jsonb, there shouldn't be a column mismatch
	if p.TagsAsJsonb {
		return tagID, err
	}

	// check for missing columns
	log.Printf("W! Possible column mismatch while inserting new tag-set: %v", err)
	retry, err := p.addMissingColumns(tagsTableName, whereColumns, whereValues)
	if err != nil {
		// missing coulmns not properly added
		log.Printf("E! Could not add missing columns: %v", err)
		return tagID, err
	}

	// We added some columns and insert might work now. Try again immediately to
	// avoid long lead time in getting metrics when there are several columns missing
	// from the original create statement and they get added in small drops.
	if retry {
		log.Printf("I! Retrying to insert new tag set")
		err := p.db.QueryRow(query, whereValues...).Scan(&tagID)
		if err != nil {
			return tagID, err
		}
	}
	return tagID, nil
}
