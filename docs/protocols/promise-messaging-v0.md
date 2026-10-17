# Promise Messaging v0

Status: frozen POC profile. This is not a frozen universal PromiseGrid
specification or SDK API.

## Purpose

This protocol records a small promise message and a recipient's response as
immutable signed evidence. It is for a two-device demonstration, not a general
chat, payment, authorization, or identity protocol.

## Protocol identity

The immutable v0 protocol descriptor is
`backend/testdata/promise-messaging-v0-descriptor.json`, whose exact UTF-8
bytes (including its final newline) have raw CIDv1 SHA-256:

```text
bafkreicutumpjz3skwoehjn3gptve5m3atgypmfbsdhbtapg2rcwt2c2hi
```

That CID is the sole v0 pCID selector accepted by this POC relay. The
descriptor is an intentionally small freeze artifact: changing it is a new
protocol version, not an edit to v0. The explanatory documents may be improved
without changing the selector; the test vector test verifies the descriptor
CID.

## Envelope

The POC uses the current *provisional* Grid-shaped profile:

```text
grid([42(protocol-pCID), payload, proof])
```

`payload` and `proof` are canonical CBOR byte strings. There are exactly three
slots. The complete encoded envelope is immutable; its message CID is computed
over its complete bytes.

The `grid(...)` wrapper and the tagged pCID selector follow current Guide
orientation. Their use here is a POC protocol choice, not a claim that this is
a frozen universal envelope API.

## Payload

The decoded payload is a closed CBOR map:

```text
{
  "kind": "promise_offer" | "promise_accept" | "promise_decline" | "observation",
  "from": 42(device-identity CID),
  "to": 42(device-identity CID),
  "body": text,
  "invokes": [42(message CID), ...],
  "created_at": unsigned integer
}
```

Rules:

- `from` and `to` are distinct CIDv1 device identities.
- `body` is UTF-8, nonempty, and at most 4,096 bytes.
- `created_at` is informational only; it does not expire a message or prove
  ordering.
- `invokes` is a sorted, duplicate-free list of zero through four CIDv1 values.
- An offer has no invocations.
- An acceptance or decline invokes exactly one offer addressed to its signer.
- An observation invokes one or more messages and records only the signer's
  observation in `body`; it does not rewrite another agent's promise.
- No payload field declares that a recipient accepted an offer. Acceptance is
  represented only by that recipient's independently signed message.

## Proof

The decoded proof is a closed CBOR map:

```text
{
  "algorithm": "Ed25519",
  "public_key": bstr32,
  "signature": bstr64
}
```

The device identity CID is derived from the canonical device-identity record
containing that public key. Version zero has one private signing key per device;
keys are generated and retained only in platform-protected storage.

The signature covers the deterministic envelope encoding with the `proof` slot
replaced by the canonical proof header containing only `algorithm` and
`public_key`. The complete envelope, including the signature, is then CIDed.

## Receiver behavior

1. Retain incoming bytes in quarantine with their retrieval source and time.
2. Reject from semantic processing if the envelope is oversized, malformed,
   non-canonical, uses the wrong pCID, has an invalid payload, has an invalid
   proof, or does not bind the proof key to `from`.
3. Record each failure separately: retrieval, format, payload, signature, or
   local identity/trust resolution.
4. A valid message is candidate evidence, not accepted state. The receiving
   user/app chooses whether to retain, accept, decline, or observe it.
5. Unknown pCIDs may be retained as opaque bytes but are never interpreted.

## Backend behavior

The backend may enforce byte-size limits, verify CIDs, and produce diagnostics.
It stores and returns exact bytes. It does not rewrite messages, synthesize
acceptance, infer promise completion, or turn a notification into evidence.

## Frozen test vectors

`backend/testdata/promise-messaging-v0-vectors.json` contains two deterministic
Ed25519 examples: an offer and the recipient's acceptance. It fixes the key
seeds, identities, payload fields, exact envelope bytes, and message CIDs.
Those private test seeds are public fixtures and must never be used by an app.

The backend test suite verifies that each fixture is byte-for-byte reproducible
and that altered bytes and an otherwise-valid unknown selector are rejected.
The relay also proves idempotent exact-byte upload/retrieval in its API tests.
Android conformance must consume these artifacts rather than duplicating their
values in a separate test fixture.

## Still outside this POC

- Key rotation, account recovery, and multi-device continuity.
- User-level identity, trust policy, authorization, and privacy policy.
- Production relay availability, anti-abuse controls, and data retention.
