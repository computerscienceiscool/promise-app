# Promise Messaging v0 frozen artifacts

`promise-messaging-v0-descriptor.json` is the exact byte sequence addressed by
the v0 protocol CID. Do not reformat it; any byte change defines a new version.

`promise-messaging-v0-vectors.json` contains public deterministic signing keys,
exact canonical envelope bytes, message CIDs, and rejection cases. It is the
shared Android/backend conformance input. The keys are test-only and never
belong in a build or a device identity store.
