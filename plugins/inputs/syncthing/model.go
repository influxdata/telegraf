package syncthing

import "time"

// /rest/svc/report
type Report struct {
	AlwaysLocalNets bool `json:"alwaysLocalNets"`
	Announce        struct {
		DefaultServersDNS int  `json:"defaultServersDNS"`
		DefaultServersIP  int  `json:"defaultServersIP"`
		GlobalEnabled     bool `json:"globalEnabled"`
		LocalEnabled      bool `json:"localEnabled"`
		OtherServers      int  `json:"otherServers"`
	} `json:"announce"`
	BlockStats struct {
	} `json:"blockStats"`
	CacheIgnoredFiles        bool `json:"cacheIgnoredFiles"`
	CustomDefaultFolderPath  bool `json:"customDefaultFolderPath"`
	CustomReleaseURL         bool `json:"customReleaseURL"`
	CustomTempIndexMinBlocks bool `json:"customTempIndexMinBlocks"`
	CustomTrafficClass       bool `json:"customTrafficClass"`
	DeviceUses               struct {
		CompressAlways   int `json:"compressAlways"`
		CompressMetadata int `json:"compressMetadata"`
		CompressNever    int `json:"compressNever"`
		CustomCertName   int `json:"customCertName"`
		DynamicAddr      int `json:"dynamicAddr"`
		Introducer       int `json:"introducer"`
		StaticAddr       int `json:"staticAddr"`
	} `json:"deviceUses"`
	FolderMaxFiles int `json:"folderMaxFiles"`
	FolderMaxMiB   int `json:"folderMaxMiB"`
	FolderUses     struct {
		AutoNormalize       int `json:"autoNormalize"`
		ExternalVersioning  int `json:"externalVersioning"`
		IgnoreDelete        int `json:"ignoreDelete"`
		IgnorePerms         int `json:"ignorePerms"`
		Receiveonly         int `json:"receiveonly"`
		Sendonly            int `json:"sendonly"`
		Sendreceive         int `json:"sendreceive"`
		SimpleVersioning    int `json:"simpleVersioning"`
		StaggeredVersioning int `json:"staggeredVersioning"`
		TrashcanVersioning  int `json:"trashcanVersioning"`
	} `json:"folderUses"`
	FolderUsesV3 struct {
		AlwaysWeakHash          int `json:"alwaysWeakHash"`
		ConflictsDisabled       int `json:"conflictsDisabled"`
		ConflictsOther          int `json:"conflictsOther"`
		ConflictsUnlimited      int `json:"conflictsUnlimited"`
		CustomWeakHashThreshold int `json:"customWeakHashThreshold"`
		DisableSparseFiles      int `json:"disableSparseFiles"`
		DisableTempIndexes      int `json:"disableTempIndexes"`
		FilesystemType          struct {
			Basic int `json:"basic"`
		} `json:"filesystemType"`
		FsWatcherDelays  []int `json:"fsWatcherDelays"`
		FsWatcherEnabled int   `json:"fsWatcherEnabled"`
		PullOrder        struct {
			Random int `json:"random"`
		} `json:"pullOrder"`
		ScanProgressDisabled int `json:"scanProgressDisabled"`
	} `json:"folderUsesV3"`
	GuiStats struct {
		Debugging                 int `json:"debugging"`
		Enabled                   int `json:"enabled"`
		InsecureAdminAccess       int `json:"insecureAdminAccess"`
		InsecureAllowFrameLoading int `json:"insecureAllowFrameLoading"`
		InsecureSkipHostCheck     int `json:"insecureSkipHostCheck"`
		ListenLocal               int `json:"listenLocal"`
		ListenUnspecified         int `json:"listenUnspecified"`
		Theme                     struct {
			Default int `json:"default"`
		} `json:"theme"`
		UseAuth int `json:"useAuth"`
		UseTLS  int `json:"useTLS"`
	} `json:"guiStats"`
	HashPerf    float64 `json:"hashPerf"`
	IgnoreStats struct {
		Deletable       int `json:"deletable"`
		DoubleStars     int `json:"doubleStars"`
		EscapedIncludes int `json:"escapedIncludes"`
		Folded          int `json:"folded"`
		Includes        int `json:"includes"`
		Inverts         int `json:"inverts"`
		Lines           int `json:"lines"`
		Rooted          int `json:"rooted"`
		Stars           int `json:"stars"`
	} `json:"ignoreStats"`
	LimitBandwidthInLan        bool   `json:"limitBandwidthInLan"`
	LongVersion                string `json:"longVersion"`
	MemorySize                 int    `json:"memorySize"`
	MemoryUsageMiB             int    `json:"memoryUsageMiB"`
	NatType                    string `json:"natType"`
	NumCPU                     int    `json:"numCPU"`
	NumDevices                 int    `json:"numDevices"`
	NumFolders                 int    `json:"numFolders"`
	OverwriteRemoteDeviceNames bool   `json:"overwriteRemoteDeviceNames"`
	Platform                   string `json:"platform"`
	ProgressEmitterEnabled     bool   `json:"progressEmitterEnabled"`
	Relays                     struct {
		DefaultServers int  `json:"defaultServers"`
		Enabled        bool `json:"enabled"`
		OtherServers   int  `json:"otherServers"`
	} `json:"relays"`
	RescanIntvs         []int   `json:"rescanIntvs"`
	RestartOnWakeup     bool    `json:"restartOnWakeup"`
	Sha256Perf          float64 `json:"sha256Perf"`
	TemporariesCustom   bool    `json:"temporariesCustom"`
	TemporariesDisabled bool    `json:"temporariesDisabled"`
	TotFiles            int     `json:"totFiles"`
	TotMiB              int     `json:"totMiB"`
	TransportStats      struct {
		TCP4 int `json:"tcp4"`
	} `json:"transportStats"`
	UniqueID             string `json:"uniqueID"`
	UpgradeAllowedAuto   bool   `json:"upgradeAllowedAuto"`
	UpgradeAllowedManual bool   `json:"upgradeAllowedManual"`
	UpgradeAllowedPre    bool   `json:"upgradeAllowedPre"`
	Uptime               int    `json:"uptime"`
	UrVersion            int    `json:"urVersion"`
	UsesRateLimit        bool   `json:"usesRateLimit"`
	Version              string `json:"version"`
}

// /rest/system/status
type SystemStatus struct {
	Alloc                   int `json:"alloc"`
	ConnectionServiceStatus map[string]interface{}
	CPUPercent              float64 `json:"cpuPercent"`
	DiscoveryEnabled        bool    `json:"discoveryEnabled"`
	DiscoveryErrors         map[string]interface{}
	DiscoveryMethods        int    `json:"discoveryMethods"`
	Goroutines              int    `json:"goroutines"`
	GuiAddressOverridden    bool   `json:"guiAddressOverridden"`
	GuiAddressUsed          string `json:"guiAddressUsed"`
	LastDialStatus          map[string]interface{}
	MyID                    string    `json:"myID"`
	PathSeparator           string    `json:"pathSeparator"`
	StartTime               time.Time `json:"startTime"`
	Sys                     int       `json:"sys"`
	Tilde                   string    `json:"tilde"`
	Uptime                  int       `json:"uptime"`
	UrVersionMax            int       `json:"urVersionMax"`
}

// /rest/db/status?folder=<ID>
type FolderStatus struct {
	Errors            int       `json:"errors"`
	GlobalBytes       int64     `json:"globalBytes"`
	GlobalDeleted     int       `json:"globalDeleted"`
	GlobalDirectories int       `json:"globalDirectories"`
	GlobalFiles       int       `json:"globalFiles"`
	GlobalSymlinks    int       `json:"globalSymlinks"`
	GlobalTotalItems  int       `json:"globalTotalItems"`
	IgnorePatterns    bool      `json:"ignorePatterns"`
	InSyncBytes       int64     `json:"inSyncBytes"`
	InSyncFiles       int       `json:"inSyncFiles"`
	Invalid           string    `json:"invalid"`
	LocalBytes        int64     `json:"localBytes"`
	LocalDeleted      int       `json:"localDeleted"`
	LocalDirectories  int       `json:"localDirectories"`
	LocalFiles        int       `json:"localFiles"`
	LocalSymlinks     int       `json:"localSymlinks"`
	LocalTotalItems   int       `json:"localTotalItems"`
	NeedBytes         int64     `json:"needBytes"`
	NeedDeletes       int       `json:"needDeletes"`
	NeedDirectories   int       `json:"needDirectories"`
	NeedFiles         int       `json:"needFiles"`
	NeedSymlinks      int       `json:"needSymlinks"`
	NeedTotalItems    int       `json:"needTotalItems"`
	PullErrors        int       `json:"pullErrors"`
	Sequence          int       `json:"sequence"`
	State             string    `json:"state"`
	StateChanged      time.Time `json:"stateChanged"`
	Version           int       `json:"version"`
}

// /rest/system/config
type Folder struct {
	ID             string `json:"id"`
	Label          string `json:"label"`
	FilesystemType string `json:"filesystemType"`
	Path           string `json:"path"`
	Type           string `json:"type"`
	Devices        []struct {
		DeviceID     string `json:"deviceID"`
		IntroducedBy string `json:"introducedBy"`
	} `json:"devices"`
	RescanIntervalS  int  `json:"rescanIntervalS"`
	FsWatcherEnabled bool `json:"fsWatcherEnabled"`
	FsWatcherDelayS  int  `json:"fsWatcherDelayS"`
	IgnorePerms      bool `json:"ignorePerms"`
	AutoNormalize    bool `json:"autoNormalize"`
	MinDiskFree      struct {
		Value int    `json:"value"`
		Unit  string `json:"unit"`
	} `json:"minDiskFree"`
	Versioning struct {
		Type   string `json:"type"`
		Params struct {
		} `json:"params"`
	} `json:"versioning"`
	Copiers                 int    `json:"copiers"`
	PullerMaxPendingKiB     int    `json:"pullerMaxPendingKiB"`
	Hashers                 int    `json:"hashers"`
	Order                   string `json:"order"`
	IgnoreDelete            bool   `json:"ignoreDelete"`
	ScanProgressIntervalS   int    `json:"scanProgressIntervalS"`
	PullerPauseS            int    `json:"pullerPauseS"`
	MaxConflicts            int    `json:"maxConflicts"`
	DisableSparseFiles      bool   `json:"disableSparseFiles"`
	DisableTempIndexes      bool   `json:"disableTempIndexes"`
	Paused                  bool   `json:"paused"`
	WeakHashThresholdPct    int    `json:"weakHashThresholdPct"`
	MarkerName              string `json:"markerName"`
	CopyOwnershipFromParent bool   `json:"copyOwnershipFromParent"`
	ModTimeWindowS          int    `json:"modTimeWindowS"`
}

// /rest/system/config
type Config struct {
	Version int      `json:"version"`
	Folders []Folder `json:"folders"`
	Devices []struct {
		DeviceID                 string        `json:"deviceID"`
		Name                     string        `json:"name"`
		Addresses                []string      `json:"addresses"`
		Compression              string        `json:"compression"`
		CertName                 string        `json:"certName"`
		Introducer               bool          `json:"introducer"`
		SkipIntroductionRemovals bool          `json:"skipIntroductionRemovals"`
		IntroducedBy             string        `json:"introducedBy"`
		Paused                   bool          `json:"paused"`
		AllowedNetworks          []interface{} `json:"allowedNetworks"`
		AutoAcceptFolders        bool          `json:"autoAcceptFolders"`
		MaxSendKbps              int           `json:"maxSendKbps"`
		MaxRecvKbps              int           `json:"maxRecvKbps"`
		IgnoredFolders           []interface{} `json:"ignoredFolders"`
		PendingFolders           []interface{} `json:"pendingFolders"`
		MaxRequestKiB            int           `json:"maxRequestKiB"`
	} `json:"devices"`
	Gui struct {
		Enabled                   bool   `json:"enabled"`
		Address                   string `json:"address"`
		User                      string `json:"user"`
		Password                  string `json:"password"`
		AuthMode                  string `json:"authMode"`
		UseTLS                    bool   `json:"useTLS"`
		APIKey                    string `json:"apiKey"`
		InsecureAdminAccess       bool   `json:"insecureAdminAccess"`
		Theme                     string `json:"theme"`
		Debugging                 bool   `json:"debugging"`
		InsecureSkipHostcheck     bool   `json:"insecureSkipHostcheck"`
		InsecureAllowFrameLoading bool   `json:"insecureAllowFrameLoading"`
	} `json:"gui"`
	Ldap struct {
		Addresd            string `json:"addresd"`
		BindDN             string `json:"bindDN"`
		Transport          string `json:"transport"`
		InsecureSkipVerify bool   `json:"insecureSkipVerify"`
	} `json:"ldap"`
	Options struct {
		ListenAddresses         []string `json:"listenAddresses"`
		GlobalAnnounceServers   []string `json:"globalAnnounceServers"`
		GlobalAnnounceEnabled   bool     `json:"globalAnnounceEnabled"`
		LocalAnnounceEnabled    bool     `json:"localAnnounceEnabled"`
		LocalAnnouncePort       int      `json:"localAnnouncePort"`
		LocalAnnounceMCAddr     string   `json:"localAnnounceMCAddr"`
		MaxSendKbps             int      `json:"maxSendKbps"`
		MaxRecvKbps             int      `json:"maxRecvKbps"`
		ReconnectionIntervalS   int      `json:"reconnectionIntervalS"`
		RelaysEnabled           bool     `json:"relaysEnabled"`
		RelayReconnectIntervalM int      `json:"relayReconnectIntervalM"`
		StartBrowser            bool     `json:"startBrowser"`
		NatEnabled              bool     `json:"natEnabled"`
		NatLeaseMinutes         int      `json:"natLeaseMinutes"`
		NatRenewalMinutes       int      `json:"natRenewalMinutes"`
		NatTimeoutSeconds       int      `json:"natTimeoutSeconds"`
		UrAccepted              int      `json:"urAccepted"`
		UrSeen                  int      `json:"urSeen"`
		UrUniqueID              string   `json:"urUniqueId"`
		UrURL                   string   `json:"urURL"`
		UrPostInsecurely        bool     `json:"urPostInsecurely"`
		UrInitialDelayS         int      `json:"urInitialDelayS"`
		RestartOnWakeup         bool     `json:"restartOnWakeup"`
		AutoUpgradeIntervalH    int      `json:"autoUpgradeIntervalH"`
		UpgradeToPreReleases    bool     `json:"upgradeToPreReleases"`
		KeepTemporariesH        int      `json:"keepTemporariesH"`
		CacheIgnoredFiles       bool     `json:"cacheIgnoredFiles"`
		ProgressUpdateIntervalS int      `json:"progressUpdateIntervalS"`
		LimitBandwidthInLan     bool     `json:"limitBandwidthInLan"`
		MinHomeDiskFree         struct {
			Value int    `json:"value"`
			Unit  string `json:"unit"`
		} `json:"minHomeDiskFree"`
		ReleasesURL                         string        `json:"releasesURL"`
		AlwaysLocalNets                     []interface{} `json:"alwaysLocalNets"`
		OverwriteRemoteDeviceNamesOnConnect bool          `json:"overwriteRemoteDeviceNamesOnConnect"`
		TempIndexMinBlocks                  int           `json:"tempIndexMinBlocks"`
		UnackedNotificationIDs              []interface{} `json:"unackedNotificationIDs"`
		TrafficClass                        int           `json:"trafficClass"`
		DefaultFolderPath                   string        `json:"defaultFolderPath"`
		SetLowPriority                      bool          `json:"setLowPriority"`
		MaxConcurrentScans                  int           `json:"maxConcurrentScans"`
		CrURL                               string        `json:"crURL"`
		CrashReportingEnabled               bool          `json:"crashReportingEnabled"`
		StunKeepaliveStartS                 int           `json:"stunKeepaliveStartS"`
		StunKeepaliveMinS                   int           `json:"stunKeepaliveMinS"`
		StunServers                         []string      `json:"stunServers"`
		DatabaseTuning                      string        `json:"databaseTuning"`
	} `json:"options"`
	RemoteIgnoredDevices []interface{} `json:"remoteIgnoredDevices"`
	PendingDevices       []interface{} `json:"pendingDevices"`
}
