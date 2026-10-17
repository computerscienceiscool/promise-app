# Phone-to-computer Promise POC

This is the canonical, reproducible demonstration path for the current POC:
one Android phone sends one signed message to a computer-owned recipient
identity. The relay stores and routes exact bytes; it does not accept, alter,
or execute the promise.

The preferred route is USB forwarding. It was used for the completed proof
because it avoids Wi-Fi client-isolation problems. It is a local implementation
choice, not a PromiseGrid transport rule.

## Requirements

- A built debug APK, an Android phone with USB debugging enabled, and a USB
  data cable.
- Go, JDK 17, Android SDK/ADB, `make`, and `curl` on the computer.
- A fresh private relay admin token. Keep it only on the computer.
- A private LAN address for the computer, even for USB forwarding: the relay
  process advertises and validates that host address while the phone uses its
  USB-local URL.

See [README installation](../README.md#install) and
[Troubleshooting](troubleshooting.md) for setup failures.

## 1. Start the relay

In terminal one, replace the sample address and secret, then leave the relay
running:

```sh
cd ~/lab/cswg/apps/promise-app
RELAY=http://YOUR-LAN-IP:8080 \
PROMISE_ADMIN_TOKEN='fresh-local-demo-secret' \
make relay
```

The terminal must report that the relay is listening on `0.0.0.0:8080`.

## 2. Install the app and prepare USB routing

In terminal two:

```sh
make install
make demo-usb-setup
```

`demo-usb-setup` verifies the local relay health, maps the phone's
`127.0.0.1:8080` to the computer over USB, and opens a QR code containing the
persistent computer recipient CID.

## 3. Send from the phone

1. Open **Promise App**.
2. Select **Use USB relay**. Confirm `Active relay: http://127.0.0.1:8080`.
3. Scan or copy the computer recipient QR, then select **Paste recipient
   identity**.
4. Enter a non-sensitive test body, such as `test`.
5. Select **Sign, queue, and upload**.
6. Record the message CID and wait for `UPLOADED` in the phone outbox.

`UPLOADED` means the phone queued immutable signed bytes and the relay accepted
them. It does not mean the computer or a person accepted the promise.

## 4. Verify on the computer

After the phone shows `UPLOADED`, run:

```sh
PROMISE_ADMIN_TOKEN='the-same-relay-token' make demo-usb-verify
```

This command fetches the computer inbox, verifies the exact envelope bytes,
and captures the correlated local relay audit evidence. It fails if there is
not exactly one verified inbox message or if the evidence cannot be matched.

For a completed example, including the phone screenshot, QR image, message
CID, verifier output, and stated limits, see [example-poc.md](example-poc.md).

## Optional private-LAN route

Use this only when the phone browser can reach the computer's private-LAN URL.
Before starting the relay, run:

```sh
make relay-preflight RELAY=http://YOUR-LAN-IP:8080
```

Start the relay as in step 1. On the phone, enter
`http://YOUR-LAN-IP:8080`, select **Save relay**, and open
`http://YOUR-LAN-IP:8080/healthz` in the phone browser. It must return
`{"status":"ok"}` before sending. Do not use `127.0.0.1` for this route;
on a phone it names the phone itself.

## Scope

This POC has physically proved the one-phone-to-computer path. It has not yet
proved two-phone or Kindle delivery, offline retry, duplicate upload handling,
production TLS, key recovery, or recipient acceptance. See
[the two-device proof runbook](poc-proof-runbook.md) for the remaining tests.
