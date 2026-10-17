# TODO 062 - Add Relay Health QR Probe

Create a QR image for the public relay `/healthz` URL so a phone can verify
LAN reachability in its browser without typing the relay address.

Completed 2026-10-17. The command generated a valid 512×512 PNG containing
the configured health URL; the Make target opens it for phone scanning.
