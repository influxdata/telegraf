## MACHBASE Input Plugin

This plugin collects statistical data from the MACHBASE server.

If this plugin is to be used, MWA must be enabled.

For more information on MWA, see http://krdoc.machbase.com/pages/viewpage.action?pageId=13436304


### Configuration

```toml
[[inputs.machbase]]
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
```

### Metrics:

* mach_session - Session information.
    * tags:
        * hostname
    * fields:
        * session_id (integer)
        * closed (integer)
        * user_id (integer)
        * login_time (string)
        * sql_logging (integer)
        * show_hidden_cols (integer)
        * feedback_append_error (integer)
        * default_date_format (string)
        * hash_bucket_size (integer)
        * max_qpx_mem (integer)
        * rs_cache_enable (integer)
        * rs_cache_time_bound_msec (integer)
        * rs_cache_max_memory_per_query (integer)
        * rs_cache_max_record_per_query (integer)
        * rs_cache_approximate_result_enable (integer)
* mach_stmt - Running query statement information.
    * tags:
        * hostname
    * fields:
        * stmt_id (integer)
        * sess_id (integer)
        * state (string)
        * record_size (integer)
        * query (string)
* mach_sysstat_general - General system statistics information.
    * tags:
        * hostname
    * fields:
        * connect_cnt (integer)
        * disconnect_cnt (integer)
        * prepare_success (integer)
        * prepare_failure (integer)
        * execute_success (integer)
        * execute_failure (integer)
        * cursor_open_cnt (integer)
        * cursor_fetch_cnt (integer)
        * cursor_close_cnt (integer)
        * append_open (integer)
        * append_data_success (integer)
        * append_data_failure (integer)
        * append_data_decompress (integer)
        * append_close (integer)
* mach_sysstat_node - Node information of system statistics.
    * tags:
        * hostname
    * fields:
        * keyvalue_scan_node_open (integer)
        * keyvalue_scan_node_fetch (integer)
        * keyvalue_scan_node_close (integer)
        * scan_node_open (integer)
        * scan_node_fetch (integer)
        * scan_node_close (integer)
        * lookup_scan_node_open (integer)
        * lookup_scan_node_fetch (integer)
        * lookup_scan_node_close (integer)
        * volatile_scan_node_open (integer)
        * volatile_scan_node_fetch (integer)
        * volatile_scan_node_close (integer)
        * index_lookup_scan_node_open (integer)
        * index_lookup_scan_node_fetch (integer)
        * index_lookup_scan_node_close (integer)
        * index_volatile_scan_node_open (integer)
        * index_volatile_scan_node_fetch (integer)
        * index_volatile_scan_node_close (integer)
        * union_all_node_open (integer)
        * union_all_node_fetch (integer)
        * union_all_node_close (integer)
        * minmax_count_node_open (integer)
        * minmax_count_node_fetch (integer)
        * minmax_count_node_close (integer)
        * limit_node_open (integer)
        * limit_node_fetch (integer)
        * limit_node_close (integer)
        * limit_sort_node_open (integer)
        * limit_sort_node_fetch (integer)
        * limit_sort_node_close (integer)
        * rid_scan_node_open (integer)
        * rid_scan_node_fetch (integer)
        * rid_scan_node_close (integer)
        * proj_node_open (integer)
        * proj_node_fetch (integer)
        * proj_node_close (integer)
        * grag_node_open (integer)
        * grag_node_fetch (integer)
        * grag_node_close (integer)
        * sort_node_open (integer)
        * sort_node_fetch (integer)
        * sort_node_close (integer)
        * join_node_open (integer)
        * join_node_fetch (integer)
        * join_node_close (integer)
        * outerjoin_node_open (integer)
        * outerjoin_node_fetch (integer)
        * outerjoin_node_close (integer)
        * cstar_time_node_open (integer)
        * cstar_time_node_fetch (integer)
        * cstar_time_node_close (integer)
        * cstar_node_open (integer)
        * cstar_node_fetch (integer)
        * cstar_node_close (integer)
        * having_node_open (integer)
        * having_node_fetch (integer)
        * having_node_close (integer)
        * inlineview_node_open (integer)
        * inlineview_node_fetch (integer)
        * inlineview_node_close (integer)
        * series_by_node_open (integer)
        * series_by_node_fetch (integer)
        * series_by_node_close (integer)
        * rownum_node_open (integer)
        * rownum_node_fetch (integer)
        * rownum_node_close (integer)
        * px_node_open (integer)
        * px_node_fetch (integer)
        * px_node_close (integer)
        * bitmap_aggr_node_open (integer)
        * bitmap_aggr_node_fetch (integer)
        * bitmap_aggr_node_close (integer)
        * bitmap_grby_node_open (integer)
        * bitmap_grby_node_fetch (integer)
        * bitmap_grby_node_close (integer)
        * bitmap_sort_node_open (integer)
        * bitmap_sort_node_fetch (integer)
        * bitmap_sort_node_close (integer)
        * tag_read_node_open (integer)
        * tag_read_node_fetch (integer)
        * tag_read_node_close (integer)
        * pivot_grby_node_open (integer)
        * pivot_grby_node_fetch (integer)
        * pivot_grby_node_close (integer)
* mach_sysstat_node - Cursor information of system statistics.
    * tags:
        * hostname
    * fields:
        * noindex_cursor_open (integer)
        * noindex_cursor_fetch (integer)
        * noindex_cursor_close (integer)
        * bitmap_cursor_open (integer)
        * bitmap_cursor_fetch (integer)
        * bitmap_cursor_close (integer)
        * bitmap_cursor_window_copy (integer)
        * bitmap_cursor_window_set_and (integer)
        * bitmap_cursor_window_set_or (integer)
        * bitmap_cursor_window_set_xor (integer)
        * bitmap_cursor_window_get_and (integer)
        * bitmap_cursor_window_get_or (integer)
        * bitmap_cursor_window_get_xor (integer)
        * bitmap_cursor_window_skip (integer)
* mach_sysstat_file - File information of system statistics.
    * tags:
        * hostname
    * fields:
        * file_create (integer)
        * file_open (integer)
        * file_close (integer)
        * file_seek (integer)
        * file_delete (integer)
        * file_rename (integer)
        * file_truncate (integer)
        * file_read_cnt (integer)
        * file_read_size (integer)
        * file_write_cnt (integer)
        * file_write_size (integer)
        * file_sync (integer)
        * file_sync_data (integer)
* mach_sysstat_etc - Remaining information of system statistic.
    * tags:
        * hostname
    * fields:
        * mtr_hash_create_cnt (integer)
        * mtr_hash_destroy_cnt (integer)
        * mtr_hash_add_cnt (integer)
        * mtr_hash_find_conflict (integer)
        * text_lexer_open (integer)
        * text_lexer_parse (integer)
        * text_lexer_close (integer)
        * minmax_cache_hit (integer)
        * minmax_cache_miss (integer)
        * page_cache_miss (integer)
        * page_cache_hit (integer)
        * keyvalue_cache_miss (integer)
        * keyvalue_cache_hit (integer)
        * keyvalue_cache_iowait (integer)
        * keyvalue_cache_flush (integer)
        * keyvalue_mem_index_search (integer)
        * minmax_part_pruning (integer)
        * minmax_part_contain (integer)
        * bloom_filter_part_pruning (integer)
        * lsmindex_level0_read_count (integer)
        * lsmindex_level1_read_count (integer)
        * lsmindex_level2_read_count (integer)
        * lsmindex_level3_read_count (integer)
        * comm_io_send_cnt (integer)
        * comm_io_recv_cnt (integer)
        * comm_io_send_size (integer)
        * comm_io_recv_size (integer)
        * accept_success (integer)
        * accept_failure (integer)
* mach_systime - System time information. Category is composed of accumulative time, (unit of action) average time, (unit of action) minimum time, (unit of action) maximum time, number of times to action.
    * tags:
        * hostname
        * category 
    * fields:
        * append (integer)
        * prepare (integer)
        * execute (integer)
        * fetch_ready (integer)
        * fetch (integer)
        * file_create (integer)
        * file_open (integer)
        * file_close (integer)
        * file_seek (integer)
        * file_delete (integer)
        * file_rename (integer)
        * file_truncate (integer)
        * file_read (integer)
        * file_write (integer)
        * file_sync (integer)
        * file_sync_data (integer)
        * table_time_range (integer)
        * table_part_access (integer)
        * table_part_pruning (integer)
        * table_part_fetch_page (integer)
        * table_part_fetch_value (integer)
        * table_part_filter_value (integer)
        * table_part_file_open (integer)
        * table_part_file_close (integer)
        * table_part_file_rd_buff (integer)
        * table_part_file_rd_buff_sz (integer)
        * table_part_file_rd_disk (integer)
        * table_part_file_rd_disk_sz (integer)
        * index_wait (integer)
        * index_mem_search (integer)
        * index_mem_read (integer)
        * index_mem_read_sz (integer)
        * index_part_access (integer)
        * index_part_pruning (integer)
        * index_part_file_open (integer)
        * index_part_file_close (integer)
        * index_part_file_rd_buff (integer)
        * index_part_file_rd_buff_sz (integer)
        * index_part_file_rd_cache (integer)
        * index_part_file_rd_cache_sz (integer)
        * index_part_file_rd_disk (integer)
        * index_part_file_rd_disk_sz (integer)
        * data_compress (integer)
        * data_compress_size (integer)
        * data_decompress (integer)
        * data_decompress_size (integer)
        * bf_part_file_open (integer)
        * bf_part_file_close (integer)
        * bf_part_file_read_buffer (integer)
        * bf_part_file_read_buffer_size (integer)
        * bf_part_file_read_disk (integer)
        * bf_part_file_read_disk_size (integer)
        * bitmapindex_part_bitvector_skip (integer)
* mach_storage - Storage system internal information.
    * tags:
        * hostname
    * fields:
        * storage_dc_table_file_size (integer)
        * storage_dc_index_file_size (integer)
        * storage_dc_tablespace_dwfile_size (integer)
        * storage_dc_kv_table_file_size (integer)
        * storage_usage_total_space (float)
        * storage_usage_used_space (float)
        * storage_usage_used_ratio (float)
        * storage_usage_ratio_cap (float)
        * storage_pagecache_max_mem_size (integer)
        * storage_pagecache_cur_mem_size (integer)
        * storage_pagecache_page_cnt (integer)
        * storage_pagecache_check_time (string)
        * storage_volatile_max_mem_size (integer)
        * storage_volatile_cur_mem_size (integer)