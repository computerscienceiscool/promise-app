# TODO 061 - Allow Private-LAN HTTP for the POC Relay

Permit cleartext HTTP in the Android app manifest because the inspectable POC
relay uses a private-LAN URL. Document that production deployments must use
HTTPS and remove this exception.

Completed 2026-10-17. The rebuilt app uploaded through the POC's HTTP relay
via the USB development route.
