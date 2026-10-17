# TODO 001 - Android/Kindle PromiseGrid Messaging POC

## Goal

Demonstrate two Android or Kindle Fire devices exchanging durable, signed
PromiseGrid-style messages through an inspectable backend. The demonstration
must show that the backend transports and stores exact evidence but does not
silently author, alter, accept, or execute a user promise.

## POC user journey

1. A user creates a device identity and sees its public identifier.
2. The user adds a peer by scanning or entering a public identifier.
3. The user composes and signs a small promise message: initially, a short
   text commitment addressed to that peer.
4. The app persists the exact signed envelope locally before upload.
5. The backend stores and relays the exact bytes; a second device learns that
   evidence is available.
6. The recipient fetches, validates, and quarantines the message before local
   acceptance.
7. The recipient can accept, decline, or retain the message as unaccepted
   evidence. Those choices create new signed evidence rather than changing the
   original sender message.
8. Both the mobile app and a backend operator screen show the message ID,
   envelope status, delivery attempts, verification result, and audit events.

## Architecture

### Android client

- Kotlin and Jetpack Compose; responsive layouts for phones and Kindle Fire
  tablets.
- No mandatory Google Mobile Services dependency; notification and platform
  integration sit behind Android interfaces so Kindle-compatible alternatives
  can be supplied.
- Android Keystore-backed device signing key; the private key never leaves the
  device in this POC.
- Local encrypted database for identity metadata, peers, exact envelope bytes,
  durable inbox/outbox state, verification reports, and user-visible audit
  history.
- Reliable outbox: sign and persist before network upload; retry the same
  immutable bytes with exponential backoff and explicit user-visible state.
- Quarantine inbox: received bytes are not displayed as accepted promises until
  parsing, size, protocol, proof, identity, and local-policy checks finish.
- Screens: onboarding/identity, peer list and peer add, compose, conversation
  and evidence detail, inbox/outbox, diagnostics, and settings.
- Accessibility, large-text, dark mode, offline operation, process-death
  recovery, and tablet layouts.

### Shared protocol core

- Deterministic builders and parsers for a POC-owned, versioned message
  protocol; no reliance on an assumed universal PromiseGrid SDK.
- Canonical CBOR, content IDs, tagged CID handling, envelope size limits,
  schema validation, and strict rejection of unknown or duplicate fields.
- A small outer envelope selected by the POC protocol identifier; the exact
  slot roles, signed projection, proof placement, payload rules, and failure
  behavior live in the POC protocol document.
- Message kinds: `promise_offer`, `promise_accept`, `promise_decline`, and
  `observation`. Every cross-message relationship is an immutable CID link.
- Separate verification outcomes for byte retrieval, envelope/payload validity,
  signature validity, and local identity/trust resolution.
- Stable test vectors shared by the Android client, backend verifier, and
  command-line diagnostic tooling.
- Version negotiation and clear unsupported-protocol handling; unsupported
  bytes may be retained as opaque evidence but are never semantically accepted.

### Backend

- A stateless authenticated API edge plus durable relational metadata store and
  immutable object store for exact envelope bytes.
- Device registration binds an account/session to a public device identity but
  never uploads private signing keys.
- Upload endpoint accepts an immutable envelope once, verifies its declared
  size and content ID, stores it idempotently, and records an append-only audit
  event.
- Retrieval endpoint returns the original bytes by content ID; no backend
  transformation, reserialization, or server-generated message content.
- Per-recipient mailbox/index records availability and delivery state, separate
  from the immutable envelope object.
- Sync endpoint supports cursors, pagination, repeated delivery, offline
  catch-up, idempotency keys, and acknowledged receipt of availability only.
- Notification adapter signals that a mailbox changed. A notification is never
  proof, acceptance, or a substitute for fetching and verifying evidence.
- Backend-side protocol verifier provides diagnostics and abuse filtering, but
  client-local verification and acceptance remain authoritative for the user.
- Admin/operator console exposes object IDs, envelope hashes, raw-byte download
  for authorized debugging, verification results, request logs, mailbox state,
  delivery attempts, and immutable audit history.
- OpenAPI contract, generated client boundary, migrations, configuration by
  environment, health/readiness endpoints, structured logs, metrics, traces,
  rate limits, quotas, and backup/restore rehearsal.
- Security controls: TLS, authenticated sessions, authorization checks,
  device/account revocation, replay/idempotency protection, input limits,
  audit log access control, secret management, dependency scanning, and threat
  model documentation.
- Privacy controls: minimal account data, retention policy, export/delete
  boundaries for account metadata, and clear distinction between deletable
  mailbox metadata and immutable protocol evidence retained for the POC.

## Protocol and PromiseGrid boundaries

- The POC protocol document owns message meaning, payload schema, proof rules,
  local acceptance, identity continuity, and failure behavior.
- The app will publish exactly which protocol document/version it implements
  and what each component claims to implement.
- The current PromiseGrid Development Guide treats the generic `grid(...)`
  envelope profile, ingress shape, final app SDK, and handler ABI as
  provisional (`DR-tuhaz`). This POC must label its envelope and relay choices
  as implementation/protocol choices, not as a universal Grid API.
- Backend transport, push delivery, UI state, caching, and scheduling are
  implementation details; they must not be presented as protocol authority.
- A phone notification, a server record, a successful upload, and message
  visibility are all distinct from recipient-local promise acceptance.
- Multi-device identity continuity, recovery, delegation, and key rotation
  require explicit protocol rules. Version one uses one non-exported signing
  key per installed device and does not silently clone identities.

## Demonstration and review evidence

- Two physical/emulated Android devices, including one Kindle-compatible
  configuration, exchange messages while one device is offline.
- Restart both app and backend during a queued delivery; prove one exact CID is
  retained and no duplicate semantic message is created.
- Show an altered envelope rejected by client and backend diagnostics.
- Show an unknown protocol version retained/quarantined without acceptance.
- Show server database/object storage has the exact bytes whose ID the clients
  display, while the server cannot forge an accepted recipient decision.
- Show audit trail from local outbox creation through upload, mailbox notice,
  fetch, verification, and recipient response.
- Automated tests: protocol vectors, malformed inputs, idempotency/replay,
  authz, migration, API integration, offline sync, process-death recovery, and
  Android UI flows.
- Produce an architecture diagram, threat model, API specification, protocol
  document, runbook, demo script, and backend-review checklist.

## Delivery order

1. Lock the POC protocol document and test vectors.
2. Create backend foundation, migrations, object store, audit ledger, and API.
3. Build Android identity, local store, envelope core, and diagnostic screen.
4. Connect outbox/inbox sync and verification.
5. Add peer onboarding, compose, response, and evidence-detail flows.
6. Add notifications, operator console, abuse controls, observability, and
   backup/restore proof.
7. Run adversarial and offline demonstrations; publish the review packet.

## Remaining mobile delivery checklist

- [ ] Finish Android canonical CID and envelope encoding.
- [ ] Sign envelopes with the persistent device key.
- [ ] Upload queued envelopes to the backend relay.
- [ ] Add upload retry, status, and visible error handling.
- [ ] Fetch recipient mailbox entries and exact envelope bytes.
- [ ] Verify incoming envelopes and show inbox/quarantine states.
- [ ] Test a two-device send/receive flow.
- [ ] Commit, push, and document the demo steps.

## Definition of done

The boss can inspect the source, API contract, database schema, exact stored
bytes, audit trail, tests, and running services; observe a signed message move
between two devices; and verify that neither relay delivery nor backend storage
is being misrepresented as user acceptance or protocol authority.
