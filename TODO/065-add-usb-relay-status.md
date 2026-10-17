# TODO 065 - Add USB Relay Status Check

Add one Make target that shows the ADB device and reverse-port mapping, then
checks the local relay health endpoint before a phone upload is retried.

Completed 2026-10-17 on the authorized host: ADB reported device `ZD222TKJ4S`,
the `tcp:8080` reverse route was installed, and the host relay returned
`{"status":"ok"}` from `/healthz`.
