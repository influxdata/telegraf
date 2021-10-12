# Oracle Input Plugin

The `oracle` plugin collects metrics from Oracle RDBMS using Dynamic Performance Views. It requires proper installation of
[python3](https://www.python.org/downloads/) with [cx_Oracle](https://cx-oracle.readthedocs.io/en/latest/user_guide/installation.html).

### Configuration:

```toml
[[inputs.oracle]]
  ## Database user with SELECT_CATALOG_ROLE role granted, required.
	username = system
	password = oracle
  ## Database SID, required.
	sid = XE

  ## python executable, python3 by default 
  python=python3

  ## Timeout for metrics collector to complete, 5s by default.
  timeout = "5s"
```

It is recommended to setup related environment variables of Oracle client libraries before running telegraf. It might be also handy 
to setup these variables in the telegraf configuration, as shown in [an example](./dev/telegraf.conf).

### Metrics

```
$ telegraf --config ./dev/telegraf.conf --input-filter oracle --test
> oracle_wait_class,instance=XE,wait_class=CPU wait_value=0.015 1634036614000000000
> oracle_wait_class,instance=XE,wait_class=CPU_OS wait_value=0.522 1634036614000000000
> oracle_wait_class,instance=XE,wait_class=Commit wait_value=0 1634036614000000000
> oracle_wait_class,instance=XE,wait_class=Concurrency wait_value=0 1634036614000000000
> oracle_wait_class,instance=XE,wait_class=Configuration wait_value=0 1634036614000000000
> oracle_wait_class,instance=XE,wait_class=Network wait_value=0 1634036614000000000
> oracle_wait_class,instance=XE,wait_class=Other wait_value=0 1634036614000000000
> oracle_wait_class,instance=XE,wait_class=Scheduler wait_value=0 1634036614000000000
> oracle_wait_class,instance=XE,wait_class=System_I/O wait_value=0.002 1634036614000000000
> oracle_wait_class,instance=XE,wait_class=User_I/O wait_value=0 1634036614000000000
> oracle_wait_event,instance=XE,wait_class=Network,wait_event=SQL*Net_message_to_client count=17,latency=0.002 1634036614000000000
> oracle_wait_event,instance=XE,wait_class=Other,wait_event=asynch_descriptor_resize count=11,latency=0.005 1634036614000000000
> oracle_wait_event,instance=XE,wait_class=System_I/O,wait_event=log_file_parallel_write count=1,latency=0.917 1634036614000000000
> oracle_wait_event,instance=XE,wait_class=System_I/O,wait_event=control_file_sequential_read count=129,latency=0.014 1634036614000000000
> oracle_wait_event,instance=XE,wait_class=System_I/O,wait_event=control_file_parallel_write count=27,latency=5.364 1634036614000000000
> oracle_wait_event,instance=XE,wait_class=User_I/O,wait_event=Disk_file_operations_I/O count=5,latency=0.035 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Buffer_Cache_Hit_Ratio metric_value=100 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Memory_Sorts_Ratio metric_value=100 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Redo_Allocation_Hit_Ratio metric_value=100 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=User_Transaction_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Reads_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Reads_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Writes_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Writes_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Reads_Direct_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Reads_Direct_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Writes_Direct_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Writes_Direct_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Reads_Direct_Lobs_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Reads_Direct_Lobs_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Writes_Direct_Lobs_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Writes_Direct_Lobs__Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Redo_Generated_Per_Sec metric_value=1.86810076422304 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Redo_Generated_Per_Txn metric_value=132 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Logons_Per_Sec metric_value=0.0141522785168412 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Logons_Per_Txn metric_value=1 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Open_Cursors_Per_Sec metric_value=0.339654684404189 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Open_Cursors_Per_Txn metric_value=24 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=User_Commits_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=User_Commits_Percentage metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=User_Rollbacks_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=User_Rollbacks_Percentage metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=User_Calls_Per_Sec metric_value=0.240588734786301 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=User_Calls_Per_Txn metric_value=17 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Recursive_Calls_Per_Sec metric_value=5.77412963487121 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Recursive_Calls_Per_Txn metric_value=408 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Logical_Reads_Per_Sec metric_value=1.40107557316728 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Logical_Reads_Per_Txn metric_value=99 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=DBWR_Checkpoints_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Background_Checkpoints_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Redo_Writes_Per_Sec metric_value=0.0141522785168412 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Redo_Writes_Per_Txn metric_value=1 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Long_Table_Scans_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Long_Table_Scans_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Total_Table_Scans_Per_Sec metric_value=0.0566091140673648 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Total_Table_Scans_Per_Txn metric_value=4 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Full_Index_Scans_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Full_Index_Scans_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Total_Index_Scans_Per_Sec metric_value=0.877441268044155 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Total_Index_Scans_Per_Txn metric_value=62 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Total_Parse_Count_Per_Sec metric_value=0.339654684404189 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Total_Parse_Count_Per_Txn metric_value=24 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Hard_Parse_Count_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Hard_Parse_Count_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Parse_Failure_Count_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Parse_Failure_Count_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Cursor_Cache_Hit_Ratio metric_value=41.6666666666667 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Disk_Sort_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Disk_Sort_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Rows_Per_Sort metric_value=134.90243902439 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Execute_Without_Parse_Ratio metric_value=48.936170212766 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Soft_Parse_Ratio metric_value=100 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=User_Calls_Ratio metric_value=4 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Host_CPU_Utilization_(%) metric_value=6.71817146112219 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Network_Traffic_Volume_Per_Sec metric_value=280.51231248231 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Enqueue_Timeouts_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Enqueue_Timeouts_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Enqueue_Waits_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Enqueue_Waits_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Enqueue_Deadlocks_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Enqueue_Deadlocks_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Enqueue_Requests_Per_Sec metric_value=7.98188508349844 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Enqueue_Requests_Per_Txn metric_value=564 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=DB_Block_Gets_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=DB_Block_Gets_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Consistent_Read_Gets_Per_Sec metric_value=1.40107557316728 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Consistent_Read_Gets_Per_Txn metric_value=99 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=DB_Block_Changes_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=DB_Block_Changes_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Consistent_Read_Changes_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Consistent_Read_Changes_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=CPU_Usage_Per_Sec metric_value=1.47788706481744 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=CPU_Usage_Per_Txn metric_value=104.4275 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=CR_Blocks_Created_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=CR_Blocks_Created_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=CR_Undo_Records_Applied_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=CR_Undo_Records_Applied_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=User_Rollback_UndoRec_Applied_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=User_Rollback_Undo_Records_Applied_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Leaf_Node_Splits_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Leaf_Node_Splits_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Branch_Node_Splits_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Branch_Node_Splits_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=PX_downgraded_1_to_25%_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=PX_downgraded_25_to_50%_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=PX_downgraded_50_to_75%_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=PX_downgraded_75_to_99%_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=PX_downgraded_to_serial_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Read_Total_IO_Requests_Per_Sec metric_value=1.82564392867252 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Read_Total_Bytes_Per_Sec metric_value=29911.3501273705 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=GC_CR_Block_Received_Per_Second metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=GC_CR_Block_Received_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=GC_Current_Block_Received_Per_Second metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=GC_Current_Block_Received_Per_Txn metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Global_Cache_Average_CR_Get_Time metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Global_Cache_Average_Current_Get_Time metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Write_Total_IO_Requests_Per_Sec metric_value=0.396263798471554 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Global_Cache_Blocks_Corrupted metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Global_Cache_Blocks_Lost metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Current_Logons_Count metric_value=22 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Current_Open_Cursors_Count metric_value=28 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=User_Limit_% metric_value=0.000000512227416157775 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=SQL_Service_Response_Time metric_value=0.345076 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Database_Wait_Time_Ratio metric_value=28.7948844005719 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Database_CPU_Time_Ratio metric_value=71.2051155994281 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Response_Time_Per_Txn metric_value=146.6573 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Row_Cache_Hit_Ratio metric_value=100 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Row_Cache_Miss_Ratio metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Library_Cache_Hit_Ratio metric_value=100 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Library_Cache_Miss_Ratio metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Shared_Pool_Free_% metric_value=94.5707676149405 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=PGA_Cache_Hit_% metric_value=100 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Process_Limit_% metric_value=28 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Session_Limit_% metric_value=16.4772727272727 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Executions_Per_Txn metric_value=47 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Executions_Per_Sec metric_value=0.665157090291537 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Txns_Per_Logon metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Database_Time_Per_Sec metric_value=2.07553495612794 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Write_Total_Bytes_Per_Sec metric_value=6267.76110953864 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Read_IO_Requests_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Read_Bytes_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Write_IO_Requests_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Physical_Write_Bytes_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=DB_Block_Changes_Per_User_Call metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=DB_Block_Gets_Per_User_Call metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Executions_Per_User_Call metric_value=2.76470588235294 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Logical_Reads_Per_User_Call metric_value=5.82352941176471 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Total_Sorts_Per_User_Call metric_value=2.41176470588235 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Total_Table_Scans_Per_User_Call metric_value=0.235294117647059 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Current_OS_Load metric_value=0.369140625 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Streams_Pool_Usage_Percentage metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=PQ_QC_Session_Count metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=PQ_Slave_Session_Count metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Queries_parallelized_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=DML_statements_parallelized_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=DDL_statements_parallelized_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=PX_operations_not_downgraded_Per_Sec metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Session_Count metric_value=29 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Average_Synchronous_Single-Block_Read_Latency metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=I/O_Megabytes_per_Second metric_value=0.0283045570336824 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=I/O_Requests_per_Second metric_value=2.22190772714407 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Average_Active_Sessions metric_value=0.0207553495612794 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Active_Serial_Sessions metric_value=2 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Active_Parallel_Sessions metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Captured_user_calls metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Replayed_user_calls metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Workload_Capture_and_Replay_status metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Background_CPU_Usage_Per_Sec metric_value=0.131868100764223 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Background_Time_Per_Sec metric_value=0.00836783187093122 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Host_CPU_Usage_Per_Sec metric_value=53.4106991225587 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Cell_Physical_IO_Interconnect_Bytes metric_value=2556416 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Temp_Space_Used metric_value=0 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Total_PGA_Allocated metric_value=55719936 1634036614000000000
> oracle_sysmetric,instance=XE,metric_name=Total_PGA_Used_by_SQL_Workareas metric_value=0 1634036614000000000
> oracle_tablespaces,instance=XE,tbs_name=SYSAUX free_space_mb=25549,max_size_mb=26154,percent_used=2.31,used_space_mb=605 1634036614000000000
> oracle_tablespaces,instance=XE,tbs_name=SYSTEM free_space_mb=247,max_size_mb=600,percent_used=58.78,used_space_mb=353 1634036614000000000
> oracle_tablespaces,instance=XE,tbs_name=TEMP free_space_mb=25534,max_size_mb=25534,percent_used=0,used_space_mb=0 1634036614000000000
> oracle_tablespaces,instance=XE,tbs_name=UNDOTBS1 free_space_mb=25538,max_size_mb=25539,percent_used=0,used_space_mb=1 1634036614000000000
> oracle_tablespaces,instance=XE,tbs_name=USERS free_space_mb=11261,max_size_mb=11264,percent_used=0.02,used_space_mb=3 1634036614000000000
> oracle_connectioncount,instance=XE,metric_name=ACTIVE metric_value=23 1634036614000000000
> oracle_status,instance=XE,metric_name=database_status metric_value=1 1634036614000000000
> oracle_status,instance=XE,metric_name=instance_status metric_value=1 1634036614000000000
```