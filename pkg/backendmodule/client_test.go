package backendmodule

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type testDriver struct {
	baseURL string
}

func (d *testDriver) Init(context.Context) error  { return nil }
func (d *testDriver) Start(context.Context) error { return nil }
func (d *testDriver) Stop(context.Context) error  { return nil }
func (d *testDriver) Health(context.Context) error {
	return nil
}
func (d *testDriver) BaseURL() string { return d.baseURL }

func TestHTTPArcAPIClientReset_PrimesIdleGameWithSecondReset(t *testing.T) {
	var resetCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/cmd/RESET", r.URL.Path)

		n := atomic.AddInt32(&resetCalls, 1)
		w.Header().Set("Content-Type", "application/json")
		switch n {
		case 1:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"guid":              "guid-1",
				"state":             "IDLE",
				"available_actions": []any{},
				"frame":             []any{[]any{0, 1}},
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"guid":              "guid-1",
				"state":             "RUNNING",
				"available_actions": []any{"ACTION1", "ACTION3"},
				"frame":             []any{[]any{1, 2}},
			})
		}
	}))
	defer srv.Close()

	client := NewHTTPArcAPIClient(&testDriver{baseURL: srv.URL}, 0, "")
	frame, err := client.Reset(context.Background(), "session-1", "bt11")
	require.NoError(t, err)
	require.EqualValues(t, 2, atomic.LoadInt32(&resetCalls))
	require.Equal(t, "RUNNING", frame["state"])
	require.Equal(t, "guid-1", frame["guid"])
	require.NotEmpty(t, frame["available_actions"])
}

func TestHTTPArcAPIClientReset_DoesNotDoubleResetWhenActive(t *testing.T) {
	var resetCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/cmd/RESET", r.URL.Path)
		atomic.AddInt32(&resetCalls, 1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"guid":              "guid-1",
			"state":             "RUNNING",
			"available_actions": []any{"ACTION1"},
			"frame":             []any{[]any{1, 2}},
		})
	}))
	defer srv.Close()

	client := NewHTTPArcAPIClient(&testDriver{baseURL: srv.URL}, 0, "")
	frame, err := client.Reset(context.Background(), "session-1", "bt11")
	require.NoError(t, err)
	require.EqualValues(t, 1, atomic.LoadInt32(&resetCalls))
	require.Equal(t, "RUNNING", frame["state"])
}
