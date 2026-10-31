# TODO 029 - Prove Duplicate Delivery Preserves One Message CID

Repeat delivery and prove the immutable object ID does not change.

Completed 2026-10-26 using the phone's Wi-Fi-uploaded CID
`bafkreifqo3nidty44ezg7hf6wcgoqnq2aq2gz2qz7xtubmsftdb7gsggly`:
`promise-peer resend` returned that identical CID with `created:false` and
`mailbox_added:false`. The subsequent proof check passed with one mailbox
entry, two uploads, and two recipient fetches.

Implementation:

- `promise-peer resend` re-submits retained, verified CBOR bytes without
  re-signing or editing them, and rejects a relay response with another CID.
- A host API test asserts that an exact retry returns the original CID with
  `created:false` and `mailbox_added:false` after the first successful upload.
- `make duplicate-local` requires exactly one retained inbox envelope, runs
  that resend through the USB relay, then requires proof-check evidence for
  one mailbox entry and at least two uploads/fetches.

Run and retain its output before checking this TODO off. Expected duplicate
response: the original CID with `created:false` and `mailbox_added:false`.
