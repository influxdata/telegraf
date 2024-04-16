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
	Device struct {
		Name     string `json:"name"`
		InfoName string `json:"info_name"`
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	} `json:"device"`
	Vendor          string `json:"vendor"`
	Product         string `json:"product"`
	ModelFamily     string `json:"model_family"`
	ModelName       string `json:"model_name"`
	SerialNumber    string `json:"serial_number"`
	FirmwareVersion string `json:"firmware_version"`
	Wwn             struct {
		Naa int   `json:"naa"`
		Oui int   `json:"oui"`
		ID  int64 `json:"id"`
	} `json:"wwn"`
	UserCapacity struct {
		Bytes int64 `json:"bytes"`
	} `json:"user_capacity"`
	SmartStatus struct {
		Passed bool `json:"passed"`
	} `json:"smart_status"`
	NvmeSmartHealthInformationLog struct {
		CriticalWarning         int64 `json:"critical_warning"`
		Temperature             int64 `json:"temperature"`
		AvailableSpare          int64 `json:"available_spare"`
		AvailableSpareThreshold int64 `json:"available_spare_threshold"`
		PercentageUsed          int64 `json:"percentage_used"`
		DataUnitsRead           int64 `json:"data_units_read"`
		DataUnitsWritten        int64 `json:"data_units_written"`
		HostReads               int64 `json:"host_reads"`
		HostWrites              int64 `json:"host_writes"`
		ControllerBusyTime      int64 `json:"controller_busy_time"`
		PowerCycles             int64 `json:"power_cycles"`
		PowerOnHours            int64 `json:"power_on_hours"`
		UnsafeShutdowns         int64 `json:"unsafe_shutdowns"`
		MediaErrors             int64 `json:"media_errors"`
		NumErrLogEntries        int64 `json:"num_err_log_entries"`
		WarningTempTime         int64 `json:"warning_temp_time"`
		CriticalCompTime        int64 `json:"critical_comp_time"`
	} `json:"nvme_smart_health_information_log"`
	Temperature struct {
		Current int `json:"current"`
	} `json:"temperature"`
	AtaSmartAttributes struct {
		Revision int `json:"revision"`
		Table    []struct {
			ID         int64  `json:"id"`
			Name       string `json:"name"`
			Value      int64  `json:"value"`
			Worst      int64  `json:"worst"`
			Thresh     int64  `json:"thresh"`
			WhenFailed string `json:"when_failed"`
			Flags      struct {
				Value         int64  `json:"value"`
				String        string `json:"string"`
				Prefailure    bool   `json:"prefailure"`
				UpdatedOnline bool   `json:"updated_online"`
				Performance   bool   `json:"performance"`
				ErrorRate     bool   `json:"error_rate"`
				EventCount    bool   `json:"event_count"`
				AutoKeep      bool   `json:"auto_keep"`
			} `json:"flags"`
			Raw struct {
				Value  int64  `json:"value"`
				String string `json:"string"`
			} `json:"raw"`
		} `json:"table"`
	} `json:"ata_smart_attributes"`
	ScsiErrorCounterLog struct {
		Read struct {
			ErrorsCorrectedByEccfast         int    `json:"errors_corrected_by_eccfast"`
			ErrorsCorrectedByEccdelayed      int    `json:"errors_corrected_by_eccdelayed"`
			ErrorsCorrectedByRereadsRewrites int    `json:"errors_corrected_by_rereads_rewrites"`
			TotalErrorsCorrected             int    `json:"total_errors_corrected"`
			CorrectionAlgorithmInvocations   int    `json:"correction_algorithm_invocations"`
			GigabytesProcessed               string `json:"gigabytes_processed"`
			TotalUncorrectedErrors           int    `json:"total_uncorrected_errors"`
		} `json:"read"`
		Write struct {
			ErrorsCorrectedByEccfast         int    `json:"errors_corrected_by_eccfast"`
			ErrorsCorrectedByEccdelayed      int    `json:"errors_corrected_by_eccdelayed"`
			ErrorsCorrectedByRereadsRewrites int    `json:"errors_corrected_by_rereads_rewrites"`
			TotalErrorsCorrected             int    `json:"total_errors_corrected"`
			CorrectionAlgorithmInvocations   int    `json:"correction_algorithm_invocations"`
			GigabytesProcessed               string `json:"gigabytes_processed"`
			TotalUncorrectedErrors           int    `json:"total_uncorrected_errors"`
		} `json:"write"`
		Verify struct {
			ErrorsCorrectedByEccfast         int    `json:"errors_corrected_by_eccfast"`
			ErrorsCorrectedByEccdelayed      int    `json:"errors_corrected_by_eccdelayed"`
			ErrorsCorrectedByRereadsRewrites int    `json:"errors_corrected_by_rereads_rewrites"`
			TotalErrorsCorrected             int    `json:"total_errors_corrected"`
			CorrectionAlgorithmInvocations   int    `json:"correction_algorithm_invocations"`
			GigabytesProcessed               string `json:"gigabytes_processed"`
			TotalUncorrectedErrors           int    `json:"total_uncorrected_errors"`
		} `json:"verify"`
	} `json:"scsi_error_counter_log"`
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
