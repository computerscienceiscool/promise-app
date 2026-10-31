# TODO 054 - Validate the Phone-Computer Host Setup

First build all host artifacts:

```sh
./scripts/build-phone-computer-host.sh --offline
```

Then run `scripts/validate-phone-computer-host.sh` after building the APK and
again with `--check-relay` after starting the LAN relay. Retain its output with
the demo evidence.

Completed 2026-10-26: the current APK built and installed, the phone scanned
the computer recipient QR, the USB relay was healthy, and the computer peer
independently verified the phone's uploaded message CID. The validator passed
with `--check-relay` against `http://10.0.99.219:8080`; its output is retained
locally at `.local/host-validation-2026-10-26.txt`. A second phone-to-computer
delivery over that direct Wi-Fi URL produced and independently verified
`bafkreifqo3nidty44ezg7hf6wcgoqnq2aq2gz2qz7xtubmsftdb7gsggly`.
