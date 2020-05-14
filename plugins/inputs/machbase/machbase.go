package machbase

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type MachDB struct {
	Drivers []string `toml:"drivers"`

	GatherSession bool `toml:"gather_mach_session"`
	GatherStmt    bool `toml:"gather_mach_stmt"`
	GatherSysStat bool `toml:"gather_mach_sysstat"`
	GatherSysTime bool `toml:"gather_mach_systime"`
	GatherStorage bool `toml:"gather_mach_storage"`

	Log telegraf.Logger
}

// queries
const (
	gSysStatQuery = `
        SELECT NAME,VALUE
        FROM V$SYSSTAT 
        ORDER BY NAME
    `
	gSysTimeQuery = `
        SELECT NAME, ACCUM_TICK, AVG_TICK, MIN_TICK, MAX_TICK, COUNT
        FROM V$SYSTIME
        ORDER BY NAME
    `
	gStorageQuery = `
        SELECT 
            DC_TABLE_FILE_SIZE, 
            DC_INDEX_FILE_SIZE, 
            DC_TABLESPACE_DWFILE_SIZE, 
            DC_KV_TABLE_FILE_SIZE 
        FROM V$STORAGE
    `
	gUsageQuery = `
        SELECT TOTAL_SPACE, USED_SPACE, USED_RATIO, RATIO_CAP  
        FROM V$STORAGE_USAGE
    `
	gPageCacheQuery = `
        SELECT MAX_MEM_SIZE, CUR_MEM_SIZE, PAGE_CNT, CHECK_TIME 
        FROM V$STORAGE_DC_PAGECACHE
    `
	gVolatileQuery = `
        SELECT MAX_MEM_SIZE, CUR_MEM_SIZE 
        FROM V$STORAGE_DC_VOLATILE_TABLE
    `
	gSessionQuery = `SELECT * FROM V$SESSION`
	gStmtQuery    = `SELECT * FROM V$STMT`
)

// sampleConfig
const sampleConfig = `
    ## An array of server connect server of the form:
    ## host:port
    ## see http://krdoc.machbase.com/display/MANUAL6/RESTful+API
    ## e.g.
    ##   127.0.0.1:5001,
    ##   192.168.0.232:5003,
    drivers = ["127.0.0.1:5001"]

    ## When true, collect per database session info
    # gather_mach_session = true

    ## When true, collect per database statments
    # gather_mach_stmt = false

    ## When true, collect per database system stats
    # gather_mach_sysstat = false

    ## When true, collect per database system time info
    # gather_mach_systime = false

    ## When true, collect per database storage info
    # gather_mach_storage = false
`

// default value
const (
	defaultDriver = "127.0.0.1:5001"

	gGENERAL = "general"
	gNODE    = "node"
	gCURSOR  = "cursor"
	gFILE    = "file"
	gETC     = "etc"
	gTIME    = "time"
	gSTORAGE = "storage"
	gSESSION = "session"
	gSTAT    = "stat"
)

var gHost = `http://{0}/machbase/`

func (m *MachDB) Description() string {
	return "Read metrics from one or many machbase servers"
}

func (m *MachDB) SampleConfig() string {
	return sampleConfig
}

func (m *MachDB) Gather(acc telegraf.Accumulator) error {
	if len(m.Drivers) == 0 {
		// use default to driver if nothing specified.
		m.Log.Debugf("driver length 0")
		return m.GatherInfo(defaultDriver, acc)
	}

	var wg sync.WaitGroup

	// Loop through each server and collect metrics
	for _, driver := range m.Drivers {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			if err := m.GatherInfo(d, acc); err != nil {
				acc.AddError(err)
			}
		}(driver)
	}
	wg.Wait()

	return nil
}

func (m *MachDB) GatherInfo(aDriver string, acc telegraf.Accumulator) error {
	sHost := aDriver
	var gatherError error

	if m.GatherSysStat {
		gatherError = m.GatherSysStatInfo(sHost, acc)
		if gatherError != nil {
			return gatherError
		}
	}
	if m.GatherSysTime {
		gatherError = m.GatherSysTimeInfo(sHost, acc)
		if gatherError != nil {
			return gatherError
		}
	}
	if m.GatherStorage {
		gatherError = m.GatherSysTimeInfo(sHost, acc)
		if gatherError != nil {
			return gatherError
		}
	}
	if m.GatherSession {
		gatherError = m.GatherSessionInfo(sHost, acc)
		if gatherError != nil {
			return gatherError
		}
	}
	if m.GatherStmt {
		gatherError = m.GatherStmtInfo(sHost, acc)
		if gatherError != nil {
			return gatherError
		}
	}

	return nil
}

func (m *MachDB) GatherSysStatInfo(aHost string, acc telegraf.Accumulator) error {
	var sError error

	sGenaralField := MakeField(gGENERAL)
	sNodeField := MakeField(gNODE)
	sCursorField := MakeField(gCURSOR)
	sFileField := MakeField(gFILE)
	sEtcField := MakeField(gETC)

	sInterface, sError := m.GetData(aHost, gSysStatQuery)
	if sError != nil {
		m.Log.Debugf(sError.Error())
		return sError
	}
	sDatas := reflect.ValueOf(sInterface)

	for i := 0; i < sDatas.Len(); i++ {
		sData := sDatas.Index(i).Interface().(map[string]interface{})
		sName := strings.ToLower(sData["NAME"].(string))
		if _, exist := sGenaralField[sName]; exist {
			sValue, _ := strconv.ParseInt(sData["VALUE"].(string), 10, 64)
			sGenaralField[sName] = sValue
		} else if _, exist := sNodeField[sName]; exist {
			sValue, _ := strconv.ParseInt(sData["VALUE"].(string), 10, 64)
			sNodeField[sName] = sValue
		} else if _, exist := sCursorField[sName]; exist {
			sValue, _ := strconv.ParseInt(sData["VALUE"].(string), 10, 64)
			sCursorField[sName] = sValue
		} else if _, exist := sFileField[sName]; exist {
			sValue, _ := strconv.ParseInt(sData["VALUE"].(string), 10, 64)
			sFileField[sName] = sValue
		} else if _, exist := sEtcField[sName]; exist {
			sValue, _ := strconv.ParseInt(sData["VALUE"].(string), 10, 64)
			sEtcField[sName] = sValue
		} else {
			m.Log.Debugf("no have key : ", sName)
		}
	}

	sTags := map[string]string{"hostname": aHost}
	acc.AddFields("mach_sysstat_general", sGenaralField, sTags)
	acc.AddFields("mach_sysstat_node", sNodeField, sTags)
	acc.AddFields("mach_sysstat_cursor", sCursorField, sTags)
	acc.AddFields("mach_sysstat_file", sFileField, sTags)
	acc.AddFields("mach_sysstat_etc", sEtcField, sTags)

	return nil
}

func (m *MachDB) GatherSysTimeInfo(aHost string, acc telegraf.Accumulator) error {

	sAccumField := MakeField(gTIME)
	sAvgField := MakeField(gTIME)
	sMinField := MakeField(gTIME)
	sMaxField := MakeField(gTIME)
	sCountField := MakeField(gTIME)

	sInterface, sError := m.GetData(aHost, gSysTimeQuery)
	if sError != nil {
		m.Log.Debugf(sError.Error())
		return sError
	}
	sDatas := reflect.ValueOf(sInterface)

	for i := 0; i < sDatas.Len(); i++ {
		sData := sDatas.Index(i).Interface().(map[string]interface{})
		sName := strings.ToLower(sData["NAME"].(string))
		sValue, _ := strconv.ParseInt(sData["ACCUM_TICK"].(string), 10, 64)
		sAccumField[sName] = sValue
		sValue, _ = strconv.ParseInt(sData["AVG_TICK"].(string), 10, 64)
		sAvgField[sName] = sValue
		sValue, _ = strconv.ParseInt(sData["MIN_TICK"].(string), 10, 64)
		sMinField[sName] = sValue
		sValue, _ = strconv.ParseInt(sData["MAX_TICK"].(string), 10, 64)
		sMaxField[sName] = sValue
		sValue, _ = strconv.ParseInt(sData["COUNT"].(string), 10, 64)
		sCountField[sName] = sValue
	}

	sTags := map[string]string{"hostname": aHost}
	sTags["category"] = "systime_accum_tick"
	acc.AddFields("mach_systime", sAccumField, sTags)
	sTags["category"] = "systime_avg_tick"
	acc.AddFields("mach_systime", sAvgField, sTags)
	sTags["category"] = "systime_min_tick"
	acc.AddFields("mach_systime", sMinField, sTags)
	sTags["category"] = "systime_max_tick"
	acc.AddFields("mach_systime", sMaxField, sTags)
	sTags["category"] = "systime_count"
	acc.AddFields("mach_systime", sCountField, sTags)

	return nil
}

func (m *MachDB) GatherStorageInfo(aHost string, acc telegraf.Accumulator) error {

	//v$strorage
	sStorageField := MakeField(gSTORAGE)
	sInterface, sError := m.GetData(aHost, gStorageQuery)
	if sError != nil {
		m.Log.Debugf(sError.Error())
		return sError
	}
	sDatas := reflect.ValueOf(sInterface)

	for i := 0; i < sDatas.Len(); i++ {
		sData := sDatas.Index(i).Interface().(map[string]interface{})
		sValue, _ := strconv.ParseInt(sData["DC_TABLE_FILE_SIZE"].(string), 10, 64)
		sStorageField["storage_dc_table_file_size"] = sValue
		sValue, _ = strconv.ParseInt(sData["DC_INDEX_FILE_SIZE"].(string), 10, 64)
		sStorageField["storage_dc_index_file_size"] = sValue
		sValue, _ = strconv.ParseInt(sData["DC_TABLESPACE_DWFILE_SIZE"].(string), 10, 64)
		sStorageField["storage_dc_tablespace_dwfile_size"] = sValue
		sValue, _ = strconv.ParseInt(sData["DC_KV_TABLE_FILE_SIZE"].(string), 10, 64)
		sStorageField["storage_dc_kv_table_file_size"] = sValue
	}

	//v$storage_usage
	sInterface, sError = m.GetData(aHost, gUsageQuery)
	if sError != nil {
		m.Log.Debugf(sError.Error())
		return sError
	}
	sDatas = reflect.ValueOf(sInterface)

	for i := 0; i < sDatas.Len(); i++ {
		sData := sDatas.Index(i).Interface().(map[string]interface{})
		sValue, _ := strconv.ParseFloat(sData["TOTAL_SPACE"].(string), 64)
		sStorageField["storage_usage_total_space"] = sValue
		sValue, _ = strconv.ParseFloat(sData["USED_SPACE"].(string), 64)
		sStorageField["storage_usage_used_space"] = sValue
		sValue, _ = strconv.ParseFloat(sData["USED_RATIO"].(string), 64)
		sStorageField["storage_usage_used_ratio"] = sValue
		sValue, _ = strconv.ParseFloat(sData["RATIO_CAP"].(string), 64)
		sStorageField["storage_usage_ratio_cap"] = sValue
	}

	//v$storage_dc_pagecache
	sInterface, sError = m.GetData(aHost, gPageCacheQuery)
	if sError != nil {
		m.Log.Debugf(sError.Error())
		return sError
	}
	sDatas = reflect.ValueOf(sInterface)

	for i := 0; i < sDatas.Len(); i++ {
		sData := sDatas.Index(i).Interface().(map[string]interface{})
		sValue, _ := strconv.ParseInt(sData["MAX_MEM_SIZE"].(string), 10, 64)
		sStorageField["storage_pagecache_max_mem_size"] = sValue
		sValue, _ = strconv.ParseInt(sData["CUR_MEM_SIZE"].(string), 10, 64)
		sStorageField["storage_pagecache_cur_mem_size"] = sValue
		sValue, _ = strconv.ParseInt(sData["PAGE_CNT"].(string), 10, 64)
		sStorageField["storage_pagecache_page_cnt"] = sValue
		sStorageField["storage_pagecache_check_time"] = sData["CHECK_TIME"].(string)
	}

	//v$storage_dc_volatile_table
	sInterface, sError = m.GetData(aHost, gVolatileQuery)
	if sError != nil {
		m.Log.Debugf(sError.Error())
		return sError
	}
	sDatas = reflect.ValueOf(sInterface)

	for i := 0; i < sDatas.Len(); i++ {
		sData := sDatas.Index(i).Interface().(map[string]interface{})
		sValue, _ := strconv.ParseInt(sData["MAX_MEM_SIZE"].(string), 10, 64)
		sStorageField["storage_volatile_max_mem_size"] = sValue
		sValue, _ = strconv.ParseInt(sData["CUR_MEM_SIZE"].(string), 10, 64)
		sStorageField["storage_volatile_cur_mem_size"] = sValue

	}

	sTags := map[string]string{"hostname": aHost}
	acc.AddFields("mach_storage", sStorageField, sTags)

	return nil
}

func (m *MachDB) GatherSessionInfo(aHost string, acc telegraf.Accumulator) error {
	sSessionField := MakeField(gSESSION)
	sTags := map[string]string{
		"hostname": aHost,
	}

	sInterface, sError := m.GetData(aHost, gSessionQuery)
	if sError != nil {
		m.Log.Debugf(sError.Error())
		return sError
	}
	sDatas := reflect.ValueOf(sInterface)

	for i := 0; i < sDatas.Len(); i++ {
		sData := sDatas.Index(i).Interface().(map[string]interface{})

		sValue, _ := strconv.ParseInt(sData["ID"].(string), 10, 64)
		sSessionField["session_id"] = sValue

		sValue, _ = strconv.ParseInt(sData["CLOSED"].(string), 10, 64)
		sSessionField["closed"] = sValue

		sValue, _ = strconv.ParseInt(sData["USER_ID"].(string), 10, 64)
		sSessionField["user_id"] = sValue

		sSessionField["login_time"] = sData["LOGIN_TIME"].(string)

		sValue, _ = strconv.ParseInt(sData["SQL_LOGGING"].(string), 10, 64)
		sSessionField["sql_logging"] = sValue

		sValue, _ = strconv.ParseInt(sData["SHOW_HIDDEN_COLS"].(string), 10, 64)
		sSessionField["show_hidden_cols"] = sValue

		sValue, _ = strconv.ParseInt(sData["FEEDBACK_APPEND_ERROR"].(string), 10, 64)
		sSessionField["feedback_append_error"] = sValue

		sSessionField["default_date_format"] = sData["DEFAULT_DATE_FORMAT"].(string)

		sValue, _ = strconv.ParseInt(sData["HASH_BUCKET_SIZE"].(string), 10, 64)
		sSessionField["hash_bucket_size"] = sValue

		sValue, _ = strconv.ParseInt(sData["MAX_QPX_MEM"].(string), 10, 64)
		sSessionField["max_qpx_mem"] = sValue

		sValue, _ = strconv.ParseInt(sData["RS_CACHE_ENABLE"].(string), 10, 64)
		sSessionField["rs_cache_enable"] = sValue

		sValue, _ = strconv.ParseInt(sData["RS_CACHE_TIME_BOUND_MSEC"].(string), 10, 64)
		sSessionField["rs_cache_time_bound_msec"] = sValue

		sValue, _ = strconv.ParseInt(sData["RS_CACHE_MAX_MEMORY_PER_QUERY"].(string), 10, 64)
		sSessionField["rs_cache_max_memory_per_query"] = sValue

		sValue, _ = strconv.ParseInt(sData["RS_CACHE_MAX_RECORD_PER_QUERY"].(string), 10, 64)
		sSessionField["rs_cache_max_record_per_query"] = sValue

		sValue, _ = strconv.ParseInt(sData["RS_CACHE_APPROXIMATE_RESULT_ENABLE"].(string), 10, 64)
		sSessionField["rs_cache_approximate_result_enable"] = sValue

		acc.AddFields("mach_session", sSessionField, sTags)
	}

	return nil
}

func (m *MachDB) GatherStmtInfo(aHost string, acc telegraf.Accumulator) error {

	sStatField := MakeField(gSTAT)
	sTags := map[string]string{
		"hostname": aHost,
	}

	sInterface, sError := m.GetData(aHost, gStmtQuery)
	if sError != nil {
		m.Log.Debugf(sError.Error())
		return sError
	}
	sDatas := reflect.ValueOf(sInterface)

	for i := 0; i < sDatas.Len(); i++ {
		sData := sDatas.Index(i).Interface().(map[string]interface{})
		sValue, _ := strconv.ParseInt(sData["ID"].(string), 10, 64)
		sStatField["stmt_id"] = sValue
		sValue, _ = strconv.ParseInt(sData["SESS_ID"].(string), 10, 64)
		sStatField["sess_id"] = sValue
		sStatField["state"] = sData["STATE"].(string)
		sValue, _ = strconv.ParseInt(sData["RECORD_SIZE"].(string), 10, 64)
		sStatField["record_size"] = sValue
		sStatField["query"] = sData["QUERY"].(string)

		acc.AddFields("mach_stmt", sStatField, sTags)

	}

	return nil
}

func (m *MachDB) GetData(aUrl string, aQuery string) (interface{}, error) {
	var sError error
	var sData interface{}
	sUrl := MakeUrl(aUrl)

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("q", aQuery)
	sError = writer.Close()
	if sError != nil {
		m.Log.Debugf("Write Field error" + sError.Error())
	} else {
		client := &http.Client{}
		req, sError := http.NewRequest("GET", sUrl, payload)
		if sError != nil {
			m.Log.Debugf("Write Field error" + sError.Error())
		} else {
			req.Header.Set("Content-Type", writer.FormDataContentType())
			res, err := client.Do(req)
			if err != nil {
				m.Log.Debugf("client.Do error" + err.Error())
				return sData, err
			}
			defer res.Body.Close()

			sBytes, _ := ioutil.ReadAll(res.Body)
			sConvertData := "[" + string(sBytes) + "]"
			sError = json.Unmarshal([]byte(sConvertData), &sData)
			if sError != nil {
				m.Log.Debugf("json unmarshal error > " + sError.Error())
			}
		}
	}
	return sData, sError
}

func MakeUrl(aHost string) string {
	return strings.Replace(gHost, "{0}", aHost, -1)
}

func MakeField(aName string) map[string]interface{} {
	sName := aName
	sField := map[string]interface{}{}
	switch sName {
	case gGENERAL:
		sField = map[string]interface{}{
			"connect_cnt":            int64(0),
			"disconnect_cnt":         int64(0),
			"prepare_success":        int64(0),
			"prepare_failure":        int64(0),
			"execute_success":        int64(0),
			"execute_failure":        int64(0),
			"cursor_open_cnt":        int64(0),
			"cursor_fetch_cnt":       int64(0),
			"cursor_close_cnt":       int64(0),
			"append_open":            int64(0),
			"append_data_success":    int64(0),
			"append_data_failure":    int64(0),
			"append_data_decompress": int64(0),
			"append_close":           int64(0),
		}
	case gNODE:
		sField = map[string]interface{}{
			"keyvalue_scan_node_open":        int64(0),
			"keyvalue_scan_node_fetch":       int64(0),
			"keyvalue_scan_node_close":       int64(0),
			"scan_node_open":                 int64(0),
			"scan_node_fetch":                int64(0),
			"scan_node_close":                int64(0),
			"lookup_scan_node_open":          int64(0),
			"lookup_scan_node_fetch":         int64(0),
			"lookup_scan_node_close":         int64(0),
			"volatile_scan_node_open":        int64(0),
			"volatile_scan_node_fetch":       int64(0),
			"volatile_scan_node_close":       int64(0),
			"index_lookup_scan_node_open":    int64(0),
			"index_lookup_scan_node_fetch":   int64(0),
			"index_lookup_scan_node_close":   int64(0),
			"index_volatile_scan_node_open":  int64(0),
			"index_volatile_scan_node_fetch": int64(0),
			"index_volatile_scan_node_close": int64(0),
			"union_all_node_open":            int64(0),
			"union_all_node_fetch":           int64(0),
			"union_all_node_close":           int64(0),
			"minmax_count_node_open":         int64(0),
			"minmax_count_node_fetch":        int64(0),
			"minmax_count_node_close":        int64(0),
			"limit_node_open":                int64(0),
			"limit_node_fetch":               int64(0),
			"limit_node_close":               int64(0),
			"limit_sort_node_open":           int64(0),
			"limit_sort_node_fetch":          int64(0),
			"limit_sort_node_close":          int64(0),
			"rid_scan_node_open":             int64(0),
			"rid_scan_node_fetch":            int64(0),
			"rid_scan_node_close":            int64(0),
			"proj_node_open":                 int64(0),
			"proj_node_fetch":                int64(0),
			"proj_node_close":                int64(0),
			"grag_node_open":                 int64(0),
			"grag_node_fetch":                int64(0),
			"grag_node_close":                int64(0),
			"sort_node_open":                 int64(0),
			"sort_node_fetch":                int64(0),
			"sort_node_close":                int64(0),
			"join_node_open":                 int64(0),
			"join_node_fetch":                int64(0),
			"join_node_close":                int64(0),
			"outerjoin_node_open":            int64(0),
			"outerjoin_node_fetch":           int64(0),
			"outerjoin_node_close":           int64(0),
			"cstar_time_node_open":           int64(0),
			"cstar_time_node_fetch":          int64(0),
			"cstar_time_node_close":          int64(0),
			"cstar_node_open":                int64(0),
			"cstar_node_fetch":               int64(0),
			"cstar_node_close":               int64(0),
			"having_node_open":               int64(0),
			"having_node_fetch":              int64(0),
			"having_node_close":              int64(0),
			"inlineview_node_open":           int64(0),
			"inlineview_node_fetch":          int64(0),
			"inlineview_node_close":          int64(0),
			"series_by_node_open":            int64(0),
			"series_by_node_fetch":           int64(0),
			"series_by_node_close":           int64(0),
			"rownum_node_open":               int64(0),
			"rownum_node_fetch":              int64(0),
			"rownum_node_close":              int64(0),
			"px_node_open":                   int64(0),
			"px_node_fetch":                  int64(0),
			"px_node_close":                  int64(0),
			"bitmap_aggr_node_open":          int64(0),
			"bitmap_aggr_node_fetch":         int64(0),
			"bitmap_aggr_node_close":         int64(0),
			"bitmap_grby_node_open":          int64(0),
			"bitmap_grby_node_fetch":         int64(0),
			"bitmap_grby_node_close":         int64(0),
			"bitmap_sort_node_open":          int64(0),
			"bitmap_sort_node_fetch":         int64(0),
			"bitmap_sort_node_close":         int64(0),
			"tag_read_node_open":             int64(0),
			"tag_read_node_fetch":            int64(0),
			"tag_read_node_close":            int64(0),
			"pivot_grby_node_open":           int64(0),
			"pivot_grby_node_fetch":          int64(0),
			"pivot_grby_node_close":          int64(0),
		}
	case gCURSOR:
		sField = map[string]interface{}{
			"noindex_cursor_open":          int64(0),
			"noindex_cursor_fetch":         int64(0),
			"noindex_cursor_close":         int64(0),
			"bitmap_cursor_open":           int64(0),
			"bitmap_cursor_fetch":          int64(0),
			"bitmap_cursor_close":          int64(0),
			"bitmap_cursor_window_copy":    int64(0),
			"bitmap_cursor_window_set_and": int64(0),
			"bitmap_cursor_window_set_or":  int64(0),
			"bitmap_cursor_window_set_xor": int64(0),
			"bitmap_cursor_window_get_and": int64(0),
			"bitmap_cursor_window_get_or":  int64(0),
			"bitmap_cursor_window_get_xor": int64(0),
			"bitmap_cursor_window_skip":    int64(0),
		}
	case gFILE:
		sField = map[string]interface{}{
			"file_create":     int64(0),
			"file_open":       int64(0),
			"file_close":      int64(0),
			"file_seek":       int64(0),
			"file_delete":     int64(0),
			"file_rename":     int64(0),
			"file_truncate":   int64(0),
			"file_read_cnt":   int64(0),
			"file_read_size":  int64(0),
			"file_write_cnt":  int64(0),
			"file_write_size": int64(0),
			"file_sync":       int64(0),
			"file_sync_data":  int64(0),
		}
	case gETC:
		sField = map[string]interface{}{
			"mtr_hash_create_cnt":        int64(0),
			"mtr_hash_destroy_cnt":       int64(0),
			"mtr_hash_add_cnt":           int64(0),
			"mtr_hash_find_conflict":     int64(0),
			"text_lexer_open":            int64(0),
			"text_lexer_parse":           int64(0),
			"text_lexer_close":           int64(0),
			"minmax_cache_hit":           int64(0),
			"minmax_cache_miss":          int64(0),
			"page_cache_miss":            int64(0),
			"page_cache_hit":             int64(0),
			"keyvalue_cache_miss":        int64(0),
			"keyvalue_cache_hit":         int64(0),
			"keyvalue_cache_iowait":      int64(0),
			"keyvalue_cache_flush":       int64(0),
			"keyvalue_mem_index_search":  int64(0),
			"minmax_part_pruning":        int64(0),
			"minmax_part_contain":        int64(0),
			"bloom_filter_part_pruning":  int64(0),
			"lsmindex_level0_read_count": int64(0),
			"lsmindex_level1_read_count": int64(0),
			"lsmindex_level2_read_count": int64(0),
			"lsmindex_level3_read_count": int64(0),
			"comm_io_send_cnt":           int64(0),
			"comm_io_recv_cnt":           int64(0),
			"comm_io_send_size":          int64(0),
			"comm_io_recv_size":          int64(0),
			"accept_success":             int64(0),
			"accept_failure":             int64(0),
		}
	case gTIME:
		sField = map[string]interface{}{
			"append":                          int64(0),
			"prepare":                         int64(0),
			"execute":                         int64(0),
			"fetch_ready":                     int64(0),
			"fetch":                           int64(0),
			"file_create":                     int64(0),
			"file_open":                       int64(0),
			"file_close":                      int64(0),
			"file_seek":                       int64(0),
			"file_delete":                     int64(0),
			"file_rename":                     int64(0),
			"file_truncate":                   int64(0),
			"file_read":                       int64(0),
			"file_write":                      int64(0),
			"file_sync":                       int64(0),
			"file_sync_data":                  int64(0),
			"table_time_range":                int64(0),
			"table_part_access":               int64(0),
			"table_part_pruning":              int64(0),
			"table_part_fetch_page":           int64(0),
			"table_part_fetch_value":          int64(0),
			"table_part_filter_value":         int64(0),
			"table_part_file_open":            int64(0),
			"table_part_file_close":           int64(0),
			"table_part_file_rd_buff":         int64(0),
			"table_part_file_rd_buff_sz":      int64(0),
			"table_part_file_rd_disk":         int64(0),
			"table_part_file_rd_disk_sz":      int64(0),
			"index_wait":                      int64(0),
			"index_mem_search":                int64(0),
			"index_mem_read":                  int64(0),
			"index_mem_read_sz":               int64(0),
			"index_part_access":               int64(0),
			"index_part_pruning":              int64(0),
			"index_part_file_open":            int64(0),
			"index_part_file_close":           int64(0),
			"index_part_file_rd_buff":         int64(0),
			"index_part_file_rd_buff_sz":      int64(0),
			"index_part_file_rd_cache":        int64(0),
			"index_part_file_rd_cache_sz":     int64(0),
			"index_part_file_rd_disk":         int64(0),
			"index_part_file_rd_disk_sz":      int64(0),
			"data_compress":                   int64(0),
			"data_compress_size":              int64(0),
			"data_decompress":                 int64(0),
			"data_decompress_size":            int64(0),
			"bf_part_file_open":               int64(0),
			"bf_part_file_close":              int64(0),
			"bf_part_file_read_buffer":        int64(0),
			"bf_part_file_read_buffer_size":   int64(0),
			"bf_part_file_read_disk":          int64(0),
			"bf_part_file_read_disk_size":     int64(0),
			"bitmapindex_part_bitvector_skip": int64(0),
		}
	case gSTORAGE:
		sField = map[string]interface{}{
			"storage_dc_table_file_size":        int64(0),
			"storage_dc_index_file_size":        int64(0),
			"storage_dc_tablespace_dwfile_size": int64(0),
			"storage_dc_kv_table_file_size":     int64(0),
			"storage_usage_total_space":         float64(0.0),
			"storage_usage_used_space":          float64(0.0),
			"storage_usage_used_ratio":          float64(0.0),
			"storage_usage_ratio_cap":           float64(0.0),
			"storage_pagecache_max_mem_size":    int64(0),
			"storage_pagecache_cur_mem_size":    int64(0),
			"storage_pagecache_page_cnt":        int64(0),
			"storage_pagecache_check_time":      string(""), //datetime
			"storage_volatile_max_mem_size":     int64(0),
			"storage_volatile_cur_mem_size":     int64(0),
		}
	case gSESSION:
		sField = map[string]interface{}{
			"session_id":                         int64(0),
			"closed":                             int64(0),
			"user_id":                            int64(0),
			"login_time":                         string(""), //datetime
			"sql_logging":                        int64(0),
			"show_hidden_cols":                   int64(0),
			"feedback_append_error":              int64(0),
			"default_date_format":                string(""),
			"hash_bucket_size":                   int64(0),
			"max_qpx_mem":                        int64(0),
			"rs_cache_enable":                    int64(0),
			"rs_cache_time_bound_msec":           int64(0),
			"rs_cache_max_memory_per_query":      int64(0),
			"rs_cache_max_record_per_query":      int64(0),
			"rs_cache_approximate_result_enable": int64(0),
		}
	case gSTAT:
		sField = map[string]interface{}{
			"stmt_id":     int64(0),
			"sess_id":     int64(0),
			"state":       string(""),
			"record_size": int64(0),
			"query":       string(""),
		}
	}

	return sField
}

func init() {
	inputs.Add("machbase", func() telegraf.Input { return &MachDB{} })
}
