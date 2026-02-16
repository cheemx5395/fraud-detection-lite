package repository

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	// Test failure with invalid URI
	os.Setenv("DB_URI", "invalid_uri")
	cfg := Config()
	assert.Nil(t, cfg)

	// Test success with dummy env
	os.Setenv("DB_URI", "postgres://user:pass@localhost:5432/db")
	cfg = Config()
	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.ConnConfig.Host)
	assert.Equal(t, uint16(5432), cfg.ConnConfig.Port)
}
