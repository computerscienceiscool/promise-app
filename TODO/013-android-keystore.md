# TODO 013 - Move Android Private Key Into Android Keystore

Replace SharedPreferences private-key persistence with Android Keystore-backed signing.

Completed: `SigningIdentity` now generates and loads a non-exportable Ed25519
key under the Android Keystore alias `promisegrid.messaging.ed25519.v1`.
The old exported SharedPreferences key material is deleted only after the new
Keystore identity is available. Android API 31 is required because this POC's
Ed25519 protocol has no secure Keystore equivalent on earlier API levels.
