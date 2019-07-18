package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConnectionStringPgEnvOverride(t *testing.T) {
	config, err := parseConnectionString("dbname=test")
	assert.NoError(t, err)
	assert.Equal(t, "test", config.Database)
	assert.Equal(t, "", config.Password)

	os.Setenv("PGPASSWORD", "pass")
	config, err = parseConnectionString("dbname=test")
	assert.NoError(t, err)
	assert.Equal(t, "test", config.Database)
	assert.Equal(t, "pass", config.Password)
}
