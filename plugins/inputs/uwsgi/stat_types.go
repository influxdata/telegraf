package uwsgi

type StatsServer struct {
	// Tags
	Url     string
	Pid     int    `json:"pid"`
	Uid     int    `json:"uid"`
	Gid     int    `json:"gid"`
	Version string `json:"version"`
	Cwd     string `json:"cwd"`

	// Fields
	ListenQueue       int `json:"listen_queue"`
	ListenQueueErrors int `json:"listen_queue_errors"`
	SignalQueue       int `json:"signal_queue"`
	Load              int `json:"load"`

	Workers []*Worker `json:"workers"`
}

type Worker struct {
	// Tags
	WorkerId int `json:"id"`
	Pid      int `json:"pid"`

	// Fields
	Accepting     int    `json:"accepting"`
	Requests      int    `json:"requests"`
	DeltaRequests int    `json:"delta_requests"`
	HarakiriCount int    `json:"harakiri_count"`
	Signals       int    `json:"signals"`
	SignalQueue   int    `json:"signal_queue"`
	Status        string `json:"status"`
	Rss           int    `json:"rss"`
	Vsz           int    `json:"vsz"`
	RunningTime   int    `json:"running_time"`
	LastSpawn     int    `json:"last_spawn"`
	RespawnCount  int    `json:"respawn_count"`
	Tx            int    `json:"tx"`
	AvgRt         int    `json:"avg_rt"`

	Apps []*App `json:"apps"`
}

type App struct {
	// Tags
	AppId      int    `json:"id"`
	MountPoint string `json:"mountpoint"`
	Chdir      string `json:"chdir"`

	// Fields
	Modifier1   int `json:"modifier1"`
	Requests    int `json:"requests"`
	StartupTime int `json:"startup_time"`
	Exceptions  int `json:"exceptions"`
}
