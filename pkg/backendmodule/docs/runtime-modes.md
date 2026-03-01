---
Title: ARC Runtime Modes
DocType: reference
Topics:
  - runtime
  - backend
Summary: "Configuration semantics for ARC runtime mode and driver options."
Order: 4
---

# ARC Runtime Modes

Key configuration knobs:

- `Driver`: `dagger` or `raw`
- `RuntimeMode`: `offline`, `normal`, or `online`
- `StartupTimeout` and `RequestTimeout`
- Runtime credentials/API key and listen address options

The module normalizes defaults on startup so composition hosts can provide minimal configuration safely.

