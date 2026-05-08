package docker

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/system"
	"github.com/google/go-cmp/cmp"
	moby_container "github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/common/docker/mock"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestInit(t *testing.T) {
	plugin := &Docker{
		Log:              testutil.Logger{},
		PerDeviceInclude: []string{"cpu", "network", "blkio"},
		TotalInclude:     []string{"cpu", "network", "blkio"},
	}
	require.NoError(t, plugin.Init())
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name      string
		perDevice []string
		total     []string
		expected  string
	}{
		{
			name:      "unsupported perdevice_include",
			perDevice: []string{"nonExistentClass"},
			total:     []string{"cpu"},
			expected:  "unknown choice nonExistentClass",
		},
		{
			name:      "unsupported total_include",
			perDevice: []string{"cpu"},
			total:     []string{"nonExistentClass"},
			expected:  "unknown choice nonExistentClass",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Docker{
				PerDeviceInclude: tt.perDevice,
				TotalInclude:     tt.total,
				Log:              testutil.Logger{},
			}
			require.ErrorContains(t, plugin.Init(), tt.expected)
		})
	}
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("docker", func() telegraf.Input {
		return &Docker{
			Endpoint:         "dummy",
			PerDeviceInclude: []string{"cpu"},
			TotalInclude:     []string{"cpu", "blkio", "network"},
			Timeout:          config.Duration(time.Second * 5),
			PodmanCacheTTL:   config.Duration(60 * time.Second),
			newEnvClient:     newEnvClient,
			newClient:        newClient,
		}
	})

	// Prepare the influx parser for expectations
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	// Comparison options
	options := []cmp.Option{
		testutil.SortMetrics(),
		// We need to ignore all time related field values
		testutil.IgnoreTime(),
		testutil.IgnoreFields("uptime_ns"),
	}

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			configFilename := filepath.Join(testcasePath, "telegraf.conf")
			expectedFilename := filepath.Join(testcasePath, "expected.out")
			expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Read the expected output if any
			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				var err error
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			// Setup the server
			server, err := mock.NewServerFromFiles(filepath.Join(testcasePath, "data"))
			require.NoError(t, err)
			server.APIVersion = "1.24"

			addr := server.Start(t)
			defer server.Close()

			// Configure and initialize the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			plugin := cfg.Inputs[0].Input.(*Docker)
			plugin.Endpoint = addr
			require.NoError(t, plugin.Init())

			// Start the plugin
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect data and test the result
			require.NoError(t, acc.GatherError(plugin.Gather))
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), options...)
		})
	}
}

func TestContainerLabels(t *testing.T) {
	var tests = []struct {
		name     string
		labels   map[string]string
		include  []string
		exclude  []string
		expected map[string]string
	}{
		{
			name: "nil filters matches all",
			labels: map[string]string{
				"a": "x",
			},
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "empty filters matches all",
			labels: map[string]string{
				"a": "x",
			},
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "must match include",
			labels: map[string]string{
				"a": "x",
				"b": "y",
			},
			include: []string{"a"},
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "must not match exclude",
			labels: map[string]string{
				"a": "x",
				"b": "y",
			},
			exclude: []string{"b"},
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "include glob",
			labels: map[string]string{
				"aa": "x",
				"ab": "y",
				"bb": "z",
			},
			include: []string{"a*"},
			expected: map[string]string{
				"aa": "x",
				"ab": "y",
			},
		},
		{
			name: "exclude glob",
			labels: map[string]string{
				"aa": "x",
				"ab": "y",
				"bb": "z",
			},
			exclude: []string{"a*"},
			expected: map[string]string{
				"bb": "z",
			},
		},
		{
			name: "excluded and includes",
			labels: map[string]string{
				"aa": "x",
				"ab": "y",
				"bb": "z",
			},
			include: []string{"a*"},
			exclude: []string{"*b"},
			expected: map[string]string{
				"aa": "x",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the server
			server, err := mock.NewServerFromFiles("testdata")
			require.NoError(t, err)
			server.APIVersion = "1.24"

			// Manipulate the data for the test
			c := server.List[0]
			c.Labels = tt.labels
			c.State = "running"
			server.List = []moby_container.Summary{c}

			addr := server.Start(t)
			defer server.Close()

			// Setup plugin
			plugin := &Docker{
				Endpoint:     addr,
				LabelInclude: tt.include,
				LabelExclude: tt.exclude,
				TotalInclude: []string{"cpu"},
				Timeout:      config.Duration(time.Second * 5),
				Log:          testutil.Logger{},
				newClient:    newClient,
				newEnvClient: newEnvClient,
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect data and check result
			require.NoError(t, acc.GatherError(plugin.Gather))
			var actual map[string]string
			for _, mt := range acc.Metrics {
				if mt.Measurement == "docker_container_cpu" {
					actual = mt.Tags
					break
				}
			}
			require.Subset(t, actual, tt.expected)
		})
	}
}

func TestContainerNames(t *testing.T) {
	var tests = []struct {
		name       string
		containers [][]string
		include    []string
		exclude    []string
		expected   []string
	}{
		{
			name:     "nil filters matches all",
			expected: []string{"etcd", "etcd2", "acme", "acme-test", "foo"},
		},
		{
			name:     "empty filters matches all",
			expected: []string{"etcd", "etcd2", "acme", "acme-test", "foo"},
		},
		{
			name:     "match all containers",
			include:  []string{"*"},
			expected: []string{"etcd", "etcd2", "acme", "acme-test", "foo"},
		},
		{
			name:     "include prefix match",
			include:  []string{"etc*"},
			expected: []string{"etcd", "etcd2"},
		},
		{
			name:     "exact match",
			include:  []string{"etcd"},
			expected: []string{"etcd"},
		},
		{
			name:     "star matches zero length",
			include:  []string{"etcd2*"},
			expected: []string{"etcd2"},
		},
		{
			name:     "exclude matches all",
			exclude:  []string{"etc*"},
			expected: []string{"acme", "acme-test", "foo"},
		},
		{
			name:     "exclude single",
			exclude:  []string{"etcd"},
			expected: []string{"etcd2", "acme", "acme-test", "foo"},
		},
		{
			name:    "exclude all",
			include: []string{"*"},
			exclude: []string{"*"},
		},
		{
			name:     "exclude item matching include",
			include:  []string{"acme*"},
			exclude:  []string{"*test*"},
			expected: []string{"acme"},
		},
		{
			name:     "exclude item no wildcards",
			include:  []string{"acme*"},
			exclude:  []string{"test"},
			expected: []string{"acme", "acme-test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the server
			server, err := mock.NewServerFromFiles("testdata")
			require.NoError(t, err)
			server.APIVersion = "1.24"

			addr := server.Start(t)
			defer server.Close()

			// Setup plugin
			plugin := &Docker{
				Endpoint:         addr,
				ContainerInclude: tt.include,
				ContainerExclude: tt.exclude,
				Timeout:          config.Duration(time.Second * 5),
				Log:              testutil.Logger{},
				newClient:        newClient,
				newEnvClient:     newEnvClient,
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect data and check the results
			require.NoError(t, acc.GatherError(plugin.Gather))
			actual := make([]string, 0)
			for _, mt := range acc.Metrics {
				if name, ok := mt.Tags["container_name"]; ok {
					actual = append(actual, name)
				}
			}
			require.Subset(t, tt.expected, actual)
		})
	}
}

func TestContainerStatus(t *testing.T) {
	var tests = []struct {
		name     string
		now      time.Time
		started  *string
		finished *string
		expected []telegraf.Metric
	}{
		{
			name: "finished_at is zero value",
			now:  time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			expected: []telegraf.Metric{
				metric.New(
					"docker_container_status",
					map[string]string{
						"container_name":    "etcd",
						"container_image":   "quay.io/coreos/etcd",
						"container_version": "v3.3.25",
						"engine_host":       "absol",
						"label1":            "test_value_1",
						"label2":            "test_value_2",
						"server_version":    "17.09.0-ce",
						"container_status":  "running",
						"source":            "e2173b9478a6",
					},
					map[string]interface{}{
						"oomkilled":     false,
						"pid":           1234,
						"restart_count": 0,
						"exitcode":      0,
						"container_id":  "e2173b9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296b7dfb",
						"started_at":    time.Date(2018, 6, 14, 5, 48, 53, 266176036, time.UTC).UnixNano(),
						"uptime_ns":     int64(3 * time.Minute),
					},
					time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
				),
			},
		},
		{
			name:     "finished_at is non-zero value",
			now:      time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			finished: new("2018-06-14T05:53:53.266176036Z"),
			expected: []telegraf.Metric{
				metric.New(
					"docker_container_status",
					map[string]string{
						"container_name":    "etcd",
						"container_image":   "quay.io/coreos/etcd",
						"container_version": "v3.3.25",
						"engine_host":       "absol",
						"label1":            "test_value_1",
						"label2":            "test_value_2",
						"server_version":    "17.09.0-ce",
						"container_status":  "running",
						"source":            "e2173b9478a6",
					},
					map[string]interface{}{
						"oomkilled":     false,
						"pid":           1234,
						"exitcode":      0,
						"restart_count": 0,
						"container_id":  "e2173b9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296b7dfb",
						"started_at":    time.Date(2018, 6, 14, 5, 48, 53, 266176036, time.UTC).UnixNano(),
						"finished_at":   time.Date(2018, 6, 14, 5, 53, 53, 266176036, time.UTC).UnixNano(),
						"uptime_ns":     int64(5 * time.Minute),
					},
					time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
				),
			},
		},
		{
			name:     "started_at is zero value",
			now:      time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			started:  new(""),
			finished: new("2018-06-14T05:53:53.266176036Z"),
			expected: []telegraf.Metric{
				metric.New(
					"docker_container_status",
					map[string]string{
						"container_name":    "etcd",
						"container_image":   "quay.io/coreos/etcd",
						"container_version": "v3.3.25",
						"engine_host":       "absol",
						"label1":            "test_value_1",
						"label2":            "test_value_2",
						"server_version":    "17.09.0-ce",
						"container_status":  "running",
						"source":            "e2173b9478a6",
					},
					map[string]interface{}{
						"oomkilled":     false,
						"pid":           1234,
						"exitcode":      0,
						"restart_count": 0,
						"container_id":  "e2173b9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296b7dfb",
						"finished_at":   time.Date(2018, 6, 14, 5, 53, 53, 266176036, time.UTC).UnixNano(),
					},
					time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
				),
			},
		},
		{
			name:     "container has been restarted",
			now:      time.Date(2019, 1, 1, 0, 0, 3, 0, time.UTC),
			started:  new("2019-01-01T00:00:02Z"),
			finished: new("2019-01-01T00:00:01Z"),
			expected: []telegraf.Metric{
				metric.New(
					"docker_container_status",
					map[string]string{
						"container_name":    "etcd",
						"container_image":   "quay.io/coreos/etcd",
						"container_version": "v3.3.25",
						"engine_host":       "absol",
						"label1":            "test_value_1",
						"label2":            "test_value_2",
						"server_version":    "17.09.0-ce",
						"container_status":  "running",
						"source":            "e2173b9478a6",
					},
					map[string]interface{}{
						"oomkilled":     false,
						"pid":           1234,
						"exitcode":      0,
						"restart_count": 0,
						"container_id":  "e2173b9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296b7dfb",
						"started_at":    time.Date(2019, 1, 1, 0, 0, 2, 0, time.UTC).UnixNano(),
						"finished_at":   time.Date(2019, 1, 1, 0, 0, 1, 0, time.UTC).UnixNano(),
						"uptime_ns":     int64(1 * time.Second),
					},
					time.Date(2019, 1, 1, 0, 0, 3, 0, time.UTC),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the time
			now = func() time.Time { return tt.now }
			defer func() { now = time.Now }()

			// Setup the server
			server, err := mock.NewServerFromFiles("testdata")
			require.NoError(t, err)
			server.APIVersion = "1.24"

			// Manipulate data for the test
			server.List = server.List[:1]
			if tt.started != nil {
				server.Inspect[server.List[0].ID].State.StartedAt = *tt.started
			}
			if tt.finished != nil {
				server.Inspect[server.List[0].ID].State.FinishedAt = *tt.finished
			}

			addr := server.Start(t)
			defer server.Close()

			// Setup plugin
			plugin := &Docker{
				Endpoint:         addr,
				IncludeSourceTag: true,
				Timeout:          config.Duration(time.Second * 5),
				Log:              testutil.Logger{},
				newClient:        newClient,
				newEnvClient:     newEnvClient,
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect data and check the result
			require.NoError(t, acc.GatherError(plugin.Gather))
			testutil.RequireMetricsSubset(t, tt.expected, acc.GetTelegrafMetrics())
		})
	}
}

func TestGatherInfo(t *testing.T) {
	// Define expected result
	expected := []telegraf.Metric{
		metric.New(
			"docker",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"n_listener_events":       int(0),
				"n_cpus":                  int(4),
				"n_used_file_descriptors": int(19),
				"n_containers":            int(108),
				"n_containers_running":    int(98),
				"n_containers_stopped":    int(6),
				"n_containers_paused":     int(3),
				"n_images":                int(199),
				"n_goroutines":            int(39),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"memory_total": int64(3840757760),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"unit":           "bytes",
			},
			map[string]interface{}{
				"pool_blocksize": int64(65540),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_data",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"unit":           "bytes",
			},
			map[string]interface{}{
				"used":      int64(17300000000),
				"total":     int64(107400000000),
				"available": int64(36530000000),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_metadata",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"unit":           "bytes",
			},
			map[string]interface{}{
				"used":      int64(20970000),
				"total":     int64(2146999999),
				"available": int64(2126999999),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_devicemapper",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"pool_name":      "docker-8:1-1182287-pool",
			},
			map[string]interface{}{
				"base_device_size_bytes":             int64(10740000000),
				"pool_blocksize_bytes":               int64(65540),
				"data_space_used_bytes":              int64(17300000000),
				"data_space_total_bytes":             int64(107400000000),
				"data_space_available_bytes":         int64(36530000000),
				"metadata_space_used_bytes":          int64(20970000),
				"metadata_space_total_bytes":         int64(2146999999),
				"metadata_space_available_bytes":     int64(2126999999),
				"thin_pool_minimum_free_space_bytes": int64(10740000000),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_cpu",
			map[string]string{
				"container_name":    "etcd2",
				"container_image":   "quay.io:4443/coreos/etcd",
				"cpu":               "cpu3",
				"container_version": "v3.3.25",
				"engine_host":       "absol",
				"ENVVAR1":           "loremipsum",
				"ENVVAR2":           "dolorsitamet",
				"ENVVAR3":           "=ubuntu:10.04",
				"ENVVAR7":           "ENVVAR8=ENVVAR9",
				"label1":            "test_value_1",
				"label2":            "test_value_2",
				"server_version":    "17.09.0-ce",
				"container_status":  "running",
			},
			map[string]interface{}{
				"usage_total":  uint64(1231652),
				"container_id": "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_mem",
			map[string]string{
				"engine_host":       "absol",
				"container_name":    "etcd2",
				"container_image":   "quay.io:4443/coreos/etcd",
				"container_version": "v3.3.25",
				"ENVVAR1":           "loremipsum",
				"ENVVAR2":           "dolorsitamet",
				"ENVVAR3":           "=ubuntu:10.04",
				"ENVVAR7":           "ENVVAR8=ENVVAR9",
				"label1":            "test_value_1",
				"label2":            "test_value_2",
				"server_version":    "17.09.0-ce",
				"container_status":  "running",
			},
			map[string]interface{}{
				"container_id":  "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
				"limit":         uint64(18935443456),
				"max_usage":     uint64(0),
				"usage":         uint64(0),
				"usage_percent": float64(0),
			},
			time.Unix(0, 0),
		),
	}

	// Setup the server
	server, err := mock.NewServerFromFiles("testdata")
	require.NoError(t, err)
	server.APIVersion = "1.24"

	addr := server.Start(t)
	defer server.Close()

	// Setup plugin
	plugin := &Docker{
		Endpoint: addr,
		TagEnvironment: []string{"ENVVAR1", "ENVVAR2", "ENVVAR3", "ENVVAR5",
			"ENVVAR6", "ENVVAR7", "ENVVAR8", "ENVVAR9"},
		PerDeviceInclude: []string{"cpu", "network", "blkio"},
		TotalInclude:     []string{"cpu", "blkio", "network"},
		Timeout:          config.Duration(time.Second * 5),
		Log:              testutil.Logger{},
		newClient:        newClient,
		newEnvClient:     newEnvClient,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Collect data and check the result
	require.NoError(t, acc.GatherError(plugin.Gather))
	testutil.RequireMetricsSubset(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherSwarmInfo(t *testing.T) {
	// Define the expected result
	expected := []telegraf.Metric{
		metric.New(
			"docker_swarm",
			map[string]string{
				"service_id":   "qolkls9g5iasdiuihcyz9rnx2",
				"service_name": "test1",
				"service_mode": "replicated",
			},
			map[string]interface{}{
				"tasks_running": int(2),
				"tasks_desired": uint64(2),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_swarm",
			map[string]string{
				"service_id":   "qolkls9g5iasdiuihcyz9rn3",
				"service_name": "test2",
				"service_mode": "global",
			},
			map[string]interface{}{
				"tasks_running": int(1),
				"tasks_desired": uint64(1),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_swarm",
			map[string]string{
				"service_id":   "rfmqydhe8cluzl9hayyrhw5ga",
				"service_name": "test3",
				"service_mode": "replicated_job",
			},
			map[string]interface{}{
				"tasks_running":     int(0),
				"max_concurrent":    uint64(2),
				"total_completions": uint64(2),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_swarm",
			map[string]string{
				"service_id":   "mp50lo68vqgkory4e26ts8f9d",
				"service_name": "test4",
				"service_mode": "global_job",
			},
			map[string]interface{}{
				"tasks_running": int(0),
			},
			time.Unix(0, 0),
		),
	}

	// Setup the server
	server, err := mock.NewServerFromFiles("testdata")
	require.NoError(t, err)
	server.APIVersion = "1.24"

	addr := server.Start(t)
	defer server.Close()

	// Setup plugin
	plugin := &Docker{
		Endpoint:       addr,
		GatherServices: true,
		Timeout:        config.Duration(time.Second * 5),
		Log:            testutil.Logger{},
		newClient:      newClient,
		newEnvClient:   newEnvClient,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Collect data and check the result
	require.NoError(t, acc.GatherError(plugin.Gather))
	testutil.RequireMetricsSubset(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherDiskUsage(t *testing.T) {
	// Define the expected result
	expected := []telegraf.Metric{
		metric.New(
			"docker_disk_usage",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"layers_size": int64(1e10),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_disk_usage",
			map[string]string{
				"container_image":   "some_image",
				"container_version": "1.0.0-alpine",
				"engine_host":       "absol",
				"server_version":    "17.09.0-ce",
				"container_name":    "some_container",
			},
			map[string]interface{}{
				"size_root_fs": int64(123456789),
				"size_rw":      int64(0)},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_disk_usage",
			map[string]string{
				"image_id":       "some_imageid",
				"image_name":     "some_image_tag",
				"image_version":  "1.0.0-alpine",
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"size":        int64(123456789),
				"shared_size": int64(0)},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_disk_usage",
			map[string]string{
				"image_id":       "7f4a1cc74046",
				"image_name":     "telegraf",
				"image_version":  "latest",
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"size":        int64(425484494),
				"shared_size": int64(0)},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_disk_usage",
			map[string]string{
				"volume_name":    "some_volume",
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"size": int64(123456789),
			},
			time.Unix(0, 0),
		),
	}

	// Setup the server
	server, err := mock.NewServerFromFiles("testdata")
	require.NoError(t, err)
	server.APIVersion = "1.24"

	addr := server.Start(t)
	defer server.Close()

	// Setup plugin
	plugin := &Docker{
		Endpoint:       addr,
		StorageObjects: []string{"container"},
		Timeout:        config.Duration(time.Second * 5),
		Log:            testutil.Logger{},
		newClient:      newClient,
		newEnvClient:   newEnvClient,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Collect data and check result
	require.NoError(t, acc.GatherError(plugin.Gather))
	testutil.RequireMetricsSubset(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestContainerStateFilter(t *testing.T) {
	var tests = []struct {
		name     string
		include  []string
		exclude  []string
		expected []string
	}{
		{
			name:     "default",
			expected: []string{"running"},
		},
		{
			name:     "include running",
			include:  []string{"running"},
			expected: []string{"running"},
		},
		{
			name:     "include glob",
			include:  []string{"r*"},
			expected: []string{"restarting", "running", "removing"},
		},
		{
			name:     "include all",
			include:  []string{"*"},
			expected: []string{"created", "restarting", "running", "removing", "paused", "exited", "dead"},
		},
		{
			name:    "exclude all",
			exclude: []string{"*"},
		},
		{
			name:     "exclude exited",
			include:  []string{"*"},
			exclude:  []string{"exited"},
			expected: []string{"created", "restarting", "running", "removing", "paused", "dead"},
		},
	}

	for _, tt := range tests {
		containerStates := []moby_container.ContainerState{
			moby_container.StateCreated,
			moby_container.StateRestarting,
			moby_container.StateRunning,
			moby_container.StateRemoving,
			moby_container.StatePaused,
			moby_container.StateExited,
			moby_container.StateDead,
		}

		t.Run(tt.name, func(t *testing.T) {
			// Setup the server
			server, err := mock.NewServerFromFiles("testdata")
			require.NoError(t, err)
			server.APIVersion = "1.24"

			// Make sure we request to list all container states
			server.ListParams = map[string]string{"all": "1"}

			// Manipulate the data for the test
			// Get the first  ID to use for gather to complete
			var id string
			for k := range server.Stats {
				id = k
				break
			}
			// Fake states data
			server.List = make([]moby_container.Summary, 0, len(containerStates))
			for _, v := range containerStates {
				server.List = append(server.List, moby_container.Summary{
					ID:    id,
					Names: []string{string(v)},
					State: v,
				})
			}

			addr := server.Start(t)
			defer server.Close()

			// Setup plugin
			plugin := &Docker{
				Endpoint:              addr,
				ContainerStateInclude: tt.include,
				ContainerStateExclude: tt.exclude,
				Timeout:               config.Duration(time.Second * 5),
				Log:                   testutil.Logger{},
				newClient:             newClient,
				newEnvClient:          newEnvClient,
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect data and check the result
			require.NoError(t, acc.GatherError(plugin.Gather))
			actual := make([]string, 0, acc.NMetrics())
			for _, mt := range acc.Metrics {
				if name, ok := mt.Tags["container_name"]; ok {
					actual = append(actual, name)
				}
			}
			require.Subset(t, actual, tt.expected)
		})
	}
}

func TestContainerName(t *testing.T) {
	tests := []struct {
		name           string
		containerNames []string
		expected       string
	}{
		{
			name:           "container stats name is preferred",
			containerNames: []string{"/logspout/foo"},
			expected:       "logspout",
		},
		{
			name:           "container stats without name uses container list name",
			containerNames: []string{"/logspout"},
			expected:       "logspout",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the server
			server, err := mock.NewServerFromFiles("testdata")
			require.NoError(t, err)
			server.APIVersion = "1.24"

			// Make sure we request to list all container states
			server.ListParams = map[string]string{"all": "1"}

			// Manipulate the data for the test
			// Get the first  ID to use for gather to complete
			var id string
			for k := range server.Stats {
				id = k
				break
			}

			// Fake the container list
			server.List = []moby_container.Summary{
				{
					ID:    id,
					Names: []string{"/logspout"},
					State: "running",
				},
			}

			addr := server.Start(t)
			defer server.Close()

			// Setup plugin
			plugin := &Docker{
				Endpoint:     addr,
				Timeout:      config.Duration(time.Second * 5),
				Log:          testutil.Logger{},
				newClient:    newClient,
				newEnvClient: newEnvClient,
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect data and check result
			require.NoError(t, acc.GatherError(plugin.Gather))

			for _, mt := range acc.Metrics {
				// This tag is set on all container measurements
				if mt.Measurement == "docker_container_mem" {
					require.Equal(t, tt.expected, mt.Tags["container_name"])
				}
			}
		})
	}
}

func TestHostnameFromID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "Real ID",
			id:       "565e3a55f5843cfdd4aa5659a1a75e4e78d47f73c3c483f782fe4a26fc8caa07",
			expected: "565e3a55f584",
		},
		{
			name:     "Short ID",
			id:       "shortid123",
			expected: "shortid123",
		},
		{
			name: "No ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, hostnameFromID(tt.id))
		})
	}
}

func TestPodmanDetection(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		engine   string
		endpoint string
		binary   string
		expected bool
	}{
		{
			name:     "Docker engine",
			version:  "28.3.2",
			engine:   "docker-desktop",
			endpoint: "unix:///var/run/docker.sock",
			binary:   "docker-init",
			expected: false,
		},
		{
			name:     "Real Podman with version number",
			version:  "5.6.1",
			engine:   "localhost.localdomain",
			endpoint: "unix:///run/podman/podman.sock",
			binary:   "crun",
			expected: true,
		},
		{
			name:     "Podman with version string containing podman",
			version:  "4.9.4-podman",
			engine:   "localhost",
			endpoint: "unix:///run/podman/podman.sock",
			expected: true,
		},
		{
			name:     "Podman with podman in name",
			version:  "4.9.4",
			engine:   "podman-machine",
			endpoint: "unix:///var/run/docker.sock",
			expected: true,
		},
		{
			name:     "Podman detected by endpoint",
			version:  "5.2.0",
			engine:   "localhost",
			endpoint: "unix:///run/podman/podman.sock",
			expected: true,
		},
		{
			name:     "Podman with crun runtime",
			version:  "5.0.1",
			engine:   "myhost.local",
			endpoint: "unix:///var/run/container.sock",
			binary:   "crun",
			expected: true,
		},
		{
			name:     "Docker with crun (should not detect as Podman)",
			version:  "20.10.7",
			engine:   "docker-host",
			endpoint: "unix:///var/run/docker.sock",
			binary:   "crun",
			expected: false,
		},
		{
			name:     "Edge case - simple version with generic name",
			version:  "4.8.2",
			engine:   "host",
			endpoint: "unix:///var/run/container.sock",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &system.Info{
				Name:          tt.engine,
				ServerVersion: tt.version,
				InitBinary:    tt.binary,
			}

			// Setup plugin
			plugin := &Docker{
				Endpoint: tt.endpoint,
				Log:      testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			// Setup plugin
			actual := plugin.detectPodman(info)
			require.Equal(t, tt.expected, actual, "Podman detection mismatch")
		})
	}
}

func TestPodmanStatsCache(t *testing.T) {
	// Create a mock Docker plugin configured as Podman
	plugin := &Docker{
		PodmanCacheTTL: config.Duration(60 * time.Second),
		Log:            testutil.Logger{},
		statsCache:     make(map[string]*cachedContainerStats),
		isPodman:       true,
	}

	// Create test stats
	testID := "test-container-123"
	stats1 := &container.StatsResponse{
		CPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage: 1000,
			},
			SystemUsage: 2000,
		},
	}

	stats2 := &container.StatsResponse{
		CPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage: 2000,
			},
			SystemUsage: 4000,
		},
		PreCPUStats: container.CPUStats{}, // Will be filled by fixPodmanCPUStats
	}

	// First call should cache the stats
	plugin.fixPodmanCPUStats(testID, stats1)
	require.Contains(t, plugin.statsCache, testID)
	require.Equal(t, stats1, plugin.statsCache[testID].stats)

	// Second call should use cached stats as PreCPUStats
	plugin.fixPodmanCPUStats(testID, stats2)
	require.Equal(t, stats1.CPUStats, stats2.PreCPUStats)

	// Test cache cleanup
	plugin.statsCache["old-container"] = &cachedContainerStats{
		stats:     stats1,
		timestamp: time.Now().Add(-3 * time.Hour),
	}
	plugin.cleanupStaleCache()
	require.NotContains(t, plugin.statsCache, "old-container")
	require.Contains(t, plugin.statsCache, testID)
}

func TestStartupErrorBehaviorError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()

	// Test that model.Start returns error when Ping fails with default "error" behavior
	// Uses the startup-error-behavior framework (TSD-006)
	plugin := &Docker{
		Endpoint:  server.URL,
		Timeout:   config.Duration(100 * time.Millisecond),
		newClient: newClient,
	}
	model := models.NewRunningInput(plugin, &models.InputConfig{
		Name:  "docker",
		Alias: "error-test",
	})
	model.StartupErrors.Set(0)
	require.NoError(t, model.Init())

	server.Close()

	// Starting the plugin will fail with an error because Ping fails
	var acc testutil.Accumulator
	require.ErrorContains(t, model.Start(&acc), "failed to ping Docker daemon")
	model.Stop()
}

func TestStartupErrorBehaviorIgnore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()

	// Test that model.Start returns fatal error with "ignore" behavior when Ping fails
	plugin := &Docker{
		Endpoint:  server.URL,
		Timeout:   config.Duration(100 * time.Millisecond),
		newClient: newClient,
	}
	model := models.NewRunningInput(plugin, &models.InputConfig{
		Name:                 "docker",
		Alias:                "ignore-test",
		StartupErrorBehavior: "ignore",
	})
	model.StartupErrors.Set(0)
	require.NoError(t, model.Init())

	server.Close()

	// Starting the plugin will fail and model should convert to fatal error
	var acc testutil.Accumulator
	require.ErrorContains(t, model.Start(&acc), "failed to ping Docker daemon")
	model.Stop()
}

func TestStartupErrorBehaviorRetry(t *testing.T) {
	failedServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer failedServer.Close()

	// Test that model.Start returns fatal error with "ignore" behavior when Ping fails
	plugin := &Docker{
		Endpoint:  failedServer.URL,
		Timeout:   config.Duration(100 * time.Millisecond),
		newClient: newClient,
	}
	model := models.NewRunningInput(plugin, &models.InputConfig{
		Name:                 "docker",
		Alias:                "retry-test",
		StartupErrorBehavior: "retry",
	})
	model.StartupErrors.Set(0)
	require.NoError(t, model.Init())

	failedServer.Close()

	// Starting the plugin will fail but model will mask this
	var acc testutil.Accumulator
	require.NoError(t, model.Start(&acc))
	defer model.Stop()

	// Check we can't gather
	require.ErrorIs(t, model.Gather(&acc), internal.ErrNotConnected)

	// Start the server again and check we can gather now
	server, err := mock.NewServerFromFiles("testdata")
	require.NoError(t, err)
	server.APIVersion = "1.24"

	addr := server.Start(t)
	defer server.Close()

	plugin.Endpoint = addr
	require.NoError(t, model.Gather(&acc))
}

func TestStartupSuccess(t *testing.T) {
	server, err := mock.NewServerFromFiles("testdata")
	require.NoError(t, err)
	server.APIVersion = "1.24"

	addr := server.Start(t)
	defer server.Close()

	// Test that Start succeeds when Docker is available
	plugin := &Docker{
		Endpoint:  addr,
		Timeout:   config.Duration(5 * time.Second),
		newClient: newClient,
	}
	model := models.NewRunningInput(plugin, &models.InputConfig{
		Name:  "docker",
		Alias: "success-test",
	})
	model.StartupErrors.Set(0)
	require.NoError(t, model.Init())

	var acc testutil.Accumulator
	require.NoError(t, model.Start(&acc))
	model.Stop()
}
