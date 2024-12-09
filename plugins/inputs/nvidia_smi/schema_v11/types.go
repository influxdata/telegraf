package schema_v11

// smi defines the structure for the output of _nvidia-smi -q -x_.
type smi struct {
	GPU           []gpu  `xml:"gpu"`
	DriverVersion string `xml:"driver_version"`
	CUDAVersion   string `xml:"cuda_version"`
}

// gpu defines the structure of the GPU portion of the smi output.
type gpu struct {
	Clocks        clockStats         `xml:"clocks"`
	ComputeMode   string             `xml:"compute_mode"`
	DisplayActive string             `xml:"display_active"`
	DisplayMode   string             `xml:"display_mode"`
	EccMode       eccMode            `xml:"ecc_mode"`
	Encoder       encoderStats       `xml:"encoder_stats"`
	FanSpeed      string             `xml:"fan_speed"` // int
	FBC           fbcStats           `xml:"fbc_stats"`
	Memory        memoryStats        `xml:"fb_memory_usage"`
	PCI           pic                `xml:"pci"`
	Power         powerReadings      `xml:"power_readings"`
	ProdName      string             `xml:"product_name"`
	PState        string             `xml:"performance_state"`
	RemappedRows  memoryRemappedRows `xml:"remapped_rows"`
	RetiredPages  memoryRetiredPages `xml:"retired_pages"`
	Serial        string             `xml:"serial"`
	Temp          tempStats          `xml:"temperature"`
	Utilization   utilizationStats   `xml:"utilization"`
	UUID          string             `xml:"uuid"`
	VbiosVersion  string             `xml:"vbios_version"`
}

// eccMode defines the structure of the ecc portions in the smi output.
type eccMode struct {
	CurrentEcc string `xml:"current_ecc"` // Enabled, Disabled, N/A
	PendingEcc string `xml:"pending_ecc"` // Enabled, Disabled, N/A
}

// memoryStats defines the structure of the memory portions in the smi output.
type memoryStats struct {
	Total    string `xml:"total"`    // int
	Used     string `xml:"used"`     // int
	Free     string `xml:"free"`     // int
	Reserved string `xml:"reserved"` // int
}

// memoryRetiredPages defines the structure of the retired pages portions in the smi output.
type memoryRetiredPages struct {
	MultipleSingleBit struct {
		Count string `xml:"retired_count"` // int
	} `xml:"multiple_single_bit_retirement"`
	DoubleBit struct {
		Count string `xml:"retired_count"` // int
	} `xml:"double_bit_retirement"`
	PendingBlacklist  string `xml:"pending_blacklist"`  // Yes/No
	PendingRetirement string `xml:"pending_retirement"` // Yes/No
}

// memoryRemappedRows defines the structure of the remapped rows portions in the smi output.
type memoryRemappedRows struct {
	Correctable   string `xml:"remapped_row_corr"`    // int
	Uncorrectable string `xml:"remapped_row_unc"`     // int
	Pending       string `xml:"remapped_row_pending"` // Yes/No
	Failure       string `xml:"remapped_row_failure"` // Yes/No
}

// tempStats defines the structure of the temperature portion of the smi output.
type tempStats struct {
	GPUTemp string `xml:"gpu_temp"` // int
}

// utilizationStats defines the structure of the utilization portion of the smi output.
type utilizationStats struct {
	GPU     string `xml:"gpu_util"`     // int
	Memory  string `xml:"memory_util"`  // int
	Encoder string `xml:"encoder_util"` // int
	Decoder string `xml:"decoder_util"` // int
}

// powerReadings defines the structure of the power_readings portion of the smi output.
type powerReadings struct {
	PowerDraw  string `xml:"power_draw"`  // float
	PowerLimit string `xml:"power_limit"` // float
}

// pic defines the structure of the pci portion of the smi output.
type pic struct {
	LinkInfo struct {
		PCIEGen struct {
			CurrentLinkGen string `xml:"current_link_gen"` // int
		} `xml:"pcie_gen"`
		LinkWidth struct {
			CurrentLinkWidth string `xml:"current_link_width"` // int
		} `xml:"link_widths"`
	} `xml:"pci_gpu_link_info"`
}

// encoderStats defines the structure of the encoder_stats portion of the smi output.
type encoderStats struct {
	SessionCount   string `xml:"session_count"`   // int
	AverageFPS     string `xml:"average_fps"`     // int
	AverageLatency string `xml:"average_latency"` // int
}

// fbcStats defines the structure of the fbc_stats portion of the smi output.
type fbcStats struct {
	SessionCount   string `xml:"session_count"`   // int
	AverageFPS     string `xml:"average_fps"`     // int
	AverageLatency string `xml:"average_latency"` // int
}

// clockStats defines the structure of the clocks portion of the smi output.
type clockStats struct {
	Graphics string `xml:"graphics_clock"` // int
	SM       string `xml:"sm_clock"`       // int
	Memory   string `xml:"mem_clock"`      // int
	Video    string `xml:"video_clock"`    // int
}
