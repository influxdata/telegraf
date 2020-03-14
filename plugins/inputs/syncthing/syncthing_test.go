package syncthing

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSystemFields(t *testing.T) {

	report := Report{
		FolderMaxFiles: 1,
		FolderMaxMiB:   2,
		MemorySize:     3,
		MemoryUsageMiB: 4,
		NumCPU:         5,
		NumDevices:     6,
		NumFolders:     7,
		TotFiles:       8,
		TotMiB:         9,
		Uptime:         0,
	}

	system := SystemStatus{
		Alloc:      1,
		CPUPercent: 2,
		Goroutines: 3,
	}

	getSystemFieldsReturn := getSystemFields(&report, &system)

	correctSystemFieldReturn := make(map[string]interface{})

	correctSystemFieldReturn["folder_max_files"] = 1
	correctSystemFieldReturn["folder_max_mib"] = 2
	correctSystemFieldReturn["memory_size"] = 3
	correctSystemFieldReturn["memory_usage_mib"] = 4
	correctSystemFieldReturn["num_cpu"] = 5
	correctSystemFieldReturn["num_devices"] = 6
	correctSystemFieldReturn["num_folders"] = 7
	correctSystemFieldReturn["total_files"] = 8
	correctSystemFieldReturn["total_mib"] = 9
	correctSystemFieldReturn["uptime_seconds"] = 0
	correctSystemFieldReturn["alloc"] = 1
	correctSystemFieldReturn["cpu_percent"] = float64(2)
	correctSystemFieldReturn["goroutines"] = 3

	require.Equal(t, true, reflect.DeepEqual(getSystemFieldsReturn, correctSystemFieldReturn))
}

func TestGetFolderFields(t *testing.T) {

	folder := FolderStatus{
		Errors:            1,
		GlobalBytes:       200000000000,
		GlobalDeleted:     3,
		GlobalDirectories: 4,
		GlobalFiles:       5,
		GlobalSymlinks:    6,
		GlobalTotalItems:  7,
		InSyncBytes:       900000000000,
		InSyncFiles:       1,
		LocalBytes:        200000000000,
		LocalDeleted:      3,
		LocalDirectories:  4,
		LocalFiles:        5,
		LocalSymlinks:     6,
		LocalTotalItems:   7,
		NeedBytes:         800000000000,
		NeedDeletes:       9,
		NeedDirectories:   1,
		NeedFiles:         2,
		NeedSymlinks:      3,
		NeedTotalItems:    4,
		PullErrors:        5,
		Sequence:          6,
		Version:           8,
	}

	getFolderFieldsReturn := getFolderFields(&folder)

	correctFolderFieldReturn := make(map[string]interface{})

	correctFolderFieldReturn["errors"] = 1
	correctFolderFieldReturn["global_bytes"] = int64(200000000000)
	correctFolderFieldReturn["global_deleted"] = 3
	correctFolderFieldReturn["global_directories"] = 4
	correctFolderFieldReturn["global_files"] = 5
	correctFolderFieldReturn["global_symlinks"] = 6
	correctFolderFieldReturn["global_total_items"] = 7
	correctFolderFieldReturn["in_sync_bytes"] = int64(900000000000)
	correctFolderFieldReturn["in_sync_files"] = 1
	correctFolderFieldReturn["local_bytes"] = int64(200000000000)
	correctFolderFieldReturn["local_deleted"] = 3
	correctFolderFieldReturn["local_directories"] = 4
	correctFolderFieldReturn["local_files"] = 5
	correctFolderFieldReturn["local_symlinks"] = 6
	correctFolderFieldReturn["local_total_items"] = 7
	correctFolderFieldReturn["need_bytes"] = int64(800000000000)
	correctFolderFieldReturn["need_deletes"] = 9
	correctFolderFieldReturn["need_directories"] = 1
	correctFolderFieldReturn["need_files"] = 2
	correctFolderFieldReturn["need_symlinks"] = 3
	correctFolderFieldReturn["need_total_items"] = 4
	correctFolderFieldReturn["pull_errors"] = 5
	correctFolderFieldReturn["sequence"] = 6
	correctFolderFieldReturn["version"] = 8

	require.Equal(t, true, reflect.DeepEqual(getFolderFieldsReturn, correctFolderFieldReturn))
}
