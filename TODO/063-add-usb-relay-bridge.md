# TODO 063 - Add USB Relay Bridge Command

Provide an ADB reverse-port shortcut for the physically connected demo phone,
allowing the private relay to be reached as `http://127.0.0.1:8080` when Wi-Fi
clients are isolated. This is a local development/demo transport, not a
PromiseGrid protocol rule or production relay configuration.

Completed 2026-10-17. `make -n usb-relay usb-relay-clear` verifies the exact
ADB port-forwarding commands; physical execution remains a demo-proof step.
