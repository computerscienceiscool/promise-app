# Promise App POC evidence record

Copy this file into a dated, private demo-notes location before running the
proof. Fill it with observed values and screenshots. Do not add identities,
message bodies, relay tokens, or private keys to the repository.

This record is POC evidence for the frozen `promise-messaging-v0` profile. It
does not claim a universal PromiseGrid mobile, relay, identity, or ingress API.
The `grid([42(pCID), payload, proof])` shape is this POC's deliberate choice
under current provisional Guide orientation (`DR-tuhaz`).

## Session

| Field | Value |
| --- | --- |
| Date/time and operator | |
| Relay LAN URL | |
| Relay data directory (local path only) | |
| Frozen pCID | `bafkreicutumpjz3skwoehjn3gptve5m3atgypmfbsdhbtapg2rcwt2c2hi` |
| Device A identity CID | |
| Device B identity CID | |
| APK build/version on each device | |

## TODO 028 — first signed message

| Evidence | Value / attachment |
| --- | --- |
| Message CID shown on device A | |
| A outbox state is `UPLOADED` | |
| B refresh result | |
| B inbox CID and `QUARANTINED` state | |
| First refresh screenshot | |

Run the verifier after the recipient fetched the message:

```sh
go run ./cmd/proof-check \
  -data-dir "$PROMISE_DATA_DIR" \
  -cid '<message CID>' \
  -recipient '<device B identity CID>'
```

Paste the `PASS` output here:

## TODOs 037–038 — two-phone send and inbox inspection

| Evidence | Value / attachment |
| --- | --- |
| Recipient CID copied from B and pasted on A | |
| Full message CID shown on A after signing | |
| A status/outbox is `UPLOADED` and attempt count | |
| B first refresh status (`fetched` / failures) | |
| B inbox full CID, `QUARANTINED`, and detail | |
| B second refresh reports already retained; inbox count unchanged | |

The device screenshots must show full CIDs. Do not put message text, private
keys, relay tokens, or identity CIDs into this repository.

## TODO 029 — duplicate delivery, one CID

| Evidence | Value / attachment |
| --- | --- |
| Original message CID | |
| Same exact CID submitted again | |
| Second upload response (`created:false`, `mailbox_added:false`) | |
| B refresh result after duplicate | |
| B inbox count remains one for that CID | |
| Second refresh screenshot | |

To submit exactly the stored immutable bytes a second time:

```sh
curl --fail -sS -X POST \
  -H 'Content-Type: application/cbor' \
  -H 'X-Device-ID: <device A identity CID>' \
  --data-binary "@$PROMISE_DATA_DIR/objects/<message CID>.cbor" \
  http://<relay LAN address>:8080/v1/envelopes
```

Run the verifier with two uploads and two fetches:

```sh
go run ./cmd/proof-check \
  -data-dir "$PROMISE_DATA_DIR" \
  -cid '<message CID>' \
  -recipient '<device B identity CID>' \
  -min-uploads 2 \
  -min-fetches 2
```

Paste the `PASS` output here:

## TODO 030 — offline retry, one CID

| Evidence | Value / attachment |
| --- | --- |
| CID queued while relay was unreachable | |
| A state/attempt count before retry | |
| A restart screenshot showing same CID | |
| CID after retry | |
| A state/attempt count after retry | |
| B fetched the same CID once | |

The relay audit can prove that it retained/fetched the recovered CID, but it
cannot by itself prove that A was offline. Preserve the phone outbox screenshots
and attempt-count values as the offline portion of this proof.

```sh
go run ./cmd/proof-check \
  -data-dir "$PROMISE_DATA_DIR" \
  -cid '<message CID>' \
  -recipient '<device B identity CID>'
```

Paste the `PASS` output here:

## TODO 031 — relay audit evidence

| Evidence | Value / attachment |
| --- | --- |
| Object file path and byte count | |
| Recipient mailbox file path | |
| Audit file path | |
| `envelope.upload` event(s) | |
| `envelope.fetch` event(s) | |
| Audit API response saved privately | |

For each recorded CID, retain the verifier output and inspect the raw files:

```sh
find "$PROMISE_DATA_DIR" -type f -maxdepth 3 -print
curl --fail -sS \
  -H "Authorization: Bearer $PROMISE_ADMIN_TOKEN" \
  http://127.0.0.1:8080/v1/admin/audit
```

The audit log is append-only POC storage, not an independently witnessed or
tamper-evident ledger. Its value in this demo is operator-visible correlation
with retained exact bytes and mailbox metadata.

## TODOs 052–053 — phone-to-computer fetch and audit evidence

The computer peer treats a fetched envelope as `verified evidence`, not as an
accepted promise. Retain these private artifacts outside the repository:

| Artifact | Private location / observed result |
| --- | --- |
| Computer peer identity CID | |
| `computer-fetch.json` verification report | |
| Retained exact `<message CID>.cbor` path | |
| `proof-check.txt` `PASS` result | |
| `relay-audit.json` response | |
| Upload and fetch event count for message CID | |

From the repository root, after the phone has uploaded and the relay remains
available, capture all artifacts in one private directory:

```sh
scripts/capture-phone-computer-evidence.sh \
  -relay http://<computer LAN address>:8080 \
  -token "$PROMISE_ADMIN_TOKEN" \
  -data-dir "$PROMISE_DATA_DIR" \
  -cid '<phone message CID>' \
  -recipient '<computer peer CID>' \
  -out-dir /private/demo-notes/phone-computer-proof
```

`proof-check.txt` proves retained bytes validate and have one recipient mailbox
entry plus at least one upload and fetch audit event. The report does not prove
acceptance, user intent, ordering, or that the audit log is tamper-evident.
The capture command also fails unless the requested message CID is fetched and
verified by the computer peer; it will not silently record unrelated mailbox
messages as proof for this run.
