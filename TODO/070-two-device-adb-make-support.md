# TODO 070 - Add Two-Device Android Make Support

Make the Android install, launch, diagnostics, screenshots, and USB relay
commands target a selected ADB serial. This removes ambiguous-device failures
when preparing the two-device delivery proof in TODO 008.

Acceptance:

- `make devices` lists attached devices.
- `make install DEVICE=<serial>` and the related phone commands target only
  that device.
- Leaving `DEVICE` empty preserves the single-device workflow.

Completed: `Makefile` now provides `make devices` and applies the optional
`DEVICE` serial to install, launch, logs, diagnostics, screenshots, and USB
relay commands. README installation instructions include a two-device example.
