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
