package qbittorrent

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMainDataPartialUpdate(t *testing.T) {

	data := MainData{
		RID: 0,
		Categories: map[string]Category{
			"test": {
				Name:     "test",
				SavePath: "download/test",
			},
		},
		Torrents: map[string]Torrent{
			"test_torrent": {
				Name:              "test_torrent",
				InfohashV2:        "infohash_v2",
				InfohashV1:        "infohash_v1",
				MagnetURI:         "magnet_uri",
				DownloadPath:      "download_path",
				AddedOn:           1000,
				ETA:               1000,
				DownloadedSession: 1000,
			},
		},
		ServerState: ServerState{
			AllTimeDownload: 1000,
			AllTimeUpload:   1000,
		},
	}
	updateValue := MainData{
		RID: 0,
		Categories: map[string]Category{
			"update-test": {
				Name:     "update-test",
				SavePath: "download/test",
			},
		},
		Torrents: map[string]Torrent{
			"test_torrent": {
				DownloadedSession: 1100,
				ETA:               900,
			},
		},
		ServerState: ServerState{
			AllTimeDownload: 1100,
			AllTimeUpload:   1200,
		},
		Tags: []string{"tag1"},
	}

	var acc testutil.Accumulator

	for k, v := range data.toMetrics() {
		for i := range v {
			acc.AddFields(k, v[i].Fields(), v[i].Tags())
		}
	}

	require.True(t, acc.HasInt64Field("torrents", "added_on"))
	addedOnValue, _ := acc.Int64Field("torrents", "added_on")
	require.Equal(t, int64(1000), addedOnValue)
	etaValue, _ := acc.Int64Field("torrents", "eta")
	require.Equal(t, int64(1000), etaValue)
	downloadedSessionValue, _ := acc.Int64Field("torrents", "downloaded_session")
	require.Equal(t, int64(1000), downloadedSessionValue)

	allTimeDownloadValue, _ := acc.Int64Field("server_state", "all_time_download")
	require.Equal(t, int64(1000), allTimeDownloadValue)
	allTimeUploadValue, _ := acc.Int64Field("server_state", "all_time_upload")
	require.Equal(t, int64(1000), allTimeUploadValue)

	categoryCount, _ := acc.Int64Field("category", "count")
	require.Equal(t, int64(1), categoryCount)

	tagsCount, _ := acc.Int64Field("tags", "count")
	require.Equal(t, int64(0), tagsCount)

	var update testutil.Accumulator

	data.partialUpdate(&updateValue)

	for k, v := range data.toMetrics() {
		for i := range v {
			update.AddFields(k, v[i].Fields(), v[i].Tags())
		}
	}

	updateAddedOnValue, _ := acc.Int64Field("torrents", "added_on")
	require.Equal(t, int64(1000), updateAddedOnValue)
	updateEtaValue, _ := update.Int64Field("torrents", "eta")
	require.Equal(t, int64(900), updateEtaValue)
	updateDownloadedSessionValue, _ := update.Int64Field("torrents", "downloaded_session")
	require.Equal(t, int64(1100), updateDownloadedSessionValue)

	updateAllTimeDownloadValue, _ := update.Int64Field("server_state", "all_time_download")
	require.Equal(t, int64(1100), updateAllTimeDownloadValue)
	updateAllTimeUploadValue, _ := update.Int64Field("server_state", "all_time_upload")
	require.Equal(t, int64(1200), updateAllTimeUploadValue)

	updateCategoryCount, _ := update.Int64Field("category", "count")
	require.Equal(t, int64(2), updateCategoryCount)

	updateTagsCount, _ := update.Int64Field("tags", "count")
	require.Equal(t, int64(1), updateTagsCount)
	require.True(t, update.HasTag("torrents", "name"))
}
