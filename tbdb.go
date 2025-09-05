package tbdb

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/qoinlyid/qore"
	tb "github.com/tigerbeetle/tigerbeetle-go"
)

// Instance defines TigerBeetle DB dependency singleton.
type Instance struct {
	// Define dependency singleton here.
	client tb.Client

	// Private field.
	cfg       *Config
	startTime time.Time
	*instanceGen
}

// New creates singleton dependency instance.
func New() *Instance {
	config := loadConfig()
	instance := &Instance{
		cfg:         config,
		instanceGen: &instanceGen{priority: config.DependencyPriority},
	}
	return instance
}

// HealthCheck returns statistics for dependency health check.
func (i *Instance) HealthCheck(ctx context.Context) *qore.DependencyStats {
	uptime := time.Since(i.startTime)
	stats := &qore.DependencyStats{
		UptimeSeconds: uptime.Seconds(),
		UptimeHuman:   uptime.String(),
	}
	if i.client == nil {
		return stats
	}

	start := time.Now()
	err := i.client.Nop()
	time.Sleep(time.Millisecond * 15)
	latency := time.Since(start)
	if err != nil {
		stats.PINGResponse = err.Error()
		return stats
	}
	stats.PINGLatencyMillis = latency.Milliseconds()
	stats.PINGLatencyHuman = latency.String()
	stats.PINGResponse = "Ok"
	return stats
}

// Open an backend connection or construct the dependency.
func (i *Instance) Open() error {
	// Setup addresses.
	var addrs []string
	for addr := range strings.SplitSeq(i.cfg.Addresses, ",") {
		if len(strings.TrimSpace(addr)) > 0 {
			addrs = append(addrs, addr)
		}
	}

	// Open connection.
	client, err := tb.NewClient(i.cfg.clusterIDTB, addrs)
	if err != nil {
		return errors.Join(ErrOpenTBConnection, err)
	}
	i.client = client

	// Set another instance field.
	i.startTime = time.Now()

	// Return.
	return nil
}

// Close an backend connection or destruct the dependency.
func (i *Instance) Close() error {
	// Close connection.
	if i.client == nil {
		return nil
	}
	i.client.Close()
	return nil
}

// Client returns TigerBeetle client interface.
func (i *Instance) Client() (tb.Client, error) {
	if err := i.validateClient(); err != nil {
		return nil, err
	}
	return i.client, nil
}
