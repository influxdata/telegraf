package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/secretstores"
)

func TestSecretConstantManually(t *testing.T) {
	mysecret := "a wonderful test"
	s := NewSecret([]byte(mysecret))
	defer s.Destroy()
	retrieved, err := s.Get()
	require.NoError(t, err)
	defer ReleaseSecret(retrieved)
	require.EqualValues(t, mysecret, retrieved)
}

func TestLinking(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	resolvers := map[string]telegraf.ResolveFunc{
		"@{referenced:secret}": func() ([]byte, bool, error) {
			return []byte("resolved secret"), false, nil
		},
	}
	s := NewSecret([]byte(mysecret))
	defer s.Destroy()
	require.NoError(t, s.Link(resolvers))
	retrieved, err := s.Get()
	require.NoError(t, err)
	defer ReleaseSecret(retrieved)
	require.EqualValues(t, "a resolved secret", retrieved)
}

func TestLinkingResolverError(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	resolvers := map[string]telegraf.ResolveFunc{
		"@{referenced:secret}": func() ([]byte, bool, error) {
			return nil, false, errors.New("broken")
		},
	}
	s := NewSecret([]byte(mysecret))
	defer s.Destroy()
	expected := `linking secrets failed: resolving "@{referenced:secret}" failed: broken`
	require.EqualError(t, s.Link(resolvers), expected)
}

func TestGettingUnlinked(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	s := NewSecret([]byte(mysecret))
	defer s.Destroy()
	_, err := s.Get()
	require.ErrorContains(t, err, "unlinked parts in secret")
}

func TestGettingMissingResolver(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	s := NewSecret([]byte(mysecret))
	defer s.Destroy()
	s.unlinked = []string{}
	s.resolvers = map[string]telegraf.ResolveFunc{
		"@{a:dummy}": func() ([]byte, bool, error) {
			return nil, false, nil
		},
	}
	_, err := s.Get()
	expected := `replacing secrets failed: no resolver for "@{referenced:secret}"`
	require.EqualError(t, err, expected)
}

func TestGettingResolverError(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	s := NewSecret([]byte(mysecret))
	defer s.Destroy()
	s.unlinked = []string{}
	s.resolvers = map[string]telegraf.ResolveFunc{
		"@{referenced:secret}": func() ([]byte, bool, error) {
			return nil, false, errors.New("broken")
		},
	}
	_, err := s.Get()
	expected := `replacing secrets failed: resolving "@{referenced:secret}" failed: broken`
	require.EqualError(t, err, expected)
}

func TestUninitializedEnclave(t *testing.T) {
	s := Secret{}
	defer s.Destroy()
	require.NoError(t, s.Link(map[string]telegraf.ResolveFunc{}))
	retrieved, err := s.Get()
	require.NoError(t, err)
	require.Empty(t, retrieved)
	ReleaseSecret(retrieved)
}

func TestEnclaveOpenError(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	s := NewSecret([]byte(mysecret))
	defer s.Destroy()
	memguard.Purge()
	err := s.Link(map[string]telegraf.ResolveFunc{})
	require.ErrorContains(t, err, "opening enclave failed")

	s.unlinked = []string{}
	_, err = s.Get()
	require.ErrorContains(t, err, "opening enclave failed")
}

func TestMissingResolver(t *testing.T) {
	mysecret := "a @{referenced:secret}"
	s := NewSecret([]byte(mysecret))
	defer s.Destroy()
	err := s.Link(map[string]telegraf.ResolveFunc{})
	require.ErrorContains(t, err, "linking secrets failed: unlinked part")
}

func TestSecretConstant(t *testing.T) {
	tests := []struct {
		name     string
		cfg      []byte
		expected string
	}{
		{
			name: "simple string",
			cfg: []byte(`
				[[inputs.mockup]]
				  secret = "a secret"
			`),
			expected: "a secret",
		},
		{
			name: "mail address",
			cfg: []byte(`
				[[inputs.mockup]]
				  secret = "someone@mock.org"
			`),
			expected: "someone@mock.org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			require.NoError(t, c.LoadConfigData(tt.cfg))
			require.Len(t, c.Inputs, 1)

			// Create a mockup secretstore
			store := &MockupSecretStore{
				Secrets: map[string][]byte{"mock": []byte("fail")},
			}
			require.NoError(t, store.Init())
			c.SecretStores["mock"] = store
			require.NoError(t, c.LinkSecrets())

			plugin := c.Inputs[0].Input.(*MockupSecretPlugin)
			secret, err := plugin.Secret.Get()
			require.NoError(t, err)
			defer ReleaseSecret(secret)

			require.EqualValues(t, tt.expected, string(secret))
		})
	}
}

func TestSecretUnquote(t *testing.T) {
	tests := []struct {
		name string
		cfg  []byte
	}{
		{
			name: "single quotes",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = 'a secret'
					expected = 'a secret'
			`),
		},
		{
			name: "double quotes",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = "a secret"
					expected = "a secret"
			`),
		},
		{
			name: "triple single quotes",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = '''a secret'''
					expected = '''a secret'''
			`),
		},
		{
			name: "triple double quotes",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = """a secret"""
					expected = """a secret"""
			`),
		},
		{
			name: "escaped double quotes",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = "\"a secret\""
					expected = "\"a secret\""
			`),
		},
		{
			name: "mix double-single quotes (single)",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = "'a secret'"
					expected = "'a secret'"
			`),
		},
		{
			name: "mix single-double quotes (single)",
			cfg: []byte(`
				[[inputs.mockup]]
				secret = '"a secret"'
				expected = '"a secret"'
			`),
		},
		{
			name: "mix double-single quotes (triple-single)",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = """'a secret'"""
					expected = """'a secret'"""
			`),
		},
		{
			name: "mix single-double quotes (triple-single)",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = '''"a secret"'''
					expected = '''"a secret"'''
			`),
		},
		{
			name: "mix double-single quotes (triple)",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = """'''a secret'''"""
					expected = """'''a secret'''"""
			`),
		},
		{
			name: "mix single-double quotes (triple)",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = '''"""a secret"""'''
					expected = '''"""a secret"""'''
			`),
		},
		{
			name: "single quotes with backslashes",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = 'Server=SQLTELEGRAF\\SQL2022;app name=telegraf;log=1;'
					expected = 'Server=SQLTELEGRAF\\SQL2022;app name=telegraf;log=1;'
			`),
		},
		{
			name: "double quotes with backslashes",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = "Server=SQLTELEGRAF\\SQL2022;app name=telegraf;log=1;"
					expected = "Server=SQLTELEGRAF\\SQL2022;app name=telegraf;log=1;"
			`),
		},
		{
			name: "triple single quotes with backslashes",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = '''Server=SQLTELEGRAF\\SQL2022;app name=telegraf;log=1;'''
					expected = '''Server=SQLTELEGRAF\\SQL2022;app name=telegraf;log=1;'''
			`),
		},
		{
			name: "triple double quotes with backslashes",
			cfg: []byte(`
				[[inputs.mockup]]
					secret = """Server=SQLTELEGRAF\\SQL2022;app name=telegraf;log=1;"""
					expected = """Server=SQLTELEGRAF\\SQL2022;app name=telegraf;log=1;"""
			`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig()
			require.NoError(t, c.LoadConfigData(tt.cfg))
			require.Len(t, c.Inputs, 1)

			// Create a mockup secretstore
			store := &MockupSecretStore{
				Secrets: map[string][]byte{},
			}
			require.NoError(t, store.Init())
			c.SecretStores["mock"] = store
			require.NoError(t, c.LinkSecrets())

			plugin := c.Inputs[0].Input.(*MockupSecretPlugin)
			secret, err := plugin.Secret.Get()
			require.NoError(t, err)
			defer ReleaseSecret(secret)

			require.EqualValues(t, plugin.Expected, string(secret))
		})
	}
}

func TestSecretEnvironmentVariable(t *testing.T) {
	cfg := []byte(`
[[inputs.mockup]]
	secret = "$SOME_ENV_SECRET"
`)
	t.Setenv("SOME_ENV_SECRET", "an env secret")

	c := NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 1)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string][]byte{},
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	plugin := c.Inputs[0].Input.(*MockupSecretPlugin)
	secret, err := plugin.Secret.Get()
	require.NoError(t, err)
	defer ReleaseSecret(secret)

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
	secret = "@{mock:a_weird_secret}"
`)

	c := NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 4)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string][]byte{
			"secret1":          []byte("Ood Bnar"),
			"secret2":          []byte("Thon"),
			"a_strange_secret": []byte("Obi-Wan Kenobi"),
			"a_weird_secret":   []byte("Arca Jeth"),
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
		ReleaseSecret(secret)
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
	secret = "@{mock:a weird secret}"
`)

	c := NewConfig()
	err := c.LoadConfigData(cfg)
	require.NoError(t, err)
	require.Len(t, c.Inputs, 4)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string][]byte{
			"":                 []byte("Ood Bnar"),
			"wild?%go":         []byte("Thon"),
			"a-strange-secret": []byte("Obi-Wan Kenobi"),
			"a weird secret":   []byte("Arca Jeth"),
		},
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	expected := []string{
		"@{mock:}",
		"@{mock:wild?%go}",
		"@{mock:a-strange-secret}",
		"@{mock:a weird secret}",
	}
	for i, input := range c.Inputs {
		plugin := input.Input.(*MockupSecretPlugin)
		secret, err := plugin.Secret.Get()
		require.NoError(t, err)
		require.EqualValues(t, expected[i], secret)
		ReleaseSecret(secret)
	}
}

func TestSecretEqualTo(t *testing.T) {
	mysecret := "a wonderful test"
	s := NewSecret([]byte(mysecret))
	defer s.Destroy()

	equal, err := s.EqualTo([]byte(mysecret))
	require.NoError(t, err)
	require.True(t, equal)

	equal, err = s.EqualTo([]byte("some random text"))
	require.NoError(t, err)
	require.False(t, equal)
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
		Secrets: map[string][]byte{"test": []byte("Arca Jeth")},
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
	defer func() { unlinkedSecrets = make([]*Secret, 0) }()

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
		Secrets: map[string][]byte{"secret": []byte("Ood Bnar")},
		Dynamic: false,
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	sequence := []string{"Ood Bnar", "Thon", "Obi-Wan Kenobi", "Arca Jeth"}
	plugin := c.Inputs[0].Input.(*MockupSecretPlugin)
	secret, err := plugin.Secret.Get()
	require.NoError(t, err)
	defer ReleaseSecret(secret)

	require.EqualValues(t, "Ood Bnar", secret)

	for _, v := range sequence {
		store.Secrets["secret"] = []byte(v)
		secret, err := plugin.Secret.Get()
		require.NoError(t, err)

		// The secret should not change as the store is marked non-dyamic!
		require.EqualValues(t, "Ood Bnar", secret)
		ReleaseSecret(secret)
	}
}

func TestSecretStoreDynamic(t *testing.T) {
	defer func() { unlinkedSecrets = make([]*Secret, 0) }()

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
		Secrets: map[string][]byte{"secret": []byte("Ood Bnar")},
		Dynamic: true,
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	sequence := []string{"Ood Bnar", "Thon", "Obi-Wan Kenobi", "Arca Jeth"}
	plugin := c.Inputs[0].Input.(*MockupSecretPlugin)
	for _, v := range sequence {
		store.Secrets["secret"] = []byte(v)
		secret, err := plugin.Secret.Get()
		require.NoError(t, err)

		// The secret should not change as the store is marked non-dynamic!
		require.EqualValues(t, v, secret)
		ReleaseSecret(secret)
	}
}

func TestSecretStoreDeclarationMissingID(t *testing.T) {
	defer func() { unlinkedSecrets = make([]*Secret, 0) }()

	cfg := []byte(`[[secretstores.mockup]]`)

	c := NewConfig()
	err := c.LoadConfigData(cfg)
	require.ErrorContains(t, err, `error parsing mockup, "mockup" secret-store without ID`)
}

func TestSecretStoreDeclarationInvalidID(t *testing.T) {
	defer func() { unlinkedSecrets = make([]*Secret, 0) }()

	invalidIDs := []string{"foo.bar", "dummy-123", "test!", "wohoo+"}
	tmpl := `
  [[secretstores.mockup]]
    id = %q
`
	for _, id := range invalidIDs {
		t.Run(id, func(t *testing.T) {
			cfg := []byte(fmt.Sprintf(tmpl, id))
			c := NewConfig()
			err := c.LoadConfigData(cfg)
			require.ErrorContains(t, err, `error parsing mockup, invalid secret-store ID`)
		})
	}
}

func TestSecretStoreDeclarationValidID(t *testing.T) {
	defer func() { unlinkedSecrets = make([]*Secret, 0) }()

	validIDs := []string{"foobar", "dummy123", "test_id", "W0Hoo_lala123"}
	tmpl := `
  [[secretstores.mockup]]
    id = %q
`
	for _, id := range validIDs {
		t.Run(id, func(t *testing.T) {
			cfg := []byte(fmt.Sprintf(tmpl, id))
			c := NewConfig()
			err := c.LoadConfigData(cfg)
			require.NoError(t, err)
		})
	}
}

func TestSecretSet(t *testing.T) {
	defer func() { unlinkedSecrets = make([]*Secret, 0) }()

	cfg := []byte(`
      [[inputs.mockup]]
	    secret = "a secret"
	`)
	c := NewConfig()
	require.NoError(t, c.LoadConfigData(cfg))
	require.Len(t, c.Inputs, 1)
	require.NoError(t, c.LinkSecrets())

	plugin := c.Inputs[0].Input.(*MockupSecretPlugin)

	secret, err := plugin.Secret.Get()
	require.NoError(t, err)
	defer ReleaseSecret(secret)
	require.EqualValues(t, "a secret", string(secret))

	require.NoError(t, plugin.Secret.Set([]byte("another secret")))
	newsecret, err := plugin.Secret.Get()
	require.NoError(t, err)
	defer ReleaseSecret(newsecret)
	require.EqualValues(t, "another secret", string(newsecret))
}

func TestSecretSetResolve(t *testing.T) {
	defer func() { unlinkedSecrets = make([]*Secret, 0) }()

	cfg := []byte(`
      [[inputs.mockup]]
	    secret = "@{mock:secret}"
	`)
	c := NewConfig()
	require.NoError(t, c.LoadConfigData(cfg))
	require.Len(t, c.Inputs, 1)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string][]byte{"secret": []byte("Ood Bnar")},
		Dynamic: true,
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	plugin := c.Inputs[0].Input.(*MockupSecretPlugin)

	secret, err := plugin.Secret.Get()
	require.NoError(t, err)
	defer ReleaseSecret(secret)
	require.EqualValues(t, "Ood Bnar", string(secret))

	require.NoError(t, plugin.Secret.Set([]byte("@{mock:secret} is cool")))
	newsecret, err := plugin.Secret.Get()
	require.NoError(t, err)
	defer ReleaseSecret(newsecret)
	require.EqualValues(t, "Ood Bnar is cool", string(newsecret))
}

func TestSecretSetResolveInvalid(t *testing.T) {
	defer func() { unlinkedSecrets = make([]*Secret, 0) }()

	cfg := []byte(`
      [[inputs.mockup]]
	    secret = "@{mock:secret}"
	`)
	c := NewConfig()
	require.NoError(t, c.LoadConfigData(cfg))
	require.Len(t, c.Inputs, 1)

	// Create a mockup secretstore
	store := &MockupSecretStore{
		Secrets: map[string][]byte{"secret": []byte("Ood Bnar")},
		Dynamic: true,
	}
	require.NoError(t, store.Init())
	c.SecretStores["mock"] = store
	require.NoError(t, c.LinkSecrets())

	plugin := c.Inputs[0].Input.(*MockupSecretPlugin)

	secret, err := plugin.Secret.Get()
	require.NoError(t, err)
	defer ReleaseSecret(secret)
	require.EqualValues(t, "Ood Bnar", string(secret))

	err = plugin.Secret.Set([]byte("@{mock:another_secret}"))
	require.ErrorContains(t, err, `linking new secrets failed: unlinked part "@{mock:another_secret}"`)
}

/*** Mockup (input) plugin for testing to avoid cyclic dependencies ***/
type MockupSecretPlugin struct {
	Secret   Secret `toml:"secret"`
	Expected string `toml:"expected"`
}

func (*MockupSecretPlugin) SampleConfig() string                { return "Mockup test secret plugin" }
func (*MockupSecretPlugin) Gather(_ telegraf.Accumulator) error { return nil }

type MockupSecretStore struct {
	Secrets map[string][]byte
	Dynamic bool
}

func (s *MockupSecretStore) Init() error {
	return nil
}
func (*MockupSecretStore) SampleConfig() string {
	return "Mockup test secret plugin"
}

func (s *MockupSecretStore) Get(key string) ([]byte, error) {
	v, found := s.Secrets[key]
	if !found {
		return nil, errors.New("not found")
	}
	return v, nil
}

func (s *MockupSecretStore) Set(key, value string) error {
	s.Secrets[key] = []byte(value)
	return nil
}

func (s *MockupSecretStore) List() ([]string, error) {
	keys := make([]string, 0, len(s.Secrets))
	for k := range s.Secrets {
		keys = append(keys, k)
	}
	return keys, nil
}
func (s *MockupSecretStore) GetResolver(key string) (telegraf.ResolveFunc, error) {
	return func() ([]byte, bool, error) {
		v, err := s.Get(key)
		return v, s.Dynamic, err
	}, nil
}

// Register the mockup plugin on loading
func init() {
	// Register the mockup input plugin for the required names
	inputs.Add("mockup", func() telegraf.Input { return &MockupSecretPlugin{} })
	secretstores.Add("mockup", func(id string) telegraf.SecretStore {
		return &MockupSecretStore{}
	})
}
