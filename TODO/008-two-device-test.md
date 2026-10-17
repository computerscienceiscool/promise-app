# TODO 008 - Test Two-Device Delivery

Prove signed send, relay, fetch, verification, and restart behavior.

Test plan (run after TODOs 003--007 are fully wired into the activity):

1. Install the debug APK on devices A and B; record each displayed device CID.
2. Start the backend on a LAN-reachable HTTPS relay and configure both apps
   with its URL.
3. On A, create a signed offer addressed to B, upload it, and save the
   response CID.
4. On B, sync the mailbox. Confirm the CID is listed, exact bytes are fetched,
   and the item is initially shown as `QUARANTINED`.
5. Run verification. Confirm a valid envelope moves only to
   `CANDIDATE_EVIDENCE`; it must not appear as accepted automatically.
6. Force-stop and reopen B. Confirm quarantine/candidate metadata and exact
   bytes persist.
7. Alter one retrieved byte in an isolated test fixture. Confirm it never
   becomes candidate evidence and is recorded as `SIGNATURE_REJECTED` (or the
   appropriate format/payload state).
8. Retry the same A upload and B sync. Confirm no duplicate local entry and
   that the backend returns the same immutable CID/bytes.
