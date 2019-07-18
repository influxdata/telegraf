package postgresql

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/groupcache/lru"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/columns"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/db"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
)

const (
	selectTagIDTemplate = "SELECT tag_id FROM %s WHERE %s"
)

// TagsCache retrieves the appropriate tagID based on the tag values
// from the database (used only when TagsAsForeignKey property selected).
// Also caches the LRU tagIDs
type tagsCache interface {
	getTagID(target *utils.TargetColumns, metric telegraf.Metric) (int, error)
	tagsTableName(measureName string) string
	setDb(db db.Wrapper)
}

type defTagsCache struct {
	cache          map[string]*lru.Cache
	tagsAsJSONb    bool
	tagTableSuffix string
	schema         string
	db             db.Wrapper
	itemsToCache   int
}

// newTagsCache returns a new implementation of the tags cache interface with LRU memoization
func newTagsCache(numItemsInCachePerMetric int, tagsAsJSONb bool, tagTableSuffix, schema string, db db.Wrapper) tagsCache {
	return &defTagsCache{
		cache:          map[string]*lru.Cache{},
		tagsAsJSONb:    tagsAsJSONb,
		tagTableSuffix: tagTableSuffix,
		schema:         schema,
		db:             db,
		itemsToCache:   numItemsInCachePerMetric,
	}
}

func (c *defTagsCache) setDb(db db.Wrapper) {
	c.db = db
}

// Checks the cache for the tag set of the metric, if present returns immediately.
// Otherwise asks the database if that tag set has already been recorded.
// If not recorded, inserts a new row to the tags table for the specific measurement.
// Re-caches the tagID after checking the DB.
func (c *defTagsCache) getTagID(target *utils.TargetColumns, metric telegraf.Metric) (int, error) {
	measureName := metric.Name()
	tags := metric.Tags()
	cacheKey := constructCacheKey(tags)
	tagID, isCached := c.checkTagCache(measureName, cacheKey)
	if isCached {
		return tagID, nil
	}

	var whereParts []string
	var whereValues []interface{}
	if c.tagsAsJSONb {
		whereParts = []string{utils.QuoteIdent(columns.TagsJSONColumn) + "= $1"}
		numTags := len(tags)
		if numTags > 0 {
			d, err := utils.BuildJsonb(tags)
			if err != nil {
				return tagID, err
			}
			whereValues = []interface{}{d}
		} else {
			whereValues = []interface{}{nil}
		}
	} else {
		whereParts = make([]string, len(target.Names)-1)
		whereValues = make([]interface{}, len(target.Names)-1)
		whereIndex := 1
		for columnIndex, tagName := range target.Names[1:] {
			if val, ok := tags[tagName]; ok {
				whereParts[columnIndex] = utils.QuoteIdent(tagName) + " = $" + strconv.Itoa(whereIndex)
				whereValues[whereIndex-1] = val
			} else {
				whereParts[whereIndex-1] = tagName + " IS NULL"
			}
			whereIndex++
		}
	}

	tagsTableName := c.tagsTableName(measureName)
	tagsTableFullName := utils.FullTableName(c.schema, tagsTableName).Sanitize()
	// SELECT tag_id FROM measure_tag WHERE t1 = v1 AND ... tN = vN
	query := fmt.Sprintf(selectTagIDTemplate, tagsTableFullName, strings.Join(whereParts, " AND "))
	err := c.db.QueryRow(query, whereValues...).Scan(&tagID)
	// tag set found in DB, cache it and return
	if err == nil {
		c.addToCache(measureName, cacheKey, tagID)
		return tagID, nil
	}

	// tag set is new, insert it, and cache the tagID
	query = utils.GenerateInsert(tagsTableFullName, target.Names[1:]) + " RETURNING " + columns.TagIDColumnName
	err = c.db.QueryRow(query, whereValues...).Scan(&tagID)
	if err == nil {
		c.addToCache(measureName, cacheKey, tagID)
	}
	return tagID, err
}

func (c *defTagsCache) tagsTableName(measureName string) string {
	return measureName + c.tagTableSuffix
}

// check the cache for the given 'measure' if it contains the
// tagID value for a given tag-set key. If the cache for that measure
// doesn't exist, creates it.
func (c *defTagsCache) checkTagCache(measure, key string) (int, bool) {
	if cacheForMeasure, ok := c.cache[measure]; ok {
		tagID, exists := cacheForMeasure.Get(key)
		if exists {
			return tagID.(int), exists
		}
		return 0, exists
	}

	c.cache[measure] = lru.New(c.itemsToCache)
	return 0, false
}

func (c *defTagsCache) addToCache(measure, key string, tagID int) {
	c.cache[measure].Add(key, tagID)
}

// cache key is constructed from the tag set as
// {tag_a:1, tag_c:2, tag_b:3}=>'tag_a 1;tag_b 3;tag_c 2;'
func constructCacheKey(tags map[string]string) string {
	numTags := len(tags)
	if numTags == 0 {
		return ""
	}
	keys := make([]string, numTags)
	i := 0
	for key := range tags {
		keys[i] = key
		i++
	}

	sort.Strings(keys)
	var whereParts strings.Builder
	for _, key := range keys {
		val := tags[key]
		whereParts.WriteString(key)
		whereParts.WriteString(" ")
		whereParts.WriteString(val)
		whereParts.WriteString(";")
	}
	return whereParts.String()
}
