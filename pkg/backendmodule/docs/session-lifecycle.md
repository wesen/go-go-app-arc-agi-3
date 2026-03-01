---
Title: ARC Session Lifecycle
DocType: guide
Topics:
  - sessions
  - timeline
  - runtime
Summary: "How ARC sessions are opened, acted on, tracked, and closed."
Order: 2
---

# ARC Session Lifecycle

Session flow:

1. Open session (`POST /sessions`)
2. Reset game state (`POST /sessions/{id}/games/{game_id}/reset`)
3. Apply action (`POST /sessions/{id}/games/{game_id}/actions`)
4. Read events (`GET /sessions/{id}/events`)
5. Read timeline (`GET /sessions/{id}/timeline`)
6. Close session (`DELETE /sessions/{id}`)

