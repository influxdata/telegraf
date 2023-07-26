package schema_v11

// SMI defines the structure for the output of _nvidia-smi -q -x_.
type smi struct {
	GPU           []GPU  `xml:"gpu"`
	DriverVersion string `xml:"driver_version"`
	CUDAVersion   string `xml:"cuda_version"`
}

// GPU defines the structure of the GPU portion of the smi output.
type GPU struct {
	FanSpeed     string             `xml:"fan_speed"` // int
	Memory       MemoryStats        `xml:"fb_memory_usage"`
	RetiredPages MemoryRetiredPages `xml:"retired_pages"`
	RemappedRows MemoryRemappedRows `xml:"remapped_rows"`
	PState       string             `xml:"performance_state"`
	Temp         TempStats          `xml:"temperature"`
	ProdName     string             `xml:"product_name"`
	UUID         string             `xml:"uuid"`
	ComputeMode  string             `xml:"compute_mode"`
	Utilization  UtilizationStats   `xml:"utilization"`
	Power        PowerReadings      `xml:"power_readings"`
	PCI          PCI                `xml:"pci"`
	Encoder      EncoderStats       `xml:"encoder_stats"`
	FBC          FBCStats           `xml:"fbc_stats"`
	Clocks       ClockStats         `xml:"clocks"`
}

// MemoryStats defines the structure of the memory portions in the smi output.
type MemoryStats struct {
	Total    string `xml:"total"`    // int
	Used     string `xml:"used"`     // int
	Free     string `xml:"free"`     // int
	Reserved string `xml:"reserved"` // int
}

// MemoryRetiredPages defines the structure of the retired pages portions in the smi output.
type MemoryRetiredPages struct {
	MultipleSingleBit struct {
		Count string `xml:"retired_count"` // int
	} `xml:"multiple_single_bit_retirement"`
	DoubleBit struct {
		Count string `xml:"retired_count"` // int
	} `xml:"double_bit_retirement"`
	PendingBlacklist  string `xml:"pending_blacklist"`  // Yes/No
	PendingRetirement string `xml:"pending_retirement"` // Yes/No
}

// MemoryRemappedRows defines the structure of the remapped rows portions in the smi output.
type MemoryRemappedRows struct {
	Correctable   string `xml:"remapped_row_corr"`    // int
	Uncorrectable string `xml:"remapped_row_unc"`     // int
	Pending       string `xml:"remapped_row_pending"` // Yes/No
	Failure       string `xml:"remapped_row_failure"` // Yes/No
}

// TempStats defines the structure of the temperature portion of the smi output.
type TempStats struct {
	GPUTemp string `xml:"gpu_temp"` // int
}

// UtilizationStats defines the structure of the utilization portion of the smi output.
type UtilizationStats struct {
	GPU     string `xml:"gpu_util"`     // int
	Memory  string `xml:"memory_util"`  // int
	Encoder string `xml:"encoder_util"` // int
	Decoder string `xml:"decoder_util"` // int
}

// PowerReadings defines the structure of the power_readings portion of the smi output.
type PowerReadings struct {
	PowerDraw string `xml:"power_draw"` // float
}

// PCI defines the structure of the pci portion of the smi output.
type PCI struct {
	LinkInfo struct {
		PCIEGen struct {
			CurrentLinkGen string `xml:"current_link_gen"` // int
		} `xml:"pcie_gen"`
		LinkWidth struct {
			CurrentLinkWidth string `xml:"current_link_width"` // int
		} `xml:"link_widths"`
	} `xml:"pci_gpu_link_info"`
}

// EncoderStats defines the structure of the encoder_stats portion of the smi output.
type EncoderStats struct {
	SessionCount   string `xml:"session_count"`   // int
	AverageFPS     string `xml:"average_fps"`     // int
	AverageLatency string `xml:"average_latency"` // int
}

// FBCStats defines the structure of the fbc_stats portion of the smi output.
type FBCStats struct {
	SessionCount   string `xml:"session_count"`   // int
	AverageFPS     string `xml:"average_fps"`     // int
	AverageLatency string `xml:"average_latency"` // int
}

// ClockStats defines the structure of the clocks portion of the smi output.
type ClockStats struct {
	Graphics string `xml:"graphics_clock"` // int
	SM       string `xml:"sm_clock"`       // int
	Memory   string `xml:"mem_clock"`      // int
	Video    string `xml:"video_clock"`    // int
}
