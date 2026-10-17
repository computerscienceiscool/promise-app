# TODO 018 - Update Android Envelope Selector to Frozen pCID

Set Android protocol selector to `bafkreicutumpjz3skwoehjn3gptve5m3atgypmfbsdhbtapg2rcwt2c2hi` and verify backend acceptance.

Completed: `PromiseEnvelope.protocolCid` uses the frozen descriptor CID, the
relay accepts only that selector, and `go test ./...` passes in `backend`.
Android compilation remains tracked separately in TODO 019 because Gradle
cannot start inside this sandbox (it cannot obtain a usable wildcard IP).
