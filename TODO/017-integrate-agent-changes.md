# TODO 017 - Integrate Android UI, Protocol, and Keystore Changes

Review and integrate TODOs 011 through 013 into one coherent Android application.

Completed: `MainActivity` now uses the Keystore-backed signing identity to
create envelopes, persists their exact bytes in the durable outbox before
upload, and presents upload/retry plus quarantine-inbox controls and states.
The backend Go suite passes for the integrated relay behavior.
