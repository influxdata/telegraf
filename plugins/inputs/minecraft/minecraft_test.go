package minecraft

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type MockClient struct {
	ConnectF func() error
	PlayersF func() ([]string, error)
	ScoresF  func(player string) ([]Score, error)
}

func (c *MockClient) Connect() error {
	return c.ConnectF()
}

func (c *MockClient) Players() ([]string, error) {
	return c.PlayersF()
}

func (c *MockClient) Scores(player string) ([]Score, error) {
	return c.ScoresF(player)
}

func TestGather(t *testing.T) {
	now := time.Unix(0, 0)

	tests := []struct {
		name    string
		client  *MockClient
		metrics []telegraf.Metric
		err     error
	}{
		{
			name: "no players",
			client: &MockClient{
				ConnectF: func() error {
					return nil
				},
				PlayersF: func() ([]string, error) {
					return []string{}, nil
				},
			},
			metrics: []telegraf.Metric{},
		},
		{
			name: "one player without scores",
			client: &MockClient{
				ConnectF: func() error {
					return nil
				},
				PlayersF: func() ([]string, error) {
					return []string{"Etho"}, nil
				},
				ScoresF: func(player string) ([]Score, error) {
					switch player {
					case "Etho":
						return []Score{}, nil
					default:
						panic("unknown player")
					}
				},
			},
			metrics: []telegraf.Metric{},
		},
		{
			name: "one player with scores",
			client: &MockClient{
				ConnectF: func() error {
					return nil
				},
				PlayersF: func() ([]string, error) {
					return []string{"Etho"}, nil
				},
				ScoresF: func(player string) ([]Score, error) {
					switch player {
					case "Etho":
						return []Score{{Name: "jumps", Value: 42}}, nil
					default:
						panic("unknown player")
					}
				},
			},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"minecraft",
					map[string]string{
						"player": "Etho",
						"server": "example.org:25575",
						"source": "example.org",
						"port":   "25575",
					},
					map[string]interface{}{
						"jumps": 42,
					},
					now,
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Minecraft{
				Server:   "example.org",
				Port:     "25575",
				Password: "xyzzy",
				client:   tt.client,
			}

			var acc testutil.Accumulator
			acc.TimeFunc = func() time.Time { return now }

			err := plugin.Gather(&acc)

			require.Equal(t, tt.err, err)
			testutil.RequireMetricsEqual(t, tt.metrics, acc.GetTelegrafMetrics())
		})
	}
}
