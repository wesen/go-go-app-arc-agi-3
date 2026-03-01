---
Title: ARC-AGI Module Overview
DocType: guide
Topics:
  - backend
  - onboarding
  - runtime
Summary: "High-level architecture and ownership boundaries for the ARC backend module."
Order: 1
---

# ARC-AGI Module Overview

The ARC backend module manages runtime lifecycle and proxies ARC gameplay APIs under:

- `/api/apps/arc-agi/...`

Core responsibilities:

- Runtime driver lifecycle (`Init`, `Start`, `Stop`, `Health`)
- Session and event timeline management
- Schema + reflection metadata exposure
- Module docs endpoints

