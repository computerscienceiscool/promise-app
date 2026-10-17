# TODO 058 - Handle Android Keystore Ed25519 Public-Key Encodings

Accept both parameter-free and legacy-NULL SubjectPublicKeyInfo forms emitted
by Android Keystore for the same Ed25519 public key, while requiring the
Ed25519 OID and an exact 32-byte BIT STRING.
