package postgresql

import "github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"

// Column names and data types for standard fields (time, tag_id, tags, and fields)
const (
	TimeColumnName       = "time"
	TimeColumnDataType   = utils.PgTimestampWithTimeZone
	TagIDColumnName      = "tag_id"
	TagIDColumnDataType  = utils.PgBigInt
	TagsJSONColumnName   = "tags"
	FieldsJSONColumnName = "fields"
	JSONColumnDataType   = utils.PgJSONb
)

var TimeColumn = utils.Column{TimeColumnName, TimeColumnDataType, utils.TimeColType}
var TagIDColumn = utils.Column{TagIDColumnName, TagIDColumnDataType, utils.TagsIDColType}
var FieldsJSONColumn = utils.Column{FieldsJSONColumnName, JSONColumnDataType, utils.FieldColType}
var TagsJSONColumn = utils.Column{TagsJSONColumnName, JSONColumnDataType, utils.TagColType}

func ColumnFromTag(key string, value interface{}) utils.Column {
	return utils.Column{key, utils.DerivePgDatatype(value), utils.TagColType}
}
func ColumnFromField(key string, value interface{}) utils.Column {
	return utils.Column{key, utils.DerivePgDatatype(value), utils.FieldColType}
}
