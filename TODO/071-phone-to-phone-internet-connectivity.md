# TODO 071 - Prove relay-mediated two-phone messaging over the Internet

**Status:** `[ ]` Open

**Transport model:** this is not a direct phone-to-phone connection. The
intended delivery path is:

```text
Phone A -> reachable relay -> Phone B
```

Phone A uploads a signed envelope to the relay. Phone B independently fetches
and verifies it later. The phones may be on different networks and do not need
to be mutually reachable. The relay stores and routes exact bytes; it does not
accept, alter, execute, or make a human decision about a promise.

**Problem:** the current office Wi-Fi isolates clients, so it cannot prove that
two Android phones can each reach the relay through the intended remote path.
The messaging proof works over a hotspot and over USB; those are useful
controls, not an Internet two-phone proof.

## Establish the test boundary

- [ ] Record the two phones, app build, relay endpoint, and network used for
  each run without recording private device data or credentials.
- [ ] Run the same signed-message flow on a hotspot and USB first as control
  cases, then compare a remote/Internet path.
- [ ] Confirm whether the blocker is office client isolation, reachability of
  the relay, Android network policy, or an app-level behavior.

## SSH path investigation

Evaluate these alternatives as bounded experiments; select none as the final
architecture merely because it works first.

- [ ] Kotlin SSH library: identify a maintained candidate, Android support,
  host-key verification behavior, background/lifecycle constraints, and how a
  local forwarded relay URL would be configured.
- [ ] Go `golang.org/x/crypto/ssh`: assess whether the existing Go-side work
  can create and supervise the needed tunnel without moving key custody into
  an unsafe location or pretending that SSH is PromiseGrid transport.
- [ ] Standalone SSH client on each phone: document the manual setup and
  evaluate it as a fastest proof path, including host-key checking and how the
  app receives the forwarded endpoint.

## Validate the chosen experiment

- [ ] Use two phones on genuinely separate networks to send, upload, fetch,
  verify, and display one signed message through the remote relay route.
- [ ] Prove the app surfaces an actionable failure when the tunnel or relay is
  unavailable; do not silently fall back to a different network path.
- [ ] Repeat after a tunnel/network interruption and document reconnect or
  recovery behavior.
- [ ] Capture redacted commands, endpoint placeholders, and observed results
  in the demo/troubleshooting documentation.

SSH is an optional experiment for reachability. It does not create a
PromiseGrid identity, trust, discovery, or protocol requirement.

**Done when:** two Android phones on separate networks complete one documented
signed-message exchange through a reproducible remote path, or the experiments
produce a clear, evidenced blocker and next decision.
