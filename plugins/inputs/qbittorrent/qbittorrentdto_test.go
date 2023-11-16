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
			"test": Category{
				Name:     "test",
				SavePath: "download/test",
			},
		},
		Torrents: map[string]Torrent{
			"test_torrent": Torrent{
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

	update_value := MainData{
		RID: 0,
		Categories: map[string]Category{
			"update-test": Category{
				Name:     "update-test",
				SavePath: "download/test",
			},
		},
		Torrents: map[string]Torrent{
			"test_torrent": Torrent{
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
	require.True(t, addedOnValue == 1000)
	etaValue, _ := acc.Int64Field("torrents", "eta")
	require.True(t, etaValue == 1000)
	downloadedSessionValue, _ := acc.Int64Field("torrents", "downloaded_session")
	require.True(t, downloadedSessionValue == 1000)

	allTimeDownloadValue, _ := acc.Int64Field("server_state", "all_time_download")
	require.True(t, allTimeDownloadValue == 1000)
	allTimeUploadValue, _ := acc.Int64Field("server_state", "all_time_upload")
	require.True(t, allTimeUploadValue == 1000)

	categoryCount, _ := acc.Int64Field("category", "count")
	require.True(t, categoryCount == 1)

	tagsCount, _ := acc.Int64Field("tags", "count")
	require.True(t, tagsCount == 0)

	var update testutil.Accumulator

	data.partialUpdate(&update_value)

	for k, v := range data.toMetrics() {
		for i := range v {
			update.AddFields(k, v[i].Fields(), v[i].Tags())
		}
	}

	updateAddedOnValue, _ := acc.Int64Field("torrents", "added_on")
	require.True(t, updateAddedOnValue == 1000)
	updateEtaValue, _ := update.Int64Field("torrents", "eta")
	require.True(t, updateEtaValue == 900)
	updateDownloadedSessionValue, _ := update.Int64Field("torrents", "downloaded_session")
	require.True(t, updateDownloadedSessionValue == 1100)

	updateAllTimeDownloadValue, _ := update.Int64Field("server_state", "all_time_download")
	require.True(t, updateAllTimeDownloadValue == 1100)
	updateAllTimeUploadValue, _ := update.Int64Field("server_state", "all_time_upload")
	require.True(t, updateAllTimeUploadValue == 1200)

	updateCategoryCount, _ := update.Int64Field("category", "count")
	require.True(t, updateCategoryCount == 2)

	updateTagsCount, _ := update.Int64Field("tags", "count")
	require.True(t, updateTagsCount == 1)
	require.True(t, update.HasTag("torrents", "name"))
}
