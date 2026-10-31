#!/usr/bin/env bash
# Generates a local Phone A/Phone B QR-card page for relay-mediated delivery.
set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)
backend_dir="$project_dir/backend"
template="$project_dir/docs/phone-to-phone-qr-cards.html"
output_dir=${PROMISE_PHONE_QR_OUTPUT_DIR:-"$project_dir/.local/promise-peer"}
output_page="$output_dir/phone-to-phone-qr-cards.html"

usage() {
  printf "%s\n" \
    "Usage:" \
    "  RELAY=https://relay.example PHONE_B_CID=PHONE_B_PUBLIC_CID ./scripts/prepare-phone-to-phone-qr-demo.sh" \
    "" \
    "It creates local QR codes for Phone A relay, Phone B relay, and Phone B public identity." \
    "Never put a relay token or private key in a QR code or this command."
}

case "${1:-}" in
  "") ;;
  -h|--help) usage; exit 0 ;;
  *) usage >&2; exit 2 ;;
esac

: "${RELAY:?Set RELAY to an HTTP(S) relay URL reachable by both phones.}"
: "${PHONE_B_CID:?Set PHONE_B_CID to Phone B public device identity CID.}"

case "$RELAY" in
  http://*|https://*) ;;
  *) printf "%s\n" "RELAY must begin with http:// or https://." >&2; exit 2 ;;
esac
case "$RELAY" in
  *localhost*|*127.0.0.1*|*0.0.0.0*)
    printf "%s\n" "RELAY must be reachable by both phones; do not use a loopback address." >&2
    exit 2
    ;;
esac
case "$PHONE_B_CID" in
  b*) ;;
  *) printf "%s\n" "PHONE_B_CID must be Phone B public CID (normally beginning with b)." >&2; exit 2 ;;
esac

mkdir -p "$output_dir"
(cd "$backend_dir" && go run ./cmd/promise-peer qr -content "$RELAY" -out "$output_dir/phone-a-relay-url-qr.png")
(cd "$backend_dir" && go run ./cmd/promise-peer qr -content "$RELAY" -out "$output_dir/phone-b-relay-url-qr.png")
(cd "$backend_dir" && go run ./cmd/promise-peer qr -content "$PHONE_B_CID" -out "$output_dir/phone-b-recipient-identity-qr.png")

sed "s|__RELAY_URL__|$RELAY|g" "$template" > "$output_page"

printf "%s\n" "Phone A/Phone B QR demo page prepared."
printf "Open: %s\n" "$output_page"
printf "%s\n" "Use Phone A -> relay -> Phone B. Both phones need relay reachability; they do not need direct reachability to each other."
