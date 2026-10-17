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

Remaining integration: expose `InboxSync.sync()` through the activity after
the device-identity CID is wired into the relay client.
