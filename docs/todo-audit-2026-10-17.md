# POC TODO audit — 2026-10-17

This audit separates work that is present and host-testable from proof that can
only be observed on the connected phone and private LAN. It avoids marking a
demo claim complete merely because supporting code exists.

## Host-side checks completed

From `backend/`, the following completed successfully in the coding
environment (with a writable Go build cache):

```sh
GOCACHE=/tmp/promise-app-go-cache go test ./...
GOCACHE=/tmp/promise-app-go-cache go build ./cmd/promise-peer
../scripts/prepare-phone-computer-demo.sh --help
```

The test suite covers the relay API/store/protocol, proof checker, and the
computer peer. `promise-peer` is therefore ready for TODOs 047--053. The
following host-side tooling is also present:

| Need | Implemented tool/document |
| --- | --- |
| Private-LAN relay configuration | `scripts/start-lan-relay.sh` and `docs/device-relay-setup.md` |
| Stable desktop recipient identity | `backend/cmd/promise-peer identity` and `scripts/prepare-phone-computer-demo.sh` |
| Fetch, strict verification, and exact-byte retention | `backend/cmd/promise-peer inbox` |
| Correlated fetch/audit evidence capture | `scripts/capture-phone-computer-evidence.sh` |
| One-phone/one-computer operator procedure | `docs/phone-computer-demo.md` |
| Evidence fields and interpretation limits | `docs/poc-proof-evidence-template.md` |

## Remaining manual proof, not missing code

The following cannot be honestly completed from this sandbox. Each requires a
real host LAN, the authorized Android phone, or an observed user action:

1. **TODOs 015, 019, 024, and 040:** build the current Android project in the
   host terminal and retain the produced APK. Gradle cannot start in the coding
   sandbox because its file-lock service cannot obtain a wildcard interface.
2. **TODOs 025, 035, and 041:** install that APK on the actual test device(s).
3. **TODOs 014, 020, 026, 033, 039, and 049:** choose the host's private LAN
   IPv4 address and a private admin token, then run the relay. These values
   must not be invented, recorded in Git, or exposed by automation.
4. **TODOs 027, 034, 042, and 050:** save the actual LAN URL and recipient CID
   in the phone UI.
5. **TODOs 028--031, 036--038, 043--045, and 051--053:** sign/send on the
   phone, retrieve through a real relay, and record observed CIDs/screenshots/
   audit output. The computer alternative in `docs/phone-computer-demo.md`
   removes the need for a second physical phone.
6. **TODOs 008, 016, 021, 029, and 030:** prove two-device or offline/retry
   behavior by changing real connectivity and observing the durable state.
7. **TODOs 022 and 032:** only check parent/demo TODOs off after the evidence
   template has real observed values.

## Current implementation correction

This dated audit originally described an earlier Android inbox. The current
Android implementation retains fetched bytes in `QUARANTINED`, then immediately
performs strict format, CID, frozen-pCID, canonical payload/proof, Ed25519,
sender-binding, and recipient checks. A successful result becomes
`CANDIDATE_EVIDENCE`; failure remains a specific local rejection state.

That implementation is present, but a physical two-phone observation remains
open in TODOs 006--008 and 021. The computer peer has physically completed the
one-phone-to-computer verification path recorded in
[`example-poc.md`](example-poc.md).

Do not state that either Android or the computer accepted, trusted, or executed
a promise. An uploaded CID proves only that immutable signed bytes were
retained by the relay; verification proves byte/protocol/signature validity;
neither proves a recipient's acceptance.
