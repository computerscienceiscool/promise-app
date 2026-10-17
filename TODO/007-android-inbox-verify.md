# TODO 007 - Verify and Quarantine Incoming Envelopes

Verify incoming bytes before accepted message state.

Implemented quarantine foundation:

- Every fetched envelope enters `QUARANTINED`; no relay response can create
  accepted state.
- Private storage keeps the original bytes separately from local metadata.
- Explicit local outcomes exist for malformed payloads/proofs, unknown
  protocols, identity failures, and `CANDIDATE_EVIDENCE`.
- A valid envelope addressed to a different device is rejected as an identity
  mismatch; a relay mailbox listing is only an availability hint.

Implemented: `InboxVerifier` now strictly decodes the fixed envelope schema,
checks canonical CBOR and the complete-envelope CID, validates the frozen
protocol selector, verifies the Ed25519 signing projection, and derives the
sender identity from the verified public key. Only then does it transition to
`CANDIDATE_EVIDENCE`; all other outcomes remain explicit local rejection
states.

The Android project now builds and installs on the host and the sender screen
has launched on the physical phone. The remaining proof is an inbound Android
recipient-device run that exercises these verification states; that belongs to
TODO 008. This TODO remains open until that run is retained as evidence.
