# TODO 007 - Verify and Quarantine Incoming Envelopes

Verify incoming bytes before accepted message state.

Implemented quarantine foundation:

- Every fetched envelope enters `QUARANTINED`; no relay response can create
  accepted state.
- Private storage keeps the original bytes separately from local metadata.
- Explicit local outcomes exist for malformed payloads/proofs, unknown
  protocols, identity failures, and `CANDIDATE_EVIDENCE`.

Remaining: implement the protocol's strict CBOR decoder, canonical-byte/CID
checks, and Ed25519 verification before any transition to
`CANDIDATE_EVIDENCE`.
