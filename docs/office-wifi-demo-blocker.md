# Office Wi-Fi Demo Blocker

## Status: office Wi-Fi is blocked; private-hotspot Wi-Fi is proven

The phone-to-computer POC has a successful direct Wi-Fi demonstration on a
private hotspot. It must not be presented as a successful direct
office-Wi-Fi demonstration until that office-network blocker is resolved and
the same proof is rerun without ADB reverse.

## Observed 2026-10-30

During an office-network test, the phone and development computer were on the
same Wi-Fi routed subnet (`10.0.98.0/23`):

| Participant | Observed address |
| --- | --- |
| Development computer | `10.0.99.219/23` |
| Demo phone | `10.0.99.64/23` |

The phone could not establish HTTP connections to a relay listening on the
computer's private address. This was observed on two separate relay ports.
The host-side health endpoint was healthy locally, and the phone app reported
the corresponding connection failure.

The USB route was independently healthy: ADB reverse forwarded phone-local
`127.0.0.1` to the computer relay and the phone received `{"status":"ok"}`.
That proves the POC application path, but does not prove office Wi-Fi ingress.

## Likely infrastructure boundary

The available evidence rules out an incorrect Wi-Fi subnet. It does not yet
distinguish between:

1. host firewall policy rejecting or dropping inbound private-LAN TCP; and
2. office wireless client isolation / east-west traffic policy blocking one
   Wi-Fi client from reaching another.

This is an office-network or host-network-policy investigation, not evidence
that the relay, QR workflow, signed-envelope protocol, or recipient verifier
failed.

## Private-hotspot confirmation

On a separate private Wi-Fi hotspot, the phone and development computer
completed the full direct network path without USB or ADB reverse:

1. The phone reached the computer relay's `/healthz` endpoint directly.
2. The app reported its configured relay as reachable.
3. The phone uploaded a message and marked it `UPLOADED`.
4. The computer recipient verifier independently reported `RECIPIENT VERIFIED`
   for that phone message's exact CID.

This controls for the application, relay, signed envelope, recipient identity,
and recipient-verification flow. The remaining failure is therefore specific
to the office network or the host policy applied while on that network.

## Requested network test

For one private demo host and one demo phone, permit TCP from the phone Wi-Fi
client to the demo host's private-LAN address on the selected relay port. Do
not expose the relay to the public Internet. The network owner should also
confirm that client isolation is disabled for the approved demo pair, or place
both devices on a private non-isolated demonstration network.

After policy changes, verify from the phone browser:

```text
http://<demo-host-private-ip>:<relay-port>/healthz
```

The expected result is:

```json
{"status":"ok"}
```

Only then rerun the complete phone upload and independent computer-recipient
verification. The same phone Message CID must appear in the computer result as
`RECIPIENT VERIFIED cid=<same CID>`.

## Current demo recommendation

Use the proven private-hotspot Wi-Fi path for a wireless POC demonstration.
State plainly that direct office-Wi-Fi delivery remains blocked pending
network-policy validation. USB is available only as a fallback demonstration
path.
