# TODO 007 - Verify and Quarantine Incoming Envelopes

Verify incoming bytes before accepted message state.

Implemented quarantine foundation:

- Every fetched envelope enters `QUARANTINED`; no relay response can create
  accepted state.
- Private storage keeps the original bytes separately from local metadata.
- Explicit local outcomes exist for malformed payloads/proofs, unknown
  protocols, identity failures, and `CANDIDATE_EVIDENCE`.

Implemented: `InboxVerifier` now strictly decodes the fixed envelope schema,
checks canonical CBOR and the complete-envelope CID, validates the frozen
protocol selector, verifies the Ed25519 signing projection, and derives the
sender identity from the verified public key. Only then does it transition to
`CANDIDATE_EVIDENCE`; all other outcomes remain explicit local rejection
states.

Verification needs an Android host build/device test before this TODO can be
checked off. The sandbox cannot start Gradle's native Linux platform library.
