package backendmodule

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeDriver struct {
	baseURL string
}

func (d *fakeDriver) Init(context.Context) error  { return nil }
func (d *fakeDriver) Start(context.Context) error { return nil }
func (d *fakeDriver) Stop(context.Context) error  { return nil }
func (d *fakeDriver) Health(context.Context) error {
	return nil
}
func (d *fakeDriver) BaseURL() string { return d.baseURL }

type fakeClient struct {
	lastActionPayload map[string]any
}

func (c *fakeClient) Health(context.Context) error { return nil }

func (c *fakeClient) ListGames(context.Context) ([]map[string]any, error) {
	return []map[string]any{
		{"game_id": "bt11", "name": "Game 11"},
	}, nil
}

func (c *fakeClient) GetGame(context.Context, string) (map[string]any, error) {
	return map[string]any{"game_id": "bt11", "name": "Game 11"}, nil
}

func (c *fakeClient) OpenSession(context.Context, map[string]any) (string, error) {
	return "s-1", nil
}

func (c *fakeClient) GetSession(context.Context, string) (map[string]any, error) {
	return map[string]any{"card_id": "s-1"}, nil
}

func (c *fakeClient) CloseSession(context.Context, string) (map[string]any, error) {
	return map[string]any{"closed": true}, nil
}

func (c *fakeClient) Reset(context.Context, string, string) (map[string]any, error) {
	return map[string]any{"guid": "guid-1", "state": "RUNNING"}, nil
}

func (c *fakeClient) Action(_ context.Context, _ string, _ string, action string, payload map[string]any) (map[string]any, error) {
	c.lastActionPayload = cloneMap(payload)
	return map[string]any{
		"guid":   "guid-1",
		"state":  "RUNNING",
		"action": action,
	}, nil
}

func newModuleForTests(t *testing.T) *Module {
	t.Helper()
	module, err := NewModuleWithRuntime(ModuleConfig{
		EnableReflection: true,
	}, &fakeDriver{baseURL: "http://127.0.0.1:7777"})
	require.NoError(t, err)
	module.client = &fakeClient{}
	return module
}

func TestModule_GameAndSessionFlow(t *testing.T) {
	module := newModuleForTests(t)
	mux := http.NewServeMux()
	require.NoError(t, module.MountRoutes(mux))

	gamesReq := httptest.NewRequest(http.MethodGet, "/games", nil)
	gamesRR := httptest.NewRecorder()
	mux.ServeHTTP(gamesRR, gamesReq)
	require.Equal(t, http.StatusOK, gamesRR.Code)
	require.Contains(t, gamesRR.Body.String(), "bt11")

	openReq := httptest.NewRequest(http.MethodPost, "/sessions", bytes.NewReader([]byte(`{"tags":["test"]}`)))
	openRR := httptest.NewRecorder()
	mux.ServeHTTP(openRR, openReq)
	require.Equal(t, http.StatusCreated, openRR.Code)

	resetReq := httptest.NewRequest(http.MethodPost, "/sessions/s-1/games/bt11/reset", bytes.NewReader([]byte(`{}`)))
	resetRR := httptest.NewRecorder()
	mux.ServeHTTP(resetRR, resetReq)
	require.Equal(t, http.StatusOK, resetRR.Code)
	require.Contains(t, resetRR.Body.String(), "guid-1")

	actionReq := httptest.NewRequest(http.MethodPost, "/sessions/s-1/games/bt11/actions", bytes.NewReader([]byte(`{"action":"ACTION3","data":{"x":10}}`)))
	actionRR := httptest.NewRecorder()
	mux.ServeHTTP(actionRR, actionReq)
	require.Equal(t, http.StatusOK, actionRR.Code)

	client := module.client.(*fakeClient)
	require.Equal(t, "guid-1", client.lastActionPayload["guid"])
	require.EqualValues(t, float64(10), client.lastActionPayload["x"])

	eventsReq := httptest.NewRequest(http.MethodGet, "/sessions/s-1/events?after_seq=0", nil)
	eventsRR := httptest.NewRecorder()
	mux.ServeHTTP(eventsRR, eventsReq)
	require.Equal(t, http.StatusOK, eventsRR.Code)
	require.Contains(t, eventsRR.Body.String(), "arc.session.opened")
	require.Contains(t, eventsRR.Body.String(), "arc.action.completed")

	timelineReq := httptest.NewRequest(http.MethodGet, "/sessions/s-1/timeline", nil)
	timelineRR := httptest.NewRecorder()
	mux.ServeHTTP(timelineRR, timelineReq)
	require.Equal(t, http.StatusOK, timelineRR.Code)
	require.Contains(t, timelineRR.Body.String(), "\"status\":\"active\"")

	closeReq := httptest.NewRequest(http.MethodDelete, "/sessions/s-1", nil)
	closeRR := httptest.NewRecorder()
	mux.ServeHTTP(closeRR, closeReq)
	require.Equal(t, http.StatusOK, closeRR.Code)

	timelineClosedReq := httptest.NewRequest(http.MethodGet, "/sessions/s-1/timeline", nil)
	timelineClosedRR := httptest.NewRecorder()
	mux.ServeHTTP(timelineClosedRR, timelineClosedReq)
	require.Equal(t, http.StatusOK, timelineClosedRR.Code)
	require.Contains(t, timelineClosedRR.Body.String(), "\"status\":\"closed\"")
}

func TestModule_ActionRequiresResetGUID(t *testing.T) {
	module := newModuleForTests(t)
	mux := http.NewServeMux()
	require.NoError(t, module.MountRoutes(mux))

	req := httptest.NewRequest(http.MethodPost, "/sessions/s-2/games/bt11/actions", bytes.NewReader([]byte(`{"action":"ACTION1"}`)))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "call reset first")
}

func TestModule_ReflectionAndSchemas(t *testing.T) {
	module := newModuleForTests(t)

	doc, err := module.Reflection(context.Background())
	require.NoError(t, err)
	require.Equal(t, AppID, doc.AppID)
	require.NotEmpty(t, doc.APIs)
	require.NotEmpty(t, doc.Schemas)

	mux := http.NewServeMux()
	require.NoError(t, module.MountRoutes(mux))

	req := httptest.NewRequest(http.MethodGet, "/schemas/arc.games.list.response.v1", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var payload map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&payload))
	require.Equal(t, "object", payload["type"])
}
