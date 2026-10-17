# Promise App backend

Runnable POC relay for exact signed Promise Messaging v0 envelopes.

```sh
PROMISE_ADMIN_TOKEN=local-demo go run ./cmd/promise-server
```

Endpoints:

- `POST /v1/envelopes` accepts canonical CBOR. `X-Device-ID` must match the
  signed sender identity.
- `GET /v1/mailbox` lists the caller's available message CIDs.
- `GET /v1/envelopes/{cid}` returns original bytes only to the addressed device.
- `GET /v1/admin/audit` requires `Authorization: Bearer <admin token>`.
- `GET /healthz` is an unauthenticated liveness check.

The filesystem data directory deliberately exposes immutable object bytes,
mailbox metadata, and append-only NDJSON audit events for POC review. It is not
a production deployment topology.

## Computer recipient peer

For a phone-to-computer demonstration, the `promise-peer` command acts as the
recipient. It creates one persistent Ed25519 device identity (mode `0600`),
prints its CID, fetches only that CID's mailbox, verifies each returned exact
CBOR envelope and CID, and retains verified bytes locally. It does **not**
accept a promise simply by fetching it.

From `backend`, first build and test the peer command:

```sh
go build ./cmd/promise-peer
go test ./cmd/promise-peer
```

Keep the computer peer state in one known directory. The private key must stay
on the computer: copy only the CID printed by this command into the phone's
Recipient field.

```sh
PEER_STATE="$PWD/.promise-peer"
go run ./cmd/promise-peer identity \
  -identity-file "$PEER_STATE/identity.ed25519"
```

Running the same command again must print the same CID. It is the computer
recipient identity. Do not delete, copy, or share `identity.ed25519`; deleting
it creates a different recipient identity and makes its prior mailbox
unavailable to this peer.

After sending from the phone, retrieve and verify the computer inbox:

```sh
go run ./cmd/promise-peer inbox \
  -relay http://192.168.1.10:8080 \
  -identity-file "$PEER_STATE/identity.ed25519" \
  -inbox-dir "$PEER_STATE/inbox" \
  -evidence-file "$PEER_STATE/fetch-evidence.json"
```

Each line is JSON containing the verified message CID, sender, recipient,
body, and `.promise-peer/inbox/<cid>.cbor` path. Keep that CBOR file as the
demo's exact signed evidence. `fetch-evidence.json` records the verified fetch
without including a private key or relay admin token.

For a repeatable evidence record, pass `-evidence-file` while fetching. The
JSON report records the relay URL, recipient identity, verified display fields,
and the retained exact-byte path; it contains neither the peer private key nor
the relay admin token:

```sh
go run ./cmd/promise-peer inbox \
  -relay http://192.168.1.10:8080 \
  -evidence-file /private/demo-notes/computer-fetch.json
```

The repository script combines a computer fetch/verification, the local
immutable-object/mailbox/audit check, and a raw authenticated audit capture.
Keep its output directory private: it may contain device CIDs and message text.

```sh
../scripts/capture-phone-computer-evidence.sh \
  -relay http://192.168.1.10:8080 \
  -token "$PROMISE_ADMIN_TOKEN" \
  -data-dir "$PROMISE_DATA_DIR" \
  -cid '<phone message CID>' \
  -recipient '<computer peer CID>' \
  -out-dir /private/demo-notes/phone-computer-proof
```

To run the same relay as a container:

```sh
docker build -t promise-relay .
docker run --rm -p 8080:8080 -e PROMISE_ADMIN_TOKEN=local-demo -v promise-data:/data promise-relay
```
