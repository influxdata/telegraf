package postgresql

import "github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"

// Define standard column types
var (
	timeColumn       = utils.Column{Name: "time", Type: PgTimestampWithoutTimeZone, Role: utils.TimeColType}
	tagIDColumn      = utils.Column{Name: "tag_id", Type: PgBigInt, Role: utils.TagsIDColType}
	fieldsJSONColumn = utils.Column{Name: "fields", Type: PgJSONb, Role: utils.FieldColType}
	tagsJSONColumn   = utils.Column{Name: "tags", Type: PgJSONb, Role: utils.TagColType}
)

func (p *Postgresql) columnFromTag(key string, value interface{}) utils.Column {
	return utils.Column{Name: key, Type: p.derivePgDatatype(value), Role: utils.TagColType}
}
func (p *Postgresql) columnFromField(key string, value interface{}) utils.Column {
	return utils.Column{Name: key, Type: p.derivePgDatatype(value), Role: utils.FieldColType}
}
