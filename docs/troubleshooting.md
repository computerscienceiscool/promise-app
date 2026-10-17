# Troubleshooting the Promise App POC

This guide covers the local Android-phone-to-computer POC. USB routing, relay
ports, QR codes, and local files are implementation details, not universal
PromiseGrid rules.

## First check

Keep `make relay` running. In another terminal run:

```sh
make usb-relay-status
```

Healthy output includes an ADB `device`, `UsbFfs tcp:8080 tcp:8080`, and:

```json
{"status":"ok"}
```

Fix the first missing result before retrying an upload.

## Phone says “failed to connect”

Wi-Fi client isolation can block a phone even on the apparent same network.
Use the USB route:

```sh
make usb-relay
make usb-relay-status
```

In the app tap **Use USB relay**. Confirm the visible setting is:

```text
Active relay: http://127.0.0.1:8080
```

Then tap **Retry pending uploads**. Do not create a replacement message just
because transport failed: retry must reuse the queued bytes.

## Relay setting appears stuck on the old URL

Install the current app:

```sh
make install
```

Then tap **Use USB relay**. It persists the value directly without a QR or
clipboard. Treat the `Active relay:` line as the confirmation. If it is wrong,
run `make diagnose-phone` and keep the output before changing settings again.

## App crashes at launch

```sh
make diagnose-phone
```

If logs mention `unsupported Ed25519 public key encoding`, install the latest
APK. It rotates the incompatible early POC P-256 key and generates an explicit
Ed25519 Android Keystore key.

## ADB does not show `device`

Unlock the phone, enable USB debugging, and accept the computer's debugging
prompt. Then:

```sh
adb kill-server
adb start-server
adb devices
```

`unauthorized` must become `device` before USB routing can work.

## QR image did not open

```sh
make peer-qr-open
```

Fallback:

```sh
google-chrome --new-window file:///home/jj/lab/cswg/apps/promise-app/.local/promise-peer/recipient-qr.png
```

The QR contains only a public CID, never a private key or relay token.

## Upload succeeded but the computer inbox is empty

Use the same persistent computer identity used to make the QR:

```sh
make peer-id
make peer-inbox-local
```

Do not delete `.local/promise-peer/identity.ed25519` during a demo.

## Audit capture refuses to run

`make evidence-local` deliberately requires exactly one retained CBOR message,
so it cannot choose the wrong proof target. For multiple messages, use the
explicit `make evidence` command with CID, recipient, output directory, relay,
and token values.

## Android build fails

Configure the SDK path, then rebuild:

```sh
cd android
printf 'sdk.dir=/home/jj/Android/Sdk\n' > local.properties
./gradlew :app:assembleDebug
```

The current project targets JVM 17. See `make diagnose-phone` for physical
device failures after a successful build.
