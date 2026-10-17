# TODO 057 - Add Phone Launch and Diagnostic Commands

Add Make targets that launch the installed Android application and collect its
filtered startup/crash logs, so physical-phone troubleshooting does not require
remembering the app package or ADB syntax.

Completed 2026-10-17. `make -n launch logcat diagnose-phone` confirms the
generated ADB commands without requiring device access from the coding sandbox.
