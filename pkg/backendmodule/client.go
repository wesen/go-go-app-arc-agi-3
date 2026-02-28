package backendmodule

import (
	"context"
	"fmt"
	"time"
)

type ArcAPIClient interface {
	Health(ctx context.Context) error
	ListGames(ctx context.Context) ([]map[string]any, error)
	GetGame(ctx context.Context, gameID string) (map[string]any, error)
	OpenSession(ctx context.Context, payload map[string]any) (string, error)
	GetSession(ctx context.Context, sessionID string) (map[string]any, error)
	CloseSession(ctx context.Context, sessionID string) (map[string]any, error)
	Reset(ctx context.Context, sessionID, gameID string) (map[string]any, error)
	Action(ctx context.Context, sessionID, gameID, action string, payload map[string]any) (map[string]any, error)
}

type HTTPArcAPIClient struct {
	driver ArcRuntimeDriver
}

func NewHTTPArcAPIClient(driver ArcRuntimeDriver, _ time.Duration) *HTTPArcAPIClient {
	return &HTTPArcAPIClient{driver: driver}
}

func (c *HTTPArcAPIClient) Health(ctx context.Context) error {
	return c.driver.Health(ctx)
}

func (c *HTTPArcAPIClient) ListGames(context.Context) ([]map[string]any, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *HTTPArcAPIClient) GetGame(context.Context, string) (map[string]any, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *HTTPArcAPIClient) OpenSession(context.Context, map[string]any) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (c *HTTPArcAPIClient) GetSession(context.Context, string) (map[string]any, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *HTTPArcAPIClient) CloseSession(context.Context, string) (map[string]any, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *HTTPArcAPIClient) Reset(context.Context, string, string) (map[string]any, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *HTTPArcAPIClient) Action(context.Context, string, string, string, map[string]any) (map[string]any, error) {
	return nil, fmt.Errorf("not implemented")
}
