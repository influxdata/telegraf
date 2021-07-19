package syncthing

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Syncthing - plugin main structure
type Syncthing struct {
	Address     string
	Token       string
	HTTPTimeout internal.Duration `toml:"http_timeout"`
	client      *http.Client
}

const sampleConfig = `
  ## Address to Syncthing
  address = "http://localhost:8384"

  ## Token/API key
  token = "1234asdf"

  ## Timeout for HTTP requests.
  # http_timeout = "5s"
`

// SampleConfig returns sample configuration for this plugin.
func (s *Syncthing) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (s *Syncthing) Description() string {
	return "Gather data and statistics from Syncthing."
}

// Create Syncthing Client
func (s *Syncthing) createSyncthingClient(ctx context.Context) (*http.Client, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: s.HTTPTimeout.Duration,
	}

	return httpClient, nil
}

func (s *Syncthing) get(endpoint string, responseData interface{}) error {
	req, err := http.NewRequest("GET", s.Address+endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", s.Token)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, responseData)
	if err != nil {
		return err
	}
	return nil
}

func (s *Syncthing) getServiceReport() (*Report, error) {
	var report Report
	err := s.get("/rest/svc/report", &report)
	if err != nil {
		return nil, err
	}

	return &report, err
}

func (s *Syncthing) getSystemStatus() (*SystemStatus, error) {
	var status SystemStatus
	err := s.get("/rest/system/status", &status)
	if err != nil {
		return nil, err
	}

	return &status, err
}

func (s *Syncthing) getFolderStatus(folderID string) (*FolderStatus, error) {
	var status FolderStatus
	url := "/rest/db/status?folder=" + folderID
	err := s.get(url, &status)
	if err != nil {
		return nil, err
	}

	return &status, err
}

func (s *Syncthing) getConfig() (*Config, error) {
	var config Config
	err := s.get("/rest/system/config", &config)
	if err != nil {
		return nil, err
	}

	return &config, err
}

// Gather Syncthing Metrics
func (s *Syncthing) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	if s.client == nil {
		syncthingClient, err := s.createSyncthingClient(ctx)

		if err != nil {
			return err
		}

		s.client = syncthingClient
	}

	// Fetch the configuration from Syncthing to know how which folders
	// we want to collect metrics for
	config, err := s.getConfig()
	if err != nil {
		acc.AddError(err)
	}

	// Create a WaitGroup for N folders + system metrics
	var wg sync.WaitGroup
	wg.Add(len(config.Folders) + 1)

	for _, folder := range config.Folders {
		go func(folder Folder, acc telegraf.Accumulator) {
			defer wg.Done()

			systemStatus, err := s.getSystemStatus()
			if err != nil {
				acc.AddError(err)
				return
			}

			folderStatus, err := s.getFolderStatus(folder.ID)
			if err != nil {
				acc.AddError(err)
				return
			}

			now := time.Now()
			tags := map[string]string{
				"instance": systemStatus.MyID,
				"folder":   folder.ID,
			}
			fields := getFolderFields(folderStatus)

			acc.AddFields("syncthing_folder", fields, tags, now)
		}(folder, acc)

	}

	go func(acc telegraf.Accumulator) {
		defer wg.Done()

		report, err := s.getServiceReport()
		if err != nil {
			acc.AddError(err)
			return
		}

		systemStatus, err := s.getSystemStatus()
		if err != nil {
			acc.AddError(err)
			return
		}

		now := time.Now()
		tags := map[string]string{
			"instance": systemStatus.MyID,
		}
		fields := getSystemFields(report, systemStatus)

		acc.AddFields("syncthing_system", fields, tags, now)
	}(acc)

	wg.Wait()
	return nil
}

func getSystemFields(report *Report, system *SystemStatus) map[string]interface{} {
	return map[string]interface{}{
		"folder_max_files": report.FolderMaxFiles,
		"folder_max_mib":   report.FolderMaxMiB,
		"memory_size":      report.MemorySize,
		"memory_usage_mib": report.MemoryUsageMiB,
		"num_cpu":          report.NumCPU,
		"num_devices":      report.NumDevices,
		"num_folders":      report.NumFolders,
		"total_files":      report.TotFiles,
		"total_mib":        report.TotMiB,
		"uptime_seconds":   report.Uptime,
		"alloc":            system.Alloc,
		"cpu_percent":      system.CPUPercent,
		"goroutines":       system.Goroutines,
	}
}

func getFolderFields(folderStatus *FolderStatus) map[string]interface{} {
	return map[string]interface{}{
		"errors":             folderStatus.Errors,
		"global_bytes":       folderStatus.GlobalBytes,
		"global_deleted":     folderStatus.GlobalDeleted,
		"global_directories": folderStatus.GlobalDirectories,
		"global_files":       folderStatus.GlobalFiles,
		"global_symlinks":    folderStatus.GlobalSymlinks,
		"global_total_items": folderStatus.GlobalTotalItems,
		"in_sync_bytes":      folderStatus.InSyncBytes,
		"in_sync_files":      folderStatus.InSyncFiles,
		"local_bytes":        folderStatus.LocalBytes,
		"local_deleted":      folderStatus.LocalDeleted,
		"local_directories":  folderStatus.LocalDirectories,
		"local_files":        folderStatus.LocalFiles,
		"local_symlinks":     folderStatus.LocalSymlinks,
		"local_total_items":  folderStatus.LocalTotalItems,
		"need_bytes":         folderStatus.NeedBytes,
		"need_deletes":       folderStatus.NeedDeletes,
		"need_directories":   folderStatus.NeedDirectories,
		"need_files":         folderStatus.NeedFiles,
		"need_symlinks":      folderStatus.NeedSymlinks,
		"need_total_items":   folderStatus.NeedTotalItems,
		"pull_errors":        folderStatus.PullErrors,
		"sequence":           folderStatus.Sequence,
		"version":            folderStatus.Version,
	}
}

func init() {
	inputs.Add("syncthing", func() telegraf.Input {
		return &Syncthing{
			HTTPTimeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
