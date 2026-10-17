# TODO 040 - Build the Latest APK Outside the Sandbox

Build the integrated Android source in the host terminal and retain the APK.

Run from the project root:

```sh
./scripts/build-phone-computer-host.sh --offline
```

The command also tests and compiles the host relay, peer, and proof tools. The
APK path printed at completion is the sole artifact for the separate install
TODO; this build command does not communicate with a phone.
