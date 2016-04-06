package uwsgi

type StatsServer struct {
	Url     string
	Workers []*Worker `json:"workers"`
}

type Worker struct {
	Id            int    `json:"id"`  // Tag
	Pid           int    `json:"pid"` // Tag
	Accepting     int    `json:"accepting"`
	Requests      int    `json:"requests"`
	DeltaRequests int    `json:"delta_requests"`
	HarakiriCount int    `json:"harakiri_count"`
	Signals       int    `json:"signals"`
	SignalQueue   int    `json:"signal_queue"`
	Status        string `json:"status"`
	RSS           int    `json:"rss"`
	VSZ           int    `json:"vsz"`
	RunningTime   int    `json:"running_time"`
	LastSpawn     int    `json:"last_spawn"`
	RespawnCount  int    `json:"respawn_count"`
	TX            int    `json:"tx"`
	AvgRT         int    `json:"avg_rt"`
}
