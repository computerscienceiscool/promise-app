# TODO 038 - Refresh and Inspect the Second Phone Inbox

Refresh the second device inbox and inspect the received quarantined evidence.
This is a manual device action; do not check it off until the values below are
recorded in private demo notes.

1. On phone B, confirm the saved relay URL matches phone A's URL.
2. Tap **Refresh inbox** once. Record the `fetched` count from the status line.
3. Inspect the Inbox section. It must show the full message CID from TODO 037,
   state `QUARANTINED`, and `Exact relay bytes retained; verification pending`.
4. Screenshot the status and Inbox section; record the full CID and state.
5. Tap **Refresh inbox** again. It should report the same object as already
   retained and the Inbox count must not increase.

Expected boundary: `QUARANTINED` means B has kept exact relay bytes locally.
It is deliberately not `CANDIDATE_EVIDENCE`, acceptance, an instruction to
perform work, or proof that the relay is authoritative.
