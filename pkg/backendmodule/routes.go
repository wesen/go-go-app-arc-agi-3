package backendmodule

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	games, err := m.client.ListGames(req.Context())
	if err != nil {
		writeArcError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"games": games})
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
	game, err := m.client.GetGame(req.Context(), gameID)
	if err != nil {
		writeArcError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, game)
}

func (m *Module) handleSessions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		payload, err := decodeOptionalObject(req)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		sessionID, err := m.client.OpenSession(req.Context(), payload)
		if err != nil {
			writeArcError(w, err)
			return
		}
		m.sessions.Ensure(sessionID)
		m.events.Append(sessionID, "", "arc.session.opened", "Session opened", map[string]any{
			"source_url": payload["source_url"],
		})
		writeJSON(w, http.StatusCreated, map[string]any{
			"session_id": sessionID,
		})
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
			sessionPayload, err := m.client.GetSession(req.Context(), sessionID)
			if err != nil {
				writeArcError(w, err)
				return
			}
			if sessionPayload == nil {
				sessionPayload = map[string]any{}
			}
			sessionPayload["session_id"] = sessionID
			sessionPayload["status"] = m.sessions.Status(sessionID)
			writeJSON(w, http.StatusOK, sessionPayload)
		case http.MethodDelete:
			sessionPayload, err := m.client.CloseSession(req.Context(), sessionID)
			if err != nil {
				writeArcError(w, err)
				return
			}
			m.sessions.MarkClosed(sessionID)
			m.events.Append(sessionID, "", "arc.session.closed", "Session closed", nil)
			if sessionPayload == nil {
				sessionPayload = map[string]any{}
			}
			sessionPayload["session_id"] = sessionID
			writeJSON(w, http.StatusOK, sessionPayload)
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

	if len(parts) == 4 && parts[1] == "games" && parts[3] == "reset" && req.Method == http.MethodPost {
		gameID := strings.TrimSpace(parts[2])
		if gameID == "" {
			http.NotFound(w, req)
			return
		}
		m.handleSessionReset(w, req, sessionID, gameID)
		return
	}
	if len(parts) == 4 && parts[1] == "games" && parts[3] == "actions" && req.Method == http.MethodPost {
		gameID := strings.TrimSpace(parts[2])
		if gameID == "" {
			http.NotFound(w, req)
			return
		}
		m.handleSessionAction(w, req, sessionID, gameID)
		return
	}

	http.NotFound(w, req)
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

func (m *Module) handleSessionReset(w http.ResponseWriter, req *http.Request, sessionID, gameID string) {
	frame, err := m.client.Reset(req.Context(), sessionID, gameID)
	if err != nil {
		writeArcError(w, err)
		return
	}
	if guid, _ := frame["guid"].(string); strings.TrimSpace(guid) != "" {
		m.sessions.UpsertGUID(sessionID, gameID, guid)
	}
	m.events.Append(sessionID, gameID, "arc.game.reset", fmt.Sprintf("Game %s reset", gameID), nil)
	frame["session_id"] = sessionID
	frame["game_id"] = gameID
	writeJSON(w, http.StatusOK, frame)
}

type sessionActionRequest struct {
	Action    string         `json:"action"`
	Data      map[string]any `json:"data"`
	Reasoning any            `json:"reasoning,omitempty"`
}

func (m *Module) handleSessionAction(w http.ResponseWriter, req *http.Request, sessionID, gameID string) {
	var actionReq sessionActionRequest
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&actionReq); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	action := normalizeActionName(actionReq.Action)
	if action == "" || action == "RESET" {
		writeJSONError(w, http.StatusBadRequest, "action must be one of ACTION1..ACTION7")
		return
	}
	payload := cloneMap(actionReq.Data)
	if payload == nil {
		payload = map[string]any{}
	}
	guid, hasGUID := m.sessions.LookupGUID(sessionID, gameID)
	if hasGUID && strings.TrimSpace(guid) != "" {
		payload["guid"] = guid
	}
	if !hasGUID || strings.TrimSpace(guid) == "" {
		writeJSONError(w, http.StatusBadRequest, "missing game guid for session/game; call reset first")
		return
	}
	if actionReq.Reasoning != nil {
		payload["reasoning"] = actionReq.Reasoning
	}

	m.events.Append(sessionID, gameID, "arc.action.requested", action, map[string]any{"action": action})
	frame, err := m.client.Action(req.Context(), sessionID, gameID, action, payload)
	if err != nil {
		m.events.Append(sessionID, gameID, "arc.action.failed", action, map[string]any{"action": action, "error": err.Error()})
		writeArcError(w, err)
		return
	}
	if newGUID, _ := frame["guid"].(string); strings.TrimSpace(newGUID) != "" {
		m.sessions.UpsertGUID(sessionID, gameID, newGUID)
	}
	m.events.Append(sessionID, gameID, "arc.action.completed", action, map[string]any{"action": action})

	frame["session_id"] = sessionID
	frame["game_id"] = gameID
	frame["action"] = action
	writeJSON(w, http.StatusOK, frame)
}

func decodeOptionalObject(req *http.Request) (map[string]any, error) {
	if req == nil || req.Body == nil {
		return map[string]any{}, nil
	}
	var payload map[string]any
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&payload); err != nil {
		if errors.Is(err, io.EOF) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return payload, nil
}

func writeArcError(w http.ResponseWriter, err error) {
	var apiErr *ArcAPIError
	if errors.As(err, &apiErr) {
		writeJSONError(w, apiErr.StatusCode, apiErr.Error())
		return
	}
	writeJSONError(w, http.StatusBadGateway, err.Error())
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
