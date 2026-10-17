# TODO 056 - Reject Placeholder Relay URLs

Reject example values such as `http://YOUR-LAN-IP:8080` in relay preflight and
startup scripts so a running relay cannot advertise an unusable phone URL.

Completed 2026-10-17. Shell syntax checks pass and startup refuses the example
placeholder in dry-run mode.
