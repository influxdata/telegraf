package machbase

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
)

// Skipping integration test in short mode
func TestGatherInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := &MachDB{
		Drivers:       []string{"127.0.0.1:5001"},
		GatherSession: true,
		GatherStmt:    false,
		GatherSysStat: false,
		GatherSysTime: false,
		GatherStorage: false,
	}

	var acc testutil.Accumulator
	err := m.GatherInfo("127.0.0.1:5001", &acc)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("mach_session"))
}

func TestMakeUrl(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			"192.168.0.232:5003",
			"http://192.168.0.232:5003/machbase/",
		},
		{
			"127.0.0.1:5001",
			"http://127.0.0.1:5001/machbase/",
		},
		{
			"192.168.0.1:20001",
			"http://192.168.0.1:20001/machbase/",
		},
	}

	for _, test := range tests {
		out := MakeUrl(test.input)
		if out != test.output {
			t.Errorf("Expected %s, got %s\n", test.output, out)
		}
	}
}

func TestMakeField(t *testing.T) {
	tests := []struct {
		input  string
		output []string
	}{
		{
			gGENERAL,
			[]string{
				"connect_cnt",
				"disconnect_cnt",
				"prepare_success",
				"prepare_failure",
				"execute_success",
				"execute_failure",
				"cursor_open_cnt",
				"cursor_fetch_cnt",
				"cursor_close_cnt",
				"append_open",
				"append_data_success",
				"append_data_failure",
				"append_data_decompress",
				"append_close",
			},
		},
		{
			gNODE,
			[]string{
				"keyvalue_scan_node_open",
				"keyvalue_scan_node_fetch",
				"keyvalue_scan_node_close",
				"scan_node_open",
				"scan_node_fetch",
				"scan_node_close",
				"lookup_scan_node_open",
				"lookup_scan_node_fetch",
				"lookup_scan_node_close",
				"volatile_scan_node_open",
				"volatile_scan_node_fetch",
				"volatile_scan_node_close",
				"index_lookup_scan_node_open",
				"index_lookup_scan_node_fetch",
				"index_lookup_scan_node_close",
				"index_volatile_scan_node_open",
				"index_volatile_scan_node_fetch",
				"index_volatile_scan_node_close",
				"union_all_node_open",
				"union_all_node_fetch",
				"union_all_node_close",
				"minmax_count_node_open",
				"minmax_count_node_fetch",
				"minmax_count_node_close",
				"limit_node_open",
				"limit_node_fetch",
				"limit_node_close",
				"limit_sort_node_open",
				"limit_sort_node_fetch",
				"limit_sort_node_close",
				"rid_scan_node_open",
				"rid_scan_node_fetch",
				"rid_scan_node_close",
				"proj_node_open",
				"proj_node_fetch",
				"proj_node_close",
				"grag_node_open",
				"grag_node_fetch",
				"grag_node_close",
				"sort_node_open",
				"sort_node_fetch",
				"sort_node_close",
				"join_node_open",
				"join_node_fetch",
				"join_node_close",
				"outerjoin_node_open",
				"outerjoin_node_fetch",
				"outerjoin_node_close",
				"cstar_time_node_open",
				"cstar_time_node_fetch",
				"cstar_time_node_close",
				"cstar_node_open",
				"cstar_node_fetch",
				"cstar_node_close",
				"having_node_open",
				"having_node_fetch",
				"having_node_close",
				"inlineview_node_open",
				"inlineview_node_fetch",
				"inlineview_node_close",
				"series_by_node_open",
				"series_by_node_fetch",
				"series_by_node_close",
				"rownum_node_open",
				"rownum_node_fetch",
				"rownum_node_close",
				"px_node_open",
				"px_node_fetch",
				"px_node_close",
				"bitmap_aggr_node_open",
				"bitmap_aggr_node_fetch",
				"bitmap_aggr_node_close",
				"bitmap_grby_node_open",
				"bitmap_grby_node_fetch",
				"bitmap_grby_node_close",
				"bitmap_sort_node_open",
				"bitmap_sort_node_fetch",
				"bitmap_sort_node_close",
				"tag_read_node_open",
				"tag_read_node_fetch",
				"tag_read_node_close",
				"pivot_grby_node_open",
				"pivot_grby_node_fetch",
				"pivot_grby_node_close",
			},
		},
		{
			gCURSOR,
			[]string{
				"noindex_cursor_open",
				"noindex_cursor_fetch",
				"noindex_cursor_close",
				"bitmap_cursor_open",
				"bitmap_cursor_fetch",
				"bitmap_cursor_close",
				"bitmap_cursor_window_copy",
				"bitmap_cursor_window_set_and",
				"bitmap_cursor_window_set_or",
				"bitmap_cursor_window_set_xor",
				"bitmap_cursor_window_get_and",
				"bitmap_cursor_window_get_or",
				"bitmap_cursor_window_get_xor",
				"bitmap_cursor_window_skip",
			},
		},
		{
			gFILE,
			[]string{
				"file_create",
				"file_open",
				"file_close",
				"file_seek",
				"file_delete",
				"file_rename",
				"file_truncate",
				"file_read_cnt",
				"file_read_size",
				"file_write_cnt",
				"file_write_size",
				"file_sync",
				"file_sync_data",
			},
		},
		{
			gETC,
			[]string{
				"mtr_hash_create_cnt",
				"mtr_hash_destroy_cnt",
				"mtr_hash_add_cnt",
				"mtr_hash_find_conflict",
				"text_lexer_open",
				"text_lexer_parse",
				"text_lexer_close",
				"minmax_cache_hit",
				"minmax_cache_miss",
				"page_cache_miss",
				"page_cache_hit",
				"keyvalue_cache_miss",
				"keyvalue_cache_hit",
				"keyvalue_cache_iowait",
				"keyvalue_cache_flush",
				"keyvalue_mem_index_search",
				"minmax_part_pruning",
				"minmax_part_contain",
				"bloom_filter_part_pruning",
				"lsmindex_level0_read_count",
				"lsmindex_level1_read_count",
				"lsmindex_level2_read_count",
				"lsmindex_level3_read_count",
				"comm_io_send_cnt",
				"comm_io_recv_cnt",
				"comm_io_send_size",
				"comm_io_recv_size",
				"accept_success",
				"accept_failure",
			},
		},
		{
			gTIME,
			[]string{
				"append",
				"prepare",
				"execute",
				"fetch_ready",
				"fetch",
				"file_create",
				"file_open",
				"file_close",
				"file_seek",
				"file_delete",
				"file_rename",
				"file_truncate",
				"file_read",
				"file_write",
				"file_sync",
				"file_sync_data",
				"table_time_range",
				"table_part_access",
				"table_part_pruning",
				"table_part_fetch_page",
				"table_part_fetch_value",
				"table_part_filter_value",
				"table_part_file_open",
				"table_part_file_close",
				"table_part_file_rd_buff",
				"table_part_file_rd_buff_sz",
				"table_part_file_rd_disk",
				"table_part_file_rd_disk_sz",
				"index_wait",
				"index_mem_search",
				"index_mem_read",
				"index_mem_read_sz",
				"index_part_access",
				"index_part_pruning",
				"index_part_file_open",
				"index_part_file_close",
				"index_part_file_rd_buff",
				"index_part_file_rd_buff_sz",
				"index_part_file_rd_cache",
				"index_part_file_rd_cache_sz",
				"index_part_file_rd_disk",
				"index_part_file_rd_disk_sz",
				"data_compress",
				"data_compress_size",
				"data_decompress",
				"data_decompress_size",
				"bf_part_file_open",
				"bf_part_file_close",
				"bf_part_file_read_buffer",
				"bf_part_file_read_buffer_size",
				"bf_part_file_read_disk",
				"bf_part_file_read_disk_size",
				"bitmapindex_part_bitvector_skip",
			},
		},
		{
			gSTORAGE,
			[]string{
				"storage_dc_table_file_size",
				"storage_dc_index_file_size",
				"storage_dc_tablespace_dwfile_size",
				"storage_dc_kv_table_file_size",
				"storage_usage_total_space",
				"storage_usage_used_space",
				"storage_usage_used_ratio",
				"storage_usage_ratio_cap",
				"storage_pagecache_max_mem_size",
				"storage_pagecache_cur_mem_size",
				"storage_pagecache_page_cnt",
				"storage_pagecache_check_time",
				"storage_volatile_max_mem_size",
				"storage_volatile_cur_mem_size",
			},
		},
		{
			gSESSION,
			[]string{
				"closed",
				"user_id",
				"login_time",
				"sql_logging",
				"show_hidden_cols",
				"feedback_append_error",
				"default_date_format",
				"hash_bucket_size",
				"max_qpx_mem",
				"rs_cache_enable",
				"rs_cache_time_bound_msec",
				"rs_cache_max_memory_per_query",
				"rs_cache_max_record_per_query",
				"rs_cache_approximate_result_enable",
			},
		},
		{
			gSTAT,
			[]string{
				"sess_id",
				"state",
				"record_size",
				"query",
			},
		},
	}

	for _, test := range tests {
		out := MakeField(test.input)
		for _, keys := range test.output {
			if _, exist := out[keys]; !exist {
				t.Errorf("Expected %s, got %s\n", test.output, out)
				break
			}
		}
	}
}
