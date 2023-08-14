package postgresql

import "github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"

func (p *Postgresql) columnFromTag(key string, value interface{}) utils.Column {
	return utils.Column{Name: key, Type: p.derivePgDatatype(value), Role: utils.TagColType}
}
func (p *Postgresql) columnFromField(key string, value interface{}) utils.Column {
	return utils.Column{Name: key, Type: p.derivePgDatatype(value), Role: utils.FieldColType}
}
