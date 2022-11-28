package postgresql

import "github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"

// Column names and data types for standard fields (time, tag_id, tags, and fields)
const (
	timeColumnName       = "time"
	timeColumnDataType   = PgTimestampWithoutTimeZone
	tagIDColumnName      = "tag_id"
	tagIDColumnDataType  = PgBigInt
	tagsJSONColumnName   = "tags"
	fieldsJSONColumnName = "fields"
	jsonColumnDataType   = PgJSONb
)

var timeColumn = utils.Column{Name: timeColumnName, Type: timeColumnDataType, Role: utils.TimeColType}
var tagIDColumn = utils.Column{Name: tagIDColumnName, Type: tagIDColumnDataType, Role: utils.TagsIDColType}
var fieldsJSONColumn = utils.Column{Name: fieldsJSONColumnName, Type: jsonColumnDataType, Role: utils.FieldColType}
var tagsJSONColumn = utils.Column{Name: tagsJSONColumnName, Type: jsonColumnDataType, Role: utils.TagColType}

func (p *Postgresql) columnFromTag(key string, value interface{}) utils.Column {
	return utils.Column{Name: key, Type: p.derivePgDatatype(value), Role: utils.TagColType}
}
func (p *Postgresql) columnFromField(key string, value interface{}) utils.Column {
	return utils.Column{Name: key, Type: p.derivePgDatatype(value), Role: utils.FieldColType}
}
