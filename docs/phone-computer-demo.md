# Phone-to-computer Promise POC

This is the shortest boss-facing demonstration: one Android/Kindle phone sends
one signed promise to a computer-owned recipient identity. The relay stores and
routes bytes; it does not accept, alter, or execute the promise.

## Before the demo

- The computer and phone must be on the same private Wi-Fi network.
- Build and install the current debug APK on the phone.
- Select the computer's private LAN IPv4 address, such as `192.168.1.10`.
  Do not use `localhost` or `127.0.0.1`; those name the phone itself when used
  on the phone.
- From `promise-app`, choose a new private admin token. Keep it on the
  computer; it is never entered in the app.

## TODO 049 — start the relay

In terminal one, from `promise-app`, start the private-LAN relay:

```sh
export PROMISE_ADMIN_TOKEN='a-new-long-private-demo-token'
export PROMISE_ADVERTISE_URL='http://192.168.1.10:8080'
./scripts/start-lan-relay.sh
```

Leave this terminal open. In a second terminal, verify the host service before
involving the phone:

```sh
curl --fail http://127.0.0.1:8080/healthz
# {"status":"ok"}
```

Before the phone steps, run the host validation. It runs the backend tests,
checks the desktop peer identity, validates an optional APK, and checks both
the loopback and LAN health endpoints without sending a message:

```sh
./scripts/validate-phone-computer-host.sh \
  -relay "$PROMISE_ADVERTISE_URL" \
  -apk android/app/build/outputs/apk/debug/app-debug.apk \
  --check-relay
```

If the computer firewall is enabled, allow TCP 8080 only from the private Wi-Fi
network for this demo. Never expose this HTTP relay to the public Internet.

## TODO 050 — configure the phone and computer identity

In terminal two, prepare a persistent computer identity and print the exact
values required by the phone:

```sh
cd promise-app
PROMISE_ADVERTISE_URL='http://192.168.1.10:8080' \
  ./scripts/prepare-phone-computer-demo.sh
```

The script creates the private computer key under `.local/` with restricted
permissions on first use and prints a public CID. It does not start a relay or
send any data.

On the phone:

1. Open **Promise App**.
2. Set **Relay URL** to the printed LAN URL, then tap **Save relay**.
3. Paste the printed **Computer peer CID** into **Recipient device identity
   CID**. Confirm it is the complete lowercase-base32 CIDv1 (it begins with
   `b`; the computer identity CID is not required to begin with `bafk`).

Optionally, open `<relay URL>/healthz` in the phone browser first. It must show
`{"status":"ok"}`. Fix Wi-Fi, firewall, VPN, or guest-network isolation before
sending if it does not.

## TODO 051 — sign and send the phone promise

Enter a distinctive, non-sensitive test message, for example:

```text
POC-PHONE-COMPUTER-2026-10-17: I will review the proposal.
```

Tap **Sign, queue, and upload**. Record the message CID shown in the phone's
outbox status and wait for its state to become `UPLOADED`. A CID and uploaded
state demonstrate that the phone persisted exact signed bytes before relay
upload; they do not imply that the computer has accepted the promise.

Continue with [TODO 052](../TODO/052-fetch-computer-promise.md) to fetch and
verify the bytes on the computer, then [TODO 053](../TODO/053-record-phone-computer-audit.md)
to preserve the evidence.
