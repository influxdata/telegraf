package qbittorrent

import (
	"fmt"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
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

	for _, m := range data.toMetrics("source1") {
		acc.AddMetric(m)
	}

	require.True(t, acc.HasTag("qbittorrent", "source"))
	require.True(t, acc.HasTag("torrent", "source"))

	require.True(t, acc.HasInt64Field("torrent", "added_on"))
	addedOnValue, _ := acc.Int64Field("torrent", "added_on")
	require.Equal(t, int64(1000), addedOnValue)
	etaValue, _ := acc.Int64Field("torrent", "eta")
	require.Equal(t, int64(1000), etaValue)
	downloadedSessionValue, _ := acc.Int64Field("torrent", "downloaded_session")
	require.Equal(t, int64(1000), downloadedSessionValue)

	allTimeDownloadValue, _ := acc.Int64Field("qbittorrent", "all_time_download")
	require.Equal(t, int64(1000), allTimeDownloadValue)
	allTimeUploadValue, _ := acc.Int64Field("qbittorrent", "all_time_upload")
	require.Equal(t, int64(1000), allTimeUploadValue)

	categoryCount, _ := acc.Int64Field("qbittorrent", "category_count")
	require.Equal(t, int64(1), categoryCount)

	tagsCount, _ := acc.Int64Field("qbittorrent", "tag_count")
	require.Equal(t, int64(0), tagsCount)

	var update testutil.Accumulator

	data.partialUpdate(&updateValue)

	for _, m := range data.toMetrics("source1") {
		update.AddMetric(m)
	}

	updateAddedOnValue, _ := acc.Int64Field("torrent", "added_on")
	require.Equal(t, int64(1000), updateAddedOnValue)
	updateEtaValue, _ := update.Int64Field("torrent", "eta")
	require.Equal(t, int64(900), updateEtaValue)
	updateDownloadedSessionValue, _ := update.Int64Field("torrent", "downloaded_session")
	require.Equal(t, int64(1100), updateDownloadedSessionValue)

	updateAllTimeDownloadValue, _ := update.Int64Field("qbittorrent", "all_time_download")
	require.Equal(t, int64(1100), updateAllTimeDownloadValue)
	updateAllTimeUploadValue, _ := update.Int64Field("qbittorrent", "all_time_upload")
	require.Equal(t, int64(1200), updateAllTimeUploadValue)

	updateCategoryCount, _ := update.Int64Field("qbittorrent", "category_count")
	require.Equal(t, int64(2), updateCategoryCount)

	updateTagsCount, _ := update.Int64Field("qbittorrent", "tag_count")
	require.Equal(t, int64(1), updateTagsCount)
	require.True(t, update.HasTag("torrent", "name"))
}

func TestGetMainData_URL(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/sync/maindata" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintln(w, mainDataJSON)
			require.NoError(t, err)
		}
	}))
	defer fakeServer.Close()

	plugin := &QBittorrent{
		URL: fakeServer.URL,
		cookie: []*http.Cookie{{
			Name:  "cookieName",
			Value: "cookieValue",
		}},
	}

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	require.Len(t, acc.Metrics, 2)

	categoriesCount, _ := acc.Int64Field("qbittorrent", "category_count")
	require.Equal(t, int64(1), categoriesCount)
	tagsCount, _ := acc.Int64Field("qbittorrent", "tag_count")
	require.Equal(t, int64(1), tagsCount)
	freeSpaceOnDisk, _ := acc.Int64Field("qbittorrent", "free_space_on_disk")
	require.Equal(t, int64(1921000090000), freeSpaceOnDisk)
	dlInfoData, _ := acc.Int64Field("qbittorrent", "dl_info_data")
	require.Equal(t, int64(1594136974832), dlInfoData)
	upInfoData, _ := acc.Int64Field("qbittorrent", "up_info_data")
	require.Equal(t, int64(300010001000), upInfoData)
	totalBuffersSize, _ := acc.Int64Field("qbittorrent", "total_buffers_size")
	require.Equal(t, int64(524288), totalBuffersSize)

	addedOnValue, _ := acc.Int64Field("torrent", "added_on")
	require.Equal(t, int64(1701232564), addedOnValue)
	etaValue, _ := acc.Int64Field("torrent", "eta")
	require.Equal(t, int64(8640000), etaValue)
	downloadedSessionValue, _ := acc.Int64Field("torrent", "downloaded_session")
	require.Equal(t, int64(483696196), downloadedSessionValue)

	require.Equal(t, "003fyd3cgjsdf6f8debc4273s4g80bhgb25hs487", acc.TagValue("torrent", "hash"))
	require.Equal(t, "https://xxxxx/announce.php?passkey=xxxxx", acc.TagValue("torrent", "tracker"))
	require.Equal(t, "/downloads/c1/torrent1", acc.TagValue("torrent", "content_path"))
}

const mainDataJSON = `
{
    "categories": {
        "c1": {
            "name": "c1",
            "savePath": "/downloads/c1"
        }
    },
    "full_update": true,
    "rid": 1,
    "server_state": {
        "alltime_dl": 34798703722659,
        "alltime_ul": 11180426792341,
        "average_time_queue": 360,
        "connection_status": "connected",
        "dht_nodes": 360,
        "dl_info_data": 1594136974832,
        "dl_info_speed": 0,
        "dl_rate_limit": 0,
        "free_space_on_disk": 1921000090000,
        "global_ratio": "1.0",
        "queued_io_jobs": 0,
        "queueing": true,
        "read_cache_hits": "0",
        "read_cache_overload": "0",
        "refresh_interval": 1500,
        "total_buffers_size": 524288,
        "total_peer_connections": 25,
        "total_queued_size": 0,
        "total_wasted_session": 1208549226,
        "up_info_data": 300010001000,
        "up_info_speed": 0,
        "up_rate_limit": 0,
        "use_alt_speed_limits": false,
        "use_subcategories": false,
        "write_cache_overload": "0"
    },
    "tags": [
        "tag1"
    ],
    "torrents": {
		"003fyd3cgjsdf6f8debc4273s4g80bhgb25hs487": {
            "added_on": 1701232564,
            "amount_left": 0,
            "auto_tmm": false,
            "availability": -1,
            "category": "",
            "completed": 483344047,
            "completion_on": 1701232687,
            "content_path": "/downloads/c1/torrent1",
            "dl_limit": 0,
            "dlspeed": 0,
            "download_path": "",
            "downloaded": 483696196,
            "downloaded_session": 483696196,
            "eta": 8640000,
            "f_l_piece_prio": false,
            "force_start": false,
            "inactive_seeding_time_limit": -2,
            "infohash_v1": "003fyd3cgjsdf6f8debc4273s4g80bhgb25hs487",
            "infohash_v2": "",
            "last_activity": 1701361789,
            "magnet_uri": "magnet:?xt=urn:btih:xxxxxxxxxxx",
            "max_inactive_seeding_time": -1,
            "max_ratio": -1,
            "max_seeding_time": -1,
            "name": "torrent1",
            "num_complete": 41,
            "num_incomplete": 88,
            "num_leechs": 0,
            "num_seeds": 0,
            "priority": 0,
            "progress": 1,
            "ratio": 0.5990,
            "ratio_limit": -2,
            "save_path": "/downloads/c1",
            "seeding_time": 182266,
            "seeding_time_limit": -2,
            "seen_complete": 1701236311,
            "seq_dl": false,
            "size": 483344047,
            "state": "stalledUP",
            "super_seeding": false,
            "tags": "tag1",
            "time_active": 182388,
            "total_size": 483344047,
            "tracker": "https://xxxxx/announce.php?passkey=xxxxx",
            "trackers_count": 1,
            "up_limit": 0,
            "uploaded": 241377164,
            "uploaded_session": 241377164,
            "upspeed": 0
        }
	}
}`
