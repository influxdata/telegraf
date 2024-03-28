package smartctl

type smartctlDeviceJSON struct {
	JSONFormatVersion []int `json:"json_format_version"`
	Smartctl          struct {
		Version      []int    `json:"version"`
		PreRelease   bool     `json:"pre_release"`
		SvnRevision  string   `json:"svn_revision"`
		PlatformInfo string   `json:"platform_info"`
		BuildInfo    string   `json:"build_info"`
		Argv         []string `json:"argv"`
		Messages     []struct {
			Severity string `json:"severity"`
			String   string `json:"string"`
		} `json:"messages"`
		ExitStatus int `json:"exit_status"`
	} `json:"smartctl"`
	LocalTime struct {
		TimeT   int    `json:"time_t"`
		Asctime string `json:"asctime"`
	} `json:"local_time"`
	Device struct {
		Name     string `json:"name"`
		InfoName string `json:"info_name"`
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	} `json:"device"`
	ModelFamily     string `json:"model_family"`
	ModelName       string `json:"model_name"`
	SerialNumber    string `json:"serial_number"`
	FirmwareVersion string `json:"firmware_version"`
	Wwn             struct {
		Naa int   `json:"naa"`
		Oui int   `json:"oui"`
		ID  int64 `json:"id"`
	} `json:"wwn"`
	NvmePciVendor struct {
		ID          int `json:"id"`
		SubsystemID int `json:"subsystem_id"`
	} `json:"nvme_pci_vendor"`
	NvmeIeeeOuiIdentifier   int   `json:"nvme_ieee_oui_identifier"`
	NvmeTotalCapacity       int64 `json:"nvme_total_capacity"`
	NvmeUnallocatedCapacity int   `json:"nvme_unallocated_capacity"`
	NvmeControllerID        int   `json:"nvme_controller_id"`
	NvmeVersion             struct {
		String string `json:"string"`
		Value  int    `json:"value"`
	} `json:"nvme_version"`
	NvmeNumberOfNamespaces int `json:"nvme_number_of_namespaces"`
	NvmeNamespaces         []struct {
		ID   int `json:"id"`
		Size struct {
			Blocks int   `json:"blocks"`
			Bytes  int64 `json:"bytes"`
		} `json:"size"`
		Capacity struct {
			Blocks int   `json:"blocks"`
			Bytes  int64 `json:"bytes"`
		} `json:"capacity"`
		Utilization struct {
			Blocks int   `json:"blocks"`
			Bytes  int64 `json:"bytes"`
		} `json:"utilization"`
		FormattedLbaSize int `json:"formatted_lba_size"`
		Eui64            struct {
			Oui   int   `json:"oui"`
			ExtID int64 `json:"ext_id"`
		} `json:"eui64"`
	} `json:"nvme_namespaces"`
	UserCapacity struct {
		Blocks int   `json:"blocks"`
		Bytes  int64 `json:"bytes"`
	} `json:"user_capacity"`
	LogicalBlockSize int `json:"logical_block_size"`
	SmartSupport     struct {
		Available bool `json:"available"`
		Enabled   bool `json:"enabled"`
	} `json:"smart_support"`
	SmartStatus struct {
		Passed bool `json:"passed"`
		Nvme   struct {
			Value int `json:"value"`
		} `json:"nvme"`
	} `json:"smart_status"`
	NvmeSmartHealthInformationLog struct {
		CriticalWarning         int `json:"critical_warning"`
		Temperature             int `json:"temperature"`
		AvailableSpare          int `json:"available_spare"`
		AvailableSpareThreshold int `json:"available_spare_threshold"`
		PercentageUsed          int `json:"percentage_used"`
		DataUnitsRead           int `json:"data_units_read"`
		DataUnitsWritten        int `json:"data_units_written"`
		HostReads               int `json:"host_reads"`
		HostWrites              int `json:"host_writes"`
		ControllerBusyTime      int `json:"controller_busy_time"`
		PowerCycles             int `json:"power_cycles"`
		PowerOnHours            int `json:"power_on_hours"`
		UnsafeShutdowns         int `json:"unsafe_shutdowns"`
		MediaErrors             int `json:"media_errors"`
		NumErrLogEntries        int `json:"num_err_log_entries"`
		WarningTempTime         int `json:"warning_temp_time"`
		CriticalCompTime        int `json:"critical_comp_time"`
	} `json:"nvme_smart_health_information_log"`
	Temperature struct {
		Current int `json:"current"`
	} `json:"temperature"`
	PowerCycleCount int `json:"power_cycle_count"`
	PowerOnTime     struct {
		Hours int `json:"hours"`
	} `json:"power_on_time"`
	NvmeErrorInformationLog struct {
		Size   int `json:"size"`
		Read   int `json:"read"`
		Unread int `json:"unread"`
		Table  []struct {
			ErrorCount        int `json:"error_count"`
			SubmissionQueueID int `json:"submission_queue_id"`
			CommandID         int `json:"command_id"`
			StatusField       struct {
				Value          int    `json:"value"`
				DoNotRetry     bool   `json:"do_not_retry"`
				StatusCodeType int    `json:"status_code_type"`
				StatusCode     int    `json:"status_code"`
				String         string `json:"string"`
			} `json:"status_field"`
			PhaseTag          bool `json:"phase_tag"`
			ParmErrorLocation int  `json:"parm_error_location"`
			Lba               struct {
				Value int `json:"value"`
			} `json:"lba"`
			Nsid int `json:"nsid"`
		} `json:"table"`
	} `json:"nvme_error_information_log"`
	NvmeSelfTestLog struct {
		CurrentSelfTestOperation struct {
			Value  int    `json:"value"`
			String string `json:"string"`
		} `json:"current_self_test_operation"`
	} `json:"nvme_self_test_log"`
	AtaSmartAttributes struct {
		Revision int `json:"revision"`
		Table    []struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Value      int    `json:"value"`
			Worst      int    `json:"worst"`
			Thresh     int    `json:"thresh"`
			WhenFailed string `json:"when_failed"`
			Flags      struct {
				Value         int    `json:"value"`
				String        string `json:"string"`
				Prefailure    bool   `json:"prefailure"`
				UpdatedOnline bool   `json:"updated_online"`
				Performance   bool   `json:"performance"`
				ErrorRate     bool   `json:"error_rate"`
				EventCount    bool   `json:"event_count"`
				AutoKeep      bool   `json:"auto_keep"`
			} `json:"flags"`
			Raw struct {
				Value  int    `json:"value"`
				String string `json:"string"`
			} `json:"raw"`
		} `json:"table"`
	} `json:"ata_smart_attributes"`
}

type smartctlScanJSON struct {
	JSONFormatVersion []int `json:"json_format_version"`
	Smartctl          struct {
		Version      []int    `json:"version"`
		PreRelease   bool     `json:"pre_release"`
		SvnRevision  string   `json:"svn_revision"`
		PlatformInfo string   `json:"platform_info"`
		BuildInfo    string   `json:"build_info"`
		Argv         []string `json:"argv"`
		ExitStatus   int      `json:"exit_status"`
	} `json:"smartctl"`
	Devices []struct {
		Name     string `json:"name"`
		InfoName string `json:"info_name"`
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	} `json:"devices"`
}
