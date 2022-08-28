package postgresql

import (
	"fmt"
	"hash/fnv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
)

type columnList struct {
	columns []utils.Column
	indices map[string]int
}

func newColumnList() *columnList {
	return &columnList{
		indices: map[string]int{},
	}
}

func (cl *columnList) Add(column utils.Column) bool {
	if _, ok := cl.indices[column.Name]; ok {
		return false
	}
	cl.columns = append(cl.columns, column)
	cl.indices[column.Name] = len(cl.columns) - 1
	return true
}

func (cl *columnList) Remove(name string) bool {
	idx, ok := cl.indices[name]
	if !ok {
		return false
	}
	cl.columns = append(cl.columns[:idx], cl.columns[idx+1:]...)
	delete(cl.indices, name)

	for i, col := range cl.columns[idx:] {
		cl.indices[col.Name] = idx + i
	}

	return true
}

// TableSource satisfies pgx.CopyFromSource
type TableSource struct {
	postgresql   *Postgresql
	metrics      []telegraf.Metric
	cursor       int
	cursorValues []interface{}
	cursorError  error
	// tagHashSalt is so that we can use a global tag cache for all tables. The salt is unique per table, and combined
	// with the tag ID when looked up in the cache.
	tagHashSalt int64

	tagColumns *columnList
	// tagSets is the list of tag IDs to tag values in use within the TableSource. The position of each value in the list
	// corresponds to the key name in the tagColumns list.
	// This data is used to build out the foreign tag table when enabled.
	tagSets map[int64][]*telegraf.Tag

	fieldColumns *columnList

	droppedTagColumns []string
}

func NewTableSources(p *Postgresql, metrics []telegraf.Metric) map[string]*TableSource {
	tableSources := map[string]*TableSource{}

	for _, m := range metrics {
		tsrc := tableSources[m.Name()]
		if tsrc == nil {
			tsrc = NewTableSource(p, m.Name())
			tableSources[m.Name()] = tsrc
		}
		tsrc.AddMetric(m)
	}

	return tableSources
}

func NewTableSource(postgresql *Postgresql, name string) *TableSource {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))

	tsrc := &TableSource{
		postgresql:  postgresql,
		cursor:      -1,
		tagSets:     make(map[int64][]*telegraf.Tag),
		tagHashSalt: int64(h.Sum64()),
	}
	if !postgresql.TagsAsJsonb {
		tsrc.tagColumns = newColumnList()
	}
	if !postgresql.FieldsAsJsonb {
		tsrc.fieldColumns = newColumnList()
	}
	return tsrc
}

func (tsrc *TableSource) AddMetric(metric telegraf.Metric) {
	if tsrc.postgresql.TagsAsForeignKeys {
		tagID := utils.GetTagID(metric)
		if _, ok := tsrc.tagSets[tagID]; !ok {
			tsrc.tagSets[tagID] = metric.TagList()
		}
	}

	if !tsrc.postgresql.TagsAsJsonb {
		for _, t := range metric.TagList() {
			tsrc.tagColumns.Add(tsrc.postgresql.columnFromTag(t.Key, t.Value))
		}
	}

	if !tsrc.postgresql.FieldsAsJsonb {
		for _, f := range metric.FieldList() {
			tsrc.fieldColumns.Add(tsrc.postgresql.columnFromField(f.Key, f.Value))
		}
	}

	tsrc.metrics = append(tsrc.metrics, metric)
}

func (tsrc *TableSource) Name() string {
	if len(tsrc.metrics) == 0 {
		return ""
	}
	return tsrc.metrics[0].Name()
}

// Returns the superset of all tags of all metrics.
func (tsrc *TableSource) TagColumns() []utils.Column {
	var cols []utils.Column

	if tsrc.postgresql.TagsAsJsonb {
		cols = append(cols, tagsJSONColumn)
	} else {
		cols = append(cols, tsrc.tagColumns.columns...)
	}

	return cols
}

// Returns the superset of all fields of all metrics.
func (tsrc *TableSource) FieldColumns() []utils.Column {
	return tsrc.fieldColumns.columns
}

// Returns the full column list, including time, tag id or tags, and fields.
func (tsrc *TableSource) MetricTableColumns() []utils.Column {
	cols := []utils.Column{
		timeColumn,
	}

	if tsrc.postgresql.TagsAsForeignKeys {
		cols = append(cols, tagIDColumn)
	} else {
		cols = append(cols, tsrc.TagColumns()...)
	}

	if tsrc.postgresql.FieldsAsJsonb {
		cols = append(cols, fieldsJSONColumn)
	} else {
		cols = append(cols, tsrc.FieldColumns()...)
	}

	return cols
}

func (tsrc *TableSource) TagTableColumns() []utils.Column {
	cols := []utils.Column{
		tagIDColumn,
	}

	cols = append(cols, tsrc.TagColumns()...)

	return cols
}

func (tsrc *TableSource) ColumnNames() []string {
	cols := tsrc.MetricTableColumns()
	names := make([]string, len(cols))
	for i, col := range cols {
		names[i] = col.Name
	}
	return names
}

// Drops the specified column.
// If column is a tag column, any metrics containing the tag will be skipped.
// If column is a field column, any metrics containing the field will have it omitted.
func (tsrc *TableSource) DropColumn(col utils.Column) error {
	switch col.Role {
	case utils.TagColType:
		return tsrc.dropTagColumn(col)
	case utils.FieldColType:
		return tsrc.dropFieldColumn(col)
	case utils.TimeColType, utils.TagsIDColType:
		return fmt.Errorf("critical column \"%s\"", col.Name)
	default:
		return fmt.Errorf("internal error: unknown column \"%s\"", col.Name)
	}
}

// Drops the tag column from conversion. Any metrics containing this tag will be skipped.
func (tsrc *TableSource) dropTagColumn(col utils.Column) error {
	if col.Role != utils.TagColType || tsrc.postgresql.TagsAsJsonb {
		return fmt.Errorf("internal error: Tried to perform an invalid tag drop. measurement=%s tag=%s", tsrc.Name(), col.Name)
	}
	tsrc.droppedTagColumns = append(tsrc.droppedTagColumns, col.Name)

	if !tsrc.tagColumns.Remove(col.Name) {
		return nil
	}

	for setID, set := range tsrc.tagSets {
		for _, tag := range set {
			if tag.Key == col.Name {
				// The tag is defined, so drop the whole set
				delete(tsrc.tagSets, setID)
				break
			}
		}
	}
	return nil
}

// Drops the field column from conversion. Any metrics containing this field will have the field omitted.
func (tsrc *TableSource) dropFieldColumn(col utils.Column) error {
	if col.Role != utils.FieldColType || tsrc.postgresql.FieldsAsJsonb {
		return fmt.Errorf("internal error: Tried to perform an invalid field drop. measurement=%s field=%s", tsrc.Name(), col.Name)
	}

	tsrc.fieldColumns.Remove(col.Name)
	return nil
}

func (tsrc *TableSource) Next() bool {
	for {
		if tsrc.cursor+1 >= len(tsrc.metrics) {
			tsrc.cursorValues = nil
			tsrc.cursorError = nil
			return false
		}
		tsrc.cursor++

		tsrc.cursorValues, tsrc.cursorError = tsrc.getValues()
		if tsrc.cursorValues != nil || tsrc.cursorError != nil {
			return true
		}
	}
}

func (tsrc *TableSource) Reset() {
	tsrc.cursor = -1
}

// getValues calculates the values for the metric at the cursor position.
// If the metric cannot be emitted, such as due to dropped tags, or all fields dropped, the return value is nil.
func (tsrc *TableSource) getValues() ([]interface{}, error) {
	metric := tsrc.metrics[tsrc.cursor]

	values := []interface{}{
		metric.Time().UTC(),
	}

	if !tsrc.postgresql.TagsAsForeignKeys {
		if !tsrc.postgresql.TagsAsJsonb {
			// tags_as_foreignkey=false, tags_as_json=false
			tagValues := make([]interface{}, len(tsrc.tagColumns.columns))
			for _, tag := range metric.TagList() {
				tagPos, ok := tsrc.tagColumns.indices[tag.Key]
				if !ok {
					// tag has been dropped, we can't emit or we risk collision with another metric
					return nil, nil
				}
				tagValues[tagPos] = tag.Value
			}
			values = append(values, tagValues...)
		} else {
			// tags_as_foreign_key=false, tags_as_json=true
			values = append(values, utils.TagListToJSON(metric.TagList()))
		}
	} else {
		// tags_as_foreignkey=true
		tagID := utils.GetTagID(metric)
		if tsrc.postgresql.ForeignTagConstraint {
			if _, ok := tsrc.tagSets[tagID]; !ok {
				// tag has been dropped
				return nil, nil
			}
		}
		values = append(values, tagID)
	}

	if !tsrc.postgresql.FieldsAsJsonb {
		// fields_as_json=false
		fieldValues := make([]interface{}, len(tsrc.fieldColumns.columns))
		fieldsEmpty := true
		for _, field := range metric.FieldList() {
			// we might have dropped the field due to the table missing the column & schema updates being turned off
			if fPos, ok := tsrc.fieldColumns.indices[field.Key]; ok {
				fieldValues[fPos] = field.Value
				fieldsEmpty = false
			}
		}
		if fieldsEmpty {
			// all fields have been dropped. Don't emit a metric with just tags and no fields.
			return nil, nil
		}
		values = append(values, fieldValues...)
	} else {
		// fields_as_json=true
		value, err := utils.FieldListToJSON(metric.FieldList())
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, nil
}

func (tsrc *TableSource) Values() ([]interface{}, error) {
	return tsrc.cursorValues, tsrc.cursorError
}

func (tsrc *TableSource) Err() error {
	return nil
}

type TagTableSource struct {
	*TableSource
	tagIDs []int64

	cursor       int
	cursorValues []interface{}
	cursorError  error
}

func NewTagTableSource(tsrc *TableSource) *TagTableSource {
	ttsrc := &TagTableSource{
		TableSource: tsrc,
		cursor:      -1,
	}

	ttsrc.tagIDs = make([]int64, 0, len(tsrc.tagSets))
	for tagID := range tsrc.tagSets {
		ttsrc.tagIDs = append(ttsrc.tagIDs, tagID)
	}

	return ttsrc
}

func (ttsrc *TagTableSource) Name() string {
	return ttsrc.TableSource.Name() + ttsrc.postgresql.TagTableSuffix
}

func (ttsrc *TagTableSource) cacheCheck(tagID int64) bool {
	// Adding the 2 hashes is good enough. It's not a perfect solution, but given that we're operating in an int64
	// space, the risk of collision is extremely small.
	key := ttsrc.tagHashSalt + tagID
	_, err := ttsrc.postgresql.tagsCache.GetInt(key)
	return err == nil
}
func (ttsrc *TagTableSource) cacheTouch(tagID int64) {
	key := ttsrc.tagHashSalt + tagID
	_ = ttsrc.postgresql.tagsCache.SetInt(key, nil, 0)
}

func (ttsrc *TagTableSource) ColumnNames() []string {
	cols := ttsrc.TagTableColumns()
	names := make([]string, len(cols))
	for i, col := range cols {
		names[i] = col.Name
	}
	return names
}

func (ttsrc *TagTableSource) Next() bool {
	for {
		if ttsrc.cursor+1 >= len(ttsrc.tagIDs) {
			ttsrc.cursorValues = nil
			return false
		}
		ttsrc.cursor++

		if ttsrc.cacheCheck(ttsrc.tagIDs[ttsrc.cursor]) {
			// tag ID already inserted
			continue
		}

		ttsrc.cursorValues = ttsrc.getValues()
		if ttsrc.cursorValues != nil {
			return true
		}
	}
}

func (ttsrc *TagTableSource) Reset() {
	ttsrc.cursor = -1
}

func (ttsrc *TagTableSource) getValues() []interface{} {
	tagID := ttsrc.tagIDs[ttsrc.cursor]
	tagSet := ttsrc.tagSets[tagID]

	var values []interface{}
	if !ttsrc.postgresql.TagsAsJsonb {
		values = make([]interface{}, len(ttsrc.TableSource.tagColumns.indices)+1)
		for _, tag := range tagSet {
			values[ttsrc.TableSource.tagColumns.indices[tag.Key]+1] = tag.Value // +1 to account for tag_id column
		}
	} else {
		values = make([]interface{}, 2)
		values[1] = utils.TagListToJSON(tagSet)
	}
	values[0] = tagID

	return values
}

func (ttsrc *TagTableSource) Values() ([]interface{}, error) {
	return ttsrc.cursorValues, ttsrc.cursorError
}

func (ttsrc *TagTableSource) UpdateCache() {
	for _, tagID := range ttsrc.tagIDs {
		ttsrc.cacheTouch(tagID)
	}
}

func (ttsrc *TagTableSource) Err() error {
	return nil
}
