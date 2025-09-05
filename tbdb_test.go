package tbdb

import (
	"testing"

	"github.com/qoinlyid/qore"
	"github.com/stretchr/testify/assert"
)

func TestNewWithoutConfig(t *testing.T) {
	i := New()
	assert.Empty(t, i.cfg.Addresses, "Addresses should be empty")
}

func TestNew(t *testing.T) {
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	assert.NotEmpty(t, i.cfg.Addresses, "Addresses must be not empty")
}

func TestOpen(t *testing.T) {
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err, "Must be no error")
	i.Close()
}

func TestClient(t *testing.T) {
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err, "Must be no error")
	defer i.Close()

	cln, err := i.Client()
	assert.NoError(t, err, "Client method must be no error")
	assert.NotNil(t, cln, "Client method must be return non-nil client iterface")
}

func TestHealthCheckHealthy(t *testing.T) {
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err)
	stats := i.HealthCheck(t.Context())
	assert.NotNil(t, stats, "DependencyStats value must be not nil")
	assert.NotEmpty(t, stats.PINGLatencyMillis, "PINGLatencyMillis value must be not empty")
}

func TestClose(t *testing.T) {
	t.Setenv(qore.CONFIG_USED_KEY, "./.env")
	i := New()
	err := i.Open()
	assert.NoError(t, err, "Open must be no error")
	err = i.Close()
	assert.NoError(t, err, "Close must be no error")
}
