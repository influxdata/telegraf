package bind

import "time"

type statistics struct {
	JSONStatsVersion string           `json:"json-stats-version"`
	BootTime         time.Time        `json:"boot-time"`
	ConfigTime       time.Time        `json:"config-time"`
	CurrentTime      time.Time        `json:"current-time"`
	Opcodes          map[string]int64 `json:"opcodes"`
	Qtypes           map[string]int64 `json:"qtypes"`
	Nsstats          map[string]int64 `json:"nsstats"`
	Views            map[string]struct {
		Zones []struct {
			Name   string `json:"name"`
			Class  string `json:"class"`
			Serial int    `json:"serial"`
		} `json:"zones"`
		Resolver struct {
			Stats      map[string]int64 `json:"stats"`
			Qtypes     map[string]int64 `json:"qtypes"`
			Cache      map[string]int64 `json:"cache"`
			Cachestats map[string]int64 `json:"cachestats"`
			Adb        map[string]int64 `json:"adb"`
		} `json:"resolver"`
	} `json:"views"`
	Sockstats map[string]int64 `json:"sockstats"`
	Socketmgr struct {
		Sockets []struct {
			ID           string   `json:"id"`
			References   int      `json:"references"`
			Type         string   `json:"type"`
			LocalAddress string   `json:"local-address,omitempty"`
			States       []string `json:"states"`
			PeerAddress  string   `json:"peer-address,omitempty"`
		} `json:"sockets"`
	} `json:"socketmgr"`
	Taskmgr struct {
		ThreadModel    string `json:"thread-model"`
		WorkerThreads  int    `json:"worker-threads"`
		DefaultQuantum int    `json:"default-quantum"`
		TasksRunning   int    `json:"tasks-running"`
		TasksReady     int    `json:"tasks-ready"`
		Tasks          []struct {
			ID         string `json:"id"`
			Name       string `json:"name,omitempty"`
			References int    `json:"references"`
			State      string `json:"state"`
			Quantum    int    `json:"quantum"`
			Events     int    `json:"events"`
		} `json:"tasks"`
	} `json:"taskmgr"`
	Memory struct {
		TotalUse    int64 `json:"TotalUse"`
		InUse       int64 `json:"InUse"`
		BlockSize   int64 `json:"BlockSize"`
		ContextSize int64 `json:"ContextSize"`
		Lost        int64 `json:"Lost"`
		Contexts    []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			References int    `json:"references"`
			Total      int    `json:"total"`
			Inuse      int    `json:"inuse"`
			Maxinuse   int    `json:"maxinuse"`
			Blocksize  int    `json:"blocksize,omitempty"`
			Pools      int    `json:"pools"`
			Hiwater    int    `json:"hiwater"`
			Lowater    int    `json:"lowater"`
		} `json:"contexts"`
	} `json:"memory"`
}
