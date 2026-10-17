# TODO 055 - Add Common Make Commands

Create a project-root `Makefile` that provides explicit, safe shortcuts for
the recurring host operations: tests, APK build/install, computer-peer
identity, relay preflight/start, host validation, and evidence capture.

Commands that can start a relay or collect authenticated evidence must require
their relay URL, token, CID, and output inputs instead of supplying defaults.

Completed 2026-10-17. `make help` documents the command surface; `make test`
passes, and the relay preflight target rejects an absent relay URL.
