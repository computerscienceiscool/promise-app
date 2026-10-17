# TODO 054 - Validate the Phone-Computer Host Setup

First build all host artifacts:

```sh
./scripts/build-phone-computer-host.sh --offline
```

Then run `scripts/validate-phone-computer-host.sh` after building the APK and
again with `--check-relay` after starting the LAN relay. Retain its output with
the demo evidence.
