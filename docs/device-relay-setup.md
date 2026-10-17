# Device-Reachable Relay Setup

This procedure starts the POC relay on the development computer so that Android
devices on the same private Wi-Fi network can reach it. It is for an observed
demo only: it intentionally exposes an unauthenticated HTTP transport on a
private network. Signed envelopes still protect their bytes and sender proof,
but HTTP does not provide confidentiality or endpoint authentication. Use HTTPS
behind a reverse proxy for any deployment outside this local demonstration.

## Preconditions

1. Build and install the **debug** APK. Its debug-only manifest permits a
   private-LAN `http://` URL. A release APK must use `https://`.
2. Put the computer and each test device on the same Wi-Fi network. Mobile data,
   guest Wi-Fi, client isolation, and a VPN can prevent device-to-computer
   traffic.
3. Select the computer's private IPv4 address on that Wi-Fi network, for example
   `192.168.1.10`. Do not use `127.0.0.1`: on a phone that address means the
   phone itself.
4. Keep the data directory outside version control. It contains immutable
   envelope bytes, mailbox metadata, and the admin audit log.

## Start the relay from source

From `promise-app`, choose a new admin token, set the host's actual private-LAN
address, and use the launcher. It creates a persistent, ignored data directory
at `.local/relay-data` unless `PROMISE_DATA_DIR` is set:

```sh
export PROMISE_ADMIN_TOKEN='replace-with-a-long-random-demo-token'
export PROMISE_ADVERTISE_URL='http://192.168.1.10:8080'
./scripts/start-lan-relay.sh
```

The startup line must say `promise relay listening on 0.0.0.0:8080`. Leave this
terminal open during the demonstration. Do not print, share, or enter the admin
token in the phone app.

Check the configuration without starting the relay:

```sh
PROMISE_ADMIN_TOKEN='replace-with-a-long-random-demo-token' \
PROMISE_ADVERTISE_URL='http://192.168.1.10:8080' \
./scripts/start-lan-relay.sh --dry-run
```

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
4. Copy each displayed device identity CID to the other phone's recipient field.

Use the two-device procedure in [TODO 021](../TODO/021-two-device-proof.md)
after both phones have saved the same relay URL.

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
