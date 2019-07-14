package postgresql

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/columns"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
)

type transformer interface {
	createRowFromMetric(numColumns int, metric telegraf.Metric, targetColumns, targetTagColumns *utils.TargetColumns) ([]interface{}, error)
}

type defTransformer struct {
	tagsAsFK      bool
	tagsAsJSONb   bool
	fieldsAsJSONb bool
	tagsCache     tagsCache
}

func newRowTransformer(tagsAsFK, tagsAsJSONb, fieldsAsJSONb bool, tagsCache tagsCache) transformer {
	return &defTransformer{
		tagsAsFK:      tagsAsFK,
		tagsAsJSONb:   tagsAsJSONb,
		fieldsAsJSONb: fieldsAsJSONb,
		tagsCache:     tagsCache,
	}
}

func (dt *defTransformer) createRowFromMetric(numColumns int, metric telegraf.Metric, targetColumns, targetTagColumns *utils.TargetColumns) ([]interface{}, error) {
	row := make([]interface{}, numColumns)
	// handle time
	row[0] = metric.Time()
	// handle tags and tag id
	if dt.tagsAsFK {
		tagID, err := dt.tagsCache.getTagID(targetTagColumns, metric)
		if err != nil {
			return nil, err
		}
		row[1] = tagID
	} else {
		if dt.tagsAsJSONb {
			jsonVal, err := utils.BuildJsonb(metric.Tags())
			if err != nil {
				return nil, err
			}
			targetIndex := targetColumns.Target[columns.TagsJSONColumn]
			row[targetIndex] = jsonVal
		} else {
			for _, tag := range metric.TagList() {
				targetIndex := targetColumns.Target[tag.Key]
				row[targetIndex] = tag.Value
			}
		}
	}

	// handle fields
	if dt.fieldsAsJSONb {
		jsonVal, err := utils.BuildJsonb(metric.Fields())
		if err != nil {
			return nil, err
		}
		targetIndex := targetColumns.Target[columns.FieldsJSONColumn]
		row[targetIndex] = jsonVal
	} else {
		for _, field := range metric.FieldList() {
			targetIndex := targetColumns.Target[field.Key]
			row[targetIndex] = field.Value
		}
	}

	return row, nil
}
