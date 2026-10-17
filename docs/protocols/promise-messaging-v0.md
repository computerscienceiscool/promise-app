# Promise Messaging v0

Status: draft POC protocol; not a frozen PromiseGrid specification.

## Purpose

This protocol records a small promise message and a recipient's response as
immutable signed evidence. It is for a two-device demonstration, not a general
chat, payment, authorization, or identity protocol.

## Protocol identity

This document must be frozen as exact bytes before implementation claims a
pCID. Until then, all v0 traffic is local test traffic and must not be called
interoperable PromiseGrid traffic. The implementation records the final pCID
in its configuration, test vectors, diagnostics, and conformance statement.

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

## Required test vectors

- Canonical offer and acceptance examples with exact encoded bytes and CIDs.
- Valid and invalid signing projections.
- Invalid CIDs, duplicate keys, unknown fields, non-canonical CBOR, oversized
  messages, incorrect pCID, altered bytes, duplicate invocations, and invalid
  response correlation.
- Idempotent backend upload/retrieval of one exact message CID.

## Open before freeze

- Exact device-identity record and CID codec.
- Envelope tag/encoding implementation and canonical-CBOR library profile.
- Final body-size and whole-envelope limits.
- Whether `observation` belongs in v0 or a separate protocol.
- Key rotation, account recovery, and multi-device continuity.
