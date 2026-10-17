# TODO 058 - Handle Android Keystore Ed25519 Public-Key Encodings

Accept both parameter-free and legacy-NULL SubjectPublicKeyInfo forms emitted
by Android Keystore for the same Ed25519 public key, while requiring the
Ed25519 OID and an exact 32-byte BIT STRING.

Also rotate the incompatible local P-256 key created by the early POC build
and generate an explicit `ECGenParameterSpec("ed25519")` Android Keystore key.

Completed 2026-10-17. The rebuilt app launched through `make diagnose-phone`
on the authorized physical device with no Android runtime crash.
