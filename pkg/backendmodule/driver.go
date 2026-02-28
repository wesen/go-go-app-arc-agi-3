package backendmodule

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ArcRuntimeDriver interface {
	Init(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health(ctx context.Context) error
	BaseURL() string
}

func newRuntimeDriver(config ModuleConfig) (ArcRuntimeDriver, error) {
	switch strings.ToLower(strings.TrimSpace(config.Driver)) {
	case "", "dagger":
		return &NoopDriver{}, nil
	case "raw":
		return &NoopDriver{}, nil
	default:
		return nil, fmt.Errorf("unsupported arc runtime driver: %q", config.Driver)
	}
}

func waitForDriverHealthy(ctx context.Context, driver ArcRuntimeDriver) error {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := driver.Health(ctx); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("arc runtime did not become healthy before timeout: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

type NoopDriver struct{}

func (d *NoopDriver) Init(context.Context) error  { return nil }
func (d *NoopDriver) Start(context.Context) error { return nil }
func (d *NoopDriver) Stop(context.Context) error  { return nil }
func (d *NoopDriver) Health(context.Context) error {
	return nil
}
func (d *NoopDriver) BaseURL() string { return "http://127.0.0.1:0" }
