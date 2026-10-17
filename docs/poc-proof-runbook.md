# Promise App POC proof runbook

This runbook describes the remaining two-device proof. It is separate from the
completed one-phone-to-computer proof in [example-poc.md](example-poc.md).
It produces evidence without treating the relay, a push/availability hint, or
an incoming byte stream as authority. The values shown on the two phones are
evidence to record.

## Preconditions

- Two Android or Kindle devices are on the same trusted Wi-Fi network.
- The current debug APK is installed on both devices.
- Go and `adb` are available on the development computer.
- Use a disposable relay data directory.  This POC relay stores envelope bytes
  and audit events on disk and has no TLS or operator authentication beyond the
  audit token; do not expose it to the public Internet.

## 1. Start and inspect the relay

From `promise-app`, start the relay on the development computer's LAN address.
Replace `192.168.1.10` with the computer's actual Wi-Fi IPv4 address.

```sh
export PROMISE_DATA_DIR=/tmp/promise-app-poc-data
export PROMISE_ADMIN_TOKEN='fresh-local-demo-secret'
RELAY=http://192.168.1.10:8080 DATA_DIR="$PROMISE_DATA_DIR" make relay
```

In another terminal, confirm the relay is alive:

```sh
curl -fsS http://127.0.0.1:8080/healthz
# expected: {"status":"ok"}
```

On both devices, open **Promise App**, enter
`http://192.168.1.10:8080` as the relay URL, and tap **Save relay**.

## 2. Prove signed, exact-byte delivery

1. On device B, copy the displayed **Device identity** CID.
2. On device A, paste that CID into **Recipient device identity CID**, enter a
   distinctive promise such as `POC-DELIVERY-<timestamp>: I will review this`,
   and tap **Sign, queue, and upload**.
3. Record the CID from A's status line and confirm its outbox state becomes
   `UPLOADED`. This demonstrates the exact signed bytes were persisted before
   upload and accepted by the relay.
4. On B, tap **Refresh inbox**. Record that the status reports one fetched
   envelope and the inbox lists the same CID as `CANDIDATE_EVIDENCE` after
   local format, CID, signature, sender, and recipient verification.
5. Tap **Refresh inbox** again. It must report the item as already retained;
   the inbox must not gain a duplicate item.

Record this proof with the private evidence template at
[`poc-proof-evidence-template.md`](poc-proof-evidence-template.md), then run:

```sh
go run ./cmd/proof-check \
  -data-dir "$PROMISE_DATA_DIR" \
  -cid '<sender CID>' \
  -recipient '<device-B identity CID>'
```

The expected `CANDIDATE_EVIDENCE` state is important: local verification does
not make relay bytes an automatic acceptance, execution, or trust decision.

## 3. Prove restart recovery

1. On B, force-stop Promise App in Android Settings, then reopen it.
2. Confirm the same inbox CID and `CANDIDATE_EVIDENCE` state remain visible.
3. On A, force-stop and reopen the app. Confirm the uploaded outbox item and
   its CID remain visible.

## 4. Prove offline retry keeps the original bytes

1. On A, temporarily replace the saved relay URL with
   `http://192.0.2.1:8080` and tap **Save relay**. (`192.0.2.0/24` is reserved
   test space and should not be routable.)
2. Address a new promise to B and tap **Sign, queue, and upload**. Wait for
   the attempt to fail; record its outbox CID, `RETRYABLE_FAILURE` state, and
   attempt count.
3. Force-stop A, reopen it, and confirm the same CID remains in its outbox as
   retryable. This proves the queue survived process restart.
4. Restore the working LAN relay URL and tap **Retry pending uploads**.
   Confirm the same CID, not a newly generated CID, reaches `UPLOADED` and its
   attempt count increases.
5. On B, tap **Refresh inbox** and confirm that same retry CID appears once as
   `CANDIDATE_EVIDENCE`. Refresh again and confirm no duplicate is retained.

The relay cannot attest that A was offline before upload. The offline part of
this proof is the device A outbox state, attempt count, and before/after
screenshots. After B fetches the recovered item, verify relay-retained evidence
with the same `proof-check` command from step 2, using the retry CID.

## 4a. Prove a duplicate upload has one immutable CID

This is separate from an inbox refresh: resubmit the exact already-uploaded
envelope bytes. From the development computer, the relay's retained object is
an appropriate diagnostic input. Substitute A's identity and the LAN URL:

```sh
curl --fail -sS -X POST \
  -H 'Content-Type: application/cbor' \
  -H 'X-Device-ID: <device-A identity CID>' \
  --data-binary "@$PROMISE_DATA_DIR/objects/<duplicate-test CID>.cbor" \
  http://192.168.1.10:8080/v1/envelopes
```

The response must report the same CID with `created:false` and
`mailbox_added:false`. Refresh B twice after that action; its local inbox must
still have one item for the CID.

```sh
go run ./cmd/proof-check \
  -data-dir "$PROMISE_DATA_DIR" \
  -cid '<duplicate-test CID>' \
  -recipient '<device-B identity CID>' \
  -min-uploads 2 \
  -min-fetches 2
```

## 5. Show backend evidence to the reviewer

With the relay still running, retrieve the append-only audit log:

```sh
curl -fsS \
  -H "Authorization: Bearer $PROMISE_ADMIN_TOKEN" \
  http://127.0.0.1:8080/v1/admin/audit
```

For each message CID recorded above, show an `envelope.upload` event and an
`envelope.fetch` event for B. The relay data directory can also be inspected:

```sh
find "$PROMISE_DATA_DIR" -maxdepth 3 -type f -print
```

The reviewer should see immutable CBOR object bytes, mailbox metadata, and the
audit NDJSON file. Do not edit this directory during the proof.

For every message CID, `proof-check` verifies the retained bytes against the
frozen profile, confirms one recipient mailbox record, and counts matching
upload/fetch events. It is a local POC checker, not a witness or a claim that
the audit log is tamper-evident.

## Acceptance record

Record the following in the demo notes before checking off TODO 021:

| Check | Evidence to capture |
| --- | --- |
| Device identities | A and B displayed CIDs |
| Delivery | sender CID, recipient CID, relay audit upload/fetch events |
| No duplicate delivery | second refresh result and unchanged inbox count |
| Restart recovery | before/after screenshots of the same outbox and inbox CIDs |
| Offline retry | same CID before and after retry, state/attempt count change |
| Backend boundary | `CANDIDATE_EVIDENCE` on B; no automatic acceptance/effect |

## Closeout gate

Only mark TODO 021 complete after all rows above have evidence. Mark TODO 022
complete only after the evidence is attached or recorded for the demo and every
completed TODO is checked in `TODO/TODO.md`. A validation failure is useful
evidence: leave the relevant TODO open and record the failure rather than
claiming the proof passed.
