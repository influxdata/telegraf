package columns

import "github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"

// a function type that generates column details for the main, and tags table in the db
type targetColumnInitializer func() (*utils.TargetColumns, *utils.TargetColumns)

// constants used for populating the 'targetColumnInit' map (for better readability)
const (
	cTagsAsFK     = true
	cTagsAsJSON   = true
	cFieldsAsJSON = true
)

// Since some of the target columns for the tables in the database don't
// depend on the metrics received, but on the plugin config, we can have
// constant initializer functions. It is always known that the 'time'
// column goes first in the main table, then if the tags are kept in a
// separate table you need to add the 'tag_id' column...
// This map contains an initializer for all the combinations
// of (tagsAsFK, tagsAsJSON, fieldsAsJSON).
func getInitialColumnsGenerator(tagsAsFK, tagsAsJSON, fieldsAsJSON bool) targetColumnInitializer {
	return standardColumns[tagsAsFK][tagsAsJSON][fieldsAsJSON]
}

var standardColumns = map[bool]map[bool]map[bool]targetColumnInitializer{
	cTagsAsFK: {
		cTagsAsJSON: {
			cFieldsAsJSON:  tagsAsFKAndJSONAndFieldsAsJSONInit,
			!cFieldsAsJSON: tagsAsFKAndJSONInit,
		},
		!cTagsAsJSON: {
			cFieldsAsJSON:  tagsAsFKFieldsAsJSONInit,
			!cFieldsAsJSON: tagsAsFKInit,
		},
	},
	!cTagsAsFK: {
		cTagsAsJSON: {
			cFieldsAsJSON:  tagsAndFieldsAsJSONInit,
			!cFieldsAsJSON: tagsAsJSONInit,
		},
		!cTagsAsJSON: {
			cFieldsAsJSON:  fieldsAsJSONInit,
			!cFieldsAsJSON: vanillaColumns,
		},
	},
}

func tagsAsFKAndJSONAndFieldsAsJSONInit() (*utils.TargetColumns, *utils.TargetColumns) {
	return &utils.TargetColumns{
			Names:     []string{TimeColumnName, TagIDColumnName, FieldsJSONColumn},
			DataTypes: []utils.PgDataType{TimeColumnDataType, TagIDColumnDataType, JSONColumnDataType},
			Target:    map[string]int{TimeColumnName: 0, TagIDColumnName: 1, FieldsJSONColumn: 2},
			Roles:     []utils.ColumnRole{utils.TimeColType, utils.TagsIDColType, utils.FieldColType},
		}, &utils.TargetColumns{
			Names:     []string{TagIDColumnName, TagsJSONColumn},
			DataTypes: []utils.PgDataType{TagIDColumnDataTypeAsPK, JSONColumnDataType},
			Target:    map[string]int{TagIDColumnName: 0, TagsJSONColumn: 1},
			Roles:     []utils.ColumnRole{utils.TagsIDColType, utils.TagColType},
		}
}

func tagsAsFKAndJSONInit() (*utils.TargetColumns, *utils.TargetColumns) {
	return &utils.TargetColumns{
			Names:     []string{TimeColumnName, TagIDColumnName},
			DataTypes: []utils.PgDataType{TimeColumnDataType, TagIDColumnDataType},
			Target:    map[string]int{TimeColumnName: 0, TagIDColumnName: 1},
			Roles:     []utils.ColumnRole{utils.TimeColType, utils.TagsIDColType},
		}, &utils.TargetColumns{
			Names:     []string{TagIDColumnName, TagsJSONColumn},
			DataTypes: []utils.PgDataType{TagIDColumnDataTypeAsPK, JSONColumnDataType},
			Target:    map[string]int{TagIDColumnName: 0, TagsJSONColumn: 1},
			Roles:     []utils.ColumnRole{utils.TagsIDColType, utils.FieldColType},
		}
}

func tagsAsFKFieldsAsJSONInit() (*utils.TargetColumns, *utils.TargetColumns) {
	return &utils.TargetColumns{
			Names:     []string{TimeColumnName, TagIDColumnName, FieldsJSONColumn},
			DataTypes: []utils.PgDataType{TimeColumnDataType, TagIDColumnDataType, JSONColumnDataType},
			Target:    map[string]int{TimeColumnName: 0, TagIDColumnName: 1, FieldsJSONColumn: 2},
			Roles:     []utils.ColumnRole{utils.TimeColType, utils.TagsIDColType, utils.FieldColType},
		}, &utils.TargetColumns{
			Names:     []string{TagIDColumnName},
			DataTypes: []utils.PgDataType{TagIDColumnDataTypeAsPK},
			Target:    map[string]int{TagIDColumnName: 0},
			Roles:     []utils.ColumnRole{utils.TagsIDColType},
		}
}

func tagsAsFKInit() (*utils.TargetColumns, *utils.TargetColumns) {
	return &utils.TargetColumns{
			Names:     []string{TimeColumnName, TagIDColumnName},
			DataTypes: []utils.PgDataType{TimeColumnDataType, TagIDColumnDataType},
			Target:    map[string]int{TimeColumnName: 0, TagIDColumnName: 1},
			Roles:     []utils.ColumnRole{utils.TimeColType, utils.TagsIDColType},
		}, &utils.TargetColumns{
			Names:     []string{TagIDColumnName},
			DataTypes: []utils.PgDataType{TagIDColumnDataTypeAsPK},
			Target:    map[string]int{TagIDColumnName: 0},
			Roles:     []utils.ColumnRole{utils.TagsIDColType},
		}
}

func tagsAndFieldsAsJSONInit() (*utils.TargetColumns, *utils.TargetColumns) {
	return &utils.TargetColumns{
		Names:     []string{TimeColumnName, TagsJSONColumn, FieldsJSONColumn},
		DataTypes: []utils.PgDataType{TimeColumnDataType, JSONColumnDataType, JSONColumnDataType},
		Target:    map[string]int{TimeColumnName: 0, TagsJSONColumn: 1, FieldsJSONColumn: 2},
		Roles:     []utils.ColumnRole{utils.TimeColType, utils.TagColType, utils.FieldColType},
	}, nil
}

func tagsAsJSONInit() (*utils.TargetColumns, *utils.TargetColumns) {
	return &utils.TargetColumns{
		Names:     []string{TimeColumnName, TagsJSONColumn},
		DataTypes: []utils.PgDataType{TimeColumnDataType, JSONColumnDataType},
		Target:    map[string]int{TimeColumnName: 0, TagsJSONColumn: 1},
		Roles:     []utils.ColumnRole{utils.TimeColType, utils.TagColType},
	}, nil
}

func fieldsAsJSONInit() (*utils.TargetColumns, *utils.TargetColumns) {
	return &utils.TargetColumns{
		Names:     []string{TimeColumnName, FieldsJSONColumn},
		DataTypes: []utils.PgDataType{TimeColumnDataType, JSONColumnDataType},
		Target:    map[string]int{TimeColumnName: 0, FieldsJSONColumn: 1},
		Roles:     []utils.ColumnRole{utils.TimeColType, utils.FieldColType},
	}, nil
}

func vanillaColumns() (*utils.TargetColumns, *utils.TargetColumns) {
	return &utils.TargetColumns{
		Names:     []string{TimeColumnName},
		DataTypes: []utils.PgDataType{TimeColumnDataType},
		Target:    map[string]int{TimeColumnName: 0},
		Roles:     []utils.ColumnRole{utils.TimeColType},
	}, nil
}
