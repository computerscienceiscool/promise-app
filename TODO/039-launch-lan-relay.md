# TODO 039 - Launch the LAN Relay With the Demo Token

Run `scripts/preflight-lan-relay.sh -relay <LAN URL>` first. It validates the
host, relay build, LAN address, and free port without requiring a token or
starting a service. Then run the validated launcher on the host and retain its
advertised LAN URL.
