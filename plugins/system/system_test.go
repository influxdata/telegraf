package system

import (
	"testing"

	"github.com/shirou/gopsutil/load"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestSystemStats(t *testing.T) {
	n := neko.Start(t)

	var mps MockPS

	n.CheckMock(&mps.Mock)

	n.It("generates metrics from the system information", func() {
		ss := &SystemStats{ps: &mps}

		lv := &load.LoadAvgStat{
			Load1:  0.3,
			Load5:  1.5,
			Load15: 0.8,
		}

		mps.On("LoadAvg").Return(lv, nil)

		msgs, err := ss.Read()
		require.NoError(t, err)

		name, ok := msgs[0].GetString("name")
		require.True(t, ok)

		assert.Equal(t, "load1", name)

		val, ok := msgs[0].GetFloat("value")
		require.True(t, ok)

		assert.Equal(t, 0.3, val)

		name, ok = msgs[1].GetString("name")
		require.True(t, ok)

		assert.Equal(t, "load5", name)

		val, ok = msgs[1].GetFloat("value")
		require.True(t, ok)

		assert.Equal(t, 1.5, val)

		name, ok = msgs[2].GetString("name")
		require.True(t, ok)

		assert.Equal(t, "load15", name)

		val, ok = msgs[2].GetFloat("value")
		require.True(t, ok)

		assert.Equal(t, 0.8, val)
	})

	n.It("adds any tags registered", func() {
		ss := &SystemStats{
			ps: &mps,
			tags: map[string]string{
				"host": "my.test",
				"dc":   "us-west-1",
			},
		}

		lv := &load.LoadAvgStat{
			Load1:  0.3,
			Load5:  1.5,
			Load15: 0.8,
		}

		mps.On("LoadAvg").Return(lv, nil)

		msgs, err := ss.Read()
		require.NoError(t, err)

		for _, m := range msgs {
			val, ok := m.GetTag("host")
			require.True(t, ok)

			assert.Equal(t, val, "my.test")

			val, ok = m.GetTag("dc")
			require.True(t, ok)

			assert.Equal(t, val, "us-west-1")
		}
	})

	n.Meow()
}
