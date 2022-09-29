package config

import (
	"errors"
	"os"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/stretchr/testify/require"
)

func TestSecretConstantManually(t *testing.T) {
	mysecret := "a wonderful test"
	s := NewSecret([]byte(mysecret))
	retrieved, err := s.Get()
	require.NoError(t, err)
	require.EqualValues(t, mysecret, retrieved)
	s.Destroy()
}

func TestLinking(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	resolvers := map[string]telegraf.ResolveFunc{
		"@{referenced:secret}": func() (string, bool, error) {
			return "resolved secret", false, nil
		},
	}
	s := NewSecret([]byte(mysecret))
	require.NoError(t, s.Link(resolvers))
	retrieved, err := s.Get()
	require.NoError(t, err)
	require.EqualValues(t, "a resolved secret", retrieved)
	s.Destroy()
}

func TestLinkingResolverError(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	resolvers := map[string]telegraf.ResolveFunc{
		"@{referenced:secret}": func() (string, bool, error) {
			return "", false, errors.New("broken")
		},
	}
	s := NewSecret([]byte(mysecret))
	err := s.Link(resolvers)
	expected := `linking secrets failed: resolving "@{referenced:secret}" failed: broken`
	require.EqualError(t, err, expected)
	s.Destroy()
}

func TestGettingUnlinked(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	s := NewSecret([]byte(mysecret))
	_, err := s.Get()
	require.ErrorContains(t, err, "unlinked parts in secret")
	s.Destroy()
}

func TestGettingMissingResolver(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	s := NewSecret([]byte(mysecret))
	s.unlinked = []string{}
	s.resolvers = map[string]telegraf.ResolveFunc{
		"@{a:dummy}": func() (string, bool, error) {
			return "", false, nil
		},
	}
	_, err := s.Get()
	expected := `replacing secrets failed: no resolver for "@{referenced:secret}"`
	require.EqualError(t, err, expected)
	s.Destroy()
}

func TestGettingResolverError(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	s := NewSecret([]byte(mysecret))
	s.unlinked = []string{}
	s.resolvers = map[string]telegraf.ResolveFunc{
		"@{referenced:secret}": func() (string, bool, error) {
			return "", false, errors.New("broken")
		},
	}
	_, err := s.Get()
	expected := `replacing secrets failed: resolving "@{referenced:secret}" failed: broken`
	require.EqualError(t, err, expected)
	s.Destroy()
}

func TestUninitializedEnclave(t *testing.T) {
	s := Secret{}
	require.NoError(t, s.Link(map[string]telegraf.ResolveFunc{}))
	retrieved, err := s.Get()
	require.NoError(t, err)
	require.Empty(t, retrieved)
	s.Destroy()
}

func TestEnclaveOpenError(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	s := NewSecret([]byte(mysecret))
	memguard.Purge()
	err := s.Link(map[string]telegraf.ResolveFunc{})
	require.ErrorContains(t, err, "opening enclave failed")

	s.unlinked = []string{}
	_, err = s.Get()
	require.ErrorContains(t, err, "opening enclave failed")
	s.Destroy()
}

func TestMissingResolver(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	s := NewSecret([]byte(mysecret))
	err := s.Link(map[string]telegraf.ResolveFunc{})
	require.ErrorContains(t, err, "linking secrets failed: unlinked part")
	s.Destroy()
}

func TestSecretConstant(t *testing.T) {
	cfg := []byte(
		`
[[inputs.mockup]]
	secret = "a secret"
`)

	c := NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 1)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string]string{},
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	plugin := c.Inputs[0].Input.(*MockupSecretPlugin)
	secret, err := plugin.Secret.Get()
	require.NoError(t, err)

	require.EqualValues(t, "a secret", secret)
}

func TestSecretEnvironmentVariable(t *testing.T) {
	cfg := []byte(`
[[inputs.mockup]]
	secret = "$SOME_ENV_SECRET"
`)
	require.NoError(t, os.Setenv("SOME_ENV_SECRET", "an env secret"))

	c := NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 1)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string]string{},
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	plugin := c.Inputs[0].Input.(*MockupSecretPlugin)
	secret, err := plugin.Secret.Get()
	require.NoError(t, err)

	require.EqualValues(t, "an env secret", secret)
}

func TestSecretStoreStatic(t *testing.T) {
	cfg := []byte(
		`
[[inputs.mockup]]
	secret = "@{mock:secret1}"
[[inputs.mockup]]
	secret = "@{mock:secret2}"
[[inputs.mockup]]
	secret = "@{mock:a_strange_secret}"
[[inputs.mockup]]
	secret = "@{mock:a_wierd_secret}"
`)

	c := NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 4)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string]string{
			"secret1":          "Ood Bnar",
			"secret2":          "Thon",
			"a_strange_secret": "Obi-Wan Kenobi",
			"a_wierd_secret":   "Arca Jeth",
		},
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	expected := []string{"Ood Bnar", "Thon", "Obi-Wan Kenobi", "Arca Jeth"}
	for i, input := range c.Inputs {
		plugin := input.Input.(*MockupSecretPlugin)
		secret, err := plugin.Secret.Get()
		require.NoError(t, err)
		require.EqualValues(t, expected[i], secret)
	}
}

func TestSecretStoreInvalidKeys(t *testing.T) {
	cfg := []byte(
		`
[[inputs.mockup]]
	secret = "@{mock:}"
[[inputs.mockup]]
	secret = "@{mock:wild?%go}"
[[inputs.mockup]]
	secret = "@{mock:a-strange-secret}"
[[inputs.mockup]]
	secret = "@{mock:a wierd secret}"
`)

	c := NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 4)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string]string{
			"":                 "Ood Bnar",
			"wild?%go":         "Thon",
			"a-strange-secret": "Obi-Wan Kenobi",
			"a wierd secret":   "Arca Jeth",
		},
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	expected := []string{
		"@{mock:}",
		"@{mock:wild?%go}",
		"@{mock:a-strange-secret}",
		"@{mock:a wierd secret}",
	}
	for i, input := range c.Inputs {
		plugin := input.Input.(*MockupSecretPlugin)
		secret, err := plugin.Secret.Get()
		require.NoError(t, err)
		require.EqualValues(t, expected[i], secret)
	}
}

func TestSecretStoreInvalidReference(t *testing.T) {
	// Make sure we clean-up our mess
	defer func() { unlinkedSecrets = make([]*Secret, 0) }()

	cfg := []byte(
		`
[[inputs.mockup]]
	secret = "@{mock:test}"
`)

	c := NewConfig()
	require.NoError(t, c.LoadConfigData(cfg))
	require.Len(t, c.Inputs, 1)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string]string{"test": "Arca Jeth"},
	}
	require.NoError(t, store.Init())
	c.SecretStores["foo"] = store
	err := c.LinkSecrets()
	require.EqualError(t, err, `unknown secret-store for "@{mock:test}"`)

	for _, input := range c.Inputs {
		plugin := input.Input.(*MockupSecretPlugin)
		secret, err := plugin.Secret.Get()
		require.EqualError(t, err, `unlinked parts in secret: @{mock:test}`)
		require.Empty(t, secret)
	}
}

func TestSecretStoreStaticChanging(t *testing.T) {
	cfg := []byte(
		`
[[inputs.mockup]]
	secret = "@{mock:secret}"
`)

	c := NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 1)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string]string{"secret": "Ood Bnar"},
		Dynamic: false,
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	sequence := []string{"Ood Bnar", "Thon", "Obi-Wan Kenobi", "Arca Jeth"}
	plugin := c.Inputs[0].Input.(*MockupSecretPlugin)
	secret, err := plugin.Secret.Get()
	require.NoError(t, err)
	require.EqualValues(t, "Ood Bnar", secret)

	for _, v := range sequence {
		store.Secrets["secret"] = v
		secret, err := plugin.Secret.Get()
		require.NoError(t, err)
		// The secret should not change as the store is marked non-dyamic!
		require.EqualValues(t, "Ood Bnar", secret)
	}
}

func TestSecretStoreDynamic(t *testing.T) {
	cfg := []byte(
		`
[[inputs.mockup]]
	secret = "@{mock:secret}"
`)

	c := NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 1)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string]string{"secret": "Ood Bnar"},
		Dynamic: true,
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	sequence := []string{"Ood Bnar", "Thon", "Obi-Wan Kenobi", "Arca Jeth"}
	plugin := c.Inputs[0].Input.(*MockupSecretPlugin)
	for _, v := range sequence {
		store.Secrets["secret"] = v
		secret, err := plugin.Secret.Get()
		require.NoError(t, err)
		// The secret should not change as the store is marked non-dyamic!
		require.EqualValues(t, v, secret)
	}
}

/*** Mockup (input) plugin for testing to avoid cyclic dependencies ***/
type MockupSecretPlugin struct {
	Secret Secret `toml:"secret"`
}

func (*MockupSecretPlugin) SampleConfig() string                  { return "Mockup test secret plugin" }
func (*MockupSecretPlugin) Gather(acc telegraf.Accumulator) error { return nil }

type MockupSecretStore struct {
	Secrets map[string]string
	Dynamic bool
}

func (s *MockupSecretStore) Init() error {
	return nil
}
func (*MockupSecretStore) SampleConfig() string { return "Mockup test secret plugin" }
func (s *MockupSecretStore) Get(key string) (string, error) {
	v, found := s.Secrets[key]
	if !found {
		return "", errors.New("not found")
	}
	return v, nil
}
func (s *MockupSecretStore) Set(key, value string) error {
	s.Secrets[key] = value
	return nil
}
func (s *MockupSecretStore) List() ([]string, error) {
	keys := []string{}
	for k := range s.Secrets {
		keys = append(keys, k)
	}
	return keys, nil
}
func (s *MockupSecretStore) GetResolver(key string) (telegraf.ResolveFunc, error) {
	return func() (string, bool, error) {
		v, err := s.Get(key)
		return v, s.Dynamic, err
	}, nil
}

// Register the mockup plugin on loading
func init() {
	// Register the mockup input plugin for the required names
	inputs.Add("mockup", func() telegraf.Input { return &MockupSecretPlugin{} })
}
