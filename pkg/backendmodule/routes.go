package backendmodule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (m *Module) handleHealth(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	if err := m.Health(req.Context()); err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (m *Module) handleGames(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSONError(w, http.StatusNotImplemented, "games endpoint not implemented yet")
}

func (m *Module) handleGamesSubresource(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	gameID := strings.Trim(strings.TrimPrefix(req.URL.Path, "/games/"), "/")
	if gameID == "" {
		http.NotFound(w, req)
		return
	}
	writeJSONError(w, http.StatusNotImplemented, fmt.Sprintf("game endpoint for %q not implemented yet", gameID))
}

func (m *Module) handleSessions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		writeJSONError(w, http.StatusNotImplemented, "session open endpoint not implemented yet")
	default:
		writeMethodNotAllowed(w)
	}
}

func (m *Module) handleSessionsSubresource(w http.ResponseWriter, req *http.Request) {
	trimmed := strings.Trim(strings.TrimPrefix(req.URL.Path, "/sessions/"), "/")
	if trimmed == "" {
		http.NotFound(w, req)
		return
	}
	parts := strings.Split(trimmed, "/")
	sessionID := parts[0]
	if sessionID == "" {
		http.NotFound(w, req)
		return
	}

	if len(parts) == 1 {
		switch req.Method {
		case http.MethodGet:
			writeJSONError(w, http.StatusNotImplemented, "session get endpoint not implemented yet")
		case http.MethodDelete:
			writeJSONError(w, http.StatusNotImplemented, "session close endpoint not implemented yet")
		default:
			writeMethodNotAllowed(w)
		}
		return
	}

	if len(parts) == 2 && parts[1] == "events" && req.Method == http.MethodGet {
		afterSeq, err := parseAfterSeq(req.URL.Query().Get("after_seq"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		events := m.events.ListAfter(sessionID, afterSeq)
		writeJSON(w, http.StatusOK, map[string]any{"session_id": sessionID, "events": events})
		return
	}

	if len(parts) == 2 && parts[1] == "timeline" && req.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, m.events.Timeline(sessionID))
		return
	}

	writeJSONError(w, http.StatusNotImplemented, "session subresource endpoint not implemented yet")
}

func parseAfterSeq(raw string) (int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid after_seq query value")
	}
	return n, nil
}

func (m *Module) handleSchemaByID(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	schemaID := strings.Trim(strings.TrimPrefix(req.URL.Path, "/schemas/"), "/")
	if schemaID == "" {
		http.NotFound(w, req)
		return
	}
	schema, ok := getSchemaByID(schemaID)
	if !ok {
		http.NotFound(w, req)
		return
	}
	writeJSON(w, http.StatusOK, schema)
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"message": strings.TrimSpace(message),
		},
	})
}
