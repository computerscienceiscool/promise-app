# TODO 030 - Prove Offline Retry Preserves One Message CID

Queue while offline, retry after connectivity returns, and prove the same CID is uploaded.

Implementation evidence is in place:

- Android persists the exact signed bytes before its first upload attempt and
  retries that record rather than regenerating an envelope.
- The relay API test verifies that posting the same verified bytes twice
  retains one object and one mailbox reference under the same CID.
- `docs/poc-proof-runbook.md` specifies the required phone screenshots,
  restart check, recipient fetch, and retained relay evidence.

This TODO remains open until those steps are performed on physical devices.
