# TODO 054 - Validate the Phone-Computer Host Setup

First build all host artifacts:

```sh
./scripts/build-phone-computer-host.sh --offline
```

Then run `scripts/validate-phone-computer-host.sh` after building the APK and
again with `--check-relay` after starting the LAN relay. Retain its output with
the demo evidence.

Current status: the host build, APK install, USB relay health check, and the
phone-to-computer proof all passed. The exact validator run with
`--check-relay` against an advertised LAN URL has not been retained, so this
TODO remains open rather than treating USB proof as LAN proof.
