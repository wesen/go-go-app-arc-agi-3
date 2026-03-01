---
Title: ARC API Reference
DocType: reference
Topics:
  - api
  - sessions
Summary: "HTTP endpoints exposed by the ARC backend module."
Order: 3
---

# ARC API Reference

Primary endpoints:

- `GET /api/apps/arc-agi/health`
- `GET /api/apps/arc-agi/games`
- `GET /api/apps/arc-agi/games/{game_id}`
- `POST /api/apps/arc-agi/sessions`
- `GET /api/apps/arc-agi/sessions/{session_id}`
- `DELETE /api/apps/arc-agi/sessions/{session_id}`
- `POST /api/apps/arc-agi/sessions/{session_id}/games/{game_id}/reset`
- `POST /api/apps/arc-agi/sessions/{session_id}/games/{game_id}/actions`
- `GET /api/apps/arc-agi/sessions/{session_id}/events`
- `GET /api/apps/arc-agi/sessions/{session_id}/timeline`
- `GET /api/apps/arc-agi/schemas/{schema_id}`
- `GET /api/apps/arc-agi/docs`
- `GET /api/apps/arc-agi/docs/{slug}`

