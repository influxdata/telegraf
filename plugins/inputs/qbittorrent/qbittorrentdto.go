package qbittorrent

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"strconv"
	"time"
)
type Category struct {
	Name     string `json:"name"`
	SavePath string `json:"savePath"`
}

func (c *Category) partialUpdate(update *Category) {
	if update.Name != "" {
		c.Name = update.Name
	}
	if update.SavePath != "" {
		c.SavePath = update.SavePath
	}
}

type ServerState struct {
	AllTimeDownload      int64  `json:"alltime_dl"`
	AllTimeUpload        int64  `json:"alltime_ul"`
	AverageTimeQueue     int64  `json:"average_time_queue"`
	ConnectionStatus     string `json:"connection_status"`
	DHTNodes             int64  `json:"dht_nodes"`
	DLInfoData           int64  `json:"dl_info_data"`
	DLInfoSpeed          int64  `json:"dl_info_speed"`
	DLRateLimit          int64  `json:"dl_rate_limit"`
	FreeSpaceOnDisk      int64  `json:"free_space_on_disk"`
	GlobalRatio          string `json:"global_ratio"`
	QueuedIOJobs         int64  `json:"queued_io_jobs"`
	Queueing             *bool  `json:"queueing"`
	ReadCacheHits        string `json:"read_cache_hits"`
	ReadCacheOverload    string `json:"read_cache_overload"`
	RefreshInterval      int64  `json:"refresh_interval"`
	TotalBuffersSize     int64  `json:"total_buffers_size"`
	TotalPeerConnections int64  `json:"total_peer_connections"`
	TotalQueuedSize      int64  `json:"total_queued_size"`
	TotalWastedSession   int64  `json:"total_wasted_session"`
	UpInfoData           int64  `json:"up_info_data"`
	UpInfoSpeed          int64  `json:"up_info_speed"`
	UpRateLimit          int64  `json:"up_rate_limit"`
	UseAltSpeedLimits    *bool  `json:"use_alt_speed_limits"`
	UseSubcategories     *bool  `json:"use_subcategories"`
	WriteCacheOverload   string `json:"write_cache_overload"`
}

func (s *ServerState) partialUpdate(update *ServerState) {
	if update.AllTimeDownload != 0 {
		s.AllTimeDownload = update.AllTimeDownload
	}
	if update.AllTimeUpload != 0 {
		s.AllTimeUpload = update.AllTimeUpload
	}
	if update.AverageTimeQueue != 0 {
		s.AverageTimeQueue = update.AverageTimeQueue
	}
	if update.ConnectionStatus != "" {
		s.ConnectionStatus = update.ConnectionStatus
	}
	if update.DHTNodes != 0 {
		s.DHTNodes = update.DHTNodes
	}
	if update.DLInfoData != 0 {
		s.DLInfoData = update.DLInfoData
	}
	if update.DLInfoSpeed != 0 {
		s.DLInfoSpeed = update.DLInfoSpeed
	}
	if update.DLRateLimit != 0 {
		s.DLRateLimit = update.DLRateLimit
	}
	if update.FreeSpaceOnDisk != 0 {
		s.FreeSpaceOnDisk = update.FreeSpaceOnDisk
	}
	if update.GlobalRatio != "" {
		s.GlobalRatio = update.GlobalRatio
	}
	if update.QueuedIOJobs != 0 {
		s.QueuedIOJobs = update.QueuedIOJobs
	}
	if update.Queueing != nil {
		s.Queueing = update.Queueing
	}
	if update.ReadCacheHits != "" {
		s.ReadCacheHits = update.ReadCacheHits
	}
	if update.ReadCacheOverload != "" {
		s.ReadCacheOverload = update.ReadCacheOverload
	}
	if update.RefreshInterval != 0 {
		s.RefreshInterval = update.RefreshInterval
	}
	if update.TotalBuffersSize != 0 {
		s.TotalBuffersSize = update.TotalBuffersSize
	}
	if update.TotalPeerConnections != 0 {
		s.TotalPeerConnections = update.TotalPeerConnections
	}
	if update.TotalQueuedSize != 0 {
		s.TotalQueuedSize = update.TotalQueuedSize
	}
	if update.TotalWastedSession != 0 {
		s.TotalWastedSession = update.TotalWastedSession
	}
	if update.UpInfoData != 0 {
		s.UpInfoData = update.UpInfoData
	}
	if update.UpInfoSpeed != 0 {
		s.UpInfoSpeed = update.UpInfoSpeed
	}
	if update.UpRateLimit != 0 {
		s.UpRateLimit = update.UpRateLimit
	}
	if update.UseAltSpeedLimits != nil {
		s.UseAltSpeedLimits = update.UseAltSpeedLimits
	}
	if update.UseSubcategories != nil {
		s.UseSubcategories = update.UseSubcategories
	}
	if update.WriteCacheOverload != "" {
		s.WriteCacheOverload = update.WriteCacheOverload
	}
}

func (s *ServerState) toFieldMap() map[string]interface{} {
	return map[string]interface{}{
		"all_time_download":      s.AllTimeDownload,
		"all_time_upload":        s.AllTimeUpload,
		"average_time_queue":     s.AverageTimeQueue,
		"connection_status":      s.ConnectionStatus,
		"dht_nodes":              s.DHTNodes,
		"dl_info_data":           s.DLInfoData,
		"dl_info_speed":          s.DLInfoSpeed,
		"dl_rate_limit":          s.DLRateLimit,
		"free_space_on_disk":     s.FreeSpaceOnDisk,
		"global_ratio":           s.GlobalRatio,
		"queued_io_jobs":         s.QueuedIOJobs,
		"queueing":               s.Queueing,
		"read_cache_hits":        s.ReadCacheHits,
		"read_cache_overload":    s.ReadCacheOverload,
		"refresh_interval":       s.RefreshInterval,
		"total_buffers_size":     s.TotalBuffersSize,
		"total_peer_connections": s.TotalPeerConnections,
		"total_queued_size":      s.TotalQueuedSize,
		"total_wasted_session":   s.TotalWastedSession,
		"up_info_data":           s.UpInfoData,
		"up_info_speed":          s.UpInfoSpeed,
		"up_rate_limit":          s.UpRateLimit,
		"use_alt_speed_limits":   s.UseAltSpeedLimits,
		"use_subcategories":      s.UseSubcategories,
		"write_cache_overload":   s.WriteCacheOverload,
	}
}

type Torrent struct {
	AddedOn                  int64   `json:"added_on"`
	AmountLeft               int64   `json:"amount_left"`
	AutoTMM                  *bool   `json:"auto_tmm"`
	Availability             int64   `json:"availability"`
	Category                 string  `json:"category"`
	Completed                int64   `json:"completed"`
	CompletionOn             int64   `json:"completion_on"`
	ContentPath              string  `json:"content_path"`
	DownloadLimit            int64   `json:"dl_limit"`
	DownloadSpeed            int64   `json:"dlspeed"`
	DownloadPath             string  `json:"download_path"`
	Downloaded               int64   `json:"downloaded"`
	DownloadedSession        int64   `json:"downloaded_session"`
	ETA                      int64   `json:"eta"`
	FLPiecePrio              *bool   `json:"f_l_piece_prio"`
	ForceStart               *bool   `json:"force_start"`
	InactiveSeedingTimeLimit int64   `json:"inactive_seeding_time_limit"`
	InfohashV1               string  `json:"infohash_v1"`
	InfohashV2               string  `json:"infohash_v2"`
	LastActivity             int64   `json:"last_activity"`
	MagnetURI                string  `json:"magnet_uri"`
	MaxInactiveSeedingTime   int64   `json:"max_inactive_seeding_time"`
	MaxRatio                 int64   `json:"max_ratio"`
	MaxSeedingTime           int64   `json:"max_seeding_time"`
	Name                     string  `json:"name"`
	NumComplete              int64   `json:"num_complete"`
	NumIncomplete            int64   `json:"num_incomplete"`
	NumLeechs                int64   `json:"num_leechs"`
	NumSeeds                 int64   `json:"num_seeds"`
	Priority                 int64   `json:"priority"`
	Progress                 float64 `json:"progress"`
	Ratio                    float64 `json:"ratio"`
	RatioLimit               int64   `json:"ratio_limit"`
	SavePath                 string  `json:"save_path"`
	SeedingTime              int64   `json:"seeding_time"`
	SeedingTimeLimit         int64   `json:"seeding_time_limit"`
	SeenComplete             int64   `json:"seen_complete"`
	SeqDownload              *bool   `json:"seq_dl"`
	Size                     int64   `json:"size"`
	State                    string  `json:"state"`
	SuperSeeding             *bool   `json:"super_seeding"`
	Tags                     string  `json:"tags"`
	TimeActive               int64   `json:"time_active"`
	TotalSize                int64   `json:"total_size"`
	Tracker                  string  `json:"tracker"`
	TrackersCount            int64   `json:"trackers_count"`
	UPLimit                  int64   `json:"up_limit"`
	Uploaded                 int64   `json:"uploaded"`
	UploadedSession          int64   `json:"uploaded_session"`
	UPSpeed                  int64   `json:"upspeed"`
}

func (t *Torrent) partialUpdate(update *Torrent) {
	if update.AddedOn != 0 {
		t.AddedOn = update.AddedOn
	}
	if update.AmountLeft != 0 {
		t.AmountLeft = update.AmountLeft
	}
	if update.AutoTMM != nil {
		t.AutoTMM = update.AutoTMM
	}
	if update.Availability != 0 {
		t.Availability = update.Availability
	}
	if update.Category != "" {
		t.Category = update.Category
	}
	if update.Completed != 0 {
		t.Completed = update.Completed
	}
	if update.CompletionOn != 0 {
		t.CompletionOn = update.CompletionOn
	}
	if update.ContentPath != "" {
		t.ContentPath = update.ContentPath
	}
	if update.DownloadLimit != 0 {
		t.DownloadLimit = update.DownloadLimit
	}
	if update.DownloadSpeed != 0 {
		t.DownloadSpeed = update.DownloadSpeed
	}
	if update.DownloadPath != "" {
		t.DownloadPath = update.DownloadPath
	}
	if update.Downloaded != 0 {
		t.Downloaded = update.Downloaded
	}
	if update.DownloadedSession != 0 {
		t.DownloadedSession = update.DownloadedSession
	}
	if update.ETA != 0 {
		t.ETA = update.ETA
	}
	if update.FLPiecePrio != nil {
		t.FLPiecePrio = update.FLPiecePrio
	}
	if update.ForceStart != nil {
		t.ForceStart = update.ForceStart
	}
	if update.InactiveSeedingTimeLimit != 0 {
		t.InactiveSeedingTimeLimit = update.InactiveSeedingTimeLimit
	}
	if update.InfohashV1 != "" {
		t.InfohashV1 = update.InfohashV1
	}
	if update.InfohashV2 != "" {
		t.InfohashV2 = update.InfohashV2
	}
	if update.LastActivity != 0 {
		t.LastActivity = update.LastActivity
	}
	if update.MagnetURI != "" {
		t.MagnetURI = update.MagnetURI
	}
	if update.MaxInactiveSeedingTime != 0 {
		t.MaxInactiveSeedingTime = update.MaxInactiveSeedingTime
	}
	if update.MaxRatio != 0 {
		t.MaxRatio = update.MaxRatio
	}
	if update.MaxSeedingTime != 0 {
		t.MaxSeedingTime = update.MaxSeedingTime
	}
	if update.Name != "" {
		t.Name = update.Name
	}
	if update.NumComplete != 0 {
		t.NumComplete = update.NumComplete
	}
	if update.NumIncomplete != 0 {
		t.NumIncomplete = update.NumIncomplete
	}
	if update.NumLeechs != 0 {
		t.NumLeechs = update.NumLeechs
	}
	if update.NumSeeds != 0 {
		t.NumSeeds = update.NumSeeds
	}
	if update.Priority != 0 {
		t.Priority = update.Priority
	}
	if update.Progress != 0 {
		t.Progress = update.Progress
	}
	if update.Ratio != 0 {
		t.Ratio = update.Ratio
	}
	if update.RatioLimit != 0 {
		t.RatioLimit = update.RatioLimit
	}
	if update.SavePath != "" {
		t.SavePath = update.SavePath
	}
	if update.SeedingTime != 0 {
		t.SeedingTime = update.SeedingTime
	}
	if update.SeedingTimeLimit != 0 {
		t.SeedingTimeLimit = update.SeedingTimeLimit
	}
	if update.SeenComplete != 0 {
		t.SeenComplete = update.SeenComplete
	}
	if update.SeqDownload != nil {
		t.SeqDownload = update.SeqDownload
	}
	if update.Size != 0 {
		t.Size = update.Size
	}
	if update.State != "" {
		t.State = update.State
	}
	if update.SuperSeeding != nil {
		t.SuperSeeding = update.SuperSeeding
	}
	if update.Tags != "" {
		t.Tags = update.Tags
	}
	if update.TimeActive != 0 {
		t.TimeActive = update.TimeActive
	}
	if update.TotalSize != 0 {
		t.TotalSize = update.TotalSize
	}
	if update.Tracker != "" {
		t.Tracker = update.Tracker
	}
	if update.TrackersCount != 0 {
		t.TrackersCount = update.TrackersCount
	}
	if update.UPLimit != 0 {
		t.UPLimit = update.UPLimit
	}
	if update.Uploaded != 0 {
		t.Uploaded = update.Uploaded
	}
	if update.UploadedSession != 0 {
		t.UploadedSession = update.UploadedSession
	}
	if update.UPSpeed != 0 {
		t.UPSpeed = update.UPSpeed
	}
}

func (t *Torrent) toFieldMap() map[string]interface{} {
	return map[string]interface{}{
		"added_on":                    t.AddedOn,
		"amount_left":                 t.AmountLeft,
		"availability":                t.Availability,
		"completed":                   t.Completed,
		"completion_on":               t.CompletionOn,
		"download_limit":              t.DownloadLimit,
		"download_speed":              t.DownloadSpeed,
		"downloaded":                  t.Downloaded,
		"downloaded_session":          t.DownloadedSession,
		"eta":                         t.ETA,
		"inactive_seeding_time_limit": t.InactiveSeedingTimeLimit,
		"last_activity":               t.LastActivity,
		"max_inactive_seeding_time":   t.MaxInactiveSeedingTime,
		"max_ratio":                   t.MaxRatio,
		"max_seeding_time":            t.MaxSeedingTime,
		"num_complete":                t.NumComplete,
		"num_incomplete":              t.NumIncomplete,
		"num_leechs":                  t.NumLeechs,
		"num_seeds":                   t.NumSeeds,
		"priority":                    t.Priority,
		"progress":                    t.Progress,
		"ratio":                       t.Ratio,
		"ratio_limit":                 t.RatioLimit,
		"seeding_time":                t.SeedingTime,
		"seeding_time_limit":          t.SeedingTimeLimit,
		"seen_complete":               t.SeenComplete,
		"size":                        t.Size,
		"time_active":                 t.TimeActive,
		"total_size":                  t.TotalSize,
		"trackers_count":              t.TrackersCount,
		"up_limit":                    t.UPLimit,
		"uploaded":                    t.Uploaded,
		"uploaded_session":            t.UploadedSession,
		"upspeed":                     t.UPSpeed,
	}
}
func (t Torrent) toTagsMap() map[string]string {
	result := map[string]string{
		"category":      t.Category,
		"auto_tmm":      strconv.FormatBool(*t.AutoTMM),
		"content_path":  t.ContentPath,
		"download_path": t.DownloadPath,
		"fl_piece_prio": strconv.FormatBool(*t.FLPiecePrio),
		"force_start":   strconv.FormatBool(*t.ForceStart),
		"infohash_v1":   t.InfohashV1,
		"infohash_v2":   t.InfohashV2,
		"magnet_uri":    t.MagnetURI,
		"name":          t.Name,
		"seq_download":  strconv.FormatBool(*t.SeqDownload),
		"super_seeding": strconv.FormatBool(*t.SuperSeeding),
		"save_path":     t.SavePath,
		"state":         t.State,
		"tags":          t.Tags,
		"tracker":       t.Tracker,
	}
	return result
}

type MainData struct {
	Categories        map[string]Category `json:"categories"`
	CategoriesRemoved []string            `json:"categories_removed"`
	FullUpdate        *bool               `json:"full_update"`
	RID               int16               `json:"rid"`
	ServerState       ServerState         `json:"server_state"`
	Tags              []string            `json:"tags"`
	TagsRemoved       []string            `json:"tags_removed"`
	Torrents          map[string]Torrent  `json:"torrents"`
	TorrentsRemoved   []string            `json:"torrents_removed"`
	Trackers          map[string][]string `json:"trackers"`
}

func (m *MainData) partialUpdate(update *MainData) {
	m.RID = update.RID
	for k, v := range update.Categories {
		category, exists := m.Categories[k]
		if exists {
			localV := v
			(&category).partialUpdate(&localV)
			m.Categories[k] = category
		} else {
			m.Categories[k] = v
		}
	}
	for _, v := range update.CategoriesRemoved {
		delete(m.Categories, v)
	}
	m.ServerState.partialUpdate(&update.ServerState)

	m.Tags = append(m.Tags, update.Tags...)

	if len(m.TagsRemoved) > 0 {
		var removeTagsResult []string
		for _, tag := range m.Tags {
			found := false
			for _, removedTag := range m.TagsRemoved {
				if tag == removedTag {
					found = true
					break
				}
			}

			if !found {
				removeTagsResult = append(removeTagsResult, tag)
			}
		}
		m.Tags = removeTagsResult
	}

	for k, v := range update.Torrents {
		torrent, exists := m.Torrents[k]
		if exists {
			localV := v
			torrent.partialUpdate(&localV)
			m.Torrents[k] = torrent
		} else {
			m.Torrents[k] = v
		}
	}
	for _, v := range update.TorrentsRemoved {
		delete(m.Torrents, v)
	}
	for k, v := range update.Trackers {
		m.Trackers[k] = v
	}
}

func (m *MainData) toMetrics() map[string][]telegraf.Metric {
	ts := time.Now().UTC()
	tags := make(map[string]string)

	var serverStateMetrics []telegraf.Metric
	serverStateMetrics = append(serverStateMetrics, metric.New("server_state", tags, m.ServerState.toFieldMap(), ts, telegraf.Gauge))

	var torrentsMetrics = make([]telegraf.Metric, 0, len(m.Torrents))
	for k, v := range m.Torrents {
		torrentTag := m.Torrents[k].toTagsMap()
		torrentTag["hash"] = k
		torrentsMetrics = append(torrentsMetrics, metric.New("torrents", torrentTag, v.toFieldMap(), ts, telegraf.Gauge))
	}

	var tagsMetrics []telegraf.Metric
	tagsMetrics = append(tagsMetrics, metric.New("tags", tags, map[string]interface{}{"count": len(m.Tags)}, ts, telegraf.Counter))

	var categoryMetrics []telegraf.Metric
	categoryMetrics = append(categoryMetrics, metric.New("category", tags, map[string]interface{}{"count": len(m.Categories)}, ts, telegraf.Counter))

	return map[string][]telegraf.Metric{
		"server_state": serverStateMetrics,
		"torrents":     torrentsMetrics,
		"tags":         tagsMetrics,
		"category":     categoryMetrics,
	}
}

type PartialUpdate[T any] interface {
	partialUpdate(*T)
}

type ToFieldMap interface {
	toFieldMap() map[string]interface{}
}
