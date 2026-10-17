# TODO 052 - Fetch and Verify the Phone Promise on the Computer

Run the peer inbox command and retain verified exact envelope bytes.

Completed 2026-10-17. `make peer-inbox-local` verified the exact message CID,
frozen POC protocol selector, sender/recipient identities, signature, and
retained the original CBOR bytes under `.local/promise-peer/inbox/`.
