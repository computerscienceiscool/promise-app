# Promise App POC

An inspectable Android-to-computer signed-message demonstration. The phone
signs and persists immutable CBOR bytes before upload. The relay keeps and
returns those exact bytes. The computer peer independently verifies and retains
them. Relay delivery or storage is not recipient acceptance of a promise.

## Scope

The local Promise Messaging v0 descriptor is frozen at:

```text
bafkreicutumpjz3skwoehjn3gptve5m3atgypmfbsdhbtapg2rcwt2c2hi
```

This is a POC implementation, not a universal PromiseGrid SDK. The Android UI,
relay, ADB bridge, local directories, and QR workflow are host-specific
implementation choices. The bundled PromiseGrid Development Guide keeps the
generic app API, envelope family, and ingress/handler boundary provisional
under `DR-tuhaz`. The POC explicitly documents its selected ingress rather
than claiming it as a universal Grid rule.

See [the protocol contract](docs/protocols/promise-messaging-v0.md).

## What a successful demo proves

- Phone private signing material remains in Android Keystore.
- Exact signed bytes are durably queued before upload and retries reuse them.
- The relay stores immutable bytes and append-only audit events.
- The computer recomputes CID, validates canonical CBOR and the frozen
  selector, verifies signature and recipient identity, then retains exact CBOR.

It does not claim a global identity registry, universal transport, recipient
acceptance, production TLS, or a completed two-phone/Kindle proof.

## Requirements

The POC is developed and tested on Linux. To build the Android app, run the
relay, and perform the connected-phone demonstration, install:

- Git and GNU Make.
- Go **1.24 or newer** for the relay, protocol tests, QR helper, and computer
  peer.
- JDK **17** for the included Gradle wrapper.
- Android SDK Platform **36**, Build Tools, and Android platform tools
  (`adb`). The app supports Android **8.0 / API 26** and newer.
- A USB data cable and an Android device with Developer options and USB
  debugging enabled. Accept the computer's RSA debugging prompt on the phone.
- `curl` for relay-health checks. A desktop opener such as `xdg-open` is
  optional, but used by the QR-opening commands.

The first Gradle build may download its dependencies unless they are already
cached. The USB demonstration does not need the phone and computer to reach
each other over Wi-Fi.

## Install

Clone the repository and enter it:

```sh
git clone https://github.com/computerscienceiscool/promise-app.git
cd promise-app
```

Point Gradle at the installed Android SDK. Change the path if your SDK is in a
different location:

```sh
printf 'sdk.dir=%s\n' "$HOME/Android/Sdk" > android/local.properties
```

Connect and authorize the phone. This must show `device`, not `unauthorized`:

```sh
adb start-server
adb devices
```

Build, install, and start the app:

```sh
make install
make launch
```

To copy the newest screenshot already saved on the connected phone into
`.local/phone-screenshots/`, run `make pull-screenshot`.

If the app does not open, run `make diagnose-phone`. It starts the app and
prints only the relevant crash/error log. See [Troubleshooting](docs/troubleshooting.md)
for the common USB, SDK, and relay cases.

## Fast connected-phone demo

Keep the relay terminal open until the demo ends.

### Start relay

```sh
cd ~/lab/cswg/apps/promise-app
RELAY=http://YOUR-LAN-IP:8080 \
PROMISE_ADMIN_TOKEN='fresh-local-demo-secret' \
make relay
```

### Build and prepare phone

In another terminal:

```sh
make install
make demo-usb-setup
```

`demo-usb-setup` uses the USB-connected phone to avoid Wi-Fi client isolation,
checks host relay health, and opens a QR for the persistent computer recipient
CID. It is a local demo transport, not a protocol requirement.

### Phone actions

1. Open **Promise App**.
2. Tap **Use USB relay** and confirm `Active relay: http://127.0.0.1:8080`.
3. Scan/copy the recipient QR, then tap **Paste recipient identity**.
4. Enter `test` and tap **Sign, queue, and upload**.
5. Confirm `UPLOADED`. Use **Retry pending uploads** after a transport error;
   it reuses the durable original bytes.

### Verify and capture audit proof

```sh
make demo-usb-verify PROMISE_ADMIN_TOKEN='the-same-relay-token'
```

The command verifies the computer inbox, captures proof-check output, and saves
authenticated relay audit JSON. It refuses ambiguous inbox contents.

## Make commands

Run `make help`. The primary commands are:

| Purpose | Command |
| --- | --- |
| Test backend/protocol/peer | `make test` |
| Build/install app | `make install` |
| Diagnose phone startup | `make diagnose-phone` |
| Check USB bridge and relay | `make usb-relay-status` |
| Full USB setup | `make demo-usb-setup` |
| Verify inbox | `make peer-inbox-local` |
| Capture audit proof | `make demo-usb-verify PROMISE_ADMIN_TOKEN=...` |

## Evidence and documentation

- `.local/promise-relay/` — relay objects, mailbox state, and audit history.
- `.local/promise-peer/identity.ed25519` — computer private key, mode `0600`.
- `.local/promise-peer/inbox/<cid>.cbor` — exact verified bytes.
- `.local/promise-evidence/<cid>/` — verified fetch, proof-check, and audit.

Do not commit private keys, relay tokens, or local evidence.

- [Phone/computer demo](docs/phone-computer-demo.md)
- [Completed phone-to-computer POC example](docs/example-poc.md)
- [Amazon Fire / Kindle installation](docs/kindle.md)
- [Device relay setup](docs/device-relay-setup.md)
- [Proof runbook](docs/poc-proof-runbook.md)
- [Troubleshooting](docs/troubleshooting.md)
