# TODO 064 - Add Pasteable USB Relay URL

Add a phone UI control that pastes a scanned relay URL from the clipboard and
a Make command that displays the USB relay URL as a QR image, avoiding manual
entry during the connected-phone demo.

Completed:

- The phone has **Paste relay URL**, which validates an `http://` or `https://`
  URL from the clipboard before persisting it.
- `make usb-relay-qr-open` renders the USB URL QR image for the connected-phone
  demo.
