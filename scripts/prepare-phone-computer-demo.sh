#!/usr/bin/env bash
# Prepares the host-side values for the one-phone/one-computer POC.
#
# This command deliberately never starts the relay and never sends a message.
# Starting a network listener and tapping Send are explicit observed-demo steps.
set -euo pipefail

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)
backend_dir="$project_dir/backend"
identity_file=${PROMISE_PEER_IDENTITY_FILE:-"$project_dir/.local/promise-peer/identity.ed25519"}

usage() {
  cat <<'EOF'
Usage:
  PROMISE_ADVERTISE_URL='http://192.168.1.10:8080' \
    ./scripts/prepare-phone-computer-demo.sh

Optional environment:
  PROMISE_PEER_IDENTITY_FILE  Host-only computer peer private-key file.

This script creates or loads the computer peer identity and prints only its
public CID plus the precise phone configuration steps. It does not start a
relay, print an admin token, or transmit a message.
EOF
}

case "${1:-}" in
  "") ;;
  -h|--help) usage; exit 0 ;;
  *) usage >&2; exit 2 ;;
esac

: "${PROMISE_ADVERTISE_URL:?Set PROMISE_ADVERTISE_URL to the computer LAN URL, e.g. http://192.168.1.10:8080.}"

case "$PROMISE_ADVERTISE_URL" in
  http://*|https://*) ;;
  *) printf '%s\n' 'PROMISE_ADVERTISE_URL must begin with http:// or https://.' >&2; exit 2 ;;
esac
case "$PROMISE_ADVERTISE_URL" in
  *127.0.0.1*|*localhost*)
    printf '%s\n' 'PROMISE_ADVERTISE_URL must name the host LAN address, not localhost.' >&2
    exit 2
    ;;
esac

command -v go >/dev/null || { printf '%s\n' 'Go is required to prepare the computer peer.' >&2; exit 127; }

peer_cid=$(cd "$backend_dir" && go run ./cmd/promise-peer identity -identity-file "$identity_file")

printf '%s\n' 'Computer peer prepared.'
printf 'Relay URL:         %s\n' "$PROMISE_ADVERTISE_URL"
printf 'Computer peer CID: %s\n' "$peer_cid"
printf 'Private key file:  %s\n' "$identity_file"
printf '\n%s\n' 'Phone steps (do these after starting the relay in another terminal):'
printf '%s\n' "  1. Open Promise App and save this Relay URL: $PROMISE_ADVERTISE_URL"
printf '%s\n' "  2. Paste this exact Computer peer CID into Recipient device identity CID: $peer_cid"
printf '%s\n' '  3. Enter a distinctive promise and tap Sign, queue, and upload.'
printf '\n%s\n' 'After upload, fetch and verify it on this computer with:'
printf '  cd %s && go run ./cmd/promise-peer inbox -relay %s -identity-file %s\n' \
  "$backend_dir" "$PROMISE_ADVERTISE_URL" "$identity_file"
