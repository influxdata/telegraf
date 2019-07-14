package columns

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
)

func mapFields(fieldList []*telegraf.Field, alreadyMapped map[string]bool, columns *utils.TargetColumns) {
	for _, field := range fieldList {
		if _, ok := alreadyMapped[field.Key]; !ok {
			alreadyMapped[field.Key] = true
			columns.Target[field.Key] = len(columns.Names)
			columns.Names = append(columns.Names, field.Key)
			columns.DataTypes = append(columns.DataTypes, utils.DerivePgDatatype(field.Value))
			columns.Roles = append(columns.Roles, utils.FieldColType)
		}
	}
}
