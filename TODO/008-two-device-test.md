# TODO 008 - Test Two-Device Delivery

Completed with two independent peers: the Android development phone and the
computer peer. Two phones were not required.

The retained POC evidence proves:

1. The phone created and signed an immutable envelope, retained it in its
   durable outbox, and uploaded it through the USB relay.
2. The computer peer, using a distinct persistent identity, listed its relay
   mailbox, fetched the exact CBOR bytes, and independently verified the
   frozen protocol selector, CID, recipient, and Ed25519 signature.
3. The proof checker retained audit evidence showing one mailbox entry,
   upload, and fetch for the same CID.

Evidence and exact commands are retained in `docs/example-poc.md`. This is
experimental POC evidence, not a claim that a second Android phone, Kindle,
or production transport has been proven.
