# TODO 046 - Use a computer as the recipient peer

- [x] Create the computer peer identity and copy its printed CID into the phone.
- [x] Run the computer mailbox command after sending from the phone.
- [x] Retain and inspect the exact verified CBOR evidence reported by the command.

The `backend/cmd/promise-peer` command implements this alternative to a second
physical phone.

Completed 2026-10-17 through the USB-connected relay path. The computer peer
verified and retained exact CBOR message
`bafkreihonloqrim7svqkpp4yz63qzwv7gxelyeotwjf73zledsutw2sbom` addressed to
its persistent CID. USB forwarding is a local POC transport choice, not a
universal PromiseGrid ingress rule.
