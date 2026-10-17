# TODO 029 - Prove Duplicate Delivery Preserves One Message CID

Repeat delivery and prove the immutable object ID does not change.

Implementation is ready but has not been physically run:

- `promise-peer resend` re-submits retained, verified CBOR bytes without
  re-signing or editing them, and rejects a relay response with another CID.
- A host API test asserts that an exact retry returns the original CID with
  `created:false` and `mailbox_added:false` after the first successful upload.
- `make duplicate-local` requires exactly one retained inbox envelope, runs
  that resend through the USB relay, then requires proof-check evidence for
  one mailbox entry and at least two uploads/fetches.

Run and retain its output before checking this TODO off. Expected duplicate
response: the original CID with `created:false` and `mailbox_added:false`.
