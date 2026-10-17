# TODO 006 - Fetch Recipient Mailbox Envelopes

Fetch the relay mailbox and exact envelope bytes.

Implementation:

- `RelayClient.mailbox()` obtains the recipient's immutable CID listing.
- `RelayClient.fetchEnvelope(cid)` gets exact `application/cbor` bytes only
  after locally validating the CID-shaped path component.
- `InboxSync` retains fetched bytes before any interpretation and reports
  per-CID failures without discarding other mailbox entries.
- `InboxStore` records relay source and retrieval time and stores the exact
  bytes in Android private storage.

Integration completed:

- **Refresh inbox** in `MainActivity` invokes `InboxSync.sync()` using the
  persisted relay URL and the Android device identity.
- The result reports fetched, already-retained, and failed counts before the
  local inbox list refreshes.

Status: implementation is complete. A physical recipient-phone inbox run is
still pending under TODO 008; the completed phone-to-computer proof does not
claim that Android recipient-device test.
