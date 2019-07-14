package columns

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
)

func mapTags(tagList []*telegraf.Tag, alreadyMapped map[string]bool, columns *utils.TargetColumns) {
	for _, tag := range tagList {
		if _, ok := alreadyMapped[tag.Key]; !ok {
			alreadyMapped[tag.Key] = true
			columns.Target[tag.Key] = len(columns.Names)
			columns.Names = append(columns.Names, tag.Key)
			columns.DataTypes = append(columns.DataTypes, utils.PgText)
			columns.Roles = append(columns.Roles, utils.TagColType)
		}
	}
}
