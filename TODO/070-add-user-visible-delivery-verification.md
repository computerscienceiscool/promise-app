# TODO 070 - Add User-Visible Delivery Verification

Make the one-phone-to-computer demonstration understandable without reading
terminal output or inferring success from a technical status string.

The phone UI and supporting host command must clearly distinguish:

1. **Relay ready** — the selected relay URL returned a health response.
2. **Message queued on this phone** — exact signed bytes and their message CID
   are durable in the phone outbox.
3. **Relay accepted upload** — the relay acknowledged those exact bytes. This
   is not recipient verification.
4. **Computer recipient verified** — the computer peer fetched the message and
   independently verified its CID, Grid-CBOR shape, frozen pCID, signature, and
   recipient binding.

Acceptance criteria:

- A person can follow one short on-screen path from QR scan to an unambiguous
  relay-ready and upload result.
- The app scans a recipient QR code directly and fills **Recipient identity**;
  it does not require a separate scanner app or clipboard copy/paste step.
- The app presents the message CID with a copy action and human-readable state,
  rather than requiring interpretation of raw log output.
- The computer-side verifier provides one concise success/failure result tied
  to that CID, suitable for showing during a demo.
- Documentation names the exact evidence each state proves and explicitly says
  that relay upload alone is not recipient acceptance or verification.
- Add deterministic tests for the new state/result mapping.

Completed 2026-10-26: the attached phone scanned the computer recipient QR,
showed `Relay check: ready`, queued and uploaded a message with visible CID
`bafkreiglye4caxojzlsltudfzdnipxh7kg3g63hdkxnpex2cxep3fez7di`, and the
computer peer independently returned `RECIPIENT VERIFIED` for that same CID.
The same phone then used the direct private-LAN Wi-Fi URL and produced the
separately verified CID
`bafkreifqo3nidty44ezg7hf6wcgoqnq2aq2gz2qz7xtubmsftdb7gsggly`.
