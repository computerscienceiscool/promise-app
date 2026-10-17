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
