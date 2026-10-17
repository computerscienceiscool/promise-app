# Device-Reachable Relay Setup

This procedure starts the POC relay on the development computer for the
optional private-LAN route. For the proven and recommended connected-phone
route, use [phone-computer-demo.md](phone-computer-demo.md): it maps the phone
to the host relay through USB and avoids Wi-Fi client-isolation problems.

This private-LAN route is for an observed POC only. It intentionally exposes
unauthenticated HTTP on a private network. Signed envelopes protect their bytes
and sender proof, but HTTP does not provide confidentiality or endpoint
authentication. Use HTTPS behind a reverse proxy for any deployment outside a
local demonstration.

## Preconditions

1. Build and install the POC APK. Its current manifest permits private-LAN
   `http://` for the demo. Do not treat the current configuration as suitable
   for a production or release deployment; production must use HTTPS.
2. Put the computer and each test device on the same Wi-Fi network. Mobile data,
   guest Wi-Fi, client isolation, and a VPN can prevent device-to-computer
   traffic.
3. Select the computer's private IPv4 address on that Wi-Fi network, for example
   `192.168.1.10`. Do not use `127.0.0.1`: on a phone that address means the
   phone itself.
4. Keep the data directory outside version control. It contains immutable
   envelope bytes, mailbox metadata, and the admin audit log.

## Start the relay from source

From `promise-app`, choose a new admin token and set the host's actual
private-LAN address. First validate the intended LAN URL without starting a
listener:

```sh
make relay-preflight RELAY=http://192.168.1.10:8080
```

Then start the relay. The Make target creates persistent, ignored data at
`.local/promise-relay` unless `DATA_DIR` is set:

```sh
RELAY=http://192.168.1.10:8080 \
PROMISE_ADMIN_TOKEN='replace-with-a-long-random-demo-token' \
make relay
```

The startup line must say `promise relay listening on 0.0.0.0:8080`. Leave this
terminal open during the demonstration. Do not print, share, or enter the admin
token in the phone app.

If the host firewall is enabled, allow **TCP 8080 only from the private Wi-Fi
subnet** for the duration of the demo. Remove that temporary rule afterwards.

## Verify before involving a phone

On the development computer, substitute the selected LAN address:

```sh
curl --fail http://192.168.1.10:8080/healthz
```

Expected response:

```json
{"status":"ok"}
```

From each phone, open `http://192.168.1.10:8080/healthz` in its browser. The
same JSON response proves routing and firewall access. A browser warning or a
failure here must be resolved before testing the app.

## Configure each phone

1. Open **Promise App**.
2. Enter `http://192.168.1.10:8080` in **Relay URL**.
3. Tap **Save relay**.
4. Tap **Copy device identity**. The app confirms that the complete CID is in
   the Android clipboard; do not try to select the shortened visual display.
5. On the other phone, tap **Paste recipient identity** beside its recipient
   field. Confirm it is the complete lowercase-base32 CIDv1 (it begins with
   `b`; identity CIDs are not required to begin with `bafk`) before sending.

The relay URL is saved on the device and a trailing slash is removed
automatically. Both phones must use the same base URL. If a phone reports
`Inbox sync failed`, first open `<relay URL>/healthz` in that phone's browser;
do not send until it returns `{"status":"ok"}`.

Use the two-device procedure in [TODO 021](../TODO/021-two-device-proof.md)
after both phones have saved the same relay URL.

### Computer recipient alternative

A second phone is optional. Use the focused
[phone-to-computer POC guide](phone-computer-demo.md) for the relay launch,
computer identity, phone configuration, signed send, and verification steps.
The computer peer verifies relay-returned bytes before it displays or retains
them. See [TODO 046](../TODO/046-computer-peer-demo.md).

## Operator inspection

The admin endpoint exposes the append-only audit events for the POC:

```sh
curl --fail \
  -H "Authorization: Bearer $PROMISE_ADMIN_TOKEN" \
  http://127.0.0.1:8080/v1/admin/audit
```

Only run the audit query from the development computer. The relay's data
directory is intentionally inspectable for the POC; delete it after the demo if
it contains test messages you no longer need.

## Container alternative

From `promise-app/backend`:

```sh
docker build -t promise-relay .
docker run --rm \
  -p 8080:8080 \
  -e PROMISE_ADMIN_TOKEN='replace-with-a-long-random-demo-token' \
  -v "$PWD/data-device-demo:/data" \
  promise-relay
```

The container defaults to `:8080`, which is device-reachable through the port
mapping. Repeat the health check and phone configuration steps above.
