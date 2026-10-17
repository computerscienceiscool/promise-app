# TODO 060 - Add Recipient QR Generator

Generate a public PNG QR code containing only the persistent computer peer CID
so a phone can scan the recipient instead of requiring manual CID entry.

Completed 2026-10-17. `make peer-qr` generated a valid 512×512 PNG from the
persistent peer identity, and `make peer-qr-open` is the display shortcut.
