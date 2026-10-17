# TODO 002 - Freeze and Test the Initial POC Message Protocol

Define the exact POC message contract before implementation: canonical bytes,
protocol identifier, payload schemas, signing projection, device identity
rules, CID calculation, validation failures, and cross-language test vectors.

The draft protocol is in `docs/protocols/promise-messaging-v0.md`. Freeze its
exact bytes and derive its pCID before an Android client or backend claims to
implement it.
