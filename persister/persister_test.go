package persister

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"

	"github.com/stretchr/testify/require"
)

const characters = "abcdedfghijklmnopqrst"
const special = "0123456789"

func generateRandomString(length int) string {
	set := characters + strings.ToUpper(characters) + special
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = set[rand.Intn(len(set))]
	}
	return string(buf)
}

func generateID(t *testing.T, name, id string) string {
	hash := sha256.New()

	_, err := hash.Write(append([]byte(name), 0))
	require.NoError(t, err, "hashing name failed")

	_, err = hash.Write(append([]byte(id), 0))
	require.NoError(t, err, "hashing state ID failed")

	return hex.EncodeToString(hash.Sum(nil))
}

func TestPersister_NotInitialized(t *testing.T) {
	plugin := MockupPluginID{
		Servers:  []string{"server A"},
		Methods:  []string{"a", "b", "c"},
		Settings: map[string]string{},
		Port:     42,
	}
	require.NoError(t, plugin.Init())

	persister := Persister{}
	err := persister.SetState("abcde", plugin.GetState())
	require.Error(t, err)
	require.Equal(t, "not initialized", err.Error())

	err = persister.Register("input.mockup", &plugin)
	require.Error(t, err)
	require.Equal(t, "not initialized", err.Error())

	state, found := persister.GetState("abcde")
	require.False(t, found)
	require.Nil(t, state)

	err = persister.Load()
	require.Error(t, err)
	require.Equal(t, "not initialized", err.Error())

	err = persister.Store()
	require.Error(t, err)
	require.Equal(t, "not initialized", err.Error())
}

func TestPersister_InvalidParams(t *testing.T) {
	persister := Persister{}
	err := persister.Init()
	require.Error(t, err)
	require.Equal(t, "init called on disabled persister", err.Error())

	persister = Persister{Enabled: true}
	err = persister.Init()
	require.Error(t, err)
	require.Equal(t, "invalid filename for \"json\" store", err.Error())

	persister = Persister{Enabled: true, Format: "json"}
	err = persister.Init()
	require.Error(t, err)
	require.Equal(t, "invalid filename for \"json\" store", err.Error())

	persister = Persister{Enabled: true, Format: "invalid"}
	err = persister.Init()
	require.Error(t, err)
	require.Equal(t, "unknown file-format \"invalid\" for persister", err.Error())
}

func TestPersister_Init(t *testing.T) {
	persister := Persister{
		Enabled:  true,
		Filename: "some random file",
	}
	require.NoError(t, persister.Init())
}

func TestPersister_IDGeneration(t *testing.T) {
	plugins := []MockupPlugin{}

	for i := 0; i < 10; i++ {
		p := MockupPlugin{
			Servers:  []string{generateRandomString(16)},
			Methods:  []string{"a", "b", "c"},
			Settings: map[string]string{},
			Port:     i,
		}

		for j := 0; j < 5; j++ {
			p.Settings[generateRandomString(10)] = generateRandomString(10)
		}

		p.Setups = make([]MockupPluginSettings, 5)
		for j := 0; j < 5; j++ {
			p.Setups[j] = MockupPluginSettings{
				Name:     fmt.Sprintf("setup_%d", j),
				Factor:   float64(j) / 100.0,
				Enabled:  (j%2 == 0),
				BitField: []int{j*10 + 1, j*10 + 2, j*10 + 3, j*10 + 4},
			}
		}
		require.NoError(t, p.Init())

		plugins = append(plugins, p)
	}

	id, err := generatePluginID("", &plugins[0])
	require.Empty(t, id)
	require.Error(t, err)
	require.Equal(t, "empty prefix", err.Error())

	// Compare generated IDs
	for i, pi := range plugins {
		ref, err := generatePluginID("input.mockup", pi)
		require.NoErrorf(t, err, "%d: generating reference ID failed for: %v", i, pi)

		// Cross-comparison
		for j, pj := range plugins {
			test, err := generatePluginID("input.mockup", pj)
			require.NoErrorf(t, err, "%d: generating testing ID failed for: %v", j, pj)

			// IDs at the same index should be identical, all others different
			if i == j {
				require.Equalf(t, ref, test, "difference for %d: \n%v (%v)\n%v (%v)", i, pi, ref, pj, test)
			} else {
				require.NotEqualf(t, ref, test, "equal for %d, %d", i, j)
			}
		}
	}
}

func TestPersister_PluginStateID(t *testing.T) {
	plugin := MockupPluginID{
		Servers:  []string{"server A"},
		Methods:  []string{"a", "b", "c"},
		Settings: map[string]string{},
		Port:     42,
	}
	require.NoError(t, plugin.Init())

	persister := Persister{
		Enabled:  true,
		Filename: "some random file",
	}
	require.NoError(t, persister.Init())

	// Compare generated IDs
	expected := generateID(t, "input.mockup", plugin.GetPluginStateID())
	id, err := generatePluginID("input.mockup", &plugin)
	require.NoError(t, err)
	require.Equal(t, expected, id)
}

func TestPersister_StateCollection(t *testing.T) {
	plugin := MockupPlugin{
		Servers:  []string{"server A"},
		Methods:  []string{"a", "b", "c"},
		Settings: map[string]string{},
		Port:     42,
	}
	require.NoError(t, plugin.Init())

	persister := Persister{
		Enabled:  true,
		Filename: "some random file",
	}
	require.NoError(t, persister.Init())
	require.NoError(t, persister.Register("input.mockup", &plugin))

	id, err := generatePluginID("input.mockup", plugin)
	require.NoError(t, err)

	// We store the state at register time
	t0, _ := time.Parse(time.RFC3339, "2021-04-24T23:42:00+02:00")
	expectedState := MockupState{
		Name:     "mockup",
		Bits:     []int{},
		Modified: t0,
	}
	state, found := persister.GetState(id)
	require.True(t, found)
	require.Equal(t, expectedState, state)

	// State should not change when collected anew
	persister.collect()
	state, found = persister.GetState(id)
	require.True(t, found)
	require.Equal(t, expectedState, state)

	// Check that we have a copy
	plugin.state.Name = "wurz"
	plugin.state.Version = 15
	state, found = persister.GetState(id)
	require.True(t, found)
	require.Equal(t, expectedState, state)

	// Check if update works
	persister.collect()
	state, found = persister.GetState(id)
	require.True(t, found)
	require.NotEqual(t, expectedState, state)
	expectedState.Name = "wurz"
	expectedState.Version = 15
	require.Equal(t, expectedState, state)
}

func TestPersister_StoreLoad(t *testing.T) {
	plugins := []telegraf.StatefulPlugin{
		&MockupPlugin{
			Servers:  []string{"server A"},
			Methods:  []string{"a", "b", "c"},
			Settings: map[string]string{},
			Port:     42,
		},
		&MockupPlugin{
			Servers:  []string{"server B"},
			Methods:  []string{"a", "b", "c"},
			Settings: map[string]string{},
			Port:     23,
		},
		&MockupPluginID{
			Servers:  []string{"server A"},
			Methods:  []string{"a", "b", "c"},
			Settings: map[string]string{},
			Port:     42,
		},
	}

	// Reserve a temporary state file
	filename := filepath.Join(t.TempDir(), "telegraf_test_state-store_load.json")
	require.NotEmpty(t, filename)
	fmt.Printf("using temporary file %q...\n", filename)

	persisterStore := Persister{
		Enabled:  true,
		Filename: filename,
	}
	require.NoError(t, persisterStore.Init())

	expected := make([]interface{}, 0, len(plugins))
	for _, plugin := range plugins {
		require.NoError(t, plugin.(telegraf.Initializer).Init())
		require.NoError(t, persisterStore.Register("input.mockup", plugin))
		// Store the state for later comparison
		expected = append(expected, plugin.GetState())
	}

	// Write state
	require.NoError(t, persisterStore.Store())

	// Modify states such that we can verify loading
	for i, plugin := range plugins {
		if p, ok := plugin.(*MockupPlugin); ok {
			appendToState(&p.state)
			require.NotEqual(t, expected[i], plugin.GetState())
			continue
		}
		if p, ok := plugin.(*MockupPluginID); ok {
			appendToState(&p.state)
			require.NotEqual(t, expected[i], plugin.GetState())
		}
	}

	persisterLoad := Persister{
		Enabled:  true,
		Filename: filename,
	}
	require.NoError(t, persisterLoad.Init())
	for i, plugin := range plugins {
		require.NoError(t, persisterLoad.Register("input.mockup", plugin))
		require.NotEqual(t, expected[i], plugin.GetState())
	}

	// Read state
	require.NoError(t, persisterLoad.Load())
	for i, plugin := range plugins {
		require.Equal(t, expected[i], plugin.GetState())
	}
}

/*** Mockup plugin for testing to avoid cyclic dependencies ***/
type MockupState struct {
	Name     string
	Version  uint64
	Offset   uint64
	Bits     []int
	Modified time.Time
}

type MockupPluginSettings struct {
	Name     string  `toml:"name"`
	Factor   float64 `toml:"factor"`
	Enabled  bool    `toml:"enabled"`
	BitField []int   `toml:"bits"`
}

type MockupPlugin struct {
	Servers          []string               `toml:"servers"`
	Methods          []string               `toml:"methods"`
	Settings         map[string]string      `toml:"wurstbrot"`
	Port             int                    `toml:"port"`
	Setups           []MockupPluginSettings `toml:"setup"`
	StateManipulator func(s *MockupState)   `toml:"-"`
	Log              telegraf.Logger        `toml:"-"`
	tls.ServerConfig

	command string
	file    string
	state   MockupState
}

func (m *MockupPlugin) Init() error {
	t0, _ := time.Parse(time.RFC3339, "2021-04-24T23:42:00+02:00")
	m.state = MockupState{
		Name:     "mockup",
		Bits:     []int{},
		Modified: t0,
	}

	return nil
}

func (m *MockupPlugin) GetState() interface{} {
	state := m.state
	if m.StateManipulator != nil {
		m.StateManipulator(&m.state)
	}
	return state
}

// SetState is called by the Persister once after loading and _after_
// the call to the plugin's Init() function if there is any.
func (m *MockupPlugin) SetState(state interface{}) error {
	s, ok := state.(MockupState)
	if !ok {
		return fmt.Errorf("invalid state type %T", state)
	}
	m.state = s

	return nil
}

func (m *MockupPlugin) SampleConfig() string                  { return "Mockup test plugin" }
func (m *MockupPlugin) Description() string                   { return "Mockup test plugin" }
func (m *MockupPlugin) Gather(acc telegraf.Accumulator) error { return nil }

type MockupPluginID MockupPlugin

func (m *MockupPluginID) GetPluginStateID() string {
	return "an identitiy"
}

func (m *MockupPluginID) Init() error {
	return (*MockupPlugin)(m).Init()
}

func (m *MockupPluginID) GetState() interface{} {
	return m.state
}

// SetState is called by the Persister once after loading and _after_
// the call to the plugin's Init() function if there is any.
func (m *MockupPluginID) SetState(state interface{}) error {
	return (*MockupPlugin)(m).SetState(state)
}

func appendToState(s *MockupState) {
	s.Name += "_a"
	s.Version++
	s.Offset += 2
	s.Bits = append(s.Bits, len(s.Bits))
	s.Modified = time.Now()
}
