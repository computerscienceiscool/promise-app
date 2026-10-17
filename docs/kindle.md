# Amazon Fire / Kindle installation guide

Promise App is an Android APK. A compatible Amazon Fire tablet can run the
same APK used for the Android phone POC; no Google Play services are required
by this app.

## Compatibility

The current APK has `minSdk = 26` (Android 8.0) and `targetSdk = 36`.

| Fire OS family | Android API level | Current APK result |
| --- | ---: | --- |
| Fire OS 8 | 29–30 | Compatible by API level; requires device test. |
| Fire OS 7 | 28 | Compatible by API level; requires device test. |
| Fire OS 6 | 25 | Not compatible: below this app's minimum API 26. |
| Fire OS 5 or older | 22 or lower | Not compatible: below this app's minimum API 26. |

Amazon documents that Fire OS is Android-based and that a normal Android app
will usually run on compatible Fire devices. Amazon's compatibility guidance
also explains that `minSdkVersion` determines whether an APK can be installed
on a device. See [Fire OS overview](https://developer.amazon.com/docs/fire-tv/fire-os-overview.html)
and [device filtering and compatibility](https://developer.amazon.com/docs/app-submission/device-filtering-and-compatibility.html).

This is compatibility analysis, not a completed Kindle proof. Before claiming
Kindle support, install and run the current APK on the actual target model and
complete the verification checklist below.

## Install by USB with ADB

This is the recommended POC path. It uses the same USB tooling as the Android
phone demonstration.

1. On the Fire tablet, open **Settings → Device Options**.
2. If **Developer Options** is absent, open **About Fire Tablet**, tap the
   tablet serial number seven times, then return to **Device Options**.
3. Open **Developer Options** and enable **USB debugging**.
4. Connect the tablet with a USB data cable. Accept its computer-authorization
   prompt if one appears.
5. On the development computer, from the repository root, confirm that ADB
   reports `device` rather than `unauthorized`:

   ```sh
   adb start-server
   adb devices
   ```

6. Build and install the debug APK:

   ```sh
   make install
   make launch
   ```

Amazon's own Fire Tablet instructions use ADB for USB testing and sideloading;
see [Connect to Fire Tablet through ADB](https://developer.amazon.com/docs/fire-tablets/connecting-adb-to-device.html)
and [Run Your App on a Fire Tablet](https://developer.amazon.com/docs/fire-tablets/ft-test-app-on-emulator-or-tablet.html).

## Run the local POC over USB

Keep the relay running on the development computer in one terminal:

```sh
RELAY=http://YOUR-LAN-IP:8080 \
PROMISE_ADMIN_TOKEN='fresh-local-demo-secret' \
make relay
```

In another terminal, connect the tablet to that relay through the USB cable
and generate the computer recipient QR code:

```sh
make demo-usb-setup
```

On the Fire tablet:

1. Open **Promise App**.
2. Select **Use USB relay** and confirm the active URL is
   `http://127.0.0.1:8080`.
3. Scan or copy the computer recipient QR code.
4. Enter a non-sensitive test promise and select **Sign, queue, and upload**.
5. Confirm the outbox shows `UPLOADED` and record its CID.

Then verify it from the computer:

```sh
PROMISE_ADMIN_TOKEN='the-same-relay-token' make demo-usb-verify
```

## Kindle proof checklist

Do not call Kindle support proven until all of these are recorded:

- [ ] `adb devices` lists the Fire tablet as `device`.
- [ ] `make install` completes and the app launches.
- [ ] The tablet creates and displays a device identity.
- [ ] **Use USB relay** saves and displays `http://127.0.0.1:8080`.
- [ ] A QR-transferred computer recipient CID can be pasted or scanned.
- [ ] A signed promise reaches `UPLOADED` with a recorded message CID.
- [ ] `make demo-usb-verify` independently verifies and retains the same CID.
- [ ] A screenshot and the verifier output are retained with the demo evidence.

## Limits and security

- The debug POC permits private-network HTTP for inspectable local testing.
  Production deployment must use HTTPS.
- The USB relay route is a host-specific demo transport; it is not a PromiseGrid
  protocol requirement.
- Do not expose the relay to the public Internet or put its admin token into
  the app.
- Current POC boundaries still include no key recovery, user-level trust policy,
  or production availability guarantees.
