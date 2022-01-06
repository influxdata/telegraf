package xtremio

type BBU struct {
	Content struct {
		Serial       string `json:"serial-number"`
		GUID         string `json:"guid"`
		PowerFeed    string `json:"power-feed"`
		Name         string `json:"Name"`
		ModelName    string `json:"model-name"`
		BBUPower     int    `json:"power"`
		BBUDailyTemp int    `json:"avg-daily-temp"`
		BBUEnabled   string `json:"enabled-state"`
		BBUNeedBat   bool   `json:"ups-need-battery-replacement,string"`
		BBULowBat    bool   `json:"is-low-battery-no-input,string"`
	}
}

type Clusters struct {
	Content struct {
		HardwarePlatform   string  `json:"hardware-platform"`
		LicenseID          string  `json:"license-id"`
		GUID               string  `json:"guid"`
		Name               string  `json:"name"`
		SerialNumber       string  `json:"sys-psnt-serial-number"`
		CompressionFactor  float64 `json:"compression-factor"`
		MemoryUsed         int     `json:"total-memory-in-use-in-percent"`
		ReadIops           int     `json:"rd-iops,string"`
		WriteIops          int     `json:"wr-iops,string"`
		NumVolumes         int     `json:"num-of-vols"`
		FreeSSDSpace       int     `json:"free-ud-ssd-space-in-percent"`
		NumSSDs            int     `json:"num-of-ssds"`
		DataReductionRatio float64 `json:"data-reduction-ratio"`
	}
}

type SSD struct {
	Content struct {
		ModelName       string `json:"model-name"`
		FirmwareVersion string `json:"fw-version"`
		SSDuid          string `json:"ssd-uid"`
		GUID            string `json:"guid"`
		SysName         string `json:"sys-name"`
		SerialNumber    string `json:"serial-number"`
		Size            int    `json:"ssd-size,string"`
		SpaceUsed       int    `json:"ssd-space-in-use,string"`
		WriteIops       int    `json:"wr-iops,string"`
		ReadIops        int    `json:"rd-iops,string"`
		WriteBandwidth  int    `json:"wr-bw,string"`
		ReadBandwidth   int    `json:"rd-bw,string"`
		NumBadSectors   int    `json:"num-bad-sectors"`
	}
}

type Volumes struct {
	Content struct {
		GUID               string  `json:"guid"`
		SysName            string  `json:"sys-name"`
		Name               string  `json:"name"`
		ReadIops           int     `json:"rd-iops,string"`
		WriteIops          int     `json:"wr-iops,string"`
		ReadLatency        int     `json:"rd-latency,string"`
		WriteLatency       int     `json:"wr-latency,string"`
		DataReductionRatio float64 `json:"data-reduction-ratio,string"`
		ProvisionedSpace   int     `json:"vol-size,string"`
		UsedSpace          int     `json:"logical-space-in-use,string"`
	}
}

type XMS struct {
	Content struct {
		GUID            string  `json:"guid"`
		Name            string  `json:"name"`
		Version         string  `json:"version"`
		IP              string  `json:"xms-ip"`
		WriteIops       int     `json:"wr-iops,string"`
		ReadIops        int     `json:"rd-iops,string"`
		EfficiencyRatio float64 `json:"overall-efficiency-ratio,string"`
		SpaceUsed       int     `json:"ssd-space-in-use,string"`
		RAMUsage        int     `json:"ram-usage,string"`
		RAMTotal        int     `json:"ram-total,string"`
		CPUUsage        float64 `json:"cpu"`
		WriteLatency    int     `json:"wr-latency,string"`
		ReadLatency     int     `json:"rd-latency,string"`
		NumAccounts     int     `json:"num-of-user-accounts"`
	}
}

type HREF struct {
	Href string `json:"href"`
}

type CollectorResponse struct {
	BBUs     []HREF `json:"bbus"`
	Clusters []HREF `json:"clusters"`
	SSDs     []HREF `json:"ssds"`
	Volumes  []HREF `json:"volumes"`
	XMS      []HREF `json:"xmss"`
}
